package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AssetContactInfo asset contact info
// swagger:model AssetContactInfo
type AssetContactInfo struct {

	// Contact ID of the point of contact for this asset. Field introduced in 20.1.1.
	ContactID *string `json:"contact_id,omitempty"`

	// Name of the point of contact for this asset. Field introduced in 20.1.1.
	Name *string `json:"name,omitempty"`
}
