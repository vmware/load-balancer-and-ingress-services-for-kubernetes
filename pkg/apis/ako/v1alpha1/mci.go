/*
 * Copyright 2022 VMware, Inc.
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

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=multiclusteringresses,shortName=mci,singular=multiclusteringress,scope=Namespaced

// MultiClusterIngress is the top-level type
type MultiClusterIngress struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +kubebuilder:validation:Required
	// spec for MultiClusterIngress Config
	Spec MultiClusterIngressSpec `json:"spec,omitempty"`
	// +optional
	Status MultiClusterIngressStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// MultiClusterIngressList is a list of GSLBConfig resources
type MultiClusterIngressList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MultiClusterIngress `json:"items"`
}

// MultiClusterIngressSpec is the spec for a MultiClusterIngress object
type MultiClusterIngressSpec struct {
	// +kubebuilder:validation:Required
	// FQDN of the ingresses of tenant clusters, from which MCI is
	// built
	Hostname string `json:"hostName,omitempty"`
	// Name of the secret object associated with the hostName in
	// the MCI object
	SecretName string `json:"secretName,omitempty"`
	// +kubebuilder:validation:Required
	Config []BackendConfig `json:"config,omitempty"`
}

// Configuration of the tenant target services
type BackendConfig struct {
	// path specified in the tenant cluster's ingress object
	Path string `json:"path,omitempty"`
	// cluster context name of the tenant cluster
	ClusterContext string `json:"cluster,omitempty"`
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=100
	// weight of this member in the resultant virtual service
	Weight  int     `json:"weight,omitempty"`
	Service Service `json:"service,omitempty"`
}

// kubernetes backend service and port configuration in the tenant cluster
type Service struct {
	// target service name in the tenant cluster
	Name string `json:"name,omitempty"`
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// port number of the target service in the tenant cluster
	Port int `json:"port,omitempty"`
	// namespace of the target service in the kubernetes cluster
	Namespace string `json:"namespace,omitempty"`
}

// MultiClusterIngressStatus represents the current status of the MultiClusterIngress object
type MultiClusterIngressStatus struct {
	// represents the load balancing properties of this object
	LoadBalancer LoadBalancer `json:"loadBalancer,omitempty"`
	// represents the ingress details like vip assigned to this
	// object
	Status AcceptedStatus `json:"status,omitempty"`
}

// AcceptedStatus represents whether the MCI object was accepted or rejected. It also
// includes the reason for rejection.
type AcceptedStatus struct {
	// represents whether the MCI object is accepted/rejected
	Accepted bool `json:"accepted,omitempty"`
	// describes the reason, if the MCI object is rejected
	Reason string `json:"reason"`
}

// LoadBalancer status is updated by AKO in the MultiClusterIngress object. It contains the
// VIP fetched from the load balancer and the host fqdn this vip is mapped to.
type LoadBalancer struct {
	Ingress []IngressStatus `json:"ingress,omitempty"`
}

// IngressStatus contains the ingress details required for the traffic
type IngressStatus struct {
	// host fqdn handled by this multi cluster ingress
	// resource
	Hostname string `json:"hostname,omitempty"`
	// virtual IP address assigned to this object by the
	// load balancer
	IP string `json:"ip,omitempty"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=clustersets,shortName=cs,singular=clusterset,scope=Namespaced
// ClusterSet is the top-level type
type ClusterSet struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +kubebuilder:validation:Required
	Spec ClusterSetSpec `json:"spec,omitempty"`
	// +optional
	Status ClusterSetStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// ClusterSetList is a list of GSLBConfig resources
type ClusterSetList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterSet `json:"items"`
}

// ClusterSetSpec has the configuration of the cluster list which form this set.
// It also has the secret which contains the kubeconfig for all the clusters defined
// in this set.
type ClusterSetSpec struct {
	// Cluster context names in the cluster set boundary
	// +optional
	Clusters []ClusterConfig `json:"clusters,omitempty"`
	// Kubeconfig secret containing the configuration for all the
	// clusters defined in the clusters field
	SecretName string `json:"secretName,omitempty"`
}

// ClusterConfig has the contains the cluster context name.
type ClusterConfig struct {
	// Context name for a kubernetes cluster as defined in
	// the kubeconfig secret
	// +optional.
	Context string `json:"context"`
}

// ClusterSetStatus has the status of the clusters
type ClusterSetStatus struct {
	// +optional
	ServiceDiscovery []ServiceDiscoveryStatus `json:"serviceDiscoveryStatus"`
}

// ServiceDiscoveryStatus contains the cluster and it's last status: connected or not.
type ServiceDiscoveryStatus struct {
	Cluster string `json:"cluster,omitempty"`
	Status  string `json:"status,omitempty"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=serviceimports,shortName=si,singular=serviceimport,scope=Namespaced
// ServiceImport is the top-level type
type ServiceImport struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// spec for MultiClusterIngress Config
	// +kubebuilder:validation:Required
	Spec ServiceImportSpec `json:"spec,omitempty"`
	// +optional
	// Status ServiceImportStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// ServiceImportList is a list of ServiceImport resources
type ServiceImportList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceImport `json:"items"`
}

type ServiceImportSpec struct {
	//Cluster from which this service is imported from
	Cluster string `json:"cluster,omitempty"`
	//namespace in the backend cluster where this service exists
	Namespace string `json:"namespace,omitempty"`
	//name of the service in the backend cluster
	Service string `json:"service,omitempty"`
	//ports specified for this service in the MCI object
	SvcPorts []BackendPort `json:"svcPorts,omitempty"`
}

type BackendPort struct {
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	//service port specified for this service in the MCI
	Port int32 `json:"port,omitempty"`
	//endpoints for this service in the backend cluster
	Endpoints []IPPort `json:"endpoints,omitempty"`
}
type IPPort struct {
	//IP address of the backend server
	IP string `json:"ip,omitempty"`
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// Port number of the backend server
	Port int32 `json:"port,omitempty"`
}
