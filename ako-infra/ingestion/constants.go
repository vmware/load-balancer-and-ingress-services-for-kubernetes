/*
 * Copyright 2024 VMware, Inc.
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

package ingestion

import (
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	AVI_ENTERPRISE                 = "ENTERPRISE"
	VSphereClusterIDLabelKey       = "vSphereClusterID"
	AviEnterpriseWithCloudServices = "ENTERPRISE_WITH_CLOUD_SERVICES"
)

// VKS cluster monitoring constants
const (
	ClusterPhaseProvisioning = "Provisioning"
	ClusterPhaseProvisioned  = "Provisioned"
	ClusterPhaseDeleting     = "Deleting"
	ClusterPhaseFailed       = "Failed"

	// Service Engine Group annotation for vSphere namespaces
	ServiceEngineGroupAnnotation = "vmware-system-csi/serviceenginegroup"

	// VKS cluster watcher configuration
	VKSClusterWorkQueue    = "vks-cluster-watcher"
	VKSClusterResyncPeriod = 30 * time.Second

	// VKS managed label for AKO deployment control
	VKSManagedLabel           = "ako.kubernetes.vmware.com/install"
	VKSManagedLabelValueTrue  = "true"
	VKSManagedLabelValueFalse = "false"

	// VKS processing annotation for cluster state tracking
	VKSProcessedAnnotation = "ako.kubernetes.vmware.com/vks-processed"
)

// VKS addon framework constants
const (
	VKSPublicNamespace  = "vmware-system-vks-public"
	AKOAddonName        = "ako"
	AKOAddonInstallName = "ako-global-installer"
)

// ClusterGVR defines the cluster.x-k8s.io/v1beta1 Cluster resource
var ClusterGVR = schema.GroupVersionResource{
	Group:    "cluster.x-k8s.io",
	Version:  "v1beta1",
	Resource: "clusters",
}

// LabelingOperation represents cluster labeling operation types
type LabelingOperation string

const (
	LabelingOperationAdd      LabelingOperation = "ADD"
	LabelingOperationRemove   LabelingOperation = "REMOVE"
	LabelingOperationUpdate   LabelingOperation = "UPDATE"
	LabelingOperationOptIn    LabelingOperation = "OPT_IN"
	LabelingOperationOptOut   LabelingOperation = "OPT_OUT"
	LabelingOperationValidate LabelingOperation = "VALIDATE"
)

// LabelingResult contains the result of a cluster labeling operation
type LabelingResult struct {
	ClusterName      string
	ClusterNamespace string
	Operation        LabelingOperation
	Success          bool
	Error            error
	PreviousValue    string
	NewValue         string
	Skipped          bool
	SkipReason       string
}
