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
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
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
	var tokenWorker *utils.FullSyncThread
	informersArg := make(map[string]interface{})

	c.informers = utils.NewInformers(utils.KubeClientIntf{ClientSet: informers.Cs}, registeredInformers, informersArg)

	var ingestionwg *sync.WaitGroup
	var graphwg *sync.WaitGroup
	var fastretrywg *sync.WaitGroup
	var slowretrywg *sync.WaitGroup
	var statusWG *sync.WaitGroup
	if len(waitGroupMap) > 0 {
		// Fetch all the waitgroups
		ingestionwg, _ = waitGroupMap[0]["ingestion"]
		graphwg, _ = waitGroupMap[0]["graph"]
		fastretrywg, _ = waitGroupMap[0]["fastretry"]
		slowretrywg, _ = waitGroupMap[0]["slowretry"]
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
	graphQueue.Run(stopCh, graphwg)

	c.SetupEventHandlers(informers)
	c.SetupGatewayApiEventHandlers(numWorkers)

	if ctrlAuthToken, ok := utils.SharedCtrlProp().AviCacheGet(utils.ENV_CTRL_AUTHTOKEN); ok && ctrlAuthToken != nil && ctrlAuthToken.(string) != "" {
		c.RefreshAuthToken()
	}
	if ctrlAuthToken, ok := utils.SharedCtrlProp().AviCacheGet(utils.ENV_CTRL_AUTHTOKEN); ok && ctrlAuthToken != nil && ctrlAuthToken.(string) != "" {
		tokenWorker = utils.NewFullSyncThread(time.Duration(utils.RefreshAuthTokenInterval) * time.Hour)
		tokenWorker.SyncFunction = c.RefreshAuthToken
		go tokenWorker.Run()
	}
	if lib.DisableSync {
		akogatewayapilib.AKOControlConfig().PodEventf(corev1.EventTypeNormal, lib.AKODeleteConfigSet, "AKO is in disable sync state")
	} else {
		akogatewayapilib.AKOControlConfig().PodEventf(corev1.EventTypeNormal, lib.AKOReady, "AKO is now listening for Object updates in the cluster")
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

	ingestionQueue.StopWorkers(stopCh)
	graphQueue.StopWorkers(stopCh)
	fastRetryQueue.StopWorkers(stopCh)
	slowRetryQueue.StopWorkers(stopCh)
	statusQueue.StopWorkers(stopCh)
}

func (c *GatewayController) addIndexers() {

	gwinformer := akogatewayapilib.AKOControlConfig().GatewayApiInformers()
	gwinformer.GatewayInformer.Informer().AddIndexers(
		cache.Indexers{
			lib.GatewayClassGatewayIndex: func(obj interface{}) ([]string, error) {
				gw, ok := obj.(*gatewayv1beta1.Gateway)
				if !ok {
					return []string{}, nil
				}
				return []string{string(gw.Spec.GatewayClassName)}, nil
			},
		},
	)
}

func (c *GatewayController) FullSyncK8s(sync bool) error {
	if c.DisableSync {
		utils.AviLog.Infof("Sync disabled, skipping full sync")
		return nil
	}
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
		nodes.DequeueIngestion(key, true)
	}

	//Gateway Section
	gatewayObjs, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayInformer.Lister().Gateways(metav1.NamespaceAll).List(labels.Set(nil).AsSelector())
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
		//TODO
		//InformerStatusUpdatesForGateway(key, gatewayObj)
		nodes.DequeueIngestion(key, true)
	}

	gwClassObjs, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayClassInformer.Lister().List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Errorf("Unable to retrieve the gatewayclasses during full sync: %s", err)
		return err
	}
	for _, gwClassObj := range gwClassObjs {
		key := lib.GatewayClass + "/" + utils.ObjKey(gwClassObj)
		nodes.DequeueIngestion(key, true)
	}

	if sync {
		c.publishAllParentVSKeysToRestLayer()
	}
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

func (c *GatewayController) FullSync() {
	aviRestClientPool := avicache.SharedAVIClients()
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
	aviclient := avicache.SharedAVIClients()
	restlayer := rest.NewRestOperations(cache, aviclient)
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
func SyncFromStatusQueue(key interface{}, wg *sync.WaitGroup) error {
	publisher := status.NewStatusPublisher()
	publisher.DequeueStatus(key)
	return nil
}

func (c *GatewayController) cleanupStaleVSes() {

	aviRestClientPool := avicache.SharedAVIClients()
	aviObjCache := avicache.SharedAviObjCache()

	delModels := k8s.DeleteConfigFromConfigmap(c.informers.ClientSet)
	if delModels {
		go k8s.SetDeleteSyncChannel()
		parentKeys := aviObjCache.VsCacheMeta.AviCacheGetAllParentVSKeys()
		k8s.DeleteAviObjects(parentKeys, aviObjCache, aviRestClientPool)
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

	if delModels && len(vsKeysPending) == 0 && lib.ConfigDeleteSyncChan != nil {
		close(lib.ConfigDeleteSyncChan)
		lib.ConfigDeleteSyncChan = nil
	}
}