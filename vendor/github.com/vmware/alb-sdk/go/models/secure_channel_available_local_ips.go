package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SecureChannelAvailableLocalIps secure channel available local ips
// swagger:model SecureChannelAvailableLocalIPs
type SecureChannelAvailableLocalIps struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Number of end.
	End *int32 `json:"end,omitempty"`

	// free_controller_ips of SecureChannelAvailableLocalIPs.
	FreeControllerIps []string `json:"free_controller_ips,omitempty"`

	// free_ips of SecureChannelAvailableLocalIPs.
	FreeIps []string `json:"free_ips,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// Number of start.
	Start *int32 `json:"start,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
