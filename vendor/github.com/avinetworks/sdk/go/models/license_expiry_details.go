package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// LicenseExpiryDetails license expiry details
// swagger:model LicenseExpiryDetails
type LicenseExpiryDetails struct {

	// Number of backend_servers.
	BackendServers *int32 `json:"backend_servers,omitempty"`

	// Number of burst_cores.
	BurstCores *int32 `json:"burst_cores,omitempty"`

	// Number of cores.
	Cores *int32 `json:"cores,omitempty"`

	// expiry_at of LicenseExpiryDetails.
	ExpiryAt *string `json:"expiry_at,omitempty"`

	// license_id of LicenseExpiryDetails.
	LicenseID *string `json:"license_id,omitempty"`

	// license_tier of LicenseExpiryDetails.
	LicenseTier []string `json:"license_tier,omitempty"`

	// license_type of LicenseExpiryDetails.
	LicenseType *string `json:"license_type,omitempty"`

	// Number of max_apps.
	MaxApps *int32 `json:"max_apps,omitempty"`

	// Number of max_ses.
	MaxSes *int32 `json:"max_ses,omitempty"`

	// Name of the object.
	Name *string `json:"name,omitempty"`

	// Number of sockets.
	Sockets *int32 `json:"sockets,omitempty"`

	// Number of throughput.
	Throughput *int32 `json:"throughput,omitempty"`
}
