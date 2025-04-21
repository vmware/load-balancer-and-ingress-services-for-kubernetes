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

package v1beta1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HTTPRule is a top-level type
type HTTPRule struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Status HTTPRuleStatus `json:"status,omitempty"`

	Spec HTTPRuleSpec `json:"spec,omitempty"`
}

// HTTPRuleSpec consists of the main HTTPRule settings
type HTTPRuleSpec struct {
	Fqdn  string          `json:"fqdn,omitempty"`
	Paths []HTTPRulePaths `json:"paths,omitempty"`
}

// HTTPRulePaths has settings for a specific target path
type HTTPRulePaths struct {
	Target                 string           `json:"target,omitempty"`
	LoadBalancerPolicy     HTTPRuleLBPolicy `json:"loadBalancerPolicy,omitempty"`
	TLS                    HTTPRuleTLS      `json:"tls,omitempty"`
	HealthMonitors         []string         `json:"healthMonitors,omitempty"`
	ApplicationPersistence string           `json:"applicationPersistence,omitempty"`
}

// HTTPRuleLBPolicy holds a path/pool's load balancer policies
type HTTPRuleLBPolicy struct {
	Algorithm  string `json:"algorithm,omitempty"`
	Hash       string `json:"hash,omitempty"`
	HostHeader string `json:"hostHeader,omitempty"`
}

// HTTPRuleTLS holds secure path/pool specific properties
type HTTPRuleTLS struct {
	Type          string `json:"type,omitempty"`
	SSLProfile    string `json:"sslProfile,omitempty"`
	PKIProfile    string `json:"pkiProfile,omitempty"`
	DestinationCA string `json:"destinationCA,omitempty"`
}

// HTTPRuleStatus holds the status of the HTTPRule
type HTTPRuleStatus struct {
	Status string `json:"status,omitempty"`
	Error  string `json:"error"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HTTPRuleList has the list of HostRule objects
type HTTPRuleList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []HTTPRule `json:"items"`
}
