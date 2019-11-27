package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CloudFlavor cloud flavor
// swagger:model CloudFlavor
type CloudFlavor struct {

	// cost of CloudFlavor.
	Cost *string `json:"cost,omitempty"`

	// Number of disk_gb.
	DiskGb *int32 `json:"disk_gb,omitempty"`

	// Placeholder for description of property enhanced_nw of obj type CloudFlavor field type str  type boolean
	EnhancedNw *bool `json:"enhanced_nw,omitempty"`

	// id of CloudFlavor.
	// Required: true
	ID *string `json:"id"`

	// If a vm flavor is recommended for requested se_usage_type.Set to True if the chosen VM flavor is recommended for requested se_usage_type.Else set to False. Field introduced in 18.1.4, 18.2.1.
	IsRecommended *bool `json:"is_recommended,omitempty"`

	// Maximum number of IPv6 addresses that can be configured per NIC. Field introduced in 18.1.1.
	MaxIP6SPerNic *int32 `json:"max_ip6s_per_nic,omitempty"`

	// Number of max_ips_per_nic.
	MaxIpsPerNic *int32 `json:"max_ips_per_nic,omitempty"`

	// Number of max_nics.
	MaxNics *int32 `json:"max_nics,omitempty"`

	// Placeholder for description of property meta of obj type CloudFlavor field type str  type object
	Meta []*CloudMeta `json:"meta,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// Placeholder for description of property public of obj type CloudFlavor field type str  type boolean
	Public *bool `json:"public,omitempty"`

	// Number of ram_mb.
	RAMMb *int32 `json:"ram_mb,omitempty"`

	// Number of vcpus.
	Vcpus *int32 `json:"vcpus,omitempty"`
}
