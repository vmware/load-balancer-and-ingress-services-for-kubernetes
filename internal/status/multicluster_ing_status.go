/*
 * Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
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
	"fmt"
	"strings"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// DeleteMultiClusterIngressStatusAndAnnotation is a wrapper function which gets the Multi-cluster ingress object and deletes the
// status/annotations.
func (l *leader) DeleteMultiClusterIngressStatusAndAnnotation(key string, option *UpdateOptions) {
	ns := option.ServiceMetadata.Namespace
	ingName := option.ServiceMetadata.IngressName
	mciObj, err := utils.GetInformers().MultiClusterIngressInformer.Lister().MultiClusterIngresses(ns).Get(ingName)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Could not get the Multi-cluster object for delete status: %s", key, err)
		return
	}
	mciObj = mciObj.DeepCopy()
	mciObj.Status.LoadBalancer.Ingress = nil
	UpdateMultiClusterIngressStatus(key, mciObj, &mciObj.Status)

	vsAnnotations := make(map[string]string)
	if err := UpdateMultiClusterIngressAnnotations(mciObj, vsAnnotations, key, option.Tenant); err != nil {
		utils.AviLog.Warnf("key: %s, msg: could not delete the Multi-cluster ingress object's annotation: err %s, %s/%s", key, err, mciObj.Namespace, mciObj.Name)
	}
}

// UpdateMultiClusterIngressStatusAndAnnotation is a wrapper function which gets the Multi-cluster ingress object and updates the
// status/annotations.
func (l *leader) UpdateMultiClusterIngressStatusAndAnnotation(key string, option *UpdateOptions) {
	ns := option.ServiceMetadata.Namespace
	ingName := option.ServiceMetadata.IngressName
	mciObj, err := utils.GetInformers().MultiClusterIngressInformer.Lister().MultiClusterIngresses(ns).Get(ingName)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Could not get the Multi-cluster ingress object for update status: %s", key, err)
		return
	}
	mciObj = mciObj.DeepCopy()
	mciObj.Status.LoadBalancer.Ingress = make([]akov1alpha1.IngressStatus, 1)
	mciObj.Status.LoadBalancer.Ingress[0].Hostname = mciObj.Spec.Hostname
	mciObj.Status.LoadBalancer.Ingress[0].IP = option.Vip[0]

	utils.AviLog.Debugf("key: %s, msg: Updating multicluster ingress object status with %+v:", key, utils.Stringify(&mciObj.Status))
	UpdateMultiClusterIngressStatus(key, mciObj, &mciObj.Status)

	vsAnnotations := make(map[string]string)
	vsAnnotations[mciObj.Spec.Hostname] = option.VirtualServiceUUID

	if err := UpdateMultiClusterIngressAnnotations(mciObj, vsAnnotations, key, option.Tenant); err != nil {
		utils.AviLog.Warnf("key: %s, msg: Could not update the Multi-cluster ingress object's annotation: err %s, %s/%s", key, err, mciObj.Namespace, mciObj.Name)
	}
}

// UpdateMultiClusterIngressStatus updates MultiClusterIngress' status
func UpdateMultiClusterIngressStatus(key string, mci *akov1alpha1.MultiClusterIngress, status *akov1alpha1.MultiClusterIngressStatus, retryNum ...int) {
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 3 {
			utils.AviLog.Errorf("key: %s, msg: UpdateMultiClusterIngressStatus retried 3 times, aborting", key)
			return
		}
	}

	patchPayload, _ := json.Marshal(map[string]interface{}{
		"status": status,
	})

	_, err := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().MultiClusterIngresses(mci.Namespace).Patch(context.TODO(), mci.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{}, "status")
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: there was an error in updating the multicluster ingress status: %+v", key, err)
		updatedMCI, err := utils.GetInformers().MultiClusterIngressInformer.Lister().MultiClusterIngresses(mci.Namespace).Get(mci.Name)
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: multicluster ingress not found %v", key, err)
			if strings.Contains(err.Error(), utils.K8S_ETIMEDOUT) {
				UpdateMultiClusterIngressStatus(key, updatedMCI, status, retry+1)
			}
			return
		}
		UpdateMultiClusterIngressStatus(key, updatedMCI, status, retry+1)
		return
	}

	utils.AviLog.Infof("key: %s, msg: Successfully updated the multicluster ingress %s/%s status %+v", key, mci.Namespace, mci.Name, utils.Stringify(status))
}

func UpdateMultiClusterIngressAnnotations(mci *akov1alpha1.MultiClusterIngress, vsAnnotations map[string]string, key, tenant string) error {

	// compare the vs annotations for this object
	required := isAnnotationsUpdateRequired(mci.GetAnnotations(), vsAnnotations, tenant, false)
	if !required {
		utils.AviLog.Debugf("annotations update not required for this multi-cluster ingress: %s/%s", mci.Namespace, mci.Name)
		return nil
	}

	patchPayloadBytes, err := getAnnotationsPayload(vsAnnotations, tenant)
	if err != nil {
		return fmt.Errorf("error in generating payload for vs annotations %v: %v", vsAnnotations, err)
	}
	_, err = lib.AKOControlConfig().CRDClientset().AkoV1alpha1().MultiClusterIngresses(mci.Namespace).Patch(context.TODO(), mci.Name, types.MergePatchType, patchPayloadBytes, metav1.PatchOptions{})
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: there was an error in updating the multicluster ingress annotation: %+v %s/%s", key, err)
		return err
	}
	return nil
}

func (f *follower) UpdateMultiClusterIngressStatusAndAnnotation(key string, option *UpdateOptions) {
	utils.AviLog.Debugf("key: %s, AKO is not a leader, not updating the Multi-Cluster Ingress status", option.Key)
}

func (f *follower) DeleteMultiClusterIngressStatusAndAnnotation(key string, option *UpdateOptions) {
	utils.AviLog.Debugf("key: %s, AKO is not a leader, not deleting the Multi-Cluster Ingress status", key)
}
