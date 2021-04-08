package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CompressionFilter compression filter
// swagger:model CompressionFilter
type CompressionFilter struct {

	//  It is a reference to an object of type StringGroup.
	DevicesRef *string `json:"devices_ref,omitempty"`

	// Number of index.
	// Required: true
	Index *int32 `json:"index"`

	// Placeholder for description of property ip_addr_prefixes of obj type CompressionFilter field type str  type object
	IPAddrPrefixes []*IPAddrPrefix `json:"ip_addr_prefixes,omitempty"`

	// Placeholder for description of property ip_addr_ranges of obj type CompressionFilter field type str  type object
	IPAddrRanges []*IPAddrRange `json:"ip_addr_ranges,omitempty"`

	// Placeholder for description of property ip_addrs of obj type CompressionFilter field type str  type object
	IPAddrs []*IPAddr `json:"ip_addrs,omitempty"`

	//  It is a reference to an object of type IpAddrGroup.
	IPAddrsRef *string `json:"ip_addrs_ref,omitempty"`

	//  Enum options - AGGRESSIVE_COMPRESSION, NORMAL_COMPRESSION, NO_COMPRESSION.
	// Required: true
	Level *string `json:"level"`

	// Whether to apply Filter when group criteria is matched or not. Enum options - IS_IN, IS_NOT_IN.
	Match *string `json:"match,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// user_agent of CompressionFilter.
	UserAgent []string `json:"user_agent,omitempty"`
}
