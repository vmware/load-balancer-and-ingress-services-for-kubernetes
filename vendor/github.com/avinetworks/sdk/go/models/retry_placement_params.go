package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// RetryPlacementParams retry placement params
// swagger:model RetryPlacementParams
type RetryPlacementParams struct {

	// Retry placement operations for all East-West services. Field introduced in 17.1.6,17.2.2.
	AllEastWest *bool `json:"all_east_west,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`

	// Indicates the vip_id that needs placement retrial. Field introduced in 17.1.2.
	// Required: true
	VipID *string `json:"vip_id"`
}
