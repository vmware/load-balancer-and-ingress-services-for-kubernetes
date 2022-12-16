/*
 * Copyright 2019-2020 VMware, Inc.
 * All Rights Reserved.
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*   http://www.apache.org/licenses/LICENSE-2.0
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*/

package k8s

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/rest"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/retry"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/status"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/clients"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/session"

	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func PopulateCache() error {
	var err error
	avi_rest_client_pool := avicache.SharedAVIClients()
	avi_obj_cache := avicache.SharedAviObjCache()
	// Randomly pickup a client.
	if avi_rest_client_pool != nil && len(avi_rest_client_pool.AviClient) > 0 {
		_, _, err = avi_obj_cache.AviObjCachePopulate(avi_rest_client_pool.AviClient, utils.CtrlVersion, utils.CloudName)
		if err != nil {
			utils.AviLog.Warnf("failed to populate avi cache with error: %v", err.Error())
			return err
		}
		if err = avicache.SetControllerClusterUUID(avi_rest_client_pool); err != nil {
			utils.AviLog.Warnf("Failed to set the controller cluster uuid with error: %v", err)
		}
	}

	return nil
}

func (c *AviController) cleanupStaleVSes() {

	aviRestClientPool := avicache.SharedAVIClients()
	aviObjCache := avicache.SharedAviObjCache()

	delModels := deleteConfigFromConfigmap(c.informers.ClientSet)
	if delModels {
		go SetDeleteSyncChannel()
		parentKeys := aviObjCache.VsCacheMeta.AviCacheGetAllParentVSKeys()
		deleteAviObjects(parentKeys, aviObjCache, aviRestClientPool)
	} else {
		status.NewStatusPublisher().ResetStatefulSetAnnotation()
	}

	// Delete Stale objects by deleting model for dummy VS
	if _, err := lib.IsClusterNameValid(); err != nil {
		utils.AviLog.Errorf("AKO cluster name is invalid.")
		return
	}
	if aviRestClientPool != nil && len(aviRestClientPool.AviClient) > 0 {
		utils.AviLog.Infof("Starting clean up of stale objects")
		restlayer := rest.NewRestOperations(aviObjCache, aviRestClientPool)
		staleVSKey := lib.GetTenant() + "/" + lib.DummyVSForStaleData
		restlayer.CleanupVS(staleVSKey, true)
		staleCacheKey := avicache.NamespaceName{
			Name:      lib.DummyVSForStaleData,
			Namespace: lib.GetTenant(),
		}
		aviObjCache.VsCacheMeta.AviCacheDelete(staleCacheKey)
	}

	vsKeysPending := aviObjCache.VsCacheMeta.AviGetAllKeys()
	if delModels {
		//Delete NPL annotations
		DeleteNPLAnnotations()
	}

	if delModels && len(vsKeysPending) == 0 && lib.ConfigDeleteSyncChan != nil {
		close(lib.ConfigDeleteSyncChan)
		lib.ConfigDeleteSyncChan = nil
	}
}

func deleteAviObjects(parentVSKeys []avicache.NamespaceName, avi_obj_cache *avicache.AviObjCache, avi_rest_client_pool *utils.AviRestClientPool) {
	for _, pvsKey := range parentVSKeys {
		// Fetch the parent VS cache and update the SNI child
		vsObj, parentFound := avi_obj_cache.VsCacheMeta.AviCacheGet(pvsKey)
		if parentFound {
			// Parent cache is already populated, just append the SNI key
			vs_cache_obj, foundvs := vsObj.(*avicache.AviVsCache)
			if foundvs {
				key := pvsKey.Namespace + "/" + pvsKey.Name
				namespace, _ := utils.ExtractNamespaceObjectName(key)
				restlayer := rest.NewRestOperations(avi_obj_cache, avi_rest_client_pool)
				restlayer.DeleteVSOper(pvsKey, vs_cache_obj, namespace, key, false, false)
			}
		}
	}
}

func PopulateNodeCache(cs *kubernetes.Clientset) {
	nodeCache := objects.SharedNodeLister()
	var nodeLabels map[string]string
	isNodePortMode := lib.IsNodePortMode()
	if isNodePortMode {
		nodeLabels = lib.GetNodePortsSelector()
	}
	nodeCache.PopulateAllNodes(cs, isNodePortMode, nodeLabels)
}

func PopulateControllerProperties(cs kubernetes.Interface) error {
	ctrlPropCache := utils.SharedCtrlProp()
	ctrlProps, err := lib.GetControllerPropertiesFromSecret(cs)
	if err != nil {
		return err
	}
	ctrlPropCache.PopulateCtrlProp(ctrlProps)
	return nil
}

func delConfigFromData(data map[string]string) bool {
	var delConf bool
	if val, ok := data[lib.DeleteConfig]; ok {
		if val == "true" {
			utils.AviLog.Infof("deleteConfig set in configmap, sync would be disabled")
			lib.AKOControlConfig().PodEventf(corev1.EventTypeNormal, lib.AKODeleteConfigSet, "DeleteConfig set in configmap, sync would be disabled")
			delConf = true
		}
	}
	lib.SetDeleteConfigMap(delConf)
	return delConf
}

func deleteConfigFromConfigmap(cs kubernetes.Interface) bool {
	cmNS := utils.GetAKONamespace()
	cm, err := cs.CoreV1().ConfigMaps(cmNS).Get(context.TODO(), lib.AviConfigMap, metav1.GetOptions{})
	if err == nil {
		return delConfigFromData(cm.Data)
	}
	utils.AviLog.Warnf("error while reading configmap, sync would be disabled: %v", err)
	return true
}

func (c *AviController) SetSEGroupCloudName() bool {
	seGroupToUse := lib.GetSEGNameEnv()
	client := avicache.SharedAVIClients().AviClient[0]
	var err error
	// 2. Marker based (only advancedL4)
	if seGroupToUse == "" && lib.GetAdvancedL4() {
		err, seGroupToUse = lib.FetchSEGroupWithMarkerSet(client)
		if err != nil {
			utils.AviLog.Infof("Setting SEGroup with markerset and no SEGroup found from env")
			return false
		}
	}

	// 3. Default-SEGroup
	if seGroupToUse == "" {
		utils.AviLog.Infof("Setting SEGroup %s for VS placement.", lib.DEFAULT_SE_GROUP)
		seGroupToUse = lib.DEFAULT_SE_GROUP
	}

	if !utils.IsVCFCluster() {
		lib.SetSEGName(seGroupToUse)
		return true
	}

	nsName := utils.GetAKONamespace()
	nsObj, err := c.informers.NSInformer.Lister().Get(nsName)
	if err != nil {
		utils.AviLog.Warnf("Failed to GET the namespace %s details due to the following error: %v", nsName, err.Error())
		return false
	}

	annotations := nsObj.GetAnnotations()
	segroup, ok := annotations[lib.WCPSEGroup]
	if !ok {
		utils.AviLog.Warnf("Failed to get SEGroup from annotation in namespace")
		return false
	}

	cloud, ok := annotations[lib.WCPCloud]
	if !ok {
		utils.AviLog.Warnf("Failed to get cloud name from annotation in namespace")
		return false
	}
	utils.AviLog.Infof("Setting SEGroup %s, cloud %s for VS placement.", segroup, cloud)
	lib.SetSEGName(segroup)
	utils.SetCloudName(cloud)

	return true
}

func (c *AviController) AddBootupNSEventHandler(k8sinfo K8sinformers, stopCh <-chan struct{}, startSyncCh chan struct{}) {
	NSHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if lib.AviSEInitialized {
				return
			}
			if c.SetSEGroupCloudName() {
				startSyncCh <- struct{}{}
				startSyncCh = nil
			}
		},
		UpdateFunc: func(old, obj interface{}) {
			if lib.AviSEInitialized {
				return
			}
			if c.SetSEGroupCloudName() {
				startSyncCh <- struct{}{}
				startSyncCh = nil
			}
		},
	}
	c.informers.NSInformer.Informer().AddEventHandler(NSHandler)

	go c.informers.NSInformer.Informer().Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, c.informers.NSInformer.Informer().HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
	} else {
		utils.AviLog.Info("Caches synced for NS informer")
	}
}

func (c *AviController) AddNCPBootstrapEventHandler(k8sinfo K8sinformers, stopCh <-chan struct{}, startSyncCh chan struct{}) {
	NCPBootstrapHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			utils.AviLog.Infof("NCP Bootstrap ADD Event")
			ctrlIP := lib.GetControllerURLFromBootstrapCR()
			if ctrlIP != "" && startSyncCh != nil {
				lib.SetControllerIP(ctrlIP)
				startSyncCh <- struct{}{}
				startSyncCh = nil
			}
		},
		UpdateFunc: func(old, obj interface{}) {
			utils.AviLog.Infof("NCP Bootstrap Update Event")
			ctrlIP := lib.GetControllerURLFromBootstrapCR()
			if ctrlIP != "" && startSyncCh != nil {
				lib.SetControllerIP(ctrlIP)
				startSyncCh <- struct{}{}
				startSyncCh = nil
			}
		},
	}
	c.dynamicInformers.NCPBootstrapInformer.Informer().AddEventHandler(NCPBootstrapHandler)

	go c.dynamicInformers.NCPBootstrapInformer.Informer().Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, c.dynamicInformers.NCPBootstrapInformer.Informer().HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
	} else {
		utils.AviLog.Info("Caches synced for NCP Bootstrap informer")
	}
}

// HandleConfigMap : initialise the controller, start informer for configmap and wait for the akc configmap to be created.
// When the configmap is created, enable sync for other k8s objects. When the configmap is disabled, disable sync.
func (c *AviController) HandleConfigMap(k8sinfo K8sinformers, ctrlCh chan struct{}, stopCh <-chan struct{}, quickSyncCh chan struct{}) error {
	cs := k8sinfo.Cs
	aviClientPool := avicache.SharedAVIClients()
	if aviClientPool == nil || len(aviClientPool.AviClient) < 1 {
		c.DisableSync = true
		lib.SetDisableSync(true)
		utils.AviLog.Errorf("could not get client to connect to Avi Controller, disabling sync")
		lib.ShutdownApi()
		return errors.New("Unable to contact the avi controller on bootup")
	}
	aviclient := aviClientPool.AviClient[0]

	validateUserInput, err := avicache.ValidateUserInput(aviclient)
	if err != nil {
		utils.AviLog.Errorf("Error while validating input: %s", err.Error())
		lib.AKOControlConfig().PodEventf(v1.EventTypeWarning, lib.SyncDisabled, "Invalid user input %s", err.Error())
	} else {
		lib.AKOControlConfig().PodEventf(v1.EventTypeNormal, lib.ValidatedUserInput, "User input validation completed.")
	}

	if !validateUserInput {
		return errors.New("sync is disabled because of configmap unavailability during bootup")
	}
	c.DisableSync = deleteConfigFromConfigmap(cs)
	lib.SetDisableSync(c.DisableSync)

	configMapEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			cm, ok := validateAviConfigMap(obj)
			if !ok {
				return
			}
			utils.AviLog.Infof("avi k8s configmap created")
			utils.AviLog.SetLevel(cm.Data[lib.LOG_LEVEL])
			lib.AKOControlConfig().EventsSetEnabled(cm.Data[lib.EnableEvents])
			// Check if AKO is configured to only use Ingress. This value can be only set during bootup and can't be edited dynamically.
			lib.SetLayer7Only(cm.Data[lib.LAYER7_ONLY])
			// Check if we need to use PGs for SNIs or not.
			lib.SetNoPGForSNI(cm.Data[lib.NO_PG_FOR_SNI])

			delModels := delConfigFromData(cm.Data)

			validateUserInput, err := avicache.ValidateUserInput(aviclient)
			if err != nil {
				utils.AviLog.Errorf("Error while validating input: %s", err.Error())
				lib.AKOControlConfig().PodEventf(v1.EventTypeWarning, lib.SyncDisabled, "Invalid user input %s", err.Error())
			} else {
				lib.AKOControlConfig().PodEventf(v1.EventTypeNormal, lib.ValidatedUserInput, "User input validation completed.")
			}
			c.DisableSync = !validateUserInput || delModels
			lib.SetDisableSync(c.DisableSync)
		},
		UpdateFunc: func(old, obj interface{}) {
			cm, ok := validateAviConfigMap(obj)
			oldcm, oldok := validateAviConfigMap(old)
			if !ok || !oldok {
				return
			}
			if oldcm.ResourceVersion == cm.ResourceVersion {
				return
			}
			// if resourceversions and loglevel change, set new loglevel
			if oldcm.Data[lib.LOG_LEVEL] != cm.Data[lib.LOG_LEVEL] {
				utils.AviLog.SetLevel(cm.Data[lib.LOG_LEVEL])
			}

			if oldcm.Data[lib.EnableEvents] != cm.Data[lib.EnableEvents] {
				lib.AKOControlConfig().EventsSetEnabled(cm.Data[lib.EnableEvents])
			}

			if oldcm.Data[lib.DeleteConfig] == cm.Data[lib.DeleteConfig] {
				return
			}
			// if DeleteConfig value has changed, then check if we need to enable/disable sync
			isValidUserInput, err := avicache.ValidateUserInput(aviclient)
			if err != nil {
				utils.AviLog.Errorf("Error while validating input: %s", err.Error())
			}
			delModels := delConfigFromData(cm.Data)
			c.DisableSync = !isValidUserInput || delModels
			lib.SetDisableSync(c.DisableSync)
			if isValidUserInput {
				if delModels {
					c.DeleteModels()
					SetDeleteSyncChannel()
					isPrimaryAKO := lib.AKOControlConfig().GetAKOInstanceFlag()
					if isPrimaryAKO && lib.GetServiceType() == "ClusterIP" {
						avicache.DeConfigureSeGroupLabels()
					}
				} else {
					status.NewStatusPublisher().ResetStatefulSetAnnotation()
					lib.AKOControlConfig().PodEventf(corev1.EventTypeNormal, lib.AKODeleteConfigUnset, "DeleteConfig unset in configmap, sync would be enabled")
					quickSyncCh <- struct{}{}
				}
			}

		},
		DeleteFunc: func(obj interface{}) {
			utils.AviLog.Warnf("avi k8s configmap deleted, shutting down api server")
			lib.ShutdownApi()
		},
	}

	c.informers.ConfigMapInformer.Informer().AddEventHandler(configMapEventHandler)

	go c.informers.ConfigMapInformer.Informer().Run(stopCh)
	if !cache.WaitForCacheSync(stopCh,
		c.informers.ConfigMapInformer.Informer().HasSynced,
	) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
	} else {
		utils.AviLog.Info("Caches synced")
	}
	return nil
}

func (c *AviController) AddBootupSecretEventHandler(k8sinfo K8sinformers, stopCh <-chan struct{}, startSyncCh chan struct{}) {
	NCPSecretHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if lib.AviSecretInitialized {
				return
			}
			data, ok := obj.(*corev1.Secret)
			if !ok || data.Namespace != utils.GetAKONamespace() {
				return
			}
			if c.ValidAviSecret() {
				startSyncCh <- struct{}{}
				startSyncCh = nil
			}
		},
		UpdateFunc: func(old, obj interface{}) {
			if lib.AviSecretInitialized {
				return
			}
			data, ok := obj.(*corev1.Secret)
			if !ok || data.Namespace != utils.GetAKONamespace() {
				return
			}
			if c.ValidAviSecret() {
				startSyncCh <- struct{}{}
				startSyncCh = nil
			}
		},
	}
	c.informers.SecretInformer.Informer().AddEventHandler(NCPSecretHandler)

	go c.informers.SecretInformer.Informer().Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, c.informers.SecretInformer.Informer().HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
	} else {
		utils.AviLog.Info("Caches synced for Avi Secret informer")
	}
}

func (c *AviController) ValidAviSecret() bool {
	cs := c.informers.ClientSet
	aviSecret, err := cs.CoreV1().Secrets(utils.GetAKONamespace()).Get(context.TODO(), lib.AviSecret, metav1.GetOptions{})
	if err == nil {
		ctrlIP := lib.GetControllerIP()
		authToken := string(aviSecret.Data["authtoken"])
		username := string(aviSecret.Data["username"])
		password := string(aviSecret.Data["password"])
		if username == "" || (password == "" && authToken == "") {
			return false
		}

		var transport *http.Transport
		if authToken == "" {
			_, err = clients.NewAviClient(ctrlIP, username,
				session.SetPassword(password), session.SetNoControllerStatusCheck, session.SetTransport(transport), session.SetInsecure)
		} else {
			_, err = clients.NewAviClient(ctrlIP, username,
				session.SetAuthToken(authToken), session.SetNoControllerStatusCheck, session.SetTransport(transport), session.SetInsecure)
		}
		if err == nil {
			utils.AviLog.Infof("Successfully connected to AVI controller using existing AKO secret")
			return true
		} else {
			utils.AviLog.Errorf("AVI controller initialization failed with err: %v", err)
		}
	} else {
		utils.AviLog.Infof("Got error while fetching avi-secret: %v", err)
	}
	return false
}

func (c *AviController) InitController(informers K8sinformers, registeredInformers []string, ctrlCh <-chan struct{}, stopCh <-chan struct{}, quickSyncCh chan struct{}, waitGroupMap ...map[string]*sync.WaitGroup) {
	// set up signals so we handle the first shutdown signal gracefully
	var worker *utils.FullSyncThread
	var tokenWorker *utils.FullSyncThread
	informersArg := make(map[string]interface{})
	informersArg[utils.INFORMERS_OPENSHIFT_CLIENT] = informers.OshiftClient
	if lib.GetNamespaceToSync() != "" {
		informersArg[utils.INFORMERS_NAMESPACE] = lib.GetNamespaceToSync()
	}
	informersArg[utils.INFORMERS_ADVANCED_L4] = lib.GetAdvancedL4()
	c.informers = utils.NewInformers(utils.KubeClientIntf{ClientSet: informers.Cs}, registeredInformers, informersArg)
	c.dynamicInformers = lib.NewDynamicInformers(informers.DynamicClient)
	var ingestionwg *sync.WaitGroup
	var graphwg *sync.WaitGroup
	var fastretrywg *sync.WaitGroup
	var slowretrywg *sync.WaitGroup
	var statusWG *sync.WaitGroup
	var leaderElectionWG *sync.WaitGroup
	if len(waitGroupMap) > 0 {
		// Fetch all the waitgroups
		ingestionwg, _ = waitGroupMap[0]["ingestion"]
		graphwg, _ = waitGroupMap[0]["graph"]
		fastretrywg, _ = waitGroupMap[0]["fastretry"]
		slowretrywg, _ = waitGroupMap[0]["slowretry"]
		statusWG, _ = waitGroupMap[0]["status"]
		leaderElectionWG, _ = waitGroupMap[0]["leaderElection"]
	}

	/** Sequence:
	  1. Initialize the graph layer queue.
	  2. Do a full sync from main thread and publish all the models.
	  3. Initialize the ingestion layer queue for partial sync.
	  **/
	// start the go routines draining the queues in various layers
	var graphQueue *utils.WorkerQueue
	// This is the first time initialization of the queue. For hostname based sharding, we don't want layer 2 to process the queue using multiple go routines.
	var retryQueueWorkers uint32
	retryQueueWorkers = 1
	slowRetryQParams := utils.WorkerQueue{NumWorkers: retryQueueWorkers, WorkqueueName: lib.SLOW_RETRY_LAYER, SlowSyncTime: lib.SLOW_SYNC_TIME}
	fastRetryQParams := utils.WorkerQueue{NumWorkers: retryQueueWorkers, WorkqueueName: lib.FAST_RETRY_LAYER}

	numWorkers := uint32(1)
	ingestionQueueParams := utils.WorkerQueue{NumWorkers: numWorkers, WorkqueueName: utils.ObjectIngestionLayer}
	numGraphWorkers := lib.GetshardSize()
	if numGraphWorkers == 0 {
		// For dedicated VSes - we will have 8 threads layer 3
		numGraphWorkers = 8
	}
	graphQueueParams := utils.WorkerQueue{NumWorkers: numGraphWorkers, WorkqueueName: utils.GraphLayer}
	statusQueueParams := utils.WorkerQueue{NumWorkers: numGraphWorkers, WorkqueueName: utils.StatusQueue}
	graphQueue = utils.SharedWorkQueue(&ingestionQueueParams, &graphQueueParams, &slowRetryQParams, &fastRetryQParams, &statusQueueParams).GetQueueByName(utils.GraphLayer)

	err := PopulateCache()
	if err != nil {
		c.DisableSync = true
		utils.AviLog.Errorf("failed to populate cache, disabling sync")
		lib.ShutdownApi()
	}
	if lib.IsIstioEnabled() {
		c.IstioBootstrap()
	}
	// Setup and start event handlers for objects.
	c.addIndexers()
	c.AddCrdIndexer()
	if lib.UseServicesAPI() {
		c.AddSvcApiIndexers()
	}
	c.Start(stopCh)

	fullSyncInterval := os.Getenv(utils.FULL_SYNC_INTERVAL)
	interval, err := strconv.ParseInt(fullSyncInterval, 10, 64)

	if lib.GetAdvancedL4() {
		// Set the error to nil
		err = nil
		interval = 300 // seconds, hardcoded for now.
	}
	// Set up the workers but don't start draining them.
	if err != nil {
		utils.AviLog.Errorf("Cannot convert full sync interval value to integer, pls correct the value and restart AKO. Error: %s", err)
	} else {
		// First boot sync
		err = c.FullSyncK8s(false)
		if ctrlAuthToken, ok := utils.SharedCtrlProp().AviCacheGet(utils.ENV_CTRL_AUTHTOKEN); ok && ctrlAuthToken != nil && ctrlAuthToken.(string) != "" {
			c.RefreshAuthToken()
		}
		if err != nil {
			// Something bad sync. We need to return and shutdown the API server
			utils.AviLog.Errorf("Couldn't run full sync successfully on bootup, going to shutdown AKO")
			lib.ShutdownApi()
			return
		}
		if interval != 0 {
			worker = utils.NewFullSyncThread(time.Duration(interval) * time.Second)
			worker.SyncFunction = c.FullSync
			worker.QuickSyncFunction = c.FullSyncK8s
			go worker.Run()
		} else {
			utils.AviLog.Warnf("Full sync interval set to 0, will not run full sync")
		}

		if ctrlAuthToken, ok := utils.SharedCtrlProp().AviCacheGet(utils.ENV_CTRL_AUTHTOKEN); ok && ctrlAuthToken != nil && ctrlAuthToken.(string) != "" {
			tokenWorker = utils.NewFullSyncThread(time.Duration(utils.RefreshAuthTokenInterval) * time.Hour)
			tokenWorker.SyncFunction = c.RefreshAuthToken
			go tokenWorker.Run()
		}
	}

	// Leader election happens after populating controller cache and fullsynck8s.
	ctx, cancel := context.WithCancel(context.Background())
	leaderElector, err := utils.NewLeaderElector(informers.Cs, c.OnStartedLeading, c.OnStoppedLeading, c.OnNewLeader)
	if err != nil {
		utils.AviLog.Fatalf("Leader election failed with error %v, shutting down AKO.", err)
	}

	leReadyCh := leaderElector.Run(ctx, leaderElectionWG)
	<-leReadyCh

	c.cleanupStaleVSes()

	// once the l3 cache is populated, we can call the updatestatus functions from here
	restlayer := rest.NewRestOperations(avicache.SharedAviObjCache(), avicache.SharedAVIClients())
	restlayer.SyncObjectStatuses()

	graphQueue.SyncFunc = SyncFromNodesLayer
	graphQueue.Run(stopCh, graphwg)

	c.SetupEventHandlers(informers)
	if lib.DisableSync {
		lib.AKOControlConfig().PodEventf(corev1.EventTypeNormal, lib.AKODeleteConfigSet, "AKO is in disable sync state")
	} else {
		lib.AKOControlConfig().PodEventf(corev1.EventTypeNormal, lib.AKOReady, "AKO is now listening for Object updates in the cluster")
	}

	ingestionQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	ingestionQueue.SyncFunc = SyncFromIngestionLayer
	ingestionQueue.Run(stopCh, ingestionwg)

	fastRetryQueue := utils.SharedWorkQueue().GetQueueByName(lib.FAST_RETRY_LAYER)
	fastRetryQueue.SyncFunc = SyncFromFastRetryLayer
	fastRetryQueue.Run(stopCh, fastretrywg)

	slowRetryQueue := utils.SharedWorkQueue().GetQueueByName(lib.SLOW_RETRY_LAYER)
	slowRetryQueue.SyncFunc = SyncFromSlowRetryLayer
	slowRetryQueue.Run(stopCh, slowretrywg)

	statusQueue := utils.SharedWorkQueue().GetQueueByName(utils.StatusQueue)
	statusQueue.SyncFunc = SyncFromStatusQueue
	statusQueue.Run(stopCh, statusWG)

LABEL:
	for {
		select {
		case <-quickSyncCh:
			worker.QuickSync()
		case <-ctrlCh:
			break LABEL
		}
	}
	if worker != nil {
		worker.Shutdown()
	}

	// Cancel the Leader election goroutines
	cancel()
	<-ctx.Done()

	ingestionQueue.StopWorkers(stopCh)
	graphQueue.StopWorkers(stopCh)
	fastRetryQueue.StopWorkers(stopCh)
	slowRetryQueue.StopWorkers(stopCh)
	statusQueue.StopWorkers(stopCh)
}

func (c *AviController) RefreshAuthToken() {
	lib.RefreshAuthToken(c.informers.KubeClientIntf.ClientSet)
}

func (c *AviController) addIndexers() {
	if c.informers.IngressClassInformer != nil {
		c.informers.IngressClassInformer.Informer().AddIndexers(
			cache.Indexers{
				lib.AviSettingIngClassIndex: func(obj interface{}) ([]string, error) {
					ingclass, ok := obj.(*networkingv1.IngressClass)
					if !ok {
						return []string{}, nil
					}
					if ingclass.Spec.Parameters != nil {
						// sample settingKey: ako.vmware.com/AviInfraSetting/avi-1
						settingKey := *ingclass.Spec.Parameters.APIGroup + "/" + ingclass.Spec.Parameters.Kind + "/" + ingclass.Spec.Parameters.Name
						return []string{settingKey}, nil
					}
					return []string{}, nil
				},
			},
		)
	}
	c.informers.ServiceInformer.Informer().AddIndexers(
		cache.Indexers{
			lib.AviSettingServicesIndex: func(obj interface{}) ([]string, error) {
				service, ok := obj.(*corev1.Service)
				if !ok {
					return []string{}, nil
				}
				if service.Spec.Type == corev1.ServiceTypeLoadBalancer {
					if val, ok := service.Annotations[lib.InfraSettingNameAnnotation]; ok && val != "" {
						return []string{val}, nil
					}
				}
				return []string{}, nil
			},
		},
	)
	if c.informers.RouteInformer != nil {
		c.informers.RouteInformer.Informer().AddIndexers(
			cache.Indexers{
				lib.AviSettingRouteIndex: func(obj interface{}) ([]string, error) {
					route, ok := obj.(*routev1.Route)
					if !ok {
						return []string{}, nil
					}
					if settingName, ok := route.Annotations[lib.InfraSettingNameAnnotation]; ok {
						return []string{settingName}, nil
					}
					return []string{}, nil
				},
			},
		)
	}

}

func (c *AviController) FullSync() {

	avi_rest_client_pool := avicache.SharedAVIClients()
	avi_obj_cache := avicache.SharedAviObjCache()
	// Randomly pickup a client.
	if len(avi_rest_client_pool.AviClient) > 0 {
		avi_obj_cache.AviClusterStatusPopulate(avi_rest_client_pool.AviClient[0])
		if !lib.GetAdvancedL4() {
			avi_obj_cache.AviCacheRefresh(avi_rest_client_pool.AviClient[0], utils.CloudName)
		} else {
			// In this case we just sync the Gateway status to the LB status
			restlayer := rest.NewRestOperations(avi_obj_cache, avi_rest_client_pool)
			restlayer.SyncObjectStatuses()
		}
		allModelsMap := objects.SharedAviGraphLister().GetAll()
		var allModels []string
		for modelName := range allModelsMap.(map[string]interface{}) {
			allModels = append(allModels, modelName)
		}
		for _, modelName := range allModels {
			utils.AviLog.Debugf("Reseting retry counter during full sync for model :%s", modelName)
			//reset retry counter in full sync
			found, avimodelIntf := objects.SharedAviGraphLister().Get(modelName)
			if found && avimodelIntf != nil {
				avimodel, ok := avimodelIntf.(*nodes.AviObjectGraph)
				if ok {
					avimodel.SetRetryCounter()
				}
			}
			// Not publishing the model anymore to layer since we don't want to support full sync for now.
			//nodes.PublishKeyToRestLayer(modelName, "fullsync", sharedQueue)
		}
	}
}

func (c *AviController) FullSyncK8s(sync bool) error {
	if c.DisableSync {
		utils.AviLog.Infof("Sync disabled, skipping full sync")
		return nil
	}
	sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
	var vrfModelName string
	if lib.GetDisableStaticRoute() && !lib.IsNodePortMode() {
		utils.AviLog.Infof("Static route sync disabled, skipping node informers")
	} else {
		isPrimaryAKO := lib.AKOControlConfig().GetAKOInstanceFlag()
		if isPrimaryAKO {
			lib.SetStaticRouteSyncHandler()
			var labelSelectorMap map[string]string
			//Apply filter to nodes in NodePort mode
			if lib.IsNodePortMode() {
				nodeLabels := lib.GetNodePortsSelector()
				if len(nodeLabels) == 2 && nodeLabels["key"] != "" {
					labelSelectorMap = make(map[string]string)
					labelSelectorMap[nodeLabels["key"]] = nodeLabels["value"]
				}
			}
			nodeObjects, _ := utils.GetInformers().NodeInformer.Lister().List(labels.Set(labelSelectorMap).AsSelector())
			for _, node := range nodeObjects {
				key := utils.NodeObj + "/" + node.Name
				meta, err := meta.Accessor(node)
				if err == nil {
					resVer := meta.GetResourceVersion()
					objects.SharedResourceVerInstanceLister().Save(key, resVer)
				}
				nodes.DequeueIngestion(key, true)
			}
			// Publish vrfcontext model now, this has to be processed first
			vrfModelName = lib.GetModelName(lib.GetTenant(), lib.GetVrf())
			utils.AviLog.Infof("Processing model for vrf context in full sync: %s", vrfModelName)
			nodes.PublishKeyToRestLayer(vrfModelName, "fullsync", sharedQueue)
			timeout := make(chan bool, 1)
			go func() {
				time.Sleep(20 * time.Second)
				timeout <- true
			}()
			select {
			case <-lib.StaticRouteSyncChan:
				utils.AviLog.Infof("Processing done for VRF")
			case <-timeout:
				utils.AviLog.Warnf("Timed out while waiting for rest layer to respond, moving on with bootup")
			}
		} else {
			utils.AviLog.Warnf("AKO is not primary instance, skipping vrf context publish in full sync.")
		}
	}

	acceptedNamespaces := make(map[string]struct{})
	allNamespaces, err := utils.GetInformers().ClientSet.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		utils.AviLog.Errorf("Error in getting all namespaces: %v", err.Error())
		return err
	}
	for _, ns := range allNamespaces.Items {
		if !lib.IsNamespaceBlocked(ns.GetName()) &&
			utils.CheckIfNamespaceAccepted(ns.GetName(), ns.GetLabels(), false) {
			acceptedNamespaces[ns.GetName()] = struct{}{}
		}
	}

	for namespace := range acceptedNamespaces {
		svcObjs, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).List(labels.Set(nil).AsSelector())
		if err != nil {
			utils.AviLog.Errorf("Unable to retrieve the services during full sync: %s", err)
			return err
		}

		for _, svcObj := range svcObjs {
			isSvcLb := isServiceLBType(svcObj)
			var key string
			if isSvcLb && !lib.GetLayer7Only() {
				key = utils.L4LBService + "/" + utils.ObjKey(svcObj)
				if svcObj.Annotations[lib.SharedVipSvcLBAnnotation] != "" {
					// mark the object type as ShareVipSvc
					// to separate these out from regulare clusterip, svclb services
					key = lib.SharedVipServiceKey + "/" + utils.ObjKey(svcObj)
				}
			} else {
				if lib.GetAdvancedL4() {
					continue
				}
				key = utils.Service + "/" + utils.ObjKey(svcObj)
			}
			meta, err := meta.Accessor(svcObj)
			if err == nil {
				resVer := meta.GetResourceVersion()
				objects.SharedResourceVerInstanceLister().Save(key, resVer)
			}
			nodes.DequeueIngestion(key, true)
		}
	}

	if lib.GetServiceType() == lib.NodePortLocal {
		podObjs, err := utils.GetInformers().PodInformer.Lister().Pods(metav1.NamespaceAll).List(labels.Everything())
		if err != nil {
			utils.AviLog.Errorf("Unable to retrieve the Pods during full sync: %s", err)
			return err
		}
		for _, podObj := range podObjs {
			podLabel := utils.ObjKey(podObj)
			ns := strings.Split(podLabel, "/")
			if lib.IsNamespaceBlocked(ns[0]) {
				continue
			}
			key := utils.Pod + "/" + podLabel
			if _, ok := podObj.GetAnnotations()[lib.NPLPodAnnotation]; !ok {
				utils.AviLog.Warnf("key : %s, msg: 'nodeportlocal.antrea.io' annotation not found, ignoring the pod", key)
				continue
			}
			meta, err := meta.Accessor(podObj)
			if err == nil {
				resVer := meta.GetResourceVersion()
				objects.SharedResourceVerInstanceLister().Save(key, resVer)
			}
			nodes.DequeueIngestion(key, true)
		}
	}

	if !lib.GetAdvancedL4() {
		hostRuleObjs, err := lib.AKOControlConfig().CRDInformers().HostRuleInformer.Lister().HostRules(metav1.NamespaceAll).List(labels.Set(nil).AsSelector())
		if err != nil {
			utils.AviLog.Errorf("Unable to retrieve the hostrules during full sync: %s", err)
		} else {
			for _, hostRuleObj := range hostRuleObjs {
				key := lib.HostRule + "/" + utils.ObjKey(hostRuleObj)
				meta, err := meta.Accessor(hostRuleObj)
				if err == nil {
					resVer := meta.GetResourceVersion()
					objects.SharedResourceVerInstanceLister().Save(key, resVer)
				}
				if err := c.GetValidator().ValidateHostRuleObj(key, hostRuleObj); err != nil {
					utils.AviLog.Warnf("key: %s, Error retrieved during validation of HostRule: %v", key, err)
				}
				nodes.DequeueIngestion(key, true)
			}
		}

		httpRuleObjs, err := lib.AKOControlConfig().CRDInformers().HTTPRuleInformer.Lister().HTTPRules(metav1.NamespaceAll).List(labels.Set(nil).AsSelector())
		if err != nil {
			utils.AviLog.Errorf("Unable to retrieve the httprules during full sync: %s", err)
		} else {
			for _, httpRuleObj := range httpRuleObjs {
				key := lib.HTTPRule + "/" + utils.ObjKey(httpRuleObj)
				meta, err := meta.Accessor(httpRuleObj)
				if err == nil {
					resVer := meta.GetResourceVersion()
					objects.SharedResourceVerInstanceLister().Save(key, resVer)
				}
				if err := c.GetValidator().ValidateHTTPRuleObj(key, httpRuleObj); err != nil {
					utils.AviLog.Warnf("key: %s, Error retrieved during validation of HTTPRule: %v", key, err)
				}
				nodes.DequeueIngestion(key, true)
			}
		}

		aviInfraObjs, err := lib.AKOControlConfig().CRDInformers().AviInfraSettingInformer.Lister().List(labels.Set(nil).AsSelector())
		if err != nil {
			utils.AviLog.Errorf("Unable to retrieve the avinfrasettings during full sync: %s", err)
		} else {
			for _, aviInfraObj := range aviInfraObjs {
				key := lib.AviInfraSetting + "/" + utils.ObjKey(aviInfraObj)
				meta, err := meta.Accessor(aviInfraObj)
				if err == nil {
					resVer := meta.GetResourceVersion()
					objects.SharedResourceVerInstanceLister().Save(key, resVer)
				}
				if err := c.GetValidator().ValidateAviInfraSetting(key, aviInfraObj); err != nil {
					utils.AviLog.Warnf("key: %s, Error retrieved during validation of AviInfraSetting: %v", key, err)
				}
				nodes.DequeueIngestion(key, true)
			}
		}

		// IngressClass Section
		if utils.GetInformers().IngressClassInformer != nil {
			ingClassObjs, err := utils.GetInformers().IngressClassInformer.Lister().List(labels.Set(nil).AsSelector())
			if err != nil {
				utils.AviLog.Errorf("Unable to retrieve the ingress classess during full sync: %s", err)
			} else {
				for _, ingClass := range ingClassObjs {
					key := utils.IngressClass + "/" + utils.ObjKey(ingClass)
					meta, err := meta.Accessor(ingClass)
					if err == nil {
						resVer := meta.GetResourceVersion()
						objects.SharedResourceVerInstanceLister().Save(key, resVer)
					}
					utils.AviLog.Debugf("Dequeue for ingressClass key: %v", key)
					nodes.DequeueIngestion(key, true)
				}
			}
		}

		// Ingress Section
		if utils.GetInformers().IngressInformer != nil {
			for namespace := range acceptedNamespaces {
				ingObjs, err := utils.GetInformers().IngressInformer.Lister().Ingresses(namespace).List(labels.Set(nil).AsSelector())
				if err != nil {
					utils.AviLog.Errorf("Unable to retrieve the ingresses during full sync: %s", err)
				} else {
					for _, ingObj := range ingObjs {
						key := utils.Ingress + "/" + utils.ObjKey(ingObj)
						// optimization to check if ingress belongs to ingressClass handled by AKO.
						if !lib.ValidateIngressForClass(key, ingObj) {
							continue
						}
						meta, err := meta.Accessor(ingObj)
						if err == nil {
							resVer := meta.GetResourceVersion()
							objects.SharedResourceVerInstanceLister().Save(key, resVer)
						}
						utils.AviLog.Debugf("Dequeue for ingress key: %v", key)
						nodes.DequeueIngestion(key, true)
					}
				}
			}
		}
		//Route Section
		if utils.GetInformers().RouteInformer != nil {
			routeObjs, err := utils.GetInformers().RouteInformer.Lister().List(labels.Set(nil).AsSelector())
			if err != nil {
				utils.AviLog.Errorf("Unable to retrieve the routes during full sync: %s", err)
			} else {
				for _, routeObj := range routeObjs {
					if _, ok := acceptedNamespaces[routeObj.Namespace]; !ok {
						continue
					}
					key := utils.OshiftRoute + "/" + utils.ObjKey(routeObj)
					meta, err := meta.Accessor(routeObj)
					if err == nil {
						resVer := meta.GetResourceVersion()
						objects.SharedResourceVerInstanceLister().Save(key, resVer)
					}
					utils.AviLog.Debugf("Dequeue for route key: %v", key)
					nodes.DequeueIngestion(key, true)
				}
			}
		}
		if lib.UseServicesAPI() {
			gatewayObjs, err := lib.AKOControlConfig().SvcAPIInformers().GatewayInformer.Lister().Gateways(metav1.NamespaceAll).List(labels.Set(nil).AsSelector())
			if err != nil {
				utils.AviLog.Errorf("Unable to retrieve the gateways during full sync: %s", err)
				return err
			}
			for _, gatewayObj := range gatewayObjs {
				gatewayLabel := utils.ObjKey(gatewayObj)
				ns := strings.Split(gatewayLabel, "/")
				if !lib.IsNamespaceBlocked(ns[0]) && utils.CheckIfNamespaceAccepted(ns[0]) {
					key := lib.Gateway + "/" + gatewayLabel
					meta, err := meta.Accessor(gatewayObj)
					if err == nil {
						resVer := meta.GetResourceVersion()
						objects.SharedResourceVerInstanceLister().Save(key, resVer)
					}
					InformerStatusUpdatesForSvcApiGateway(key, gatewayObj)
					nodes.DequeueIngestion(key, true)
				}
			}

			gwClassObjs, err := lib.AKOControlConfig().SvcAPIInformers().GatewayClassInformer.Lister().List(labels.Set(nil).AsSelector())
			if err != nil {
				utils.AviLog.Errorf("Unable to retrieve the gatewayclasses during full sync: %s", err)
				return err
			}
			for _, gwClassObj := range gwClassObjs {
				key := lib.GatewayClass + "/" + utils.ObjKey(gwClassObj)
				meta, err := meta.Accessor(gwClassObj)
				if err == nil {
					resVer := meta.GetResourceVersion()
					objects.SharedResourceVerInstanceLister().Save(key, resVer)
				}
				nodes.DequeueIngestion(key, true)
			}
		}
		if utils.IsMultiClusterIngressEnabled() {
			mciObjs, err := utils.GetInformers().MultiClusterIngressInformer.Lister().MultiClusterIngresses(metav1.NamespaceAll).List(labels.Set(nil).AsSelector())
			if err != nil {
				utils.AviLog.Errorf("Unable to retrieve the multi-cluster ingresses during full sync: %s", err)
				return err
			}
			for _, mciObj := range mciObjs {
				mciLabel := utils.ObjKey(mciObj)
				ns := strings.Split(mciLabel, "/")
				if !lib.IsNamespaceBlocked(ns[0]) && utils.CheckIfNamespaceAccepted(ns[0]) {
					key := lib.MultiClusterIngress + "/" + mciLabel
					meta, err := meta.Accessor(mciObj)
					if err == nil {
						resVer := meta.GetResourceVersion()
						objects.SharedResourceVerInstanceLister().Save(key, resVer)
					}
					nodes.DequeueIngestion(key, true)
				}
			}
			siObjs, err := utils.GetInformers().ServiceImportInformer.Lister().ServiceImports(metav1.NamespaceAll).List(labels.Set(nil).AsSelector())
			if err != nil {
				utils.AviLog.Errorf("Unable to retrieve the service imports during full sync: %s", err)
				return err
			}
			for _, siObj := range siObjs {
				siLabel := utils.ObjKey(siObj)
				ns := strings.Split(siLabel, "/")
				if !lib.IsNamespaceBlocked(ns[0]) && utils.CheckIfNamespaceAccepted(ns[0]) {
					key := lib.MultiClusterIngress + "/" + siLabel
					meta, err := meta.Accessor(siObj)
					if err == nil {
						resVer := meta.GetResourceVersion()
						objects.SharedResourceVerInstanceLister().Save(key, resVer)
					}
					nodes.DequeueIngestion(key, true)
				}
			}
		}
	} else {
		//Gateway Section

		gatewayObjs, err := lib.AKOControlConfig().AdvL4Informers().GatewayInformer.Lister().Gateways(metav1.NamespaceAll).List(labels.Set(nil).AsSelector())
		if err != nil {
			utils.AviLog.Errorf("Unable to retrieve the gateways during full sync: %s", err)
			return err
		}
		for _, gatewayObj := range gatewayObjs {
			gatewayLabel := utils.ObjKey(gatewayObj)
			ns := strings.Split(gatewayLabel, "/")
			if lib.IsNamespaceBlocked(ns[0]) {
				continue
			}
			key := lib.Gateway + "/" + utils.ObjKey(gatewayObj)
			InformerStatusUpdatesForGateway(key, gatewayObj)
			nodes.DequeueIngestion(key, true)
		}

		gwClassObjs, err := lib.AKOControlConfig().AdvL4Informers().GatewayClassInformer.Lister().List(labels.Set(nil).AsSelector())
		if err != nil {
			utils.AviLog.Errorf("Unable to retrieve the gatewayclasses during full sync: %s", err)
			return err
		}
		for _, gwClassObj := range gwClassObjs {
			key := lib.GatewayClass + "/" + utils.ObjKey(gwClassObj)
			nodes.DequeueIngestion(key, true)
		}
	}
	if sync {
		c.publishAllParentVSKeysToRestLayer()
	}
	return nil
}

func (c *AviController) publishAllParentVSKeysToRestLayer() {
	cache := avicache.SharedAviObjCache()
	vsKeys := cache.VsCacheMeta.AviCacheGetAllParentVSKeys()
	utils.AviLog.Debugf("Got the VS keys: %s", vsKeys)
	allModelsMap := objects.SharedAviGraphLister().GetAll()
	allModels := make(map[string]struct{})
	vrfModelName := lib.GetModelName(lib.GetTenant(), lib.GetVrf())
	for modelName := range allModelsMap.(map[string]interface{}) {
		// ignore vrf model, as it has been published already
		if modelName != vrfModelName && !lib.IsIstioKey(modelName) {
			allModels[modelName] = struct{}{}
		}
	}
	sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
	syncNamespace := lib.GetNamespaceToSync()
	for _, vsCacheKey := range vsKeys {
		modelName := vsCacheKey.Namespace + "/" + vsCacheKey.Name
		// Reverse map the model key from this.
		if syncNamespace != "" {
			shardVsPrefix := lib.ShardVSPrefix
			if shardVsPrefix != "" {
				if strings.HasPrefix(vsCacheKey.Name, shardVsPrefix) {
					delete(allModels, modelName)
					utils.AviLog.Infof("Model published L7 VS during namespace based sync: %s", modelName)
					nodes.PublishKeyToRestLayer(modelName, "fullsync", sharedQueue)
				}
			}
			// For namespace based syncs, the L4 VSes would be named: clusterName + "--" + namespace
			if strings.HasPrefix(vsCacheKey.Name, lib.GetNamePrefix()+syncNamespace) {
				delete(allModels, modelName)
				utils.AviLog.Infof("Model published L4 VS during namespace based sync: %s", modelName)
				nodes.PublishKeyToRestLayer(modelName, "fullsync", sharedQueue)
			}
		} else {
			delete(allModels, modelName)
			utils.AviLog.Infof("Model published in full sync %s", modelName)
			nodes.PublishKeyToRestLayer(modelName, "fullsync", sharedQueue)
		}
	}
	// Now also publish the newly generated models (if any)
	// Publish all the models to REST layer.
	utils.AviLog.Debugf("Newly generated models that do not exist in cache %s", utils.Stringify(allModels))
	for modelName := range allModels {
		nodes.PublishKeyToRestLayer(modelName, "fullsync", sharedQueue)
	}
}

// DeleteModels : Delete models and add the model name in the queue.
// The rest layer would pick up the model key and delete the objects in Avi
func (c *AviController) DeleteModels() {
	utils.AviLog.Infof("Deletion of all avi objects triggered")
	publisher := status.NewStatusPublisher()
	publisher.AddStatefulSetAnnotation(lib.ObjectDeletionStartStatus)
	allModels := objects.SharedAviGraphLister().GetAll()
	allModelsMap := allModels.(map[string]interface{})
	if len(allModelsMap) == 0 {
		utils.AviLog.Infof("No Avi Object to delete, status would be updated in Statefulset")
		publisher.AddStatefulSetAnnotation(lib.ObjectDeletionDoneStatus)
		return
	}
	sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
	for modelName, avimodelIntf := range allModelsMap {
		objects.SharedAviGraphLister().Save(modelName, nil)
		if avimodelIntf != nil {
			avimodel := avimodelIntf.(*nodes.AviObjectGraph)
			// for vrf, delete all static routes
			if avimodel.IsVrf {
				newAviModel := nodes.NewAviObjectGraph()
				newAviModel.IsVrf = true
				aviVrfNode := &nodes.AviVrfNode{
					Name: lib.GetVrf(),
				}
				newAviModel.AddModelNode(aviVrfNode)
				newAviModel.CalculateCheckSum()
				objects.SharedAviGraphLister().Save(modelName, newAviModel)
			}
		}
		bkt := utils.Bkt(modelName, sharedQueue.NumWorkers)
		utils.AviLog.Infof("Deleting objects for model: %s", modelName)
		sharedQueue.Workqueue[bkt].AddRateLimited(modelName)
	}

	DeleteNPLAnnotations()
}

func SetDeleteSyncChannel() {
	// Wait for maximum 30 minutes for the sync to get completed
	if lib.ConfigDeleteSyncChan == nil {
		lib.SetConfigDeleteSyncChan()
	}

	select {
	case <-lib.ConfigDeleteSyncChan:
		status.NewStatusPublisher().AddStatefulSetAnnotation(lib.ObjectDeletionDoneStatus)
		utils.AviLog.Infof("Processing done for deleteConfig, user would be notified through statefulset update")
		lib.AKOControlConfig().PodEventf(corev1.EventTypeNormal, lib.AKODeleteConfigDone, "AKO has removed all objects from Avi Controller")

	case <-time.After(lib.AviObjDeletionTime * time.Minute):
		status.NewStatusPublisher().AddStatefulSetAnnotation(lib.ObjectDeletionTimeoutStatus)
		utils.AviLog.Warnf("Timed out while waiting for rest layer to respond for delete config")
		lib.AKOControlConfig().PodEventf(corev1.EventTypeNormal, lib.AKODeleteConfigTimeout, "Timed out while waiting for rest layer to respond for delete config")
	}

}

func DeleteNPLAnnotations() {
	if !lib.AutoAnnotateNPLSvc() {
		return
	}
	publisher := status.NewStatusPublisher()
	// Delete NPL annotations from the Services
	allSvcIntf := objects.SharedClusterIpLister().GetAll()
	allSvcs, ok := allSvcIntf.(map[string]interface{})
	if !ok {
		utils.AviLog.Infof("Can not delete NPL annotations, wrong type of object in ClusterIpLister: %T", allSvcIntf)
	} else {
		for nsSvc := range allSvcs {
			ns, _, svc := lib.ExtractTypeNameNamespace(nsSvc)
			publisher.DeleteNPLAnnotation(nsSvc, ns, svc)
		}
	}
	objects.SharedlbLister().GetAll()
	allLBSvcIntf := objects.SharedlbLister().GetAll()
	allLBSvcs, ok := allLBSvcIntf.(map[string]interface{})
	if !ok {
		utils.AviLog.Infof("Can not delete NPL annotations, wrong type of object in lbLister: %T", allLBSvcIntf)
	} else {
		for nsSvc := range allLBSvcs {
			ns, _, svc := lib.ExtractTypeNameNamespace(nsSvc)
			publisher.DeleteNPLAnnotation(nsSvc, ns, svc)
		}
	}
}

func SyncFromIngestionLayer(key interface{}, wg *sync.WaitGroup) error {
	// This method will do all necessary graph calculations on the Graph Layer
	// Let's route the key to the graph layer.
	// NOTE: There's no error propagation from the graph layer back to the workerqueue. We will evaluate
	// This condition in the future and visit as needed. But right now, there's no necessity for it.
	// sharedQueue := SharedWorkQueueWrappers().GetQueueByName(queue.GraphLayer)

	keyStr, ok := key.(string)
	if !ok {
		utils.AviLog.Warnf("Unexpected object type: expected string, got %T", key)
		return nil
	}
	nodes.DequeueIngestion(keyStr, false)
	return nil
}

func SyncFromFastRetryLayer(key interface{}, wg *sync.WaitGroup) error {
	keyStr, ok := key.(string)
	if !ok {
		utils.AviLog.Warnf("Unexpected object type: expected string, got %T", key)
		return nil
	}
	retry.DequeueFastRetry(keyStr)
	return nil
}

func SyncFromSlowRetryLayer(key interface{}, wg *sync.WaitGroup) error {
	keyStr, ok := key.(string)
	if !ok {
		utils.AviLog.Warnf("Unexpected object type: expected string, got %T", key)
		return nil
	}
	retry.DequeueSlowRetry(keyStr)
	return nil
}

func SyncFromNodesLayer(key interface{}, wg *sync.WaitGroup) error {
	keyStr, ok := key.(string)
	if !ok {
		utils.AviLog.Warnf("Unexpected object type: expected string, got %T", key)
		return nil
	}
	cache := avicache.SharedAviObjCache()
	aviclient := avicache.SharedAVIClients()
	restlayer := rest.NewRestOperations(cache, aviclient)
	restlayer.DequeueNodes(keyStr)
	return nil
}

func SyncFromStatusQueue(key interface{}, wg *sync.WaitGroup) error {
	publisher := status.NewStatusPublisher()
	publisher.DequeueStatus(key)
	return nil
}

// Controller Specific method
func (c *AviController) InitializeNamespaceSync() {
	nsLabelToSyncKey, nsLabelToSyncVal := lib.GetLabelToSyncNamespace()
	if nsLabelToSyncKey != "" {
		utils.AviLog.Debugf("Initializing Namespace Sync. Received namespace label: %s = %s", nsLabelToSyncKey, nsLabelToSyncVal)
		utils.InitializeNSSync(nsLabelToSyncKey, nsLabelToSyncVal)
	}
	nsFilterObj := utils.GetGlobalNSFilter()
	if !nsFilterObj.EnableMigration {
		utils.AviLog.Info("Namespace Sync is disabled.")
		return
	}
	populateNamespaceList()
	utils.AviLog.Info("Namespace Sync is enabled")
}

// Add namespaces with correct labels to list of valid Namespaces. This is used while populating status of k8s objects.
func populateNamespaceList() {
	k8sclient := utils.GetInformers().ClientSet
	allNamespaces, err := k8sclient.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		utils.AviLog.Errorf("Error en getting all namespaces: %v", err.Error())
		return
	}
	for _, ns := range allNamespaces.Items {
		if utils.CheckIfNamespaceAccepted(ns.GetName(), ns.GetLabels(), false) {
			utils.AddNamespaceToFilter(ns.GetName())
			utils.AviLog.Debugf("Namespace passed filter, added to valid Namespace list: %s", ns.GetName())
		}
	}
}

func (c *AviController) IstioBootstrap() {
	cs := c.informers.ClientSet
	istioSecret, err := cs.CoreV1().Secrets(utils.GetAKONamespace()).Get(context.TODO(), lib.IstioSecret, metav1.GetOptions{})
	if err == nil {
		rootCA := istioSecret.Data["root-cert"]
		sslKey := istioSecret.Data["key"]
		sslCert := istioSecret.Data["cert-chain"]
		newAviModel := nodes.NewAviObjectGraph()
		newAviModel.IsVrf = false
		newAviModel.Name = lib.IstioModel
		pkinode := &nodes.AviPkiProfileNode{
			Name:   lib.GetIstioPKIProfileName(),
			Tenant: lib.GetTenant(),
			CACert: string(rootCA),
		}
		newAviModel.AddModelNode(pkinode)
		sslNode := &nodes.AviTLSKeyCertNode{
			Name:   lib.GetIstioWorkloadCertificateName(),
			Tenant: lib.GetTenant(),
			Type:   lib.CertTypeVS,
			Cert:   sslCert,
			Key:    sslKey,
		}
		newAviModel.AddModelNode(sslNode)

		cache := avicache.SharedAviObjCache()
		aviclient := avicache.SharedAVIClients()
		restlayer := rest.NewRestOperations(cache, aviclient)

		key := utils.Secret + "/" + utils.GetAKONamespace() + "/" + lib.IstioSecret
		restlayer.IstioCU(key, newAviModel)
		lib.SetIstioInitialized(true)

	} else {
		utils.AviLog.Fatalf("Could not fetch secret: %s, %v", lib.IstioSecret, err)
	}
}
