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
	// This is where you can define
	// your own custom spec
	Spec HostRuleSpec `json:"spec,omitempty"`
}

// custom spec
type HostRuleSpec struct {
	VirtualHost HostRuleVirtualHost `json:"virtualhost,omitempty"`
}

type HostRuleVirtualHost struct {
	Fqdn                  string      `json:"fqdn,omitempty"`
	TLS                   HostRuleTLS `json:"tls,omitempty"`
	HTTPPolicySet         []string    `json:"httpPolicySet,omitempty"`
	NetworkSecurityPolicy string      `json:"networkSecurityPolicy,omitempty"`
	WAFPolicy             string      `json:"wafPolicy,omitempty"`
	ApplicationProfile    string      `json:"applicationProfile,omitempty"`
}

type HostRuleTLS struct {
	SSLKeyCertificate string `json:"sslKeyCertificate,omitempty"`
}

// HostRuleStatus holds the status of the HostRule
type HostRuleStatus struct {
	Status string `json:"status,omitempty"`
	Error  string `json:"error,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// no client needed for list as it's been created in above
type HostRuleList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []HostRule `json:"items"`
}
