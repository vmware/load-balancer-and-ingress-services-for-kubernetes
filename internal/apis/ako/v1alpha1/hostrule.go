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

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HostRule is a top-level type
type HostRule struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Status HostRuleStatus `json:"status,omitempty"`

	Spec HostRuleSpec `json:"spec,omitempty"`
}

// HostRuleSpec consists of the main HostRule settings
type HostRuleSpec struct {
	VirtualHost HostRuleVirtualHost `json:"virtualhost,omitempty"`
}

// HostRuleVirtualHost defines properties for a host
type HostRuleVirtualHost struct {
	AnalyticsProfile   string             `json:"analyticsProfile,omitempty"`
	ApplicationProfile string             `json:"applicationProfile,omitempty"`
	Datascripts        []string           `json:"datascripts,omitempty"`
	EnableVirtualHost  *bool              `json:"enableVirtualHost,omitempty"`
	ErrorPageProfile   string             `json:"errorPageProfile,omitempty"`
	Fqdn               string             `json:"fqdn,omitempty"`
	HTTPPolicy         HostRuleHTTPPolicy `json:"httpPolicy,omitempty"`
	TLS                HostRuleTLS        `json:"tls,omitempty"`
	WAFPolicy          string             `json:"wafPolicy,omitempty"`
}

// HostRuleTLS holds secure host specific properties
type HostRuleTLS struct {
	SSLKeyCertificate HostRuleSecret `json:"sslKeyCertificate,omitempty"`
	SSLProfile        string         `json:"sslProfile,omitempty"`
	Termination       string         `json:"termination,omitempty"`
}

// HostRuleSecret is required to provide distinction between Avi SSLKeyCertificate
// or K8s Secret Objects
type HostRuleSecret struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}

// HostRuleHTTPPolicy holds knobs and refs for httpPolicySets
type HostRuleHTTPPolicy struct {
	PolicySets []string `json:"policySets,omitempty"`
	Overwrite  bool     `json:"overwrite,omitempty"`
}

// HostRuleStatus holds the status of the HostRule
type HostRuleStatus struct {
	Status string `json:"status,omitempty"`
	Error  string `json:"error,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HostRuleList has the list of HostRule objects
type HostRuleList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []HostRule `json:"items"`
}
