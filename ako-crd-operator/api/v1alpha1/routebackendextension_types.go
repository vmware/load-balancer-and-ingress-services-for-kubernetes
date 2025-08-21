/*
Copyright 2019-2025 VMware, Inc.
All Rights Reserved.

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

// HealthMonitorKind defines the type of HealthMonitor object.
type HealthMonitorKind string

const (
	HealthMonitorKindAVIREF HealthMonitorKind = "AVIREF"
)

// BackendHealthMonitor defines the desired state of BackendHealthMontior.
type BackendHealthMonitor struct {
	// Defines the type of HealthMonitor object
	// +kubebuilder:validation:Enum=AVIREF
	// +required
	Kind HealthMonitorKind `json:"kind,omitempty"`
	// Defines the name of HealthMonitor object. HealthMonitor object should be in the same namespace as that of RouteBackendExtension object
	// +required
	Name string `json:"name,omitempty"`
}

// LBAlgorithmType defines the type of LB algorithm.
// +kubebuilder:validation:Enum=LB_ALGORITHM_LEAST_CONNECTIONS;LB_ALGORITHM_ROUND_ROBIN;LB_ALGORITHM_FASTEST_RESPONSE;LB_ALGORITHM_CONSISTENT_HASH;LB_ALGORITHM_LEAST_LOAD;LB_ALGORITHM_FEWEST_SERVERS;LB_ALGORITHM_RANDOM;LB_ALGORITHM_FEWEST_TASKS;LB_ALGORITHM_NEAREST_SERVER;LB_ALGORITHM_CORE_AFFINITY;LB_ALGORITHM_TOPOLOGY
type LBAlgorithmType string

const (
	LBAlgorithmLeastConnections LBAlgorithmType = "LB_ALGORITHM_LEAST_CONNECTIONS"
	LBAlgorithmRoundRobin       LBAlgorithmType = "LB_ALGORITHM_ROUND_ROBIN"
	LBAlgorithmFastestResponse  LBAlgorithmType = "LB_ALGORITHM_FASTEST_RESPONSE"
	LBAlgorithmConsistentHash   LBAlgorithmType = "LB_ALGORITHM_CONSISTENT_HASH"
	LBAlgorithmLeastLoad        LBAlgorithmType = "LB_ALGORITHM_LEAST_LOAD"
	LBAlgorithmFewestServers    LBAlgorithmType = "LB_ALGORITHM_FEWEST_SERVERS"
	LBAlgorithmRandom           LBAlgorithmType = "LB_ALGORITHM_RANDOM"
	LBAlgorithmFewestTasks      LBAlgorithmType = "LB_ALGORITHM_FEWEST_TASKS"
	LBAlgorithmNearestServer    LBAlgorithmType = "LB_ALGORITHM_NEAREST_SERVER"
	LBAlgorithmCoreAffinity     LBAlgorithmType = "LB_ALGORITHM_CORE_AFFINITY"
	LBAlgorithmTopology         LBAlgorithmType = "LB_ALGORITHM_TOPOLOGY"
)

// LBAlgorithmHashType defines criteria used as a key for determining the hash between the client and server
// +kubebuilder:validation:Enum=LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS;LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS_AND_PORT;LB_ALGORITHM_CONSISTENT_HASH_URI;LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER;LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_STRING;LB_ALGORITHM_CONSISTENT_HASH_CALLID
type LBAlgorithmHashType string

const (
	LBAlgorithmConsistentHashSourceIPAddress        LBAlgorithmHashType = "LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS"
	LBAlgorithmConsistentHashSourceIPAddressAndPort LBAlgorithmHashType = "LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS_AND_PORT"
	LBAlgorithmConsistentHashURI                    LBAlgorithmHashType = "LB_ALGORITHM_CONSISTENT_HASH_URI"
	LBAlgorithmConsistentHashCustomHeader           LBAlgorithmHashType = "LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER"
	LBAlgorithmConsistentHashCustomString           LBAlgorithmHashType = "LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_STRING"
	LBAlgorithmConsistentHashCallID                 LBAlgorithmHashType = "LB_ALGORITHM_CONSISTENT_HASH_CALLID"
)

// RouteBackendExtensionSpec defines the desired state of RouteBackendExtension
// +kubebuilder:validation:XValidation:rule="(self.lbAlgorithm == 'LB_ALGORITHM_CONSISTENT_HASH') && has(self.lbAlgorithmHash)",message="lbAlgorithmHash must be set if and only if lbAlgorithm is LB_ALGORITHM_CONSISTENT_HASH"
// +kubebuilder:validation:XValidation:rule="!has(self.lbAlgorithmHash) || (self.lbAlgorithmHash == 'LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER') && has(self.lbAlgorithmConsistentHashHdr)",message="lbAlgorithmConsistentHashHdr must be set if and only if lbAlgorithmHash is LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER"

type RouteBackendExtensionSpec struct {
	// Defines LB algorithm on Pool
	// +optional
	// +kubebuilder:default=LB_ALGORITHM_LEAST_CONNECTIONS
	LBAlgorithm LBAlgorithmType `json:"lbAlgorithm,omitempty"`
	// HTTP header name to be used for the hash key.
	// This field should be specified only when lbAlgorithm is LB_ALGORITHM_CONSISTENT_HASH and lbAlgorithmHash is LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER.
	// +optional
	LBAlgorithmConsistentHashHdr string `json:"lbAlgorithmConsistentHashHdr,omitempty"`
	// Criteria used as a key for determining the hash between the client and server
	// +optional
	// +kubebuilder:default=LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS
	LBAlgorithmHash LBAlgorithmHashType `json:"lbAlgorithmHash,omitempty"`
	// Represents health monitor objects
	// +optional
	HealthMonitor []BackendHealthMonitor `json:"healthMonitor,omitempty"`
}

// RouteBackendExtensionStatus defines the observed state of RouteBackendExtension.
type RouteBackendExtensionStatus struct {
	// Field is populated by AKO CRD operator as ako-crd-operator
	// +optional
	Controller string `json:"controller,omitempty"`
	Error      string `json:"error,omitempty"`
	Status     string `json:"status,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=routebackendextension,scope=Namespaced
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=routebackendextensions,shortName=rbe,singular=routebackendextension,scope=Namespaced
// RouteBackendExtension is the Schema for the routebackendextensions API.
type RouteBackendExtension struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RouteBackendExtensionSpec   `json:"spec,omitempty"`
	Status RouteBackendExtensionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RouteBackendExtensionList contains a list of RouteBackendExtension.
type RouteBackendExtensionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RouteBackendExtension `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RouteBackendExtension{}, &RouteBackendExtensionList{})
}

func (rb *RouteBackendExtension) SetRouteBackendExtensionController(controller string) {
	rb.Status.Controller = controller
}
