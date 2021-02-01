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
	"strings"

	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/apis/ako/v1alpha1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// UpdateCRDStatusOptions CRD Status Update Options
type UpdateCRDStatusOptions struct {
	Status string
	Error  string
}

// UpdateHostRuleStatus HostRule status updates
func UpdateHostRuleStatus(key string, hr *akov1alpha1.HostRule, updateStatus UpdateCRDStatusOptions, retryNum ...int) {
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 3 {
			utils.AviLog.Errorf("key: %s, msg: UpdateHostRuleStatus retried 3 times, aborting", key)
			return
		}
	}

	hr.Status.Status = updateStatus.Status
	hr.Status.Error = updateStatus.Error

	_, err := lib.GetCRDClientset().AkoV1alpha1().HostRules(hr.Namespace).UpdateStatus(context.TODO(), hr, metav1.UpdateOptions{})
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: there was an error in updating the hostrule status: %+v", key, err)
		updatedHr, err := lib.GetCRDClientset().AkoV1alpha1().HostRules(hr.Namespace).Get(context.TODO(), hr.Name, metav1.GetOptions{})
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: hostrule not found %v", key, err)
			if strings.Contains(err.Error(), utils.K8S_ETIMEDOUT) {
				UpdateHostRuleStatus(key, updatedHr, updateStatus, retry+1)
			}
			return
		}
		UpdateHostRuleStatus(key, updatedHr, updateStatus, retry+1)
	}

	utils.AviLog.Infof("key: %s, msg: Successfully updated the hostrule %s/%s status %+v", key, hr.Namespace, hr.Name, utils.Stringify(updateStatus))
	return
}

// UpdateHTTPRuleStatus HostRule status updates
func UpdateHTTPRuleStatus(key string, rr *akov1alpha1.HTTPRule, updateStatus UpdateCRDStatusOptions, retryNum ...int) {
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 3 {
			utils.AviLog.Errorf("key: %s, msg: UpdateHTTPRuleStatus retried 3 times, aborting", key)
			return
		}
	}

	rr.Status.Status = updateStatus.Status
	rr.Status.Error = updateStatus.Error

	_, err := lib.GetCRDClientset().AkoV1alpha1().HTTPRules(rr.Namespace).UpdateStatus(context.TODO(), rr, metav1.UpdateOptions{})
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: %d there was an error in updating the httprule status: %+v", key, retry, err)
		updatedRr, err := lib.GetCRDClientset().AkoV1alpha1().HTTPRules(rr.Namespace).Get(context.TODO(), rr.Name, metav1.GetOptions{})
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: httprule not found %v", key, err)
			if strings.Contains(err.Error(), utils.K8S_ETIMEDOUT) {
				UpdateHTTPRuleStatus(key, updatedRr, updateStatus, retry+1)
			}
			return
		}
		UpdateHTTPRuleStatus(key, updatedRr, updateStatus, retry+1)
	}

	utils.AviLog.Infof("key: %s, msg: Successfully updated the httprule %s/%s status %+v", key, rr.Namespace, rr.Name, utils.Stringify(updateStatus))
	return
}
