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

package status

import (
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func msgFoundInStatus(conditions []appsv1.StatefulSetCondition, msg string) bool {
	for _, c := range conditions {
		if c.Type == lib.AKOConditionType && c.Message == msg {
			return true
		}
	}
	return false
}

// ResetStatefulSetStatus removes the condition set by AKO from AKO statefulset
func ResetStatefulSetStatus() {
	ss, err := utils.GetInformers().ClientSet.AppsV1().StatefulSets(lib.GetAKONamespace()).Get(lib.AKOStatefulSet, metav1.GetOptions{})
	if err != nil {
		utils.AviLog.Warnf("Error in getting ako statefulset: %v", err)
		return
	}

	var foundCondition bool
	for i, c := range ss.Status.Conditions {
		if c.Type == lib.AKOConditionType {
			ss.Status.Conditions = append(ss.Status.Conditions[:i], ss.Status.Conditions[i+1:]...)
			foundCondition = true
			break
		}
	}
	if !foundCondition {
		return
	}

	u, err := utils.GetInformers().ClientSet.AppsV1().StatefulSets(lib.GetAKONamespace()).UpdateStatus(ss)
	if err != nil {
		utils.AviLog.Warnf("Error in updating ako statefulset: %v", err)
		return
	}
	utils.AviLog.Debugf("Successfully reset ako statefulset: %v", u)
}

// AddStatefulSetStatus sets a condition in status of AKO statefulset to the desired value
func AddStatefulSetStatus(msg string) {
	ss, err := utils.GetInformers().ClientSet.AppsV1().StatefulSets(lib.GetAKONamespace()).Get(lib.AKOStatefulSet, metav1.GetOptions{})
	if err != nil {
		utils.AviLog.Warnf("Error in getting ako statefulset: %v", err)
		return
	}

	if msgFoundInStatus(ss.Status.Conditions, msg) {
		return
	}

	var foundCondition bool
	for i, c := range ss.Status.Conditions {
		if c.Type == lib.AKOConditionType {
			ss.Status.Conditions[i].Message = msg
			foundCondition = true
			break
		}
	}

	if !foundCondition {
		cond := appsv1.StatefulSetCondition{
			Type:    lib.AKOConditionType,
			Status:  v1.ConditionTrue,
			Message: msg,
		}
		ss.Status.Conditions = append(ss.Status.Conditions, cond)
	}

	u, err := utils.GetInformers().ClientSet.AppsV1().StatefulSets(lib.GetAKONamespace()).UpdateStatus(ss)
	if err != nil {
		utils.AviLog.Warnf("Error in patching ako statefulset: %v", err)
		return
	}
	utils.AviLog.Debugf("Successfully updated ako statefulset: %v", u)
}
