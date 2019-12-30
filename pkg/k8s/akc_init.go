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
	"time"

	avicache "gitlab.eng.vmware.com/orion/akc/pkg/cache"
	"gitlab.eng.vmware.com/orion/akc/pkg/nodes"
	"gitlab.eng.vmware.com/orion/akc/pkg/rest"
	"gitlab.eng.vmware.com/orion/container-lib/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

func PopulateCache() {
	avi_rest_client_pool := avicache.SharedAVIClients()
	avi_obj_cache := avicache.SharedAviObjCache()
	// Randomly pickup a client.
	if len(avi_rest_client_pool.AviClient) > 0 {
		avi_obj_cache.AviObjCachePopulate(avi_rest_client_pool.AviClient[0],
			utils.CtrlVersion, utils.CloudName)
	}
}

func InitController(informers K8sinformers) {
	// set up signals so we handle the first shutdown signal gracefully
	stopCh := utils.SetupSignalHandler()
	c := SharedAviController()

	c.SetupEventHandlers(informers)

	// start the go routines draining the queues in various layers
	ingestionQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	ingestionQueue.SyncFunc = SyncFromIngestionLayer
	ingestionQueue.Run(stopCh)
	graphQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
	graphQueue.SyncFunc = SyncFromNodesLayer
	graphQueue.Run(stopCh)
	c.Start(stopCh)
	// TODO (sudswas): Remove hard coding.
	worker := utils.NewFullSyncThread(50000 * time.Second)
	worker.SyncFunction = FullSync
	go worker.Run()
	<-stopCh
	worker.Shutdown()
	ingestionQueue.StopWorkers(stopCh)
	graphQueue.StopWorkers(stopCh)
}

func FullSync(istioEnabled string) {
	avi_obj_cache := utils.SharedAviObjCache()
	avi_rest_client_pool := utils.SharedAVIClients()
	avi_obj_cache.AviObjCachePopulate(avi_rest_client_pool.AviClient[0],
		utils.CtrlVersion, utils.CloudName)
	FullSyncK8s()
}

func FullSyncK8s() {
	// List all the kubernetes resources
	epObjs, err := utils.GetInformers().EpInformer.Lister().List(labels.Set(nil).AsSelector())
	if err == nil {
		utils.AviLog.Trace.Printf("Obtained all the endpoints :%s", utils.Stringify(epObjs))
	} else {
		utils.AviLog.Warning.Printf("Unable to fetch the endpoints, will not process them as a part of full sync")
	}
	svcObjs, err := utils.GetInformers().ServiceInformer.Lister().List(labels.Set(nil).AsSelector())
	if err == nil {
		utils.AviLog.Trace.Printf("Obtained all the Services :%s", utils.Stringify(svcObjs))
	} else {
		utils.AviLog.Info.Printf("Unable to fetch Services, will not process them as a part of full sync")
	}
	secretObjs, err := utils.GetInformers().SecretInformer.Lister().List(labels.Set(nil).AsSelector())
	if err == nil {
		utils.AviLog.Trace.Printf("Obtained all the Secrets :%s", utils.Stringify(secretObjs))
	} else {
		utils.AviLog.Warning.Printf("Unable to fetch Secrets, will not process them as a part of full sync")
	}
	PublishK8sKeys(svcObjs, epObjs, secretObjs)
}

func PublishK8sKeys(svcObjs []*v1.Service, epObjs []*v1.Endpoints, secretObjs []*v1.Secret) {
	ingestionQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	for _, svcObj := range svcObjs {
		namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(svcObj))
		key := "Service/" + utils.ObjKey(svcObj)
		bkt := utils.Bkt(namespace, ingestionQueue.NumWorkers)
		ingestionQueue.Workqueue[bkt].AddRateLimited(key)
	}

	for _, epObj := range epObjs {
		namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(epObj))
		key := "Endpoints/" + utils.ObjKey(epObj)
		bkt := utils.Bkt(namespace, ingestionQueue.NumWorkers)
		ingestionQueue.Workqueue[bkt].AddRateLimited(key)
	}

}

func SyncFromIngestionLayer(key string) error {
	// This method will do all necessary graph calculations on the Graph Layer
	// Let's route the key to the graph layer.
	// NOTE: There's no error propagation from the graph layer back to the workerqueue. We will evaluate
	// This condition in the future and visit as needed. But right now, there's no necessity for it.
	//sharedQueue := SharedWorkQueueWrappers().GetQueueByName(queue.GraphLayer)
	nodes.DequeueIngestion(key)
	return nil
}

func SyncFromNodesLayer(key string) error {
	cache := avicache.SharedAviObjCache()
	aviclient := avicache.SharedAVIClients()
	restlayer := rest.NewRestOperations(cache, aviclient)
	restlayer.DeQueueNodes(key)
	return nil
}
