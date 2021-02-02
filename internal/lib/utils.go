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

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
)

type NPLAnnotation struct {
	PodPort  int    `json:"podPort"`
	NodeIP   string `json:"nodeIP"`
	NodePort int    `json:"nodePort"`
}

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
		utils.AviLog.Errorf("Unable to retrieve the services : %s", err)
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

func GetPodsFromService(namespace, serviceName string) []utils.NamespaceName {
	var pods []utils.NamespaceName
	svcKey := namespace + "/" + serviceName
	svc, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).Get(serviceName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return pods
		}
		if found, podsIntf := objects.SharedSvcToPodLister().Get(svcKey); found {
			savedPods, ok := podsIntf.([]utils.NamespaceName)
			if ok {
				return savedPods
			}
		}
		return pods
	}

	if len(svc.Spec.Selector) == 0 {
		return pods
	}

	podList, err := utils.GetInformers().PodInformer.Lister().Pods(namespace).List(labels.SelectorFromSet(labels.Set(svc.Spec.Selector)))
	if err != nil {
		utils.AviLog.Warnf("Got error while listing Pods with selector %v: %v", svc.Spec.Selector, err)
		return pods
	}
	for _, pod := range podList {
		pods = append(pods, utils.NamespaceName{Namespace: pod.Namespace, Name: pod.Name})
	}

	objects.SharedSvcToPodLister().Save(svcKey, pods)
	return pods
}

func GetServicesForPod(pod *corev1.Pod) []string {
	var svcList []string
	services, err := utils.GetInformers().ServiceInformer.Lister().List(labels.Everything())
	if err != nil {
		utils.AviLog.Warnf("Got error while listing Services with NPL annotation: %v", err)
		return svcList
	}

	for _, svc := range services {
		if svc.Spec.Type != corev1.ServiceTypeNodePort {
			if matchSvcSelectorPodLabels(svc.Spec.Selector, pod.GetLabels()) {
				svcList = append(svcList, svc.Namespace+"/"+svc.Name)
			}
		}
	}
	return svcList
}

func matchSvcSelectorPodLabels(svcSelector, podLabel map[string]string) bool {
	if len(svcSelector) == 0 {
		return false
	}

	for selectorKey, selectorVal := range svcSelector {
		if labelVal, ok := podLabel[selectorKey]; !ok || selectorVal != labelVal {
			return false
		}
	}
	return true
}
