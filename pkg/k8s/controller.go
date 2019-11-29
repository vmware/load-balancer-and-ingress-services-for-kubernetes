/*
 * [2013] - [2018] Avi Networks Incorporated
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
	"fmt"
	"reflect"
	"sync"

	"gitlab.eng.vmware.com/orion/container-lib/utils"
	corev1 "k8s.io/api/core/v1"
	extensionv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

var controllerInstance *AviController
var ctrlonce sync.Once

type AviController struct {
	worker_id       uint32
	worker_id_mutex sync.Mutex
	//recorder        record.EventRecorder
	informers *utils.Informers
	workqueue []workqueue.RateLimitingInterface
}

type K8sinformers struct {
	Cs kubernetes.Interface
}

func SharedAviController() *AviController {
	ctrlonce.Do(func() {
		controllerInstance = &AviController{
			worker_id: (uint32(1) << utils.NumWorkersIngestion) - 1,
			//recorder:  recorder,
			informers: utils.GetInformers(),
		}
	})
	return controllerInstance
}

func (c *AviController) SetupEventHandlers(k8sinfo K8sinformers) {
	cs := k8sinfo.Cs
	utils.AviLog.Info.Printf("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(utils.AviLog.Info.Printf)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: cs.CoreV1().Events("")})
	//recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "avi-k8s-controller"})

	mcpQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	c.workqueue = mcpQueue.Workqueue
	numWorkers := mcpQueue.NumWorkers

	ep_event_handler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ep := obj.(*corev1.Endpoints)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(ep))
			key := "Endpoints/" + utils.ObjKey(ep)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Info.Printf("ADD Endpoint key: %s", key)
		},
		DeleteFunc: func(obj interface{}) {
			ep, ok := obj.(*corev1.Endpoints)
			if !ok {
				// endpoints was deleted but its final state is unrecorded.
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Error.Printf("couldn't get object from tombstone %#v", obj)
					return
				}
				ep, ok = tombstone.Obj.(*corev1.Endpoints)
				if !ok {
					utils.AviLog.Error.Printf("Tombstone contained object that is not an Endpoints: %#v", obj)
					return
				}
			}
			ep = obj.(*corev1.Endpoints)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(ep))
			key := "Endpoints/" + utils.ObjKey(ep)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Info.Printf("DELETE Endpoint key: %s", key)
		},
		UpdateFunc: func(old, cur interface{}) {
			oep := old.(*corev1.Endpoints)
			cep := cur.(*corev1.Endpoints)
			if !reflect.DeepEqual(cep.Subsets, oep.Subsets) {
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(cep))
				key := "Endpoints/" + utils.ObjKey(cep)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				utils.AviLog.Info.Printf("UPDATE Endpoint key: %s", key)
			}
		},
	}

	svc_event_handler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			svc := obj.(*corev1.Service)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(svc))
			key := "Service/" + utils.ObjKey(svc)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Info.Printf("ADD Service key: %s", key)
		},
		DeleteFunc: func(obj interface{}) {
			svc, ok := obj.(*corev1.Service)
			if !ok {
				// endpoints was deleted but its final state is unrecorded.
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Error.Printf("couldn't get object from tombstone %#v", obj)
					return
				}
				svc, ok = tombstone.Obj.(*corev1.Service)
				if !ok {
					utils.AviLog.Error.Printf("Tombstone contained object that is not an Service: %#v", obj)
					return
				}
			}
			svc = obj.(*corev1.Service)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(svc))
			key := "Service/" + utils.ObjKey(svc)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Info.Printf("DELETE Service key: %s", key)
		},
		UpdateFunc: func(old, cur interface{}) {
			oldobj := old.(*corev1.Service)
			svc := cur.(*corev1.Service)
			if oldobj.ResourceVersion != svc.ResourceVersion {
				// Only add the key if the resource versions have changed.
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(svc))
				key := "Service/" + utils.ObjKey(svc)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				utils.AviLog.Info.Printf("UPDATE service key: %s", key)
			}
		},
	}

	ingress_event_handler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ingress := obj.(*extensionv1beta1.Ingress)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(ingress))
			key := "Ingress/" + utils.ObjKey(ingress)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Info.Printf("ADD Ingress key: %s", key)
		},
		DeleteFunc: func(obj interface{}) {
			ingress, ok := obj.(*extensionv1beta1.Ingress)
			if !ok {
				// endpoints was deleted but its final state is unrecorded.
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Error.Printf("couldn't get object from tombstone %#v", obj)
					return
				}
				ingress, ok = tombstone.Obj.(*extensionv1beta1.Ingress)
				if !ok {
					utils.AviLog.Error.Printf("Tombstone contained object that is not an Ingress: %#v", obj)
					return
				}
			}
			ingress = obj.(*extensionv1beta1.Ingress)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(ingress))
			key := "Ingress/" + utils.ObjKey(ingress)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Info.Printf("DELETE ingress key: %s", key)
		},
		UpdateFunc: func(old, cur interface{}) {
			oldobj := old.(*extensionv1beta1.Ingress)
			ingress := cur.(*extensionv1beta1.Ingress)
			if oldobj.ResourceVersion != ingress.ResourceVersion {
				// Only add the key if the resource versions have changed.
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(ingress))
				key := "Ingress/" + utils.ObjKey(ingress)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				utils.AviLog.Info.Printf("UPDATE ingress key: %s", key)
			}
		},
	}

	c.informers.EpInformer.Informer().AddEventHandler(ep_event_handler)
	c.informers.ServiceInformer.Informer().AddEventHandler(svc_event_handler)
	c.informers.IngressInformer.Informer().AddEventHandler(ingress_event_handler)

}

func (c *AviController) Start(stopCh <-chan struct{}) {
	go c.informers.ServiceInformer.Informer().Run(stopCh)
	go c.informers.EpInformer.Informer().Run(stopCh)
	go c.informers.IngressInformer.Informer().Run(stopCh)

	if !cache.WaitForCacheSync(stopCh,
		c.informers.EpInformer.Informer().HasSynced,
		c.informers.ServiceInformer.Informer().HasSynced,
		c.informers.IngressInformer.Informer().HasSynced,
	) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
	} else {
		utils.AviLog.Info.Print("Caches synced")
	}
}

// // Run will set up the event handlers for types we are interested in, as well
// // as syncing informer caches and starting workers. It will block until stopCh
// // is closed, at which point it will shutdown the workqueue and wait for
// // workers to finish processing their current work items.
func (c *AviController) Run(stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()

	utils.AviLog.Info.Print("Started the Kubernetes Controller")
	<-stopCh
	utils.AviLog.Info.Print("Shutting down the Kubernetes Controller")

	return nil
}
