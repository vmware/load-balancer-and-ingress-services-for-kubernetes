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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
)

func PopulateCache() error {
	avi_rest_client_pool := avicache.SharedAVIClients()
	avi_obj_cache := avicache.SharedAviObjCache()
	// Randomly pickup a client.
	if avi_rest_client_pool != nil && len(avi_rest_client_pool.AviClient) > 0 {
		_, _, err := avi_obj_cache.AviObjCachePopulate(avi_rest_client_pool.AviClient[0], utils.CtrlVersion, utils.CloudName)
		if err != nil {
			utils.AviLog.Warnf("failed to populate avi cache with error: %v", err.Error())
			return err
		}
		// once the l3 cache is populated, we can call the updatestatus functions from here
		restlayer := rest.NewRestOperations(avi_obj_cache, avi_rest_client_pool)
		restlayer.SyncObjectStatuses()
	}

	// Delete Stale objects by deleting model for dummy VS
	aviclient := avicache.SharedAVIClients()
	restlayer := rest.NewRestOperations(avi_obj_cache, aviclient)
	staleVSKey := lib.GetTenant() + "/" + lib.DummyVSForStaleData
	if lib.IsClusterNameValid() && aviclient != nil && len(aviclient.AviClient) > 0 {
		utils.AviLog.Infof("Starting clean up of stale objects")
		restlayer.CleanupVS(staleVSKey, true)
		staleCacheKey := avicache.NamespaceName{
			Name:      lib.DummyVSForStaleData,
			Namespace: lib.GetTenant(),
		}
		avi_obj_cache.VsCacheMeta.AviCacheDelete(staleCacheKey)
	}
	return nil
}

func PopulateNodeCache(cs *kubernetes.Clientset) {
	nodeCache := objects.SharedNodeLister()
	nodeCache.PopulateAllNodes(cs)
}

func delConfigFromData(data map[string]string) bool {
	if val, ok := data[lib.DeleteConfig]; ok {
		if val == "true" {
			utils.AviLog.Infof("deleteConfig set in configmap, sync would be disabled")
			return true
		}
	}
	return false
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
	c.DisableSync = !avicache.ValidateUserInput(aviclient) || deleteConfigFromConfigmap(cs)
	if c.DisableSync {
		return errors.New("Sync is disabled because of configmap unavailability during bootup")
	}
	lib.SetDisableSync(c.DisableSync)

	utils.AviLog.Infof("Creating event broadcaster for handling configmap")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(utils.AviLog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: cs.CoreV1().Events("")})

	configMapEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			cm, ok := validateAviConfigMap(obj)
			if !ok {
				return
			}
			utils.AviLog.Infof("avi k8s configmap created")
			utils.AviLog.SetLevel(cm.Data[lib.LOG_LEVEL])
			delModels := delConfigFromData(cm.Data)
			if !delModels {
				status.ResetStatefulSetStatus()
			}
			c.DisableSync = !avicache.ValidateUserInput(aviclient) || delModels
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

			if oldcm.Data[lib.DeleteConfig] == cm.Data[lib.DeleteConfig] {
				return
			}
			// if DeleteConfig value has changed, then check if we need to enable/disable sync
			isValidUserInput := avicache.ValidateUserInput(aviclient)
			c.DisableSync = !isValidUserInput || delConfigFromData(cm.Data)
			lib.SetDisableSync(c.DisableSync)
			if isValidUserInput {
				if delConfigFromData(cm.Data) {
					c.DeleteModels()
				} else {
					status.ResetStatefulSetStatus()
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

func (c *AviController) InitController(informers K8sinformers, registeredInformers []string, ctrlCh <-chan struct{}, stopCh <-chan struct{}, quickSyncCh chan struct{}, waitGroupMap ...map[string]*sync.WaitGroup) {
	// set up signals so we handle the first shutdown signal gracefully
	var worker *utils.FullSyncThread
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
	if len(waitGroupMap) > 0 {
		// Fetch all the waitgroups
		ingestionwg, _ = waitGroupMap[0]["ingestion"]
		graphwg, _ = waitGroupMap[0]["graph"]
		fastretrywg, _ = waitGroupMap[0]["fastretry"]
		slowretrywg, _ = waitGroupMap[0]["slowretry"]
	}
	c.Start(stopCh)
	/** Sequence:
	  1. Initialize the graph layer queue.
	  2. Do a full sync from main thread and publish all the models.
	  3. Initialize the ingestion layer queue for partial sync.
	  **/
	// start the go routines draining the queues in various layers
	var graphQueue *utils.WorkerQueue
	shardScheme := lib.GetShardScheme()
	// This is the first time initialization of the queue. For hostname based sharding, we don't want layer 2 to process the queue using multiple go routines.
	var retryQueueWorkers uint32
	retryQueueWorkers = 1
	slowRetryQParams := utils.WorkerQueue{NumWorkers: retryQueueWorkers, WorkqueueName: lib.SLOW_RETRY_LAYER, SlowSyncTime: lib.SLOW_SYNC_TIME}
	fastRetryQParams := utils.WorkerQueue{NumWorkers: retryQueueWorkers, WorkqueueName: lib.FAST_RETRY_LAYER}
	var numWorkers uint32
	if shardScheme == lib.HOSTNAME_SHARD_SCHEME {
		numWorkers = 1
		ingestionQueueParams := utils.WorkerQueue{NumWorkers: numWorkers, WorkqueueName: utils.ObjectIngestionLayer}
		numGraphWorkers := lib.GetshardSize()
		graphQueueParams := utils.WorkerQueue{NumWorkers: numGraphWorkers, WorkqueueName: utils.GraphLayer}
		graphQueue = utils.SharedWorkQueue(ingestionQueueParams, graphQueueParams, slowRetryQParams, fastRetryQParams).GetQueueByName(utils.GraphLayer)

	} else {
		// Namespace sharding.
		if lib.IsNodePortMode() {
			// Setting the numWorkers to 1 as single node update in L2 affects multiple ingresses.
			// Cannot have multiple workers working on ingress and node updates.
			numWorkers = 1
		} else {
			numWorkers = utils.NumWorkersIngestion
		}
		ingestionQueueParams := utils.WorkerQueue{NumWorkers: numWorkers, WorkqueueName: utils.ObjectIngestionLayer}
		graphQueueParams := utils.WorkerQueue{NumWorkers: utils.NumWorkersGraph, WorkqueueName: utils.GraphLayer}
		graphQueue = utils.SharedWorkQueue(ingestionQueueParams, graphQueueParams, slowRetryQParams, fastRetryQParams).GetQueueByName(utils.GraphLayer)
	}

	graphQueue.SyncFunc = SyncFromNodesLayer
	graphQueue.Run(stopCh, graphwg)
	fullSyncInterval := os.Getenv(utils.FULL_SYNC_INTERVAL)
	interval, err := strconv.ParseInt(fullSyncInterval, 10, 64)
	if lib.GetAdvancedL4() {
		// Set the error to nil
		err = nil
		interval = 300 // seconds, hardcoded for now.
	}
	// Set up the workers but don't start draining them.
	c.SetupEventHandlers(informers)
	if err != nil {
		utils.AviLog.Errorf("Cannot convert full sync interval value to integer, pls correct the value and restart AKO. Error: %s", err)
	} else {
		// First boot sync
		err = c.FullSyncK8s()
		if err != nil {
			// Something bad sync. We need to return and shutdown the API server
			utils.AviLog.Errorf("Couldn't run full sync successfully on bootup, going to shutdown AKO: %s", err)
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

	ingestionQueue.StopWorkers(stopCh)
	graphQueue.StopWorkers(stopCh)
	fastRetryQueue.StopWorkers(stopCh)
	slowRetryQueue.StopWorkers(stopCh)
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
		for modelName, _ := range allModelsMap.(map[string]interface{}) {
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

func (c *AviController) FullSyncK8s() error {
	if c.DisableSync {
		utils.AviLog.Infof("Sync disabled, skipping full sync")
		return nil
	}
	sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
	var vrfModelName string
	if lib.GetDisableStaticRoute() && !lib.IsNodePortMode() {
		utils.AviLog.Infof("Static route sync disabled, skipping node informers")
	} else {
		lib.SetStaticRouteSyncHandler()
		nodeObjects, _ := utils.GetInformers().NodeInformer.Lister().List(labels.Set(nil).AsSelector())
		for _, node := range nodeObjects {
			key := utils.NodeObj + "/" + node.Name
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
	}

	svcObjs, err := utils.GetInformers().ServiceInformer.Lister().Services("").List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Errorf("Unable to retrieve the services during full sync: %s", err)
		return err
	} else {
		for _, svcObj := range svcObjs {
			isSvcLb := isServiceLBType(svcObj)
			var key string
			if isSvcLb {
				key = utils.L4LBService + "/" + utils.ObjKey(svcObj)
			} else {
				if lib.GetAdvancedL4() {
					continue
				}
				key = utils.Service + "/" + utils.ObjKey(svcObj)
			}
			nodes.DequeueIngestion(key, true)
		}
	}

	if !lib.GetAdvancedL4() {
		hostRuleObjs, err := lib.GetCRDInformers().HostRuleInformer.Lister().HostRules("").List(labels.Set(nil).AsSelector())
		if err != nil {
			utils.AviLog.Errorf("Unable to retrieve the hostrules during full sync: %s", err)
		} else {
			for _, hostRuleObj := range hostRuleObjs {
				key := lib.HostRule + "/" + utils.ObjKey(hostRuleObj)
				nodes.DequeueIngestion(key, true)
			}
		}

		httpRuleObjs, err := lib.GetCRDInformers().HTTPRuleInformer.Lister().HTTPRules("").List(labels.Set(nil).AsSelector())
		if err != nil {
			utils.AviLog.Errorf("Unable to retrieve the httprules during full sync: %s", err)
		} else {
			for _, httpRuleObj := range httpRuleObjs {
				key := lib.HTTPRule + "/" + utils.ObjKey(httpRuleObj)
				nodes.DequeueIngestion(key, true)
			}
		}

		// Ingress Section
		if utils.GetInformers().IngressInformer != nil {
			ingObjs, err := utils.GetInformers().IngressInformer.Lister().Ingresses("").List(labels.Set(nil).AsSelector())
			if err != nil {
				utils.AviLog.Errorf("Unable to retrieve the ingresses during full sync: %s", err)
			} else {
				for _, ingObj := range ingObjs {
					ingLabel := utils.ObjKey(ingObj)
					ns := strings.Split(ingLabel, "/")
					if utils.CheckIfNamespaceAccepted(ns[0], utils.GetGlobalNSFilter(), nil, true) {
						key := utils.Ingress + "/" + ingLabel
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
					// to do move to container-lib
					routeLabel := utils.ObjKey(routeObj)
					ns := strings.Split(routeLabel, "/")
					if utils.CheckIfNamespaceAccepted(ns[0], utils.GetGlobalNSFilter(), nil, true) {
						key := utils.OshiftRoute + "/" + routeLabel
						utils.AviLog.Debugf("Dequeue for route key: %v", key)
						nodes.DequeueIngestion(key, true)
					}
				}
			}
		}
	} else {
		//Gateway Section

		gatewayObjs, err := lib.GetAdvL4Informers().GatewayInformer.Lister().Gateways("").List(labels.Set(nil).AsSelector())
		if err != nil {
			utils.AviLog.Errorf("Unable to retrieve the gateways during full sync: %s", err)
			return err
		} else {
			for _, gatewayObj := range gatewayObjs {
				key := lib.Gateway + "/" + utils.ObjKey(gatewayObj)
				InformerStatusUpdatesForGateway(key, gatewayObj)
				nodes.DequeueIngestion(key, true)
			}
		}

		gwClassObjs, err := lib.GetAdvL4Informers().GatewayClassInformer.Lister().List(labels.Set(nil).AsSelector())
		if err != nil {
			utils.AviLog.Errorf("Unable to retrieve the gatewayclasses during full sync: %s", err)
			return err
		} else {
			for _, gwClassObj := range gwClassObjs {
				key := lib.GatewayClass + "/" + utils.ObjKey(gwClassObj)
				nodes.DequeueIngestion(key, true)
			}
		}
	}

	cache := avicache.SharedAviObjCache()
	vsKeys := cache.VsCacheMeta.AviCacheGetAllParentVSKeys()
	utils.AviLog.Debugf("Got the VS keys: %s", vsKeys)
	allModelsMap := objects.SharedAviGraphLister().GetAll()
	var allModels []string
	for modelName, _ := range allModelsMap.(map[string]interface{}) {
		// ignore vrf model, as it has been published already
		if modelName != vrfModelName {
			allModels = append(allModels, modelName)
		}
	}
	if len(vsKeys) != 0 {
		for _, vsCacheKey := range vsKeys {
			// Reverse map the model key from this.
			if lib.GetNamespaceToSync() != "" {
				shardVsPrefix := lib.ShardVSPrefix
				if shardVsPrefix != "" {
					if strings.HasPrefix(vsCacheKey.Name, shardVsPrefix) {
						modelName := vsCacheKey.Namespace + "/" + vsCacheKey.Name
						if utils.HasElem(allModels, modelName) {
							allModels = utils.Remove(allModels, modelName)
						}
						utils.AviLog.Infof("Model published L7 VS during namespace based sync: %s", modelName)
						nodes.PublishKeyToRestLayer(modelName, "fullsync", sharedQueue)
					}
				}
				// For namespace based syncs, the L4 VSes would be named: clusterName + "--" + namespace
				if strings.HasPrefix(vsCacheKey.Name, lib.GetNamePrefix()+lib.GetNamespaceToSync()) {
					modelName := vsCacheKey.Namespace + "/" + vsCacheKey.Name
					if utils.HasElem(allModels, modelName) {
						allModels = utils.Remove(allModels, modelName)
					}
					utils.AviLog.Infof("Model published L4 VS during namespace based sync: %s", modelName)
					nodes.PublishKeyToRestLayer(modelName, "fullsync", sharedQueue)
				}
			} else {
				modelName := vsCacheKey.Namespace + "/" + vsCacheKey.Name
				if utils.HasElem(allModels, modelName) {
					allModels = utils.Remove(allModels, modelName)
				}
				utils.AviLog.Infof("Model published in full sync %s", modelName)
				nodes.PublishKeyToRestLayer(modelName, "fullsync", sharedQueue)
			}
		}
	}
	// Now also publish the newly generated models (if any)
	// Publish all the models to REST layer.
	utils.AviLog.Debugf("Newly generated models that do not exist in cache %s", utils.Stringify(allModels))
	if allModels != nil {
		for _, modelName := range allModels {
			nodes.PublishKeyToRestLayer(modelName, "fullsync", sharedQueue)
		}
	}
	return nil
}

// DeleteModels : Delete models and add the model name in the queue.
// The rest layer would pick up the model key and delete the objects in Avi
func (c *AviController) DeleteModels() {
	utils.AviLog.Infof("Deletion of all avi objects triggered")
	status.AddStatefulSetStatus(lib.ObjectDeletionStartStatus, corev1.ConditionTrue)
	allModels := objects.SharedAviGraphLister().GetAll()
	sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
	for modelName, avimodelIntf := range allModels.(map[string]interface{}) {
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

	// Wait for maximum 30 minutes for the sync to get completed
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(lib.AviObjDeletionTime * time.Minute)
		timeout <- true
	}()
	lib.SetConfigDeleteSyncChan()
	select {
	case <-lib.ConfigDeleteSyncChan:
		status.AddStatefulSetStatus(lib.ObjectDeletionDoneStatus, corev1.ConditionFalse)
		utils.AviLog.Infof("Processing done for deleteConfig, user would be notified through statefulset update")
	case <-timeout:
		status.AddStatefulSetStatus(lib.ObjectDeletionTimeoutStatus, corev1.ConditionUnknown)
		utils.AviLog.Warnf("Timed out while waiting for rest layer to respond for delete config")
	}
}

func SyncFromIngestionLayer(key string, wg *sync.WaitGroup) error {
	// This method will do all necessary graph calculations on the Graph Layer
	// Let's route the key to the graph layer.
	// NOTE: There's no error propagation from the graph layer back to the workerqueue. We will evaluate
	// This condition in the future and visit as needed. But right now, there's no necessity for it.
	//sharedQueue := SharedWorkQueueWrappers().GetQueueByName(queue.GraphLayer)
	nodes.DequeueIngestion(key, false)
	return nil
}

func SyncFromFastRetryLayer(key string, wg *sync.WaitGroup) error {
	retry.DequeueFastRetry(key)
	return nil
}

func SyncFromSlowRetryLayer(key string, wg *sync.WaitGroup) error {
	retry.DequeueSlowRetry(key)
	return nil
}

func SyncFromNodesLayer(key string, wg *sync.WaitGroup) error {
	cache := avicache.SharedAviObjCache()
	aviclient := avicache.SharedAVIClients()
	restlayer := rest.NewRestOperations(cache, aviclient)
	restlayer.DeQueueNodes(key)
	return nil
}

//Controller Specific method
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
	utils.AviLog.Info("Namespace Sync is enabled")
}
