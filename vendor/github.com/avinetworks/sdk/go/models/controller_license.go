package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ControllerLicense controller license
// swagger:model ControllerLicense
type ControllerLicense struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// List of active burst core license in use. Field introduced in 17.2.5.
	ActiveBurstResources []*BurstResource `json:"active_burst_resources,omitempty"`

	// Total number of Service Engine cores for burst core based licenses. Field introduced in 17.2.5.
	BurstCores *int32 `json:"burst_cores,omitempty"`

	// Number of Service Engine cores in non-container clouds.
	Cores *int32 `json:"cores,omitempty"`

	// customer_name of ControllerLicense.
	// Required: true
	CustomerName *string `json:"customer_name"`

	//  Field introduced in 17.2.5.
	DisableEnforcement *bool `json:"disable_enforcement,omitempty"`

	// List of used or expired burst core licenses. Field introduced in 17.2.5.
	ExpiredBurstResources []*BurstResource `json:"expired_burst_resources,omitempty"`

	//  Field introduced in 17.2.5.
	LicenseID *string `json:"license_id,omitempty"`

	// license_tier of ControllerLicense.
	LicenseTier []string `json:"license_tier,omitempty"`

	//  Field introduced in 17.2.5.
	LicenseTiers []*CumulativeLicense `json:"license_tiers,omitempty"`

	// Placeholder for description of property licenses of obj type ControllerLicense field type str  type object
	Licenses []*SingleLicense `json:"licenses,omitempty"`

	// Number of Service Engines hosts in container clouds.
	MaxSes *int32 `json:"max_ses,omitempty"`

	// Name of the object.
	Name *string `json:"name,omitempty"`

	// Service Engine bandwidth limits for bandwidth based licenses. Field introduced in 17.2.5.
	SeBandwidthLimits []*SEBandwidthLimit `json:"se_bandwidth_limits,omitempty"`

	// Number of physical cpu sockets across Service Engines in no access and linux server clouds.
	Sockets *int32 `json:"sockets,omitempty"`

	// start_on of ControllerLicense.
	StartOn *string `json:"start_on,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`

	// valid_until of ControllerLicense.
	// Required: true
	ValidUntil *string `json:"valid_until"`
}
