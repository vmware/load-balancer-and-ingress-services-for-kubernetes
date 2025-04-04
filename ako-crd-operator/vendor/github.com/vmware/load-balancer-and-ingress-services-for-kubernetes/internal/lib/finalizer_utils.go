/*
 * Copyright 2022 VMware, Inc.
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
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	svcapiv1alpha1 "sigs.k8s.io/service-apis/apis/v1alpha1"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

// Utility functions to add/remove finalizers on AdvancedL4 Gateways synced by AKO.
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

	_, err := AKOControlConfig().AdvL4Clientset().NetworkingV1alpha1pre1().Gateways(gw.Namespace).Patch(context.TODO(), gw.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{})
	if err != nil {
		utils.AviLog.Warnf("error while patching the gateway with updated finalizers, %v", err)
		return
	}

	utils.AviLog.Debugf("Successfully patched the gateway with finalizers: %v", gw.GetFinalizers())
}

// Utility functions to add/remove finalizers on Gateways synced by AKO.
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

	_, err := AKOControlConfig().ServicesAPIClientset().NetworkingV1alpha1().Gateways(gw.Namespace).Patch(context.TODO(), gw.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{})
	if err != nil {
		utils.AviLog.Warnf("error while patching the gateway with updated finalizers, %v", err)
		return
	}

	utils.AviLog.Debugf("Successfully patched the gateway with finalizers: %v", gw.GetFinalizers())
}

// Utility functions to add/remove finalizers on Ingresses synced by AKO.
func RemoveIngressFinalizer(ing *networkingv1.Ingress) {
	finalizers := utils.Remove(ing.GetFinalizers(), IngressFinalizer)
	ing.SetFinalizers(finalizers)
	UpdateIngressFinalizer(ing)
}

func CheckAndSetIngressFinalizer(ing *networkingv1.Ingress) {
	if !ContainsFinalizer(ing, IngressFinalizer) {
		finalizers := append(ing.GetFinalizers(), IngressFinalizer)
		ing.SetFinalizers(finalizers)
		UpdateIngressFinalizer(ing)
	}
}

func UpdateIngressFinalizer(ing *networkingv1.Ingress) {
	patchPayload, _ := json.Marshal(map[string]interface{}{
		"metadata": map[string][]string{
			"finalizers": ing.GetFinalizers(),
		},
	})

	_, err := utils.GetInformers().ClientSet.NetworkingV1().Ingresses(ing.Namespace).Patch(context.TODO(), ing.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{})
	if err != nil {
		utils.AviLog.Warnf("error while patching the ingress with updated finalizers, %v", err)
		return
	}

	utils.AviLog.Debugf("Successfully patched the ingress with finalizers: %v", ing.GetFinalizers())
}
