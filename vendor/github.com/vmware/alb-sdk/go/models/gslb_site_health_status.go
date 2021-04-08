package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbSiteHealthStatus gslb site health status
// swagger:model GslbSiteHealthStatus
type GslbSiteHealthStatus struct {

	// Controller retrieved GSLB service operational info based of virtual service state. .
	ControllerGsinfo []*GslbPoolMemberRuntimeInfo `json:"controller_gsinfo,omitempty"`

	// Controller retrieved GSLB service operational info based of dns datapath resolution. This information is generated only on those sites that have DNS-VS participating in GSLB.
	DatapathGsinfo []*GslbPoolMemberRuntimeInfo `json:"datapath_gsinfo,omitempty"`

	// DNS info at the site.
	DNSInfo *GslbDNSInfo `json:"dns_info,omitempty"`

	// GSLB application persistence profile state at member. Field introduced in 17.1.1.
	GapTable []*CfgState `json:"gap_table,omitempty"`

	// GSLB Geo Db profile state at member. Field introduced in 17.1.1.
	GeoTable []*CfgState `json:"geo_table,omitempty"`

	// GSLB health monitor state at member.
	GhmTable []*CfgState `json:"ghm_table,omitempty"`

	// GSLB state at member.
	GlbTable []*CfgState `json:"glb_table,omitempty"`

	// GSLB service state at member.
	GsTable []*CfgState `json:"gs_table,omitempty"`

	// Current Software version of the site.
	SwVersion *string `json:"sw_version,omitempty"`

	// Timestamp of Health-Status generation.
	Timestamp *float32 `json:"timestamp,omitempty"`
}
