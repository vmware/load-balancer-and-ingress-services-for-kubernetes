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

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/status"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servicesapi "sigs.k8s.io/service-apis/apis/v1alpha1"
)

// Services API related functions. Parking the functions on this file instead of creating a new one since most of the functionality is same with v1alpha1pre1

func InformerStatusUpdatesForGatewayV1Alpha1(key string, gateway *servicesapi.Gateway) {
	gwStatus := gateway.Status.DeepCopy()
	defer status.UpdateSvcApiGatewayStatusObject(gateway, gwStatus)

	status.InitializeSvcApiGatewayConditions(gwStatus, &gateway.Spec, false)
	gwClassObj, err := lib.GetAdvL4Informers().GatewayClassInformer.Lister().Get(gateway.Spec.GatewayClassName)
	if err != nil {
		status.UpdateSvcApiGatewayStatusGWCondition(gwStatus, &status.UpdateSvcApiGWStatusConditionOptions{
			Type:    "Pending",
			Status:  metav1.ConditionTrue,
			Message: fmt.Sprintf("Corresponding networking.x-k8s.io/gatewayclass not found %s", gateway.Spec.GatewayClassName),
			Reason:  "InvalidGatewayClass",
		})
		utils.AviLog.Warnf("key: %s, msg: Corresponding networking.x-k8s.io/gatewayclass not found %s %v",
			key, gateway.Spec.GatewayClassName, err)
		return
	}

	for _, listener := range gateway.Spec.Listeners {
		gwName, nameOk := listener.Routes.Selector.MatchLabels[lib.GatewayNameLabelKey]
		gwNamespace, nsOk := listener.Routes.Selector.MatchLabels[lib.GatewayNamespaceLabelKey]
		if !nameOk || !nsOk ||
			(nameOk && gwName != gateway.Name) ||
			(nsOk && gwNamespace != gateway.Namespace) {
			status.UpdateSvcApiGatewayStatusGWCondition(gwStatus, &status.UpdateSvcApiGWStatusConditionOptions{
				Type:    "Pending",
				Status:  metav1.ConditionTrue,
				Message: "Incorrect gateway matchLabels configuration",
				Reason:  "InvalidMatchLabels",
			})
			return
		}
	}

	// Additional check to see if the gatewayclass is a valid avi gateway class or not.
	if gwClassObj.Spec.Controller != lib.AviGatewayController {
		// Return an error since this is not our object.
		status.UpdateSvcApiGatewayStatusGWCondition(gwStatus, &status.UpdateSvcApiGWStatusConditionOptions{
			Type:    "Pending",
			Status:  metav1.ConditionTrue,
			Message: fmt.Sprintf("Unable to identify controller %s", gwClassObj.Spec.Controller),
			Reason:  "UnidentifiedController",
		})
	}
}
