package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeGeoDbDetails se geo db details
// swagger:model SeGeoDbDetails
type SeGeoDbDetails struct {

	// Geo Db file name.
	FileName *string `json:"file_name,omitempty"`

	// Name of the Gslb Geo Db Profile.
	GeoDbProfileName *string `json:"geo_db_profile_name,omitempty"`

	// UUID of the Gslb Geo Db Profile. It is a reference to an object of type GslbGeoDbProfile.
	GeoDbProfileRef *string `json:"geo_db_profile_ref,omitempty"`

	// Reason for Gslb Geo Db failure. Enum options - NO_ERROR, FILE_ERROR, FILE_FORMAT_ERROR, FILE_NO_RESOURCES.
	Reason *string `json:"reason,omitempty"`

	// UUID of the SE responsible for this event. It is a reference to an object of type ServiceEngine.
	SeRef *string `json:"se_ref,omitempty"`

	// VIP id.
	VipID *string `json:"vip_id,omitempty"`

	// Virtual Service name. It is a reference to an object of type VirtualService.
	VirtualService *string `json:"virtual_service,omitempty"`
}
