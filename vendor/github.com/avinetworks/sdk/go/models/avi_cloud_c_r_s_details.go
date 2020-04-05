package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AviCloudCRSDetails avi cloud c r s details
// swagger:model AviCloudCRSDetails
type AviCloudCRSDetails struct {

	// Name of the CRS release. Field introduced in 18.2.6.
	Name *string `json:"name,omitempty"`

	// CRS release date. Field introduced in 18.2.6.
	ReleaseDate *string `json:"release_date,omitempty"`

	// Download link of the CRS release. Field introduced in 18.2.6.
	URL *string `json:"url,omitempty"`

	// Version of the CRS release. Field introduced in 18.2.6.
	Version *string `json:"version,omitempty"`
}
