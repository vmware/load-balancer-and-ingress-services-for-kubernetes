/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:validation:Enum=INFO;DEBUG;WARN;ERROR
type LogLevelType string

// +kubebuilder:validation:Enum=NodePort;ClusterIP
type ServiceTypeStr string

// +kubebuilder:validation:Enum=LARGE;MEDIUM;SMALL
type VSSize string

type NamespaceSelector struct {
	LabelKey   string `json:"labelKey,omitempty"`
	LabelValue string `json:"labelValue,omitempty"`
}

// AKOSettings defines the settings requried for the AKO controller
type AKOSettings struct {
	// LogLevel defines the log level to be used by the AKO controller
	LogLevel LogLevelType `json:"logLevel,omitempty"`
	// FullSyncFrequency defines the interval at which full sync is carried out by the AKO controller
	FullSyncFrequency string `json:"fullSyncFrequency,omitempty"`
	// APIServerPort is the port at which the AKO API server runs
	APIServerPort int `json:"apiServerPort,omitempty"`
	// DeleteConfig is set if clean up is required by AKO
	DeleteConfig bool `json:"deleteConfig,omitempty"`
	// DisableStaticRouteSync is set if the static route sync is not required
	DisableStaticRouteSync bool `json:"disableStaticRouteSync,omitempty"`
	// ClusterName is used to identify a cluster
	ClusterName string `json:"clusterName,omitempty"`
	// CNIPlugin specifies the CNI to be used
	CNIPlugin string `json:"cniPlugin,omitempty"`
	// Namespace selector specifies the namespace labels from which AKO should sync from
	NSSelector NamespaceSelector `json:"namespaceSelector,omitempty"`
}

type NodeNetwork struct {
	NetworkName string   `json:"networkName,omitempty"`
	Cidrs       []string `json:"cidrs,omiempty"`
}

// NetworkSettings defines the network details required for the AKO controller
type NetworkSettings struct {
	// NodeNetworkList is the list of networks and their cidrs used in pool placement network for vcenter
	// cloud. This is not required for either of these cases:
	// 1. nodeport is enabled
	// 2. static routes are disabled
	// 3. non vcenter clouds
	NodeNetworkList []NodeNetwork `json:"nodeNetworkList,omitempty"`
	// SubnetIP is the Network IP for the subnet to be used
	SubnetIP string `json:"subnetIP,omitemmpty"`
	// SubnetPrefix is the netmask for the subnet
	SubnetPrefix string `json:"subnetPrefix,omitempty"`
	// NetworkName is the name of the network as specified in Avi
	NetworkName string `json:"networkName,omitempty"`
	// EnableRHI is a cluster wide setting for BGP peering
	EnableRHI bool `json:"enableRHI,omitempty"`
}

// L7Settings defines the L7 configuration for the AKO controller
type L7Settings struct {
	// DefaultIngController specifies whether AKO controller is the default ingress controller
	DefaultIngController bool `json:"defaultIngController,omitempty"`
	// ShardingScheme specifies how the ingress objects are sharded
	ShardingScheme string `json:"shardingScheme,omitempty"`
	// ServiceType defines the service type: ClusterIP or NodePort
	ServiceType ServiceTypeStr `json:"serviceType,omitempty"`
	// ShardVSSize specifies the number of shard VSs to be created
	ShardVSSize VSSize `json:"shardVSSize,omitempty"`
	// PassthroughShardSize specifies the number of shard VSs to be created for passthrough routes
	PassthroughShardSize VSSize `json:"passthroughShardSize,omitempty"`
	// SyncNamespace takes in a namespace from which AKO will sync the objects
	SyncNamespace string `json:"syncNamespace,omitempty"`
}

// L4Settings defines the L4 configuration for the AKO controller
type L4Settings struct {
	// AdvancedL4 specifies whether the AKO controller should listen for the Gateway objects
	AdvancedL4 bool `json:"advancedL4,omitempty"`
	// DefaultDomain is the default domain
	DefaultDomain string `json:"defaultDomain,omitempty"`
}

// ControllerSettings defines the Avi Controller parameters
type ControllerSettings struct {
	// ServiceEngineGroupName is the name of the Serviceengine group in Avi
	ServiceEngineGroupName string `json:"serviceEngineGroupName,omitempty"`
	// ControllerVersion is the Avi controller version
	ControllerVersion string `json:"controllerVersion,omitempty"`
	// CloudName is the name of the cloud to be used in Avi
	CloudName string `json:"cloudName,omitempty"`
	// ControllerIP is the IP address of the Avi Controller
	ControllerIP string `json:"controllerIP,omitempty"`
	// TenantsPerCluster if set to true, AKO will map each k8s cluster uniquely to a tenant
	// in Avi
	TenantsPerCluster bool `json:"tenantsPerCluster,omitempty"`
	// TenantName is the name of the tenant where all AKO objects will be created in Avi.
	TenantName string `json:"tenantName,omitempty"`
}

// NodePortSelector defines the node port settings, to be used only if the serviceTYpe is selected
// NodePort
type NodePortSelector struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

// ResourceLimits defines the limits on cpu and memory for the AKO controller
type ResourceLimits struct {
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

// ResourceRequests defines the requests for cpu and memory by the AKO controller
type ResourceRequests struct {
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

// Resources defines the limits and requests for cpu and memory to be used by the AKO controller
type Resources struct {
	Limits   ResourceLimits   `json:"limits,omitempty"`
	Requests ResourceRequests `json:"requests,omitempty"`
}

type Rbac struct {
	PSPEnable bool `json:"pspEnable,omitempty"`
}

// AKOConfigSpec defines the desired state of AKOConfig
type AKOConfigSpec struct {
	// ImageRepository is where the AKO controller resides.
	ImageRepository string `json:"imageRepository,omitempty"`
	// ImagePullPolicy defines when the AKO controller image gets pulled.
	ImagePullPolicy       string `json:"imagePullPolicy,omitempty"`
	AKOSettings           `json:"akoSettings,omitempty"`
	NetworkSettings       `json:"networkSettings,omitempty"`
	L7Settings            `json:"l7Settings,omitempty"`
	L4Settings            `json:"l4Settings,omitempty"`
	ControllerSettings    `json:"controllerSettings,omitempty"`
	NodePortSelector      `json:"nodePortSelector,omitempty"`
	Resources             `json:"resources,omitempty"`
	Rbac                  `json:"rbac,omitempty"`
	PersistentVolumeClaim string `json:"pvc,omitempty"`
	MountPath             string `json:"mountPath,omitempty"`
	LogFile               string `json:"logFile,omitempty"`
}

// AKOConfigStatus defines the observed state of AKOConfig
type AKOConfigStatus struct {
	State string `json:"state,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// AKOConfig is the Schema for the akoconfigs API
type AKOConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AKOConfigSpec   `json:"spec,omitempty"`
	Status AKOConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AKOConfigList contains a list of AKOConfig
type AKOConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AKOConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AKOConfig{}, &AKOConfigList{})
}
