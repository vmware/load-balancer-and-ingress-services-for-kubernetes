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
	"context"
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

const ObjectDeletionStatus = "AviObjectDeletionStatus"

// ResetStatefulSetStatus removes the condition set by AKO from AKO statefulset
func ResetStatefulSetStatus() {
	ss, err := utils.GetInformers().ClientSet.AppsV1().StatefulSets(utils.GetAKONamespace()).Get(context.TODO(), lib.AKOStatefulSet, metav1.GetOptions{})
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

	u, err := utils.GetInformers().ClientSet.AppsV1().StatefulSets(utils.GetAKONamespace()).UpdateStatus(context.TODO(), ss, metav1.UpdateOptions{})
	if err != nil {
		utils.AviLog.Warnf("Error in updating ako statefulset: %v", err)
		return
	}
	utils.AviLog.Debugf("Successfully reset ako statefulset: %v", u)
}

func ResetStatefulSetAnnotation() {
	ss, err := utils.GetInformers().ClientSet.AppsV1().StatefulSets(utils.GetAKONamespace()).Get(context.TODO(), lib.AKOStatefulSet, metav1.GetOptions{})
	if err != nil {
		utils.AviLog.Warnf("Error in getting ako statefulset: %v", err)
		return
	}
	ann := ss.GetAnnotations()
	if ann == nil {
		return
	}
	if _, ok := ann[ObjectDeletionStatus]; !ok {
		return
	}
	payloadValue := make(map[string]*string)
	// To delete an annotation with patch call, the value has to be set to nil
	payloadValue[ObjectDeletionStatus] = nil

	patchPayload := map[string]interface{}{
		"metadata": map[string]map[string]*string{
			"annotations": payloadValue,
		},
	}
	payloadBytes, _ := json.Marshal(patchPayload)
	_, err = utils.GetInformers().ClientSet.AppsV1().StatefulSets(utils.GetAKONamespace()).Patch(context.TODO(), ss.Name, types.MergePatchType, payloadBytes, metav1.PatchOptions{})

	if err != nil {
		utils.AviLog.Warnf("Error in patching ako statefulset: %v", err)
		return
	}
	utils.AviLog.Infof("Successfully removed annotation %s from ako statefulset", ObjectDeletionStatus)
	lib.AKOControlConfig().PodEventf(corev1.EventTypeNormal, lib.AKODeleteConfigUnset, "DeleteConfig unset in configmap, sync would be enabled")

	//Remove any status from previous versions of AKO
	ResetStatefulSetStatus()
}

func AddStatefulSetAnnotation(reason string) {

	if !lib.AKOControlConfig().IsLeader() {
		utils.AviLog.Debug("AKO is running as a follower, not updating the annotation")
		return
	}

	ss, err := utils.GetInformers().ClientSet.AppsV1().StatefulSets(utils.GetAKONamespace()).Get(context.TODO(), lib.AKOStatefulSet, metav1.GetOptions{})
	if err != nil {
		utils.AviLog.Warnf("Error in getting ako statefulset: %v", err)
		return
	}

	ann := ss.GetAnnotations()
	if ann == nil {
		ann = make(map[string]string)
	}
	if val, ok := ann[ObjectDeletionStatus]; ok {
		if val == reason {
			return
		}
	}
	ann[ObjectDeletionStatus] = reason
	patchPayload := map[string]interface{}{
		"metadata": map[string]map[string]string{
			"annotations": ann,
		},
	}
	payloadBytes, _ := json.Marshal(patchPayload)
	_, err = utils.GetInformers().ClientSet.AppsV1().StatefulSets(utils.GetAKONamespace()).Patch(context.TODO(), ss.Name, types.MergePatchType, payloadBytes, metav1.PatchOptions{})

	if err != nil {
		utils.AviLog.Warnf("Error in patching ako statefulset annotation: %v", err)
		return
	}
	utils.AviLog.Debugf("Successfully updated annotation %s in ako statefulset", ObjectDeletionStatus)
}
