// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NsxtConfiguration nsxt configuration
// swagger:model NsxtConfiguration
type NsxtConfiguration struct {

	// Automatically create DFW rules for VirtualService in NSX-T Manager. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Basic edition(Allowed values- false), Essentials, Enterprise with Cloud Services edition.
	AutomateDfwRules *bool `json:"automate_dfw_rules,omitempty"`

	// Data network configuration for Avi Service Engines. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	DataNetworkConfig *DataNetworkConfig `json:"data_network_config"`

	// Domain where NSGroup objects belongs to. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DomainID *string `json:"domain_id,omitempty"`

	// Enforcement point is where the rules of a policy to apply. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnforcementpointID *string `json:"enforcementpoint_id,omitempty"`

	// Management network configuration for Avi Service Engines. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	ManagementNetworkConfig *ManagementNetworkConfig `json:"management_network_config"`

	// Credentials to access NSX-T manager. It is a reference to an object of type CloudConnectorUser. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	NsxtCredentialsRef *string `json:"nsxt_credentials_ref"`

	// NSX-T manager hostname or IP address. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	NsxtURL *string `json:"nsxt_url"`

	// Site where transport zone belongs to. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SiteID *string `json:"site_id,omitempty"`
}
