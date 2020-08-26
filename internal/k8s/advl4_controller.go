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
	"reflect"
	"time"

	"github.com/avinetworks/ako/internal/lib"
	"github.com/avinetworks/ako/pkg/utils"
	"k8s.io/client-go/tools/cache"

	advl4v1alpha1pre1 "github.com/vmware-tanzu/service-apis/apis/v1alpha1pre1"
	advl4crd "github.com/vmware-tanzu/service-apis/pkg/client/clientset/versioned"
	advl4informer "github.com/vmware-tanzu/service-apis/pkg/client/informers/externalversions"
)

func NewAdvL4Informers(cs advl4crd.Interface) {
	var advl4InformerFactory advl4informer.SharedInformerFactory

	advl4InformerFactory = advl4informer.NewSharedInformerFactoryWithOptions(cs, time.Second*30)
	gatewayInformer := advl4InformerFactory.Networking().V1alpha1pre1().Gateways()
	gatewayClassInformer := advl4InformerFactory.Networking().V1alpha1pre1().GatewayClasses()

	lib.SetAdvL4Informers(&lib.AdvL4Informers{
		GatewayInformer:      gatewayInformer,
		GatewayClassInformer: gatewayClassInformer,
	})
}

// SetupAdvL4EventHandlers handles setting up of AdvL4 event handlers
func (c *AviController) SetupAdvL4EventHandlers(numWorkers uint32) {
	utils.AviLog.Infof("Setting up AdvL4 Event handlers")
	informer := lib.GetAdvL4Informers()

	gatewayEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			gw := obj.(*advl4v1alpha1pre1.Gateway)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(gw))
			key := lib.Gateway + "/" + utils.ObjKey(gw)
			utils.AviLog.Debugf("key: %s, msg: ADD", key)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
		},
		UpdateFunc: func(old, new interface{}) {
			if c.DisableSync {
				return
			}
			oldObj := old.(*advl4v1alpha1pre1.Gateway)
			gw := new.(*advl4v1alpha1pre1.Gateway)
			if !reflect.DeepEqual(oldObj.Spec, gw.Spec) {
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(gw))
				key := lib.Gateway + "/" + utils.ObjKey(gw)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			gw := obj.(*advl4v1alpha1pre1.Gateway)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(gw))
			key := lib.Gateway + "/" + utils.ObjKey(gw)
			utils.AviLog.Debugf("key: %s, msg: DELETE", key)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
		},
	}

	gatewayClassEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			gwclass := obj.(*advl4v1alpha1pre1.GatewayClass)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(gwclass))
			key := lib.GatewayClass + "/" + utils.ObjKey(gwclass)
			utils.AviLog.Debugf("key: %s, msg: ADD", key)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
		},
		UpdateFunc: func(old, new interface{}) {
			if c.DisableSync {
				return
			}
			oldObj := old.(*advl4v1alpha1pre1.GatewayClass)
			gwclass := new.(*advl4v1alpha1pre1.GatewayClass)
			if !reflect.DeepEqual(oldObj.Spec, gwclass.Spec) {
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(gwclass))
				key := lib.GatewayClass + "/" + utils.ObjKey(gwclass)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			gwclass := obj.(*advl4v1alpha1pre1.GatewayClass)
			key := lib.GatewayClass + "/" + utils.ObjKey(gwclass)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(gwclass))
			utils.AviLog.Debugf("key: %s, msg: DELETE", key)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
		},
	}

	informer.GatewayInformer.Informer().AddEventHandler(gatewayEventHandler)
	informer.GatewayClassInformer.Informer().AddEventHandler(gatewayClassEventHandler)

	return
}
