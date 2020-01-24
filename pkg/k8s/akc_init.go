/*
 * [2013] - [2019] Avi Networks Incorporated
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
	"time"

	avicache "gitlab.eng.vmware.com/orion/akc/pkg/cache"
	"gitlab.eng.vmware.com/orion/akc/pkg/nodes"
	"gitlab.eng.vmware.com/orion/akc/pkg/objects"
	"gitlab.eng.vmware.com/orion/akc/pkg/rest"
	"gitlab.eng.vmware.com/orion/container-lib/utils"
	"k8s.io/apimachinery/pkg/labels"
)

func PopulateCache() {
	avi_rest_client_pool := avicache.SharedAVIClients()
	avi_obj_cache := avicache.SharedAviObjCache()
	// Randomly pickup a client.
	if len(avi_rest_client_pool.AviClient) > 0 {
		avi_obj_cache.AviObjCachePopulate(avi_rest_client_pool.AviClient[0],
			utils.CtrlVersion, utils.CloudName)
	}
	nodeCache := objects.SharedNodeLister()
	nodeCache.PopulateAllNodes()
}

func InitController(informers K8sinformers) {
	// set up signals so we handle the first shutdown signal gracefully
	var worker *utils.FullSyncThread
	stopCh := utils.SetupSignalHandler()
	c := SharedAviController()

	PopulateCache()
	c.Start(stopCh)
	/** Sequence:
	  1. Initialize the graph layer queue.
	  2. Do a full sync from main thread and publish all the models.
	  3. Initialize the ingestion layer queue for partial sync.
	  **/
	// start the go routines draining the queues in various layers
	graphQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
	graphQueue.SyncFunc = SyncFromNodesLayer
	graphQueue.Run(stopCh)
	fullSyncInterval := os.Getenv(utils.FULL_SYNC_INTERVAL)
	interval, err := strconv.ParseInt(fullSyncInterval, 10, 64)
	if err != nil {
		utils.AviLog.Error.Printf("Cannot convert full sync interval value to integer, pls correct the value and restart AKC. Error: %s", err)
	} else {
		// First boot sync
		FullSyncK8s()
		worker = utils.NewFullSyncThread(time.Duration(interval) * time.Second)
		worker.SyncFunction = FullSync
		go worker.Run()
	}
	c.SetupEventHandlers(informers)
	ingestionQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	ingestionQueue.SyncFunc = SyncFromIngestionLayer
	ingestionQueue.Run(stopCh)
	<-stopCh
	if worker != nil {
		worker.Shutdown()
	}
	ingestionQueue.StopWorkers(stopCh)
	graphQueue.StopWorkers(stopCh)
}

func FullSync() {
	avi_rest_client_pool := avicache.SharedAVIClients()
	avi_obj_cache := avicache.SharedAviObjCache()
	// Randomly pickup a client.
	if len(avi_rest_client_pool.AviClient) > 0 {
		avi_obj_cache.AviObjCachePopulate(avi_rest_client_pool.AviClient[0],
			utils.CtrlVersion, utils.CloudName)
	}
	// Not handling any full sync error right now.
	FullSyncK8s()
}

func FullSyncK8s() error {
	// List all the kubernetes resources
	namespaces, err := utils.GetInformers().NSInformer.Lister().List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Error.Printf("Unable to list the namespaces")
		return err
	}
	for _, nsObj := range namespaces {
		svcObjs, err := utils.GetInformers().ServiceInformer.Lister().Services(nsObj.ObjectMeta.Name).List(labels.Set(nil).AsSelector())
		if err != nil {
			utils.AviLog.Error.Printf("Unable to retrieve the services during full sync: %s", err)
			continue
		}
		for _, svcObj := range svcObjs {
			key := utils.Service + "/" + utils.ObjKey(svcObj)
			nodes.DequeueIngestion(key, true)
		}
		ingObjs, err := utils.GetInformers().IngressInformer.Lister().Ingresses(nsObj.ObjectMeta.Name).List(labels.Set(nil).AsSelector())
		if err != nil {
			utils.AviLog.Error.Printf("Unable to retrieve the ingresses during full sync: %s", err)
			continue
		}
		for _, ingObj := range ingObjs {
			key := utils.Ingress + "/" + utils.ObjKey(ingObj)
			nodes.DequeueIngestion(key, true)
		}
	}
	// Publish all the models to REST layer.
	allModels := objects.SharedAviGraphLister().GetAll()
	sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
	if allModels != nil {
		for modelName, aviModel := range allModels.(map[string]interface{}) {
			nodes.PublishKeyToRestLayer(aviModel.(*nodes.AviObjectGraph), modelName, "fullsync", sharedQueue, false)
		}
	}
	return nil
}

func SyncFromIngestionLayer(key string) error {
	// This method will do all necessary graph calculations on the Graph Layer
	// Let's route the key to the graph layer.
	// NOTE: There's no error propagation from the graph layer back to the workerqueue. We will evaluate
	// This condition in the future and visit as needed. But right now, there's no necessity for it.
	//sharedQueue := SharedWorkQueueWrappers().GetQueueByName(queue.GraphLayer)
	nodes.DequeueIngestion(key, false)
	return nil
}

func SyncFromNodesLayer(key string) error {
	cache := avicache.SharedAviObjCache()
	aviclient := avicache.SharedAVIClients()
	restlayer := rest.NewRestOperations(cache, aviclient)
	restlayer.DeQueueNodes(key)
	return nil
}
