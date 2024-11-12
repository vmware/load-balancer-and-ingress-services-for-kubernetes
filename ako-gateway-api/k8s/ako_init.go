/*
 * Copyright 2023-2024 VMware, Inc.
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
	"fmt"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	akogatewayapinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/nodes"
	akogatewayapistatus "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/status"
	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/rest"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/retry"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/status"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func (c *GatewayController) InitController(informers k8s.K8sinformers, registeredInformers []string, ctrlCh <-chan struct{}, stopCh <-chan struct{}, quickSyncCh chan struct{}, waitGroupMap ...map[string]*sync.WaitGroup) {
	// set up signals so we handle the first shutdown signal gracefully
	var worker *utils.FullSyncThread
	informersArg := make(map[string]interface{})

	c.informers = utils.NewInformers(utils.KubeClientIntf{ClientSet: informers.Cs}, registeredInformers, informersArg)

	var ingestionWG *sync.WaitGroup
	var graphWG *sync.WaitGroup
	var fastretryWG *sync.WaitGroup
	var slowretryWG *sync.WaitGroup
	var statusWG *sync.WaitGroup
	if len(waitGroupMap) > 0 {
		// Fetch all the waitgroups
		ingestionWG, _ = waitGroupMap[0]["ingestion"]
		graphWG, _ = waitGroupMap[0]["graph"]
		fastretryWG, _ = waitGroupMap[0]["fastretry"]
		slowretryWG, _ = waitGroupMap[0]["slowretry"]
		statusWG, _ = waitGroupMap[0]["status"]
	}

	/** Sequence:
	  1. Initialize the graph layer queue.
	  2. Do a full sync from main thread and publish all the models.
	  3. Initialize the ingestion layer queue for partial sync.
	  **/
	// start the go routines draining the queues in various layers
	var graphQueue *utils.WorkerQueue
	// This is the first time initialization of the queue. For hostname based sharding, we don't want layer 2 to process the queue using multiple go routines.
	retryQueueWorkers := uint32(1)
	slowRetryQParams := utils.WorkerQueue{NumWorkers: retryQueueWorkers, WorkqueueName: lib.SLOW_RETRY_LAYER, SlowSyncTime: lib.SLOW_SYNC_TIME}
	fastRetryQParams := utils.WorkerQueue{NumWorkers: retryQueueWorkers, WorkqueueName: lib.FAST_RETRY_LAYER}

	//TODO Parallelize workers
	//Every worker can work with a single graph object
	//Each graph object corresponds to a single gateway
	//HTTPRoutes can be attached to multiple gateways
	//This will make HTTPRoute updates affect multiple graphs
	numWorkers := uint32(1)
	ingestionQueueParams := utils.WorkerQueue{NumWorkers: numWorkers, WorkqueueName: utils.ObjectIngestionLayer}

	numGraphWorkers := uint32(8)

	graphQueueParams := utils.WorkerQueue{NumWorkers: numGraphWorkers, WorkqueueName: utils.GraphLayer}
	statusQueueParams := utils.WorkerQueue{NumWorkers: numGraphWorkers, WorkqueueName: utils.StatusQueue}
	graphQueue = utils.SharedWorkQueue(&ingestionQueueParams, &graphQueueParams, &slowRetryQParams, &fastRetryQParams, &statusQueueParams).GetQueueByName(utils.GraphLayer)

	err := k8s.PopulateCache()
	if err != nil {
		c.DisableSync = true
		utils.AviLog.Errorf("failed to populate cache, disabling sync")
		lib.ShutdownApi()
	}

	// Setup and start event handlers for objects.
	c.addIndexers()
	c.Start(stopCh)

	fullSyncInterval := os.Getenv(utils.FULL_SYNC_INTERVAL)
	interval, err := strconv.ParseInt(fullSyncInterval, 10, 64)

	// Set up the workers but don't start draining them.
	if err != nil {
		utils.AviLog.Errorf("Cannot convert full sync interval value to integer, pls correct the value and restart AKO. Error: %s", err)
	} else {
		// First boot sync
		err = c.FullSyncK8s(false)
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

	}

	c.cleanupStaleVSes()

	graphQueue.SyncFunc = SyncFromNodesLayer
	graphQueue.Run(stopCh, graphWG)

	c.SetupEventHandlers(informers)
	c.SetupGatewayApiEventHandlers(numWorkers)

	if lib.DisableSync {
		akogatewayapilib.AKOControlConfig().PodEventf(corev1.EventTypeNormal, lib.AKODeleteConfigSet, "AKO is in disable sync state")
	} else {
		akogatewayapilib.AKOControlConfig().PodEventf(corev1.EventTypeNormal, lib.AKOReady, "AKO is now listening for Object updates in the cluster")
	}

	ingestionQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	ingestionQueue.SyncFunc = SyncFromIngestionLayer
	ingestionQueue.Run(stopCh, ingestionWG)

	fastRetryQueue := utils.SharedWorkQueue().GetQueueByName(lib.FAST_RETRY_LAYER)
	fastRetryQueue.SyncFunc = SyncFromFastRetryLayer
	fastRetryQueue.Run(stopCh, fastretryWG)

	slowRetryQueue := utils.SharedWorkQueue().GetQueueByName(lib.SLOW_RETRY_LAYER)
	slowRetryQueue.SyncFunc = SyncFromSlowRetryLayer
	slowRetryQueue.Run(stopCh, slowretryWG)

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

	ingestionQueue.StopWorkers(stopCh)
	graphQueue.StopWorkers(stopCh)
	fastRetryQueue.StopWorkers(stopCh)
	slowRetryQueue.StopWorkers(stopCh)
	statusQueue.StopWorkers(stopCh)
}

func (c *GatewayController) addIndexers() {

	if lib.AKOControlConfig().GetEndpointSlicesEnabled() {
		c.informers.EpSlicesInformer.Informer().AddIndexers(
			cache.Indexers{
				discovery.LabelServiceName: func(obj interface{}) ([]string, error) {
					eps, ok := obj.(*discovery.EndpointSlice)
					if !ok {
						utils.AviLog.Debugf("Error indexing epslice object by service name")
						return []string{}, nil
					}
					if val, ok := eps.Labels[discovery.LabelServiceName]; ok && val != "" {
						return []string{eps.Namespace + "/" + val}, nil
					}
					return []string{}, nil
				},
			},
		)
	}
	gwinformer := akogatewayapilib.AKOControlConfig().GatewayApiInformers()
	gwinformer.GatewayInformer.Informer().AddIndexers(
		cache.Indexers{
			lib.GatewayClassGatewayIndex: func(obj interface{}) ([]string, error) {
				gw, ok := obj.(*gatewayv1.Gateway)
				if !ok {
					return []string{}, nil
				}
				return []string{string(gw.Spec.GatewayClassName)}, nil
			},
		},
	)
	gwinformer.GatewayClassInformer.Informer().AddIndexers(
		cache.Indexers{
			akogatewayapilib.GatewayClassGatewayControllerIndex: func(obj interface{}) ([]string, error) {
				gwClass, ok := obj.(*gatewayv1.GatewayClass)
				if !ok {
					return []string{}, nil
				}
				if gwClass.Spec.ControllerName == akogatewayapilib.GatewayController {
					return []string{akogatewayapilib.GatewayController}, nil
				}
				return []string{}, nil
			},
		},
	)
}

func (c *GatewayController) FullSyncK8s(sync bool) error {
	if c.DisableSync {
		utils.AviLog.Infof("Sync disabled, skipping full sync")
		return nil
	}

	// GatewayClass Section
	gwClassObjs, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayClassInformer.Lister().List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Errorf("Unable to retrieve the gatewayclasses during full sync: %s", err)
		return err
	}

	var filteredGatewayClasses []*gatewayv1.GatewayClass
	for _, gwClassObj := range gwClassObjs {
		key := lib.GatewayClass + "/" + utils.ObjKey(gwClassObj)
		meta, err := meta.Accessor(gwClassObj)
		if err == nil {
			resVer := meta.GetResourceVersion()
			objects.SharedResourceVerInstanceLister().Save(key, resVer)
		}
		if IsGatewayClassValid(key, gwClassObj) {
			filteredGatewayClasses = append(filteredGatewayClasses, gwClassObj)
		}
	}
	for _, filteredGatewayClass := range filteredGatewayClasses {
		key := lib.GatewayClass + "/" + utils.ObjKey(filteredGatewayClass)

		akogatewayapinodes.DequeueIngestion(key, true)
	}

	// Gateway Section
	var filteredGateways []*gatewayv1.Gateway
	gatewayObjs, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayInformer.Lister().Gateways(metav1.NamespaceAll).List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Errorf("Unable to retrieve the gateways during full sync: %s", err)
		return err
	}

	for _, gatewayObj := range gatewayObjs {
		key := lib.Gateway + "/" + utils.ObjKey(gatewayObj)
		meta, err := meta.Accessor(gatewayObj)
		if err == nil {
			resVer := meta.GetResourceVersion()
			objects.SharedResourceVerInstanceLister().Save(key, resVer)
		}
		if valid, _ := IsValidGateway(key, gatewayObj); valid {
			filteredGateways = append(filteredGateways, gatewayObj)
		}
	}
	sort.Slice(filteredGateways, func(i, j int) bool {
		if filteredGateways[i].GetCreationTimestamp().Unix() == filteredGateways[j].GetCreationTimestamp().Unix() {
			return filteredGateways[i].Namespace+"/"+filteredGateways[i].Name < filteredGateways[j].Namespace+"/"+filteredGateways[j].Name
		}
		return filteredGateways[i].GetCreationTimestamp().Unix() < filteredGateways[j].GetCreationTimestamp().Unix()
	})
	for _, filteredGateway := range filteredGateways {
		key := lib.Gateway + "/" + utils.ObjKey(filteredGateway)

		akogatewayapinodes.DequeueIngestion(key, true)
	}

	// HTTPRoute Section
	var filteredHTTPRoutes []*gatewayv1.HTTPRoute
	httpRouteObjs, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().HTTPRouteInformer.Lister().HTTPRoutes(metav1.NamespaceAll).List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Errorf("Unable to retrieve the httproutes during full sync: %s", err)
		return err
	}

	for _, httpRouteObj := range httpRouteObjs {
		key := lib.HTTPRoute + "/" + utils.ObjKey(httpRouteObj)
		meta, err := meta.Accessor(httpRouteObj)
		if err == nil {
			resVer := meta.GetResourceVersion()
			objects.SharedResourceVerInstanceLister().Save(key, resVer)
		}
		if IsHTTPRouteConfigValid(key, httpRouteObj) {
			filteredHTTPRoutes = append(filteredHTTPRoutes, httpRouteObj)
		}
	}
	sort.Slice(filteredHTTPRoutes, func(i, j int) bool {
		if filteredHTTPRoutes[i].GetCreationTimestamp().Unix() == filteredHTTPRoutes[j].GetCreationTimestamp().Unix() {
			return filteredHTTPRoutes[i].Namespace+"/"+filteredHTTPRoutes[i].Name < filteredHTTPRoutes[j].Namespace+"/"+filteredHTTPRoutes[j].Name
		}
		return filteredHTTPRoutes[i].GetCreationTimestamp().Unix() < filteredHTTPRoutes[j].GetCreationTimestamp().Unix()
	})
	for _, filteredHTTPRoute := range filteredHTTPRoutes {
		key := lib.HTTPRoute + "/" + utils.ObjKey(filteredHTTPRoute)

		akogatewayapinodes.DequeueIngestion(key, true)
	}

	// Service Section
	svcObjs, err := utils.GetInformers().ServiceInformer.Lister().Services(metav1.NamespaceAll).List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Errorf("Unable to retrieve the services during full sync: %s", err)
		return err
	}

	for _, svcObj := range svcObjs {
		key := utils.Service + "/" + utils.ObjKey(svcObj)
		meta, err := meta.Accessor(svcObj)
		if err == nil {
			resVer := meta.GetResourceVersion()
			objects.SharedResourceVerInstanceLister().Save(key, resVer)
		}
		// Not pushing the service to the next layer as it is
		// not required since we don't create a model out of service
	}
	if lib.GetServiceType() == lib.NodePortLocal {
		podObjs, err := utils.GetInformers().PodInformer.Lister().Pods(metav1.NamespaceAll).List(labels.Everything())
		if err != nil {
			utils.AviLog.Errorf("Unable to retrieve the Pods during full sync: %s", err)
			return err
		}
		for _, podObj := range podObjs {
			podLabel := utils.ObjKey(podObj)
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
			akogatewayapinodes.DequeueIngestion(key, true)
		}
	}

	c.publishAllParentVSKeysToRestLayer()

	return nil
}

func (c *GatewayController) publishAllParentVSKeysToRestLayer() {
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

	for _, vsCacheKey := range vsKeys {
		modelName := vsCacheKey.Namespace + "/" + vsCacheKey.Name
		delete(allModels, modelName)
		utils.AviLog.Infof("Model published in full sync %s", modelName)
		nodes.PublishKeyToRestLayer(modelName, "fullsync", sharedQueue)

	}
	// Now also publish the newly generated models (if any)
	// Publish all the models to REST layer.
	utils.AviLog.Debugf("Newly generated models that do not exist in cache %s", utils.Stringify(allModels))
	for modelName := range allModels {
		nodes.PublishKeyToRestLayer(modelName, "fullsync", sharedQueue)
	}
}

func (c *GatewayController) FullSync() {
	aviRestClientPool := avicache.SharedAVIClients(lib.GetTenant())
	aviObjCache := avicache.SharedAviObjCache()

	// Randomly pickup a client.
	if len(aviRestClientPool.AviClient) > 0 {
		aviObjCache.AviClusterStatusPopulate(aviRestClientPool.AviClient[0])

		aviObjCache.AviCacheRefresh(aviRestClientPool.AviClient[0], utils.CloudName)

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

func SyncFromNodesLayer(key interface{}, wg *sync.WaitGroup) error {
	keyStr, ok := key.(string)
	if !ok {
		utils.AviLog.Warnf("Unexpected object type: expected string, got %T", key)
		return nil
	}
	cache := avicache.SharedAviObjCache()
	restlayer := rest.NewRestOperations(cache)
	restlayer.DequeueNodes(keyStr)
	return nil
}

func (c *GatewayController) RefreshAuthToken() {
	lib.RefreshAuthToken(c.informers.KubeClientIntf.ClientSet)
}

func SyncFromIngestionLayer(key interface{}, wg *sync.WaitGroup) error {
	// This method will do all necessary graph calculations on the Graph Layer
	// Let's route the key to the graph layer.
	// NOTE: There's no error propagation from the graph layer back to the workerqueue. We will evaluate
	// This condition in the future and visit as needed. But right now, there's no necessity for it.

	keyStr, ok := key.(string)
	if !ok {
		utils.AviLog.Warnf("Unexpected object type: expected string, got %T", key)
		return nil
	}
	akogatewayapinodes.DequeueIngestion(keyStr, false)
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
func SyncFromStatusQueue(key interface{}, wg *sync.WaitGroup) error {
	akogatewayapistatus.DequeueStatus(key)
	return nil
}

func (c *GatewayController) cleanupStaleVSes() {

	aviObjCache := avicache.SharedAviObjCache()

	delModels, err := DeleteConfigFromConfigmap(c.informers.ClientSet)
	if err != nil {
		c.DisableSync = true
		utils.AviLog.Errorf("Error occurred while fetching values from configmap. Err: %s", utils.Stringify(err))
		return
	}
	if delModels {
		go SetDeleteSyncChannel()
		parentKeys := aviObjCache.VsCacheMeta.AviCacheGetAllParentVSKeys()
		k8s.DeleteAviObjects(parentKeys, aviObjCache)
	}

	// Delete Stale objects by deleting model for dummy VS
	if _, err := lib.IsClusterNameValid(); err != nil {
		utils.AviLog.Errorf("AKO cluster name is invalid.")
		return
	}
	utils.AviLog.Infof("Starting clean up of stale objects")
	restlayer := rest.NewRestOperations(aviObjCache)
	staleVSKey := lib.GetTenant() + "/" + lib.DummyVSForStaleData
	restlayer.CleanupVS(staleVSKey, true)
	staleCacheKey := avicache.NamespaceName{
		Name:      lib.DummyVSForStaleData,
		Namespace: lib.GetTenant(),
	}
	aviObjCache.VsCacheMeta.AviCacheDelete(staleCacheKey)

	vsKeysPending := aviObjCache.VsCacheMeta.AviGetAllKeys()

	if delModels {
		//Delete NPL annotations
		k8s.DeleteNPLAnnotations()
	}
	if delModels && len(vsKeysPending) == 0 && lib.ConfigDeleteSyncChan != nil {
		close(lib.ConfigDeleteSyncChan)
		lib.ConfigDeleteSyncChan = nil
	}
}

// HandleConfigMap : initialise the controller, start informer for configmap and wait for the ako configmap to be created.
// When the configmap is created, enable sync for other k8s objects. When the configmap is disabled, disable sync.
func (c *GatewayController) HandleConfigMap(k8sinfo k8s.K8sinformers, ctrlCh chan struct{}, stopCh <-chan struct{}, quickSyncCh chan struct{}) error {
	cs := k8sinfo.Cs
	aviClientPool := avicache.SharedAVIClients(lib.GetTenant())
	if aviClientPool == nil || len(aviClientPool.AviClient) < 1 {
		c.DisableSync = true
		lib.SetDisableSync(true)
		utils.AviLog.Errorf("could not get client to connect to Avi Controller, disabling sync")
	}
	aviclient := aviClientPool.AviClient[0]

	var err error

	c.DisableSync, err = DeleteConfigFromConfigmap(cs)
	lib.SetDisableSync(c.DisableSync)
	if err != nil {
		return err
	}

	configMapEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			cm, ok := validateAviConfigMap(obj)
			if !ok {
				return
			}
			utils.AviLog.Infof("avi k8s configmap created")
			utils.AviLog.SetLevel(cm.Data[lib.LOG_LEVEL])
			akogatewayapilib.AKOControlConfig().EventsSetEnabled(cm.Data[lib.EnableEvents])

			delModels := delConfigFromData(cm.Data)

			validateUserInput, err := avicache.ValidateUserInput(aviclient, true)
			if err != nil {
				utils.AviLog.Errorf("Error while validating input: %s", err.Error())
				akogatewayapilib.AKOControlConfig().PodEventf(corev1.EventTypeWarning, lib.SyncDisabled, "Invalid user input %s", err.Error())
			} else {
				akogatewayapilib.AKOControlConfig().PodEventf(corev1.EventTypeNormal, lib.ValidatedUserInput, "User input validation completed.")
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
				akogatewayapilib.AKOControlConfig().EventsSetEnabled(cm.Data[lib.EnableEvents])
			}

			if oldcm.Data[lib.DeleteConfig] == cm.Data[lib.DeleteConfig] {
				return
			}
			// if DeleteConfig value has changed, then check if we need to enable/disable sync
			isValidUserInput, err := avicache.ValidateUserInput(aviclient, true)
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

				} else {
					status.NewStatusPublisher().ResetStatefulSetAnnotation(status.GatewayObjectDeletionStatus)
					akogatewayapilib.AKOControlConfig().PodEventf(corev1.EventTypeNormal, lib.AKODeleteConfigUnset, "DeleteConfig unset in configmap, sync would be enabled")
					quickSyncCh <- struct{}{}
				}
			}

		},
		DeleteFunc: func(obj interface{}) {
			utils.AviLog.Warnf("avi k8s configmap deleted, shutting down api server")
		},
	}

	c.informers.ConfigMapInformer.Informer().AddEventHandler(configMapEventHandler)
	go c.informers.ConfigMapInformer.Informer().Run(stopCh)
	if !cache.WaitForCacheSync(stopCh,
		c.informers.ConfigMapInformer.Informer().HasSynced,
	) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
	} else {
		utils.AviLog.Infof("Caches synced")
	}
	return nil
}

func DeleteConfigFromConfigmap(cs kubernetes.Interface) (bool, error) {
	cmNS := utils.GetAKONamespace()
	cm, err := cs.CoreV1().ConfigMaps(cmNS).Get(context.TODO(), lib.AviConfigMap, metav1.GetOptions{})
	if err == nil {
		return delConfigFromData(cm.Data), err
	}
	utils.AviLog.Warnf("error while reading configmap, sync would be disabled: %v", err)
	return true, err
}
func delConfigFromData(data map[string]string) bool {
	var delConf bool
	if val, ok := data[lib.DeleteConfig]; ok {
		if val == "true" {
			utils.AviLog.Infof("deleteConfig set in configmap, sync would be disabled")
			akogatewayapilib.AKOControlConfig().PodEventf(corev1.EventTypeNormal, lib.AKODeleteConfigSet, "DeleteConfig set in configmap, sync would be disabled")
			delConf = true
		}
	}
	lib.SetDeleteConfigMap(delConf)
	return delConf
}

// DeleteModels : Delete models and add the model name in the queue.
// The rest layer would pick up the model key and delete the objects in Avi
func (c *GatewayController) DeleteModels() {
	utils.AviLog.Infof("Deletion of all avi objects triggered")
	publisher := status.NewStatusPublisher()
	publisher.AddStatefulSetAnnotation(status.GatewayObjectDeletionStatus, lib.ObjectDeletionStartStatus)
	allModels := objects.SharedAviGraphLister().GetAll()
	allModelsMap := allModels.(map[string]interface{})
	if len(allModelsMap) == 0 {
		utils.AviLog.Infof("No Avi Object to delete, status would be updated in Statefulset")
		publisher.AddStatefulSetAnnotation(status.GatewayObjectDeletionStatus, lib.ObjectDeletionDoneStatus)
		return
	}
	sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
	for modelName := range allModelsMap {
		objects.SharedAviGraphLister().Save(modelName, nil)
		bkt := utils.Bkt(modelName, sharedQueue.NumWorkers)
		utils.AviLog.Infof("Deleting objects for model: %s", modelName)
		//graph queue prometheus
		sharedQueue.Workqueue[bkt].AddRateLimited(modelName)
	}
}

func SetDeleteSyncChannel() {
	// Wait for maximum 30 minutes for the sync to get completed
	if lib.ConfigDeleteSyncChan == nil {
		lib.SetConfigDeleteSyncChan()
	}

	select {
	case <-lib.ConfigDeleteSyncChan:
		status.NewStatusPublisher().AddStatefulSetAnnotation(status.GatewayObjectDeletionStatus, lib.ObjectDeletionDoneStatus)
		utils.AviLog.Infof("Processing done for deleteConfig, user would be notified through statefulset update")
		akogatewayapilib.AKOControlConfig().PodEventf(corev1.EventTypeNormal, lib.AKODeleteConfigDone, "AKO has removed all objects from Avi Controller")

	case <-time.After(lib.AviObjDeletionTime * time.Minute):
		status.NewStatusPublisher().AddStatefulSetAnnotation(status.GatewayObjectDeletionStatus, lib.ObjectDeletionTimeoutStatus)
		utils.AviLog.Warnf("Timed out while waiting for rest layer to respond for delete config")
		akogatewayapilib.AKOControlConfig().PodEventf(corev1.EventTypeNormal, lib.AKODeleteConfigTimeout, "Timed out while waiting for rest layer to respond for delete config")
	}

}
