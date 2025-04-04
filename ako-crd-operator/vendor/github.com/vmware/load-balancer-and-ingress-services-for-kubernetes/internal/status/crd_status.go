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
	"fmt"
	"strings"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	akov1alpha2 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha2"
	akov1beta1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1beta1"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	backoff "github.com/cenkalti/backoff/v4"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// UpdateCRDStatusOptions CRD Status Update Options
type UpdateCRDStatusOptions struct {
	Status string
	Error  string
}

// UpdateHostRuleStatus HostRule status updates
func UpdateHostRuleStatus(key string, hr *akov1beta1.HostRule, updateStatus UpdateCRDStatusOptions, retryNum ...int) {
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 3 {
			utils.AviLog.Errorf("key: %s, msg: UpdateHostRuleStatus retried 3 times, aborting", key)
			return
		}
	}

	patchPayload, _ := json.Marshal(map[string]interface{}{
		"status": akov1beta1.HostRuleStatus(updateStatus),
	})

	hrFromK8sClient, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(hr.Namespace).Patch(context.TODO(), hr.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{}, "status")
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: there was an error in updating the hostrule status: %+v", key, err)
		updatedHr, err := lib.AKOControlConfig().CRDInformers().HostRuleInformer.Lister().HostRules(hr.Namespace).Get(hr.Name)
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: hostrule not found %v", key, err)
			if strings.Contains(err.Error(), utils.K8S_ETIMEDOUT) {
				UpdateHostRuleStatus(key, updatedHr, updateStatus, retry+1)
			}
			return
		}
		UpdateHostRuleStatus(key, updatedHr, updateStatus, retry+1)
	}
	// wait for crdinformer to be updated
	constantBackoff := backoff.WithMaxRetries(backoff.NewConstantBackOff(500*time.Millisecond), 5)

	operation := func() error {
		hrFromInformer, err := lib.AKOControlConfig().CRDInformers().HostRuleInformer.Lister().HostRules(hr.Namespace).Get(hr.Name)
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: Unable to get the hostrule %s/%s from informer", key, hr.Namespace, hr.Name)
			return fmt.Errorf("hostrule not found. err: [%s]", err.Error())
		}
		if hrFromInformer.Status.Status != hrFromK8sClient.Status.Status {
			return fmt.Errorf("hostrule not updated in cache")
		}
		return nil
	}

	err = backoff.Retry(operation, constantBackoff)
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: Hostrule %s/%s lister cache not updated with patch. error: [%s]", key, hr.Namespace, hr.Name, err.Error())
	}

	utils.AviLog.Infof("key: %s, msg: Successfully updated the hostrule %s/%s status %+v", key, hr.Namespace, hr.Name, utils.Stringify(updateStatus))
}

// HostRuleEventBroadcast is responsible from broadcasting HostRule specific events when the VS Cache is Added/Updated/Deleted.
func HostRuleEventBroadcast(vsName string, vsCacheMetadataOld, vsMetadataNew lib.CRDMetadata) {
	if vsCacheMetadataOld.Value != vsMetadataNew.Value {
		oldHRNamespaceName := strings.Split(vsCacheMetadataOld.Value, "/")
		newHRNamespaceName := strings.Split(vsMetadataNew.Value, "/")

		if len(oldHRNamespaceName) != 2 || len(newHRNamespaceName) != 2 {
			return
		}

		oldHostRule, _ := lib.AKOControlConfig().CRDInformers().HostRuleInformer.Lister().HostRules(oldHRNamespaceName[0]).Get(oldHRNamespaceName[1])
		newHostRule, _ := lib.AKOControlConfig().CRDInformers().HostRuleInformer.Lister().HostRules(newHRNamespaceName[0]).Get(newHRNamespaceName[1])
		if oldHostRule == nil || newHostRule == nil {
			return
		}

		lib.AKOControlConfig().EventRecorder().Eventf(oldHostRule, corev1.EventTypeNormal, lib.Attached, "Configuration removed from VirtualService %s", vsName)
		lib.AKOControlConfig().EventRecorder().Eventf(newHostRule, corev1.EventTypeNormal, lib.Attached, "Configuration applied to VirtualService %s", vsName)
	}

	hrNamespaceName := strings.Split(vsMetadataNew.Value, "/")
	if len(hrNamespaceName) != 2 {
		return
	}
	hostrule, _ := lib.AKOControlConfig().CRDInformers().HostRuleInformer.Lister().HostRules(hrNamespaceName[0]).Get(hrNamespaceName[1])
	if hostrule == nil {
		return
	}

	if (vsCacheMetadataOld.Status == lib.CRDInactive || vsCacheMetadataOld.Status == "") && vsMetadataNew.Status == lib.CRDActive {
		// CRD was added, INACTIVE -> ACTIVE transitions
		lib.AKOControlConfig().EventRecorder().Eventf(hostrule, corev1.EventTypeNormal, lib.Attached, "Configuration applied to VirtualService %s", vsName)
	} else if vsCacheMetadataOld.Status == lib.CRDActive && (vsMetadataNew.Status == "" || vsMetadataNew.Status == lib.CRDInactive) {
		// CRD was removed, ACTIVE -> INACTIVE transitions
		lib.AKOControlConfig().EventRecorder().Eventf(hostrule, corev1.EventTypeNormal, lib.Attached, "Configuration removed from VirtualService %s", vsName)
	}
}

// UpdateHTTPRuleStatus HttpRule status updates
func UpdateHTTPRuleStatus(key string, rr *akov1beta1.HTTPRule, updateStatus UpdateCRDStatusOptions, retryNum ...int) {
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 3 {
			utils.AviLog.Errorf("key: %s, msg: UpdateHTTPRuleStatus retried 3 times, aborting", key)
			return
		}
	}

	patchPayload, _ := json.Marshal(map[string]interface{}{
		"status": akov1beta1.HTTPRuleStatus(updateStatus),
	})

	_, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HTTPRules(rr.Namespace).Patch(context.TODO(), rr.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{}, "status")
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: %d there was an error in updating the httprule status: %+v", key, retry, err)
		updatedRr, err := lib.AKOControlConfig().CRDInformers().HTTPRuleInformer.Lister().HTTPRules(rr.Namespace).Get(rr.Name)
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
}

// HttpRuleEventBroadcast is responsible from broadcasting HttpRule specific events when the Pool Cache is Added/Updated/Deleted.
func HttpRuleEventBroadcast(poolName string, poolCacheMetadataOld, vsMetadataNew lib.CRDMetadata) {
	if poolCacheMetadataOld.Value != vsMetadataNew.Value {
		oldHRNamespaceName := strings.SplitN(poolCacheMetadataOld.Value, "/", 3)
		newHRNamespaceName := strings.SplitN(vsMetadataNew.Value, "/", 3)

		if len(oldHRNamespaceName) != 3 || len(newHRNamespaceName) != 3 {
			return
		}

		oldHttpRule, _ := lib.AKOControlConfig().CRDInformers().HTTPRuleInformer.Lister().HTTPRules(oldHRNamespaceName[0]).Get(oldHRNamespaceName[1])
		newHttpRule, _ := lib.AKOControlConfig().CRDInformers().HTTPRuleInformer.Lister().HTTPRules(newHRNamespaceName[0]).Get(newHRNamespaceName[1])
		if oldHttpRule == nil || newHttpRule == nil {
			return
		}

		lib.AKOControlConfig().EventRecorder().Eventf(oldHttpRule, corev1.EventTypeNormal, lib.Attached, "Configuration for target path %s removed from Pool %s", oldHRNamespaceName[2], poolName)
		lib.AKOControlConfig().EventRecorder().Eventf(newHttpRule, corev1.EventTypeNormal, lib.Attached, "Configuration for target path %s applied to Pool %s", newHRNamespaceName[2], poolName)
	}

	hrNamespaceName := strings.SplitN(vsMetadataNew.Value, "/", 3)
	if len(hrNamespaceName) != 3 {
		return
	}
	httprule, _ := lib.AKOControlConfig().CRDInformers().HTTPRuleInformer.Lister().HTTPRules(hrNamespaceName[0]).Get(hrNamespaceName[1])
	if httprule == nil {
		return
	}

	if (poolCacheMetadataOld.Status == lib.CRDInactive || poolCacheMetadataOld.Status == "") && vsMetadataNew.Status == lib.CRDActive {
		// CRD was added, INACTIVE -> ACTIVE transitions
		lib.AKOControlConfig().EventRecorder().Eventf(httprule, corev1.EventTypeNormal, lib.Attached, "Configuration for target path %s applied to Pool %s", hrNamespaceName[2], poolName)
	} else if poolCacheMetadataOld.Status == lib.CRDActive && (vsMetadataNew.Status == "" || vsMetadataNew.Status == lib.CRDInactive) {
		// CRD was removed, ACTIVE -> INACTIVE transitions
		lib.AKOControlConfig().EventRecorder().Eventf(httprule, corev1.EventTypeNormal, lib.Attached, "Configuration for target path %s removed from Pool %s", hrNamespaceName[2], poolName)
	}
}

// UpdateAviInfraSettingStatus AviInfraSetting status updates
func UpdateAviInfraSettingStatus(key string, infraSetting *akov1beta1.AviInfraSetting, updateStatus UpdateCRDStatusOptions, retryNum ...int) {

	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 3 {
			utils.AviLog.Errorf("key: %s, msg: UpdateAviInfraSettingStatus retried 3 times, aborting", key)
			return
		}
	}

	patchPayload, _ := json.Marshal(map[string]interface{}{
		"status": akov1beta1.AviInfraSettingStatus(updateStatus),
	})

	_, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Patch(context.TODO(), infraSetting.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{}, "status")
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: %d there was an error in updating the aviinfrasetting status: %+v", key, retry, err)
		updatedInfraSetting, err := lib.AKOControlConfig().CRDInformers().AviInfraSettingInformer.Lister().Get(infraSetting.Name)
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: aviinfrasetting not found %v", key, err)
			if strings.Contains(err.Error(), utils.K8S_ETIMEDOUT) {
				UpdateAviInfraSettingStatus(key, updatedInfraSetting, updateStatus, retry+1)
			}
			return
		}
		UpdateAviInfraSettingStatus(key, updatedInfraSetting, updateStatus, retry+1)
	}

	utils.AviLog.Infof("key: %s, msg: Successfully updated the aviinfrasetting %s status %+v", key, infraSetting.Name, utils.Stringify(updateStatus))
}

// UpdateL4RuleStatus updates the L4Rule status
func UpdateL4RuleStatus(key string, l4Rule *akov1alpha2.L4Rule, updateStatus UpdateCRDStatusOptions, retryNum ...int) {
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 3 {
			utils.AviLog.Errorf("key: %s, msg: UpdateL4RuleStatus retried 3 times, aborting", key)
			return
		}
	}

	patchPayload, _ := json.Marshal(map[string]interface{}{
		"status": akov1alpha2.L4RuleStatus(updateStatus),
	})

	_, err := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(l4Rule.Namespace).Patch(context.TODO(), l4Rule.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{}, "status")
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: %d there was an error in updating the L4Rule status: %+v", key, retry, err)
		updatedL4RuleObj, err := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(l4Rule.Namespace).Get(context.TODO(), l4Rule.Name, metav1.GetOptions{})
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: L4Rule not found %v", key, err)
			if strings.Contains(err.Error(), utils.K8S_ETIMEDOUT) {
				UpdateL4RuleStatus(key, updatedL4RuleObj, updateStatus, retry+1)
			}
			return
		}
		UpdateL4RuleStatus(key, updatedL4RuleObj, updateStatus, retry+1)
	}

	utils.AviLog.Infof("key: %s, msg: Successfully updated the L4Rule %s status %+v", key, l4Rule.Name, utils.Stringify(updateStatus))
}

// L4RuleEventBroadcast is responsible from broadcasting L4Rule specific events when the VS Cache is Added/Updated/Deleted.
func L4RuleEventBroadcast(vsName string, vsCacheMetadataOld, vsMetadataNew lib.CRDMetadata) {
	if vsCacheMetadataOld.Value != vsMetadataNew.Value {
		oldLRNamespaceName := strings.Split(vsCacheMetadataOld.Value, "/")
		newLRNamespaceName := strings.Split(vsMetadataNew.Value, "/")

		if len(oldLRNamespaceName) != 2 || len(newLRNamespaceName) != 2 {
			return
		}

		oldL4Rule, _ := lib.AKOControlConfig().CRDInformers().L4RuleInformer.Lister().L4Rules(oldLRNamespaceName[0]).Get(oldLRNamespaceName[1])
		newL4Rule, _ := lib.AKOControlConfig().CRDInformers().L4RuleInformer.Lister().L4Rules(newLRNamespaceName[0]).Get(newLRNamespaceName[1])
		if oldL4Rule == nil || newL4Rule == nil {
			return
		}

		lib.AKOControlConfig().EventRecorder().Eventf(oldL4Rule, corev1.EventTypeNormal, lib.Attached, "Configuration removed from VirtualService %s", vsName)
		lib.AKOControlConfig().EventRecorder().Eventf(newL4Rule, corev1.EventTypeNormal, lib.Attached, "Configuration applied to VirtualService %s", vsName)
	}

	lrNamespaceName := strings.Split(vsMetadataNew.Value, "/")
	if len(lrNamespaceName) != 2 {
		return
	}

	l4Rule, _ := lib.AKOControlConfig().CRDInformers().L4RuleInformer.Lister().L4Rules(lrNamespaceName[0]).Get(lrNamespaceName[1])
	if l4Rule == nil {
		return
	}

	if (vsCacheMetadataOld.Status == lib.CRDInactive || vsCacheMetadataOld.Status == "") && vsMetadataNew.Status == lib.CRDActive {
		// CRD was added, INACTIVE -> ACTIVE transitions
		lib.AKOControlConfig().EventRecorder().Eventf(l4Rule, corev1.EventTypeNormal, lib.Attached, "Configuration applied to VirtualService %s", vsName)
	} else if vsCacheMetadataOld.Status == lib.CRDActive && (vsMetadataNew.Status == "" || vsMetadataNew.Status == lib.CRDInactive) {
		// CRD was removed, ACTIVE -> INACTIVE transitions
		lib.AKOControlConfig().EventRecorder().Eventf(l4Rule, corev1.EventTypeNormal, lib.Attached, "Configuration removed from VirtualService %s", vsName)
	}
}

// UpdateSSORuleStatus SSORule status updates
func UpdateSSORuleStatus(key string, sr *akov1alpha2.SSORule, updateStatus UpdateCRDStatusOptions, retryNum ...int) {
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 3 {
			utils.AviLog.Errorf("key: %s, msg: UpdateSSORuleStatus retried 3 times, aborting", key)
			return
		}
	}

	patchPayload, _ := json.Marshal(map[string]interface{}{
		"status": akov1alpha2.SSORuleStatus(updateStatus),
	})

	_, err := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().SSORules(sr.Namespace).Patch(context.TODO(), sr.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{}, "status")
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: there was an error in updating the SSORule status: %+v", key, err)
		updatedSr, err := lib.AKOControlConfig().CRDInformers().SSORuleInformer.Lister().SSORules(sr.Namespace).Get(sr.Name)
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: SSORule not found %v", key, err)
			if strings.Contains(err.Error(), utils.K8S_ETIMEDOUT) {
				UpdateSSORuleStatus(key, updatedSr, updateStatus, retry+1)
			}
			return
		}
		UpdateSSORuleStatus(key, updatedSr, updateStatus, retry+1)
	}

	utils.AviLog.Infof("key: %s, msg: Successfully updated the SSORule %s/%s status %+v", key, sr.Namespace, sr.Name, utils.Stringify(updateStatus))
}

// SSORuleEventBroadcast is responsible for broadcasting SSORule specific events when the VS Cache is Added/Updated/Deleted.
func SSORuleEventBroadcast(vsName string, vsCacheMetadataOld, vsMetadataNew lib.CRDMetadata) {
	if vsCacheMetadataOld.Value != vsMetadataNew.Value {
		oldSRNamespaceName := strings.Split(vsCacheMetadataOld.Value, "/")
		newSRNamespaceName := strings.Split(vsMetadataNew.Value, "/")

		if len(oldSRNamespaceName) != 2 || len(newSRNamespaceName) != 2 {
			return
		}

		oldSSORule, _ := lib.AKOControlConfig().CRDInformers().SSORuleInformer.Lister().SSORules(oldSRNamespaceName[0]).Get(oldSRNamespaceName[1])
		newSSORule, _ := lib.AKOControlConfig().CRDInformers().SSORuleInformer.Lister().SSORules(newSRNamespaceName[0]).Get(newSRNamespaceName[1])
		if oldSSORule == nil || newSSORule == nil {
			return
		}

		lib.AKOControlConfig().EventRecorder().Eventf(oldSSORule, corev1.EventTypeNormal, lib.Attached, "Configuration removed from VirtualService %s", vsName)
		lib.AKOControlConfig().EventRecorder().Eventf(newSSORule, corev1.EventTypeNormal, lib.Attached, "Configuration applied to VirtualService %s", vsName)
	}

	srNamespaceName := strings.Split(vsMetadataNew.Value, "/")
	if len(srNamespaceName) != 2 {
		return
	}
	ssoRule, _ := lib.AKOControlConfig().CRDInformers().SSORuleInformer.Lister().SSORules(srNamespaceName[0]).Get(srNamespaceName[1])
	if ssoRule == nil {
		return
	}

	if (vsCacheMetadataOld.Status == lib.CRDInactive || vsCacheMetadataOld.Status == "") && vsMetadataNew.Status == lib.CRDActive {
		// CRD was added, INACTIVE -> ACTIVE transitions
		lib.AKOControlConfig().EventRecorder().Eventf(ssoRule, corev1.EventTypeNormal, lib.Attached, "Configuration applied to VirtualService %s", vsName)
	} else if vsCacheMetadataOld.Status == lib.CRDActive && (vsMetadataNew.Status == "" || vsMetadataNew.Status == lib.CRDInactive) {
		// CRD was removed, ACTIVE -> INACTIVE transitions
		lib.AKOControlConfig().EventRecorder().Eventf(ssoRule, corev1.EventTypeNormal, lib.Attached, "Configuration removed from VirtualService %s", vsName)
	}
}

// UpdateL7RuleStatus updates the L7Rule status
func UpdateL7RuleStatus(key string, l7Rule *akov1alpha2.L7Rule, updateStatus UpdateCRDStatusOptions, retryNum ...int) {
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 3 {
			utils.AviLog.Errorf("key: %s, msg: UpdateL7RuleStatus retried 3 times, aborting", key)
			return
		}
	}

	patchPayload, _ := json.Marshal(map[string]interface{}{
		"status": akov1alpha2.L7RuleStatus(updateStatus),
	})

	_, err := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L7Rules(l7Rule.Namespace).Patch(context.TODO(), l7Rule.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{}, "status")
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: %d there was an error in updating the L7Rule status: %+v", key, retry, err)
		updatedL7RuleObj, err := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L7Rules(l7Rule.Namespace).Get(context.TODO(), l7Rule.Name, metav1.GetOptions{})
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: L7Rule not found %v", key, err)
			if strings.Contains(err.Error(), utils.K8S_ETIMEDOUT) {
				UpdateL7RuleStatus(key, updatedL7RuleObj, updateStatus, retry+1)
			}
			return
		}
		UpdateL7RuleStatus(key, updatedL7RuleObj, updateStatus, retry+1)
	}

	utils.AviLog.Infof("key: %s, msg: Successfully updated the L7Rule %s status %+v", key, l7Rule.Name, utils.Stringify(updateStatus))
}

// L7RuleEventBroadcast is responsible from broadcasting L7Rule specific events when the VS Cache is Added/Updated/Deleted.
func L7RuleEventBroadcast(vsName string, vsCacheMetadataOld, vsMetadataNew lib.CRDMetadata) {
	if vsCacheMetadataOld.Value != vsMetadataNew.Value {
		oldLRNamespaceName := strings.Split(vsCacheMetadataOld.Value, "/")
		newLRNamespaceName := strings.Split(vsMetadataNew.Value, "/")

		if len(oldLRNamespaceName) != 2 || len(newLRNamespaceName) != 2 {
			return
		}

		oldL7Rule, _ := lib.AKOControlConfig().CRDInformers().L7RuleInformer.Lister().L7Rules(oldLRNamespaceName[0]).Get(oldLRNamespaceName[1])
		newL7Rule, _ := lib.AKOControlConfig().CRDInformers().L7RuleInformer.Lister().L7Rules(newLRNamespaceName[0]).Get(newLRNamespaceName[1])
		if oldL7Rule == nil || newL7Rule == nil {
			return
		}

		lib.AKOControlConfig().EventRecorder().Eventf(oldL7Rule, corev1.EventTypeNormal, lib.Attached, "Configuration removed from VirtualService %s", vsName)
		lib.AKOControlConfig().EventRecorder().Eventf(newL7Rule, corev1.EventTypeNormal, lib.Attached, "Configuration applied to VirtualService %s", vsName)
	}

	lrNamespaceName := strings.Split(vsMetadataNew.Value, "/")
	if len(lrNamespaceName) != 2 {
		return
	}

	l7Rule, _ := lib.AKOControlConfig().CRDInformers().L7RuleInformer.Lister().L7Rules(lrNamespaceName[0]).Get(lrNamespaceName[1])
	if l7Rule == nil {
		return
	}

	if (vsCacheMetadataOld.Status == lib.CRDInactive || vsCacheMetadataOld.Status == "") && vsMetadataNew.Status == lib.CRDActive {
		// CRD was added, INACTIVE -> ACTIVE transitions
		lib.AKOControlConfig().EventRecorder().Eventf(l7Rule, corev1.EventTypeNormal, lib.Attached, "Configuration applied to VirtualService %s", vsName)
	} else if vsCacheMetadataOld.Status == lib.CRDActive && (vsMetadataNew.Status == "" || vsMetadataNew.Status == lib.CRDInactive) {
		// CRD was removed, ACTIVE -> INACTIVE transitions
		lib.AKOControlConfig().EventRecorder().Eventf(l7Rule, corev1.EventTypeNormal, lib.Attached, "Configuration removed from VirtualService %s", vsName)
	}
}
