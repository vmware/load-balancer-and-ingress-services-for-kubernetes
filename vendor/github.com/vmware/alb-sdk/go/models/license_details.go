package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// LicenseDetails license details
// swagger:model LicenseDetails
type LicenseDetails struct {

	// Number of backend_servers.
	BackendServers *int32 `json:"backend_servers,omitempty"`

	// expiry_at of LicenseDetails.
	ExpiryAt *string `json:"expiry_at,omitempty"`

	// license_id of LicenseDetails.
	LicenseID *string `json:"license_id,omitempty"`

	// license_type of LicenseDetails.
	LicenseType *string `json:"license_type,omitempty"`

	// Name of the object.
	Name *string `json:"name,omitempty"`
}
