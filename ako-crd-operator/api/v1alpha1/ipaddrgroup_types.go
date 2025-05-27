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

type IpAddrGroupSpec struct {
	// Configure IP address(es)
	//TODO (f_create_heat_resource) = true,
	//TODO (f_udiff) = true
	Addrs []IpAddr `json:"addrs,omitempty"`

	// Configure IP address range(s)
	//TODO (f_udiff) = true
	Ranges []IpAddrRange `json:"ranges,omitempty"`

	// Configure IP address prefix(es)
	//TODO (f_udiff) = true
	Prefixes []IpAddrPrefix `json:"prefixes,omitempty"`

	// Populate the IP address ranges from the geo database for this country
	//TODO (f_udiff) = true
	CountryCodes []string `json:"country_codes,omitempty"`

	// Configure (IP address, port) tuple(s)
	//TODO (f_udiff) = true
	IpPorts []IpAddrPort `json:"ip_ports,omitempty"`

	// Populate IP addresses from tasks of this Marathon app
	// +optional
	MarathonAppName string `json:"marathon_app_name,omitempty"`

	// Task port associated with marathon service port. If Marathon app has multiple service ports, this is required. Else, the first task port is used
	// +optional
	MarathonServicePort *uint32 `json:"marathon_service_port,omitempty"`

	// List of labels to be used for granular RBAC.
	// TODO in BASIC and ESSENTIALS allow_any is true
	Markers []RoleFilterMatchLabel `json:"markers,omitempty"`

	// +optional
	Description string `json:"description,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Reference:type=Tenant
	TenantUuid string `json:"tenant_uuid"`
}

type IpAddrPort struct {
	// IP Address of host. One of IP address or hostname should be set.
	// +optional
	Ip *IpAddr `json:"ip,omitempty"`

	// Port number of server.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port uint32 `json:"port"`

	// Hostname of server. One of IP address or hostname should be set.
	// +optional
	Hostname string `json:"hostname,omitempty"`

	// +optional
	Name string `json:"name,omitempty"`
}

// IpAddrGroupStatus defines the observed state of IpAddrGroup.
type IpAddrGroupStatus struct {
	// Status of the application profile
	Status string `json:"status,omitempty"`
	// Error if any error was encountered
	Error string `json:"error"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=ipaddrgroups,scope=Namespaced
// +kubebuilder:subresource:status
// IpAddrGroup is the Schema for the ipaddrgroups API
type IpAddrGroup struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of IpAddrGroup
	// +optional
	Spec IpAddrGroupSpec `json:"spec,omitempty"`

	// Status defines the observed state of IpAddrGroup
	// +optional
	Status IpAddrGroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IpAddrGroupList contains a list of IpAddrGroup.
type IpAddrGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IpAddrGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IpAddrGroup{}, &IpAddrGroupList{})
}
