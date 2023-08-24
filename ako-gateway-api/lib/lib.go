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

package lib

import (
	"k8s.io/client-go/kubernetes"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func InformersToRegister(kclient *kubernetes.Clientset) ([]string, error) {
	// Initialize the following informers in all AKO deployments. Provide AKO the ability to watch over
	// Services, Endpoints, Secrets, ConfigMaps and Namespaces.
	allInformers := []string{
		utils.ServiceInformer,
		utils.EndpointInformer,
		utils.SecretInformer,
		utils.ConfigMapInformer,
		utils.NSInformer,
	}

	return allInformers, nil
}

// parent vs name format - clustername--gatewayNs-gatewayName-EVH
func GetGatewayParentName(namespace, gwName string) string {
	//clustername > gateway namespace > Gateway-name
	//Adding -EVH prefix to reuse rest layer
	return lib.GetNamePrefix() + namespace + "-" + gwName + "-EVH"
}

// child vs name format - clustername--encoded value of parentNs-parentName-childNs-childName-encodedStr
func GetChildName(parentNs, parentName, routeNs, routeName, matchName string) string {
	name := lib.GetNamePrefix() + parentNs + "-" + parentName + "-" + routeNs + "-" + routeName + "-" + utils.Stringify(utils.Hash(matchName))
	return lib.Encode(name, lib.EVHVS)
}

func GetPoolName(parentNs, parentName, routeNs, routeName, matchName, backendNs, backendName, backendPort string) string {
	name := lib.GetNamePrefix() + parentNs + "-" + parentName + "-" + routeNs + "-" + routeName + "-" + utils.Stringify(utils.Hash(matchName)) + "-" + backendNs + "-" + backendName + "-" + backendPort
	return lib.Encode(name, lib.Pool)
}

func GetPoolGroupName(parentNs, parentName, routeNs, routeName, matchName string) string {
	name := lib.GetNamePrefix() + parentNs + "-" + parentName + "-" + routeNs + "-" + routeName + "-" + utils.Stringify(utils.Hash(matchName))
	return lib.Encode(name, lib.PG)
}

func CheckGatewayClassController(controllerName string) bool {
	return controllerName == lib.AviIngressController
}

func FindListenerByName(name string, listener []gatewayv1beta1.Listener) int {
	for i := range listener {
		if string(listener[i].Name) == name {
			return i
		}
	}
	return -1
}

func FindListenerStatusByName(name string, status []gatewayv1beta1.ListenerStatus) int {
	for i := range status {
		if string(status[i].Name) == name {
			return i
		}
	}
	return -1
}
