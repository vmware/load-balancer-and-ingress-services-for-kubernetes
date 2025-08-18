// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbSiteHealthStatus gslb site health status
// swagger:model GslbSiteHealthStatus
type GslbSiteHealthStatus struct {

	// Controller retrieved GSLB service operational info based of virtual service state. . Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ControllerGsinfo []*GslbPoolMemberRuntimeInfo `json:"controller_gsinfo,omitempty"`

	// Controller retrieved GSLB service operational info based of dns datapath resolution. This information is generated only on those sites that have DNS-VS participating in GSLB. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DatapathGsinfo []*GslbPoolMemberRuntimeInfo `json:"datapath_gsinfo,omitempty"`

	// DNS info at the site. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DNSInfo *GslbDNSInfo `json:"dns_info,omitempty"`

	// GSLB application persistence profile state at member. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GapTable []*CfgState `json:"gap_table,omitempty"`

	// GSLB Geo Db profile state at member. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GeoTable []*CfgState `json:"geo_table,omitempty"`

	// GSLB health monitor state at member. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GhmTable []*CfgState `json:"ghm_table,omitempty"`

	// GSLB state at member. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GlbTable []*CfgState `json:"glb_table,omitempty"`

	// GSLB service state at member. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GsTable []*CfgState `json:"gs_table,omitempty"`

	// Current Software version of the site. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SwVersion *string `json:"sw_version,omitempty"`

	// Timestamp of Health-Status generation. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Timestamp *float32 `json:"timestamp,omitempty"`
}
