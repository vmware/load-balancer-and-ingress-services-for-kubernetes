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
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"k8s.io/apimachinery/pkg/labels"
)

func ExtractTypeNameNamespace(key string) (string, string, string) {
	segments := strings.Split(key, "/")
	if len(segments) == 3 {
		return segments[0], segments[1], segments[2]
	}
	if len(segments) == 2 {
		return segments[0], "", segments[1]
	}
	return "", "", segments[0]
}

func isServiceLBType(svcObj *corev1.Service) bool {
	// If we don't find a service or it is not of type loadbalancer - return false.
	if svcObj.Spec.Type == "LoadBalancer" {
		return true
	}
	return false
}

func IsServiceNodPortType(svcObj *corev1.Service) bool {
	if svcObj.Spec.Type == NodePort {
		return true
	}
	return false
}

func IsServiceClusterIPType(svcObj *corev1.Service) bool {
	if svcObj.Spec.Type == "ClusterIP" {
		return true
	}
	return false
}

func GetSvcKeysForNodeCRUD() (svcl4Keys []string, svcl7Keys []string) {
	// For NodePort if the node matches the  selector update all L4 services.

	svcObjs, err := utils.GetInformers().ServiceInformer.Lister().Services("").List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Errorf("Unable to retrieve the services during full sync: %s", err)
		return
	}
	for _, svc := range svcObjs {
		var key string
		if isServiceLBType(svc) {
			key = utils.L4LBService + "/" + utils.ObjKey(svc)
			svcl4Keys = append(svcl4Keys, key)
		}
		if IsServiceNodPortType(svc) {
			key = utils.Service + "/" + utils.ObjKey(svc)
			svcl7Keys = append(svcl7Keys, key)
		}
	}
	return svcl4Keys, svcl7Keys

}
