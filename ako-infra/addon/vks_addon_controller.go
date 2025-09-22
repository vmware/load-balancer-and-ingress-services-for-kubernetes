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

package addon

import (
	"context"
	"fmt"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-infra/webhook"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	VKSPublicNamespace  = "vmware-system-vks-public"
	AKOAddonName        = "ako"
	AKOAddonInstallName = "ako-global-installer"
)

// EnsureGlobalAddonInstall creates the global AddonInstall resource for VKS
func EnsureGlobalAddonInstall() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dynamicClient := lib.GetDynamicClientSet()
	if dynamicClient == nil {
		return fmt.Errorf("dynamic client is nil")
	}

	// Check if AddonInstall already exists
	existingAddon, err := dynamicClient.Resource(lib.AddonInstallGVR).Namespace(VKSPublicNamespace).Get(
		ctx, AKOAddonInstallName, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to check existing AddonInstall %s/%s: %w",
			VKSPublicNamespace, AKOAddonInstallName, err)
	}

	if existingAddon != nil {
		utils.AviLog.Infof("VKS global AddonInstall %s/%s already exists", VKSPublicNamespace, AKOAddonInstallName)
		return nil
	}

	addonInstall := createAddonInstallSpec()

	_, err = dynamicClient.Resource(lib.AddonInstallGVR).Namespace(VKSPublicNamespace).Create(
		ctx, addonInstall, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create AddonInstall %s/%s: %w",
			VKSPublicNamespace, AKOAddonInstallName, err)
	}

	utils.AviLog.Infof("VKS global AddonInstall %s/%s created successfully", VKSPublicNamespace, AKOAddonInstallName)
	return nil
}

// createAddonInstallSpec creates the AddonInstall resource specification
func createAddonInstallSpec() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "addons.kubernetes.vmware.com/v1alpha1",
			"kind":       "AddonInstall",
			"metadata": map[string]interface{}{
				"name":      AKOAddonInstallName,
				"namespace": VKSPublicNamespace,
				"labels": map[string]interface{}{
					"app.kubernetes.io/name":       AKOAddonName,
					"app.kubernetes.io/managed-by": "ako-infra",
				},
			},
			"spec": map[string]interface{}{
				"addonRef": map[string]interface{}{
					"name": AKOAddonName,
				},
				"crossNamespaceSelection": "Allowed",
				"clusters": []interface{}{
					map[string]interface{}{
						"selector": map[string]interface{}{
							"matchLabels": map[string]interface{}{
								webhook.VKSManagedLabel: webhook.VKSManagedLabelValueTrue,
							},
						},
					},
				},
				"paused": false,
			},
		},
	}
}

func CleanupGlobalAddonInstall() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dynamicClient := lib.GetDynamicClientSet()
	if dynamicClient == nil {
		return fmt.Errorf("dynamic client not available")
	}

	utils.AviLog.Infof("VKS addon: cleaning up global AddonInstall %s/%s", VKSPublicNamespace, AKOAddonInstallName)

	err := dynamicClient.Resource(lib.AddonInstallGVR).Namespace(VKSPublicNamespace).Delete(
		ctx, AKOAddonInstallName, metav1.DeleteOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			utils.AviLog.Infof("VKS addon: global AddonInstall %s/%s already deleted", VKSPublicNamespace, AKOAddonInstallName)
			return nil
		}
		return fmt.Errorf("failed to delete global AddonInstall %s/%s: %v", VKSPublicNamespace, AKOAddonInstallName, err)
	}

	utils.AviLog.Infof("VKS addon: successfully deleted global AddonInstall %s/%s", VKSPublicNamespace, AKOAddonInstallName)
	return nil
}

// EnsureGlobalAddonInstallWithRetry ensures global addon install with infinite retry
func EnsureGlobalAddonInstallWithRetry(stopCh <-chan struct{}) {
	utils.AviLog.Infof("VKS addon: starting global addon install with infinite retry")

	retryInterval := 10 * time.Second

	for {
		if err := EnsureGlobalAddonInstall(); err != nil {
			utils.AviLog.Warnf("VKS addon: failed to ensure global addon install, will retry in %v: %v", retryInterval, err)

			// Wait before retry, but also check for shutdown
			select {
			case <-stopCh:
				utils.AviLog.Infof("VKS addon: shutdown signal received during retry wait")
				return
			case <-time.After(retryInterval):
				// Continue to next retry
				continue
			}
		} else {
			utils.AviLog.Infof("VKS addon: global addon install ensured successfully")
			return
		}
	}
}
