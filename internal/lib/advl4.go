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

package lib

import (
	"context"
	"encoding/json"

	advl4v1alpha1pre1 "github.com/vmware-tanzu/service-apis/apis/v1alpha1pre1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	advl4crd "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/service-apis/client/clientset/versioned"
	advl4informer "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/service-apis/client/informers/externalversions/apis/v1alpha1pre1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var AdvL4Clientset advl4crd.Interface

func SetAdvL4Clientset(cs advl4crd.Interface) {
	AdvL4Clientset = cs
}

func GetAdvL4Clientset() advl4crd.Interface {
	return AdvL4Clientset
}

var AKOAdvL4Informers *AdvL4Informers

type AdvL4Informers struct {
	GatewayInformer      advl4informer.GatewayInformer
	GatewayClassInformer advl4informer.GatewayClassInformer
}

func SetAdvL4Informers(c *AdvL4Informers) {
	AKOAdvL4Informers = c
}

func GetAdvL4Informers() *AdvL4Informers {
	return AKOAdvL4Informers
}

func RemoveGatewayFinalizer(gw *advl4v1alpha1pre1.Gateway) {
	finalizers := utils.Remove(gw.GetFinalizers(), GatewayFinalizer)
	gw.SetFinalizers(finalizers)
	UpdateGatewayFinalizer(gw)
}

func CheckAndSetGatewayFinalizer(gw *advl4v1alpha1pre1.Gateway) {
	if !ContainsFinalizer(gw, GatewayFinalizer) {
		finalizers := append(gw.GetFinalizers(), GatewayFinalizer)
		gw.SetFinalizers(finalizers)
		UpdateGatewayFinalizer(gw)
	}
}

func UpdateGatewayFinalizer(gw *advl4v1alpha1pre1.Gateway) {
	patchPayload, _ := json.Marshal(map[string]interface{}{
		"metadata": map[string][]string{
			"finalizers": gw.GetFinalizers(),
		},
	})

	_, err := GetAdvL4Clientset().NetworkingV1alpha1pre1().Gateways(gw.Namespace).Patch(context.TODO(), gw.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{})
	if err != nil {
		utils.AviLog.Warnf("error while patching the gateway with updated finalizers, %v", err)
		return
	}

	utils.AviLog.Infof("Successfully patched the gateway with finalizers: %v", gw.GetFinalizers())
}
