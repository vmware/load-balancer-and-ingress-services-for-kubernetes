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

// +kubebuilder:rbac:groups=addons.kubernetes.vmware.com,resources=addoninstalls,verbs=get;list;watch;create;update;patch;delete

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
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	VKSPublicNamespace  = "vmware-system-vks-public"
	AKOAddonName        = "ako"
	AKOAddonInstallName = "ako-global-installer"
)

var AddonInstallGVR = schema.GroupVersionResource{
	Group:    "addons.kubernetes.vmware.com",
	Version:  "v1alpha1",
	Resource: "addoninstalls",
}

// EnsureGlobalAddonInstall creates the global AddonInstall resource for VKS
func EnsureGlobalAddonInstall() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dynamicClient := lib.GetDynamicClientSet()
	if dynamicClient == nil {
		return fmt.Errorf("dynamic client is nil")
	}

	// Check if AddonInstall already exists
	existingAddon, err := dynamicClient.Resource(AddonInstallGVR).Namespace(VKSPublicNamespace).Get(
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

	_, err = dynamicClient.Resource(AddonInstallGVR).Namespace(VKSPublicNamespace).Create(
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
					"app.kubernetes.io/name":       "ako",
					"app.kubernetes.io/managed-by": "ako-infra",
				},
			},
			"spec": map[string]interface{}{
				"addonName":               AKOAddonName,
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
				"releases": map[string]interface{}{
					"selector": map[string]interface{}{
						"matchLabels": map[string]interface{}{
							"addon.kubernetes.vmware.com/addon-name": AKOAddonName,
						},
					},
					"resolutionRule": "PreferLatest",
				},
				"paused": false,
			},
		},
	}
}
