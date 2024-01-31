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

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

type FqdnType string

const (
	// Matches the string character by character to the VS FQDNs,
	// in an exact match fashion.
	Exact FqdnType = "Exact"

	// Matches the string to multiple VS FQDNs, and matches the FQDNs
	// with the provided string as the suffix. The string must start with
	// a '*' to qualify for wildcard matching.
	// fqdn: *.alb.vmware.com
	Wildcard FqdnType = "Wildcard"

	// Matches the string to multiple VS FQDNs, and matches the FQDNs
	// with the provided string as a substring of any possible FQDNs
	// programmed by AKO.
	// fqdn: Shared-VS-1
	Contains FqdnType = "Contains"
)

// HostRuleVirtualHost defines properties for a host
type HostRuleVirtualHost struct {
	AnalyticsProfile      string                   `json:"analyticsProfile,omitempty"`
	ApplicationProfile    string                   `json:"applicationProfile,omitempty"`
	Datascripts           []string                 `json:"datascripts,omitempty"`
	EnableVirtualHost     *bool                    `json:"enableVirtualHost,omitempty"`
	ErrorPageProfile      string                   `json:"errorPageProfile,omitempty"`
	Fqdn                  string                   `json:"fqdn,omitempty"`
	FqdnType              FqdnType                 `json:"fqdnType,omitempty"`
	HTTPPolicy            HostRuleHTTPPolicy       `json:"httpPolicy,omitempty"`
	Gslb                  HostRuleGSLB             `json:"gslb,omitempty"`
	TLS                   HostRuleTLS              `json:"tls,omitempty"`
	WAFPolicy             string                   `json:"wafPolicy,omitempty"`
	AnalyticsPolicy       *HostRuleAnalyticsPolicy `json:"analyticsPolicy,omitempty"`
	TCPSettings           *HostRuleTCPSettings     `json:"tcpSettings,omitempty"`
	Aliases               []string                 `json:"aliases,omitempty"`
	ICAPProfile           []string                 `json:"icapProfile,omitempty"`
	NetworkSecurityPolicy string                   `json:"networkSecurityPolicy,omitempty"`
	L7Rule                string                   `json:"l7Rule,omitempty"`
}

// HostRuleTCPSettings allows for customizing TCP settings
type HostRuleTCPSettings struct {
	Listeners      []HostRuleTCPListeners `json:"listeners,omitempty"`
	LoadBalancerIP string                 `json:"loadBalancerIP,omitempty"`
}

// HostRuleTCPListeners holds fields to program listener settings
// like port to be exposed and enableSsl/enableHttp2 on the port
type HostRuleTCPListeners struct {
	Port      int  `json:"port,omitempty"`
	EnableSSL bool `json:"enableSSL,omitempty"`
}

// HostRuleTLS holds secure host specific properties
type HostRuleTLS struct {
	SSLKeyCertificate HostRuleSSLKeyCertificate `json:"sslKeyCertificate,omitempty"`
	SSLProfile        string                    `json:"sslProfile,omitempty"`
	Termination       string                    `json:"termination,omitempty"`
}

// HostRuleSecret is required to provide distinction between Avi SSLKeyCertificate
// or K8s Secret Objects
type HostRuleSecret struct {
	Name string             `json:"name,omitempty"`
	Type HostRuleSecretType `json:"type,omitempty"`
}
type HostRuleSSLKeyCertificate struct {
	Name                 string             `json:"name,omitempty"`
	Type                 HostRuleSecretType `json:"type,omitempty"`
	AlternateCertificate HostRuleSecret     `json:"alternateCertificate,omitempty"`
}

type HostRuleSecretType string

const (
	HostRuleSecretTypeAviReference    HostRuleSecretType = "ref"
	HostRuleSecretTypeSecretReference HostRuleSecretType = "secret"
)

// HostRuleHTTPPolicy holds knobs and refs for httpPolicySets
type HostRuleHTTPPolicy struct {
	PolicySets []string `json:"policySets,omitempty"`
	Overwrite  bool     `json:"overwrite,omitempty"`
}

// HostRuleHTTPPolicy holds knobs and refs for httpPolicySets
type HostRuleGSLB struct {
	Fqdn           string `json:"fqdn,omitempty"`
	IncludeAliases bool   `json:"includeAliases,omitempty"`
}

// HostRuleStatus holds the status of the HostRule
type HostRuleStatus struct {
	Status string `json:"status,omitempty"`
	Error  string `json:"error"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HostRuleList has the list of HostRule objects
type HostRuleList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []HostRule `json:"items"`
}

// HostRuleAnalyticsPolicy holds analytics policy objects
type HostRuleAnalyticsPolicy struct {
	FullClientLogs *FullClientLogs `json:"fullClientLogs,omitempty"`
	LogAllHeaders  *bool           `json:"logAllHeaders,omitempty"`
}

// FullClientLogs hold the client log properties
type FullClientLogs struct {
	Enabled  *bool  `json:"enabled,omitempty"`
	Throttle string `json:"throttle,omitempty"`
}
