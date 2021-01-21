/*
 * Copyright 2020-2021 VMware, Inc.
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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	svcapiv1alpha1 "sigs.k8s.io/service-apis/apis/v1alpha1"
	svcapi "sigs.k8s.io/service-apis/pkg/client/clientset/versioned"
	svcInformer "sigs.k8s.io/service-apis/pkg/client/informers/externalversions/apis/v1alpha1"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

var svcAPICS svcapi.Interface
var svcAPIInformers *ServicesAPIInformers

func SetServicesAPIClientset(cs svcapi.Interface) {
	svcAPICS = cs
}

func GetServicesAPIClientset() svcapi.Interface {
	return svcAPICS
}

type ServicesAPIInformers struct {
	GatewayInformer      svcInformer.GatewayInformer
	GatewayClassInformer svcInformer.GatewayClassInformer
}

func SetSvcAPIsInformers(c *ServicesAPIInformers) {
	svcAPIInformers = c
}

func GetSvcAPIInformers() *ServicesAPIInformers {
	return svcAPIInformers
}

func RemoveSvcApiGatewayFinalizer(gw *svcapiv1alpha1.Gateway) {
	finalizers := utils.Remove(gw.GetFinalizers(), GatewayFinalizer)
	gw.SetFinalizers(finalizers)
	UpdateSvcApiGatewayFinalizer(gw)
}

func CheckAndSetSvcApiGatewayFinalizer(gw *svcapiv1alpha1.Gateway) {
	if !ContainsFinalizer(gw, GatewayFinalizer) {
		finalizers := append(gw.GetFinalizers(), GatewayFinalizer)
		gw.SetFinalizers(finalizers)
		UpdateSvcApiGatewayFinalizer(gw)
	}
}

func UpdateSvcApiGatewayFinalizer(gw *svcapiv1alpha1.Gateway) {
	patchPayload, _ := json.Marshal(map[string]interface{}{
		"metadata": map[string][]string{
			"finalizers": gw.GetFinalizers(),
		},
	})

	_, err := GetServicesAPIClientset().NetworkingV1alpha1().Gateways(gw.Namespace).Patch(context.TODO(), gw.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{})
	if err != nil {
		utils.AviLog.Warnf("error while patching the gateway with updated finalizers, %v", err)
		return
	}

	utils.AviLog.Infof("Successfully patched the gateway with finalizers: %v", gw.GetFinalizers())
}
