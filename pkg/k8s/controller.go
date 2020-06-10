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
	"fmt"
	"os"
	"reflect"
	"sync"

	"ako/pkg/lib"

	"github.com/avinetworks/container-lib/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
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
	informers        *utils.Informers
	dynamicInformers *lib.DynamicInformers
	workqueue        []workqueue.RateLimitingInterface
	DisableSync      bool
}

type K8sinformers struct {
	Cs            kubernetes.Interface
	DynamicClient dynamic.Interface
}

func SharedAviController() *AviController {
	ctrlonce.Do(func() {
		controllerInstance = &AviController{
			worker_id: (uint32(1) << utils.NumWorkersIngestion) - 1,
			//recorder:  recorder,
			informers:        utils.GetInformers(),
			dynamicInformers: lib.GetDynamicInformers(),
			DisableSync:      true,
		}
	})
	return controllerInstance
}

func isNodeUpdated(oldNode, newNode *corev1.Node) bool {
	if oldNode.ResourceVersion == newNode.ResourceVersion {
		return false
	}
	var oldaddr, newaddr string

	oldAddrs := oldNode.Status.Addresses
	newAddrs := newNode.Status.Addresses
	if len(oldAddrs) != len(newAddrs) {
		return true
	}

	for _, addr := range oldAddrs {
		if addr.Type == "InternalIP" {
			oldaddr = addr.Address
			break
		}
	}
	for _, addr := range newAddrs {
		if addr.Type == "InternalIP" {
			newaddr = addr.Address
			break
		}
	}
	if oldaddr != newaddr {
		return true
	}
	if oldNode.Spec.PodCIDR != newNode.Spec.PodCIDR {
		return true
	}

	return false
}

// Consider an ingress has been updated only if spec/annotation is updated
func isIngressUpdated(oldIngress, newIngress *v1beta1.Ingress) bool {
	if oldIngress.ResourceVersion == newIngress.ResourceVersion {
		return false
	}

	oldSpecHash := utils.Hash(utils.Stringify(oldIngress.Spec))
	oldAnnotationHash := utils.Hash(utils.Stringify(oldIngress.Annotations))
	newSpecHash := utils.Hash(utils.Stringify(newIngress.Spec))
	newAnnotationHash := utils.Hash(utils.Stringify(newIngress.Annotations))

	if oldSpecHash != newSpecHash || oldAnnotationHash != newAnnotationHash {
		return true
	}

	return false
}

func (c *AviController) SetupEventHandlers(k8sinfo K8sinformers) {
	cs := k8sinfo.Cs
	utils.AviLog.Debugf("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(utils.AviLog.Debugf)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: cs.CoreV1().Events("")})
	mcpQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	c.workqueue = mcpQueue.Workqueue
	numWorkers := mcpQueue.NumWorkers

	ep_event_handler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				utils.AviLog.Debugf("Sync disabled, skipping sync for endpoint add")
				return
			}
			ep := obj.(*corev1.Endpoints)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(ep))
			key := utils.Endpoints + "/" + utils.ObjKey(ep)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Debugf("key: %s, msg: ADD", key)
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				utils.AviLog.Debugf("Sync disabled, skipping sync for endpoint delete")
				return
			}
			ep, ok := obj.(*corev1.Endpoints)
			if !ok {
				// endpoints was deleted but its final state is unrecorded.
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				ep, ok = tombstone.Obj.(*corev1.Endpoints)
				if !ok {
					utils.AviLog.Errorf("Tombstone contained object that is not an Endpoints: %#v", obj)
					return
				}
			}
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(ep))
			key := utils.Endpoints + "/" + utils.ObjKey(ep)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Debugf("key: %s, msg: DELETE", key)
		},
		UpdateFunc: func(old, cur interface{}) {
			if c.DisableSync {
				utils.AviLog.Debugf("Sync disabled, skipping sync for endpoint update")
				return
			}
			oep := old.(*corev1.Endpoints)
			cep := cur.(*corev1.Endpoints)
			if !reflect.DeepEqual(cep.Subsets, oep.Subsets) {
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(cep))
				key := utils.Endpoints + "/" + utils.ObjKey(cep)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
			}
		},
	}

	svc_event_handler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				utils.AviLog.Debugf("Sync disabled, skipping sync for svc add")
				return
			}
			svc := obj.(*corev1.Service)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(svc))
			isSvcLb := isServiceLBType(svc)
			var key string
			if isSvcLb {
				key = utils.L4LBService + "/" + utils.ObjKey(svc)
			} else {
				key = utils.Service + "/" + utils.ObjKey(svc)
			}
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Debugf("key: %s, msg: ADD", key)
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				utils.AviLog.Debugf("Sync disabled, skipping sync for svc delete")
				return
			}
			svc, ok := obj.(*corev1.Service)
			if !ok {
				// endpoints was deleted but its final state is unrecorded.
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				svc, ok = tombstone.Obj.(*corev1.Service)
				if !ok {
					utils.AviLog.Errorf("Tombstone contained object that is not an Service: %#v", obj)
					return
				}
			}
			isSvcLb := isServiceLBType(svc)
			var key string
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(svc))
			if isSvcLb {
				key = utils.L4LBService + "/" + utils.ObjKey(svc)
			} else {
				key = utils.Service + "/" + utils.ObjKey(svc)
			}
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Debugf("key: %s, msg: DELETE", key)
		},
		UpdateFunc: func(old, cur interface{}) {
			if c.DisableSync {
				utils.AviLog.Debugf("Sync disabled, skipping sync for svc update")
				return
			}
			oldobj := old.(*corev1.Service)
			svc := cur.(*corev1.Service)
			if oldobj.ResourceVersion != svc.ResourceVersion {
				// Only add the key if the resource versions have changed.
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(svc))
				isSvcLb := isServiceLBType(svc)
				var key string
				if isSvcLb {
					key = utils.L4LBService + "/" + utils.ObjKey(svc)
				} else {
					key = utils.Service + "/" + utils.ObjKey(svc)
				}

				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
			}
		},
	}

	ingress_event_handler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				utils.AviLog.Debugf("Sync disabled, skipping sync for ingress add")
				return
			}
			ingress, ok := utils.ToNetworkingIngress(obj)
			if !ok {
				utils.AviLog.Errorf("Unable to convert obj type interface to networking/v1beta1 ingress")
			}

			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(ingress))
			key := utils.Ingress + "/" + utils.ObjKey(ingress)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Debugf("key: %s, msg: ADD", key)
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				utils.AviLog.Debugf("Sync disabled, skipping sync for ingress delete")
				return
			}
			ingress, ok := utils.ToNetworkingIngress(obj)
			if !ok {
				// ingress was deleted but its final state is unrecorded.
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				ingress, ok = tombstone.Obj.(*v1beta1.Ingress)
				if !ok {
					utils.AviLog.Errorf("Tombstone contained object that is not an Ingress: %#v", obj)
					return
				}
			}
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(ingress))
			key := utils.Ingress + "/" + utils.ObjKey(ingress)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Debugf("key: %s, msg: DELETE", key)
		},
		UpdateFunc: func(old, cur interface{}) {
			if c.DisableSync {
				utils.AviLog.Debugf("Sync disabled, skipping sync for ingress update")
				return
			}
			oldobj, okOld := utils.ToNetworkingIngress(old)
			ingress, okNew := utils.ToNetworkingIngress(cur)
			if !okOld || !okNew {
				utils.AviLog.Errorf("Unable to convert obj type interface to networking/v1beta1 ingress")
			}

			if isIngressUpdated(oldobj, ingress) {
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(ingress))
				key := utils.Ingress + "/" + utils.ObjKey(ingress)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
			}
		},
	}

	secret_event_handler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				utils.AviLog.Debugf("Sync disabled, skipping sync for secret add")
				return
			}
			secret := obj.(*corev1.Secret)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(secret))
			key := "Secret" + "/" + utils.ObjKey(secret)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Debugf("key: %s, msg: ADD", key)
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				utils.AviLog.Debugf("Sync disabled, skipping sync for secret delete")
				return
			}
			secret, ok := obj.(*corev1.Secret)
			if !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				secret, ok = tombstone.Obj.(*corev1.Secret)
				if !ok {
					utils.AviLog.Errorf("Tombstone contained object that is not an Ingress: %#v", obj)
					return
				}
			}
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(secret))
			key := "Secret" + "/" + utils.ObjKey(secret)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Debugf("key: %s, msg: DELETE", key)
		},
		UpdateFunc: func(old, cur interface{}) {
			if c.DisableSync {
				utils.AviLog.Debugf("Sync disabled, skipping sync for secret update")
				return
			}
			oldobj := old.(*corev1.Secret)
			secret := cur.(*corev1.Secret)
			if oldobj.ResourceVersion != secret.ResourceVersion {
				// Only add the key if the resource versions have changed.
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(secret))
				key := "Secret" + "/" + utils.ObjKey(secret)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
			}
		},
	}

	node_event_handler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				utils.AviLog.Debugf("Sync disabled, skipping sync for node add")
				return
			}
			node := obj.(*corev1.Node)
			key := utils.NodeObj + "/" + node.Name
			bkt := utils.Bkt(lib.GetTenant(), numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Debugf("key: %s, msg: ADD", key)
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				utils.AviLog.Debugf("Sync disabled, skipping sync for node delete")
				return
			}
			node, ok := obj.(*corev1.Node)
			if !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				node, ok = tombstone.Obj.(*corev1.Node)
				if !ok {
					utils.AviLog.Errorf("Tombstone contained object that is not an Node: %#v", obj)
					return
				}
			}

			key := utils.NodeObj + "/" + node.Name
			bkt := utils.Bkt(lib.GetTenant(), numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Debugf("key: %s, msg: DELETE", key)
		},
		UpdateFunc: func(old, cur interface{}) {
			if c.DisableSync {
				utils.AviLog.Debugf("Sync disabled, skipping sync for node update")
				return
			}
			oldobj := old.(*corev1.Node)
			node := cur.(*corev1.Node)
			key := utils.NodeObj + "/" + node.Name
			if isNodeUpdated(oldobj, node) {
				bkt := utils.Bkt(lib.GetTenant(), numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
			} else {
				utils.AviLog.Debugf("key: %s, msg: node object did not change\n", key)
			}
		},
	}

	if lib.GetCNIPlugin() == lib.CALICO_CNI {
		block_affinity_handler := cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				utils.AviLog.Debugf("calico blockaffinity ADD Event")
				if c.DisableSync {
					utils.AviLog.Debugf("Sync disabled, skipping sync for calico blockaffinity add")
					return
				}
				crd := obj.(*unstructured.Unstructured)
				specJSON, found, err := unstructured.NestedStringMap(crd.UnstructuredContent(), "spec")
				if err != nil || !found {
					utils.AviLog.Warnf("calico blockaffinity spec not found: %+v", err)
					return
				}
				key := utils.NodeObj + "/" + specJSON["name"]
				bkt := utils.Bkt(lib.GetTenant(), numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
			},
			DeleteFunc: func(obj interface{}) {
				utils.AviLog.Debugf("calico blockaffinity DELETE Event")
				if c.DisableSync {
					utils.AviLog.Debugf("Sync disabled, skipping sync for calico blockaffinity delete")
					return
				}
				crd := obj.(*unstructured.Unstructured)
				specJSON, found, err := unstructured.NestedStringMap(crd.UnstructuredContent(), "spec")
				if err != nil || !found {
					utils.AviLog.Warnf("calico blockaffinity spec not found: %+v", err)
					return
				}
				key := utils.NodeObj + "/" + specJSON["name"]
				bkt := utils.Bkt(lib.GetTenant(), numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
			},
		}

		c.dynamicInformers.CalicoBlockAffinityInformer.Informer().AddEventHandler(block_affinity_handler)
	}

	c.informers.EpInformer.Informer().AddEventHandler(ep_event_handler)
	c.informers.ServiceInformer.Informer().AddEventHandler(svc_event_handler)
	c.informers.IngressInformer.Informer().AddEventHandler(ingress_event_handler)
	c.informers.SecretInformer.Informer().AddEventHandler(secret_event_handler)
	if os.Getenv(lib.DISABLE_STATIC_ROUTE_SYNC) == "true" {
		utils.AviLog.Infof("Static route sync disabled, skipping node informers")
	} else {
		c.informers.NodeInformer.Informer().AddEventHandler(node_event_handler)
	}
}

func validateAviConfigMap(obj interface{}) (*corev1.ConfigMap, bool) {
	configMap, ok := obj.(*corev1.ConfigMap)
	if ok && lib.GetNamespaceToSync() != "" {
		// AKO is running for a particular namespace, look for the Avi config map here
		if configMap.Name == lib.AviConfigMap {
			return configMap, true
		}
	} else if ok && configMap.Namespace == lib.AviNS && configMap.Name == lib.AviConfigMap {
		return configMap, true
	}
	return nil, false
}

func (c *AviController) Start(stopCh <-chan struct{}) {
	go c.informers.ServiceInformer.Informer().Run(stopCh)
	go c.informers.EpInformer.Informer().Run(stopCh)
	go c.informers.IngressInformer.Informer().Run(stopCh)
	go c.informers.SecretInformer.Informer().Run(stopCh)
	go c.informers.NodeInformer.Informer().Run(stopCh)
	go c.informers.NSInformer.Informer().Run(stopCh)

	if lib.GetCNIPlugin() == lib.CALICO_CNI {
		go c.dynamicInformers.CalicoBlockAffinityInformer.Informer().Run(stopCh)
	}

	if !cache.WaitForCacheSync(stopCh,
		c.informers.EpInformer.Informer().HasSynced,
		c.informers.ServiceInformer.Informer().HasSynced,
		c.informers.IngressInformer.Informer().HasSynced,
		c.informers.SecretInformer.Informer().HasSynced,
		c.informers.NodeInformer.Informer().HasSynced,
		c.informers.NSInformer.Informer().HasSynced,
	) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
	} else {
		utils.AviLog.Info("Caches synced")
	}
}

func isServiceLBType(svcObj *corev1.Service) bool {
	// If we don't find a service or it is not of type loadbalancer - return false.
	if svcObj.Spec.Type == "LoadBalancer" {
		return true
	}
	return false
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *AviController) Run(stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()

	utils.AviLog.Info("Started the Kubernetes Controller")
	<-stopCh
	utils.AviLog.Info("Shutting down the Kubernetes Controller")

	return nil
}
