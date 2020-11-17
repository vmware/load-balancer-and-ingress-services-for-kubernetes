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
	"encoding/json"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const akoStatefulset = "ako"

// RemoveConfigmapFinalizer : Remove the ako finaliser from configmap. After this the configmap can be deleted by the user
// This can be used to notify the user that all AVI objects have been deleted by AKO.
func RemoveStatefulsetFinalizer() {
	ss, err := utils.GetInformers().ClientSet.AppsV1().StatefulSets(lib.AviNS).Get(akoStatefulset, metav1.GetOptions{})
	//currConfig, err := utils.GetInformers().ConfigMapInformer.Lister().ConfigMaps(lib.AviNS).Get(lib.AviConfigMap)
	if err != nil {
		utils.AviLog.Warnf("Error in getting configmap: %v", err)
		return
	}
	ss.SetFinalizers([]string{})
	UpdateStatefulsetFinalizer(ss, []string{})
	utils.AviLog.Infof("Removed the finalizer %s from avi CM", lib.ConfigmapFinalizer)
}

// SetConfigmapFinalizer : update from configmap with ako finaliser.
// After this the configmap cannot be deleted by the user without clearing the finaliser
func AddStatefulSetFinalizer() {
	ss, err := utils.GetInformers().ClientSet.AppsV1().StatefulSets(lib.AviNS).Get(akoStatefulset, metav1.GetOptions{})
	if err != nil {
		utils.AviLog.Warnf("Error in getting configmap: %v", err)
		return
	}

	if lib.ContainsFinalizer(ss, lib.ConfigmapFinalizer) {
		utils.AviLog.Warnf("Avi configmap already has the finaliser: %s", lib.ConfigmapFinalizer)
		return
	}

	UpdateStatefulsetFinalizer(ss, []string{lib.ConfigmapFinalizer})
	utils.AviLog.Infof("Successfully patched the CM with finalizers: %v", ss.GetFinalizers())
}

func UpdateStatefulsetFinalizer(ss *appsv1.StatefulSet, finalizerStr []string) {
	ss.SetFinalizers(finalizerStr)
	patchPayload, _ := json.Marshal(map[string]interface{}{
		"metadata": map[string][]string{
			"finalizers": finalizerStr,
		},
	})

	_, err := utils.GetInformers().ClientSet.AppsV1().StatefulSets(lib.AviNS).Patch(akoStatefulset, types.MergePatchType, patchPayload)
	if err != nil {
		utils.AviLog.Warnf("Error in updating configmap: %v", err)
	}
}
