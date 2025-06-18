/*
Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.

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

// +kubebuilder:validation:Enum=NodePort;ClusterIP;NodePortLocal
type ServiceTypeStr string

// +kubebuilder:validation:Enum=LARGE;MEDIUM;SMALL;DEDICATED
type VSSize string

// +kubebuilder:validation:Enum=LARGE;MEDIUM;SMALL
type PassthroughVSSize string

type NamespaceSelector struct {
	LabelKey   string `json:"labelKey,omitempty"`
	LabelValue string `json:"labelValue,omitempty"`
}

// AKOSettings defines the settings required for the AKO controller
type AKOSettings struct {
	// LogLevel defines the log level to be used by the AKO controller
	LogLevel LogLevelType `json:"logLevel,omitempty"`
	// EnableEvents controls whether AKO broadcasts Events in the cluster or not
	EnableEvents bool `json:"enableEvents,omitempty"`
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
	// EnableEVH enables the Enhanced Virtual Hosting Model in Avi Controller for the Virtual Services
	EnableEVH bool `json:"enableEVH,omitempty"`
	// Layer7Only enables AKO to do Layer 7 loadbalancing only
	Layer7Only bool `json:"layer7Only,omitempty"`
	// ServicesAPI enables AKO to do Layer 4 loadbalancing using Services API
	ServicesAPI bool `json:"servicesAPI,omitempty"`
	// VipPerNamespace enables AKO to create Parent VS per Namespace in EVH mode
	VipPerNamespace bool `json:"vipPerNamespace,omitempty"`
	// IstioEnabled flag needs to be enabled when AKO is be to brought up in an Istio environment
	IstioEnabled bool `json:"istioEnabled,omitempty"`
	// BlockedNamespaceList is the list of system namespaces from which AKO will not listen any Kubernetes or Openshift object event.
	BlockedNamespaceList []string `json:"blockedNamespaceList,omitempty"`
	// IPFamily specifies IP family to be used. This flag can take values V4 or V6 (default V4). This is for the backend pools to use ipv6 or ipv4. For frontside VS, use v6cidr.
	IPFamily string `json:"ipFamily,omitempty"`
	// UseDefaultSecretsOnly flag if set to true, AKO will only handle default secrets from the namespace where AKO is installed. This flag is applicable only to Openshift clusters.
	UseDefaultSecretsOnly bool `json:"useDefaultSecretsOnly,omitempty"`
}

type NodeNetwork struct {
	NetworkName string   `json:"networkName,omitempty"`
	Cidrs       []string `json:"cidrs,omitempty"`
	NetworkUUID string   `json:"networkUUID,omitempty"`
}

type VipNetwork struct {
	NetworkName string `json:"networkName,omitempty"`
	Cidr        string `json:"cidr,omitempty"`
	//V6Cidr will enable the VS networks to use ipv6
	V6Cidr      string `json:"v6cidr,omitempty"`
	NetworkUUID string `json:"networkUUID,omitempty"`
}

// NetworkSettings defines the network details required for the AKO controller
type NetworkSettings struct {
	// NodeNetworkList is the list of networks and their cidrs used in pool placement network for vcenter
	// cloud. Either networkName or networkUUID should be specified. If duplicate networks are present for
	// the network name, networkUUID should be used for appropriate network. This is not required for either of these cases:
	// 1. nodeport is enabled
	// 2. static routes are disabled
	// 3. non vcenter clouds
	NodeNetworkList []NodeNetwork `json:"nodeNetworkList,omitempty"`
	// EnableRHI is a cluster wide setting for BGP peering
	EnableRHI bool `json:"enableRHI,omitempty"`
	// NsxtT1LR is the unique ID (not display name) of the T1 Logical Router for Service Engine connectivity. Only applies to NSX-T cloud.
	// For eg : nsxtT1LR: "/infra/tier-1s/avi-t1".
	NsxtT1LR string `json:"nsxtT1LR,omitempty"`
	// BGPPeerLabels enable selection of BGP peers, for selective VsVip advertisement.
	BGPPeerLabels []string `json:"bgpPeerLabels,omitempty"`
	// VipNetworkList holds the names and subnet information of networks as specified in Avi.
	// Either networkName or networkUUID should be specified. If duplicate networks are present
	// for the network name, networkUUID should be used for appropriate network.
	VipNetworkList []VipNetwork `json:"vipNetworkList,omitempty"`
}

// L7Settings defines the L7 configuration for the AKO controller
type L7Settings struct {
	// DefaultIngController specifies whether AKO controller is the default ingress controller
	DefaultIngController bool `json:"defaultIngController,omitempty"`
	// ServiceType defines the service type: ClusterIP, NodePort or NodePortLocal
	ServiceType ServiceTypeStr `json:"serviceType,omitempty"`
	// ShardVSSize specifies the number of shard VSs to be created
	ShardVSSize VSSize `json:"shardVSSize,omitempty"`
	// PassthroughShardSize specifies the number of shard VSs to be created for passthrough routes
	PassthroughShardSize PassthroughVSSize `json:"passthroughShardSize,omitempty"`
	// SyncNamespace takes in a namespace from which AKO will sync the objects
	SyncNamespace string `json:"syncNamespace,omitempty"`
	// NoPGForSNI removes Avi PoolGroups from SNI VSes
	NoPGForSNI bool `json:"noPGForSNI,omitempty"`
}

// L4Settings defines the L4 configuration for the AKO controller
type L4Settings struct {
	// AdvancedL4 specifies whether the AKO controller should listen for the Gateway objects
	AdvancedL4 bool `json:"advancedL4,omitempty"`
	// DefaultDomain is the default domain
	DefaultDomain string `json:"defaultDomain,omitempty"`
	//Specifies the FQDN pattern - default, flat or disabled
	AutoFQDN string `json:"autoFQDN,omitempty"`
	// DefaultLBController enables ako to check if it is the default LoadBalancer controller.
	DefaultLBController bool `json:"defaultLBController,omitempty"`
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
	// VRFName is the name of the VRFContext. All Avi objects will be under this VRF. Applicable only in Vcenter Cloud.
	VRFName string `json:"vrfName,omitempty"`
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

type ImagePullSecret struct {
	Name string `json:"name,omitempty"`
}

// FeatureGates is to enable or disable experimental features
type FeatureGates struct {
	// GatewayAPI enables/disables processing of Kubernetes Gateway API CRDs
	GatewayAPI bool `json:"gatewayAPI,omitempty"`
	// EnablePrometheus enables/disables prometheus scraping for AKO container
	EnablePrometheus bool `json:"enablePrometheus,omitempty"`
}

// GatewayAPI defines settings for AKO Gateway API container
type GatewayAPI struct {
	// Image defines image related settings for AKO Gateway API container
	Image Image `json:"image,omitempty"`
}

type Image struct {
	Repository string `json:"repository,omitempty"`
	PullPolicy string `json:"pullPolicy,omitempty"`
}

// AKOConfigSpec defines the desired state of AKOConfig
type AKOConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//+operator-sdk:csv:customresourcedefinitions:type=spec
	// ImageRepository is where the AKO controller resides
	ImageRepository string `json:"imageRepository,omitempty"`
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	// ImagePullPolicy defines when the AKO controller image gets pulled
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	// ImagePullSecrets will add pull secrets to the statefulset for AKO. Required if using secure private container image registry for AKO image
	ImagePullSecrets []ImagePullSecret `json:"imagePullSecrets,omitempty"`
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	// ReplicaCount defines the number of replicas for AKO Statefulset
	ReplicaCount int `json:"replicaCount,omitempty"`
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="AKO Settings"
	// AKOSettings defines the settings required for the AKO controller
	AKOSettings AKOSettings `json:"akoSettings,omitempty"`
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	// NetworkSettings defines the network details required for the AKO controller
	NetworkSettings NetworkSettings `json:"networkSettings,omitempty"`
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Layer 7 Settings"
	// L7Settings defines the L7 configuration for the AKO controller
	L7Settings L7Settings `json:"l7Settings,omitempty"`
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Layer 4 Settings"
	// L4Settings defines the L4 configuration for the AKO controller
	L4Settings L4Settings `json:"l4Settings,omitempty"`
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	// ControllerSettings defines the Avi Controller parameters
	ControllerSettings ControllerSettings `json:"controllerSettings,omitempty"`
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="NodePort Selector"
	// NodePortSelector defines the node port settings, to be used only if the serviceTYpe is selected NodePort
	NodePortSelector NodePortSelector `json:"nodePortSelector,omitempty"`
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	// Resources defines the limits and requests for cpu and memory to be used by the AKO controller
	Resources Resources `json:"resources,omitempty"`
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	// Rbac enables the pod security policy for AKO
	Rbac Rbac `json:"rbac,omitempty"`
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="PVC"
	// PersistentVolumeClaim where the logs need to be stored
	PersistentVolumeClaim string `json:"pvc,omitempty"`
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	// MountPath is where the logFile will be mounted on
	MountPath string `json:"mountPath,omitempty"`
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	// LogFile is the name of the file where AKO will dump its logs
	LogFile string `json:"logFile,omitempty"`
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	// AKOGatewayLogFile is the name of the file where ako-gateway-api container will dump its logs
	AKOGatewayLogFile string `json:"akoGatewayLogFile,omitempty"`
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	// FeatureGates is to enable or disable experimental features
	FeatureGates FeatureGates `json:"featureGates,omitempty"`
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	// GatewayAPI defines settings for AKO Gateway API container
	GatewayAPI GatewayAPI `json:"gatewayAPI,omitempty"`
}

// AKOConfigStatus defines the observed state of AKOConfig
type AKOConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="State",xDescriptors="urn:alm:descriptor:urn:alm:descriptor:io.kubernetes.phase"
	// State defines the current Kubernetes phase of AKOConfig object
	State string `json:"state,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:deprecatedversion:warning="The v1alpha1 version is deprecated for AKOConfig CRD, please use v1beta1 version"

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
