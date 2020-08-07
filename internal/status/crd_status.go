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
	"ako/internal/lib"
	"errors"

	akov1alpha1 "ako/internal/apis/ako/v1alpha1"

	"ako/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// UpdateCRDStatusOptions CRD Status Update Options
type UpdateCRDStatusOptions struct {
	Status string
	Error  string
}

// UpdateHostRuleStatus HostRule status updates
func UpdateHostRuleStatus(hr *akov1alpha1.HostRule, updateStatus UpdateCRDStatusOptions, retryNum ...int) error {
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 2 {
			return errors.New("msg: UpdateHostRuleStatus retried 3 times, aborting")
		}
	}

	hr.Status.Status = updateStatus.Status
	hr.Status.Error = updateStatus.Error

	_, err := lib.GetCRDClientset().AkoV1alpha1().HostRules(hr.Namespace).UpdateStatus(hr)
	if err != nil {
		utils.AviLog.Errorf("msg: there was an error in updating the hostrule status: %+v", err)
		updatedHr, err := lib.GetCRDClientset().AkoV1alpha1().HostRules(hr.Namespace).Get(hr.Name, metav1.GetOptions{})
		if err != nil {
			utils.AviLog.Warnf("hostrule not found %v", err)
			return err
		}
		return UpdateHostRuleStatus(updatedHr, updateStatus, retry+1)
	}

	utils.AviLog.Infof("msg: Successfully updated the hostrule %s/%s status %+v", hr.Namespace, hr.Name, utils.Stringify(updateStatus))
	return nil
}

// UpdateHTTPRuleStatus HostRule status updates
func UpdateHTTPRuleStatus(rr *akov1alpha1.HTTPRule, updateStatus UpdateCRDStatusOptions, retryNum ...int) error {
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 2 {
			return errors.New("msg: UpdateHTTPRuleStatus retried 3 times, aborting")
		}
	}

	rr.Status.Status = updateStatus.Status
	rr.Status.Error = updateStatus.Error

	_, err := lib.GetCRDClientset().AkoV1alpha1().HTTPRules(rr.Namespace).UpdateStatus(rr)
	if err != nil {
		utils.AviLog.Errorf("msg: %d there was an error in updating the httprule status: %+v", retry, err)
		updatedRr, err := lib.GetCRDClientset().AkoV1alpha1().HTTPRules(rr.Namespace).Get(rr.Name, metav1.GetOptions{})
		if err != nil {
			utils.AviLog.Warnf("httprule not found %v", err)
			return err
		}
		return UpdateHTTPRuleStatus(updatedRr, updateStatus, retry+1)
	}

	utils.AviLog.Infof("msg: Successfully updated the httprule %s/%s status %+v", rr.Namespace, rr.Name, utils.Stringify(updateStatus))
	return nil
}
