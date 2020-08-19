package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NsxtSegmentRuntime nsxt segment runtime
// swagger:model NsxtSegmentRuntime
type NsxtSegmentRuntime struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Nsxt segment belongs to cloud. It is a reference to an object of type Cloud. Field introduced in 20.1.1.
	CloudRef *string `json:"cloud_ref,omitempty"`

	// V6 DHCP ranges configured in Nsxt. Field introduced in 20.1.1.
	Dhcp6Ranges []string `json:"dhcp6_ranges,omitempty"`

	// IP address management scheme for this Segment associated network. Field introduced in 20.1.1.
	DhcpEnabled *bool `json:"dhcp_enabled,omitempty"`

	// DHCP ranges configured in Nsxt. Field introduced in 20.1.1.
	DhcpRanges []string `json:"dhcp_ranges,omitempty"`

	// Segment object name. Field introduced in 20.1.1.
	Name *string `json:"name,omitempty"`

	// Network Name. Field introduced in 20.1.1.
	NwName *string `json:"nw_name,omitempty"`

	// Corresponding network object in Avi. It is a reference to an object of type Network. Field introduced in 20.1.1.
	NwRef *string `json:"nw_ref,omitempty"`

	// Opaque network Id. Field introduced in 20.1.1.
	OpaqueNetworkID *string `json:"opaque_network_id,omitempty"`

	// Segment Gateway. Field introduced in 20.1.1.
	SegmentGw *string `json:"segment_gw,omitempty"`

	// V6 segment Gateway. Field introduced in 20.1.1.
	SegmentGw6 *string `json:"segment_gw6,omitempty"`

	// Segment Id. Field introduced in 20.1.1.
	SegmentID *string `json:"segment_id,omitempty"`

	// Segment name. Field introduced in 20.1.1.
	Segname *string `json:"segname,omitempty"`

	// Segment Cidr. Field introduced in 20.1.1.
	Subnet *string `json:"subnet,omitempty"`

	// V6 Segment Cidr. Field introduced in 20.1.1.
	Subnet6 *string `json:"subnet6,omitempty"`

	// Nsxt segment belongs to tenant. It is a reference to an object of type Tenant. Field introduced in 20.1.1.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Tier1 router Id. Field introduced in 20.1.1.
	Tier1ID *string `json:"tier1_id,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Uuid. Field introduced in 20.1.1.
	UUID *string `json:"uuid,omitempty"`

	// Corresponding vrf context object in Avi. It is a reference to an object of type VrfContext. Field introduced in 20.1.1.
	VrfContextRef *string `json:"vrf_context_ref,omitempty"`
}
