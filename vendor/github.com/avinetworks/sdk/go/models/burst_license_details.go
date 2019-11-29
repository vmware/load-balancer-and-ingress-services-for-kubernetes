package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// BurstLicenseDetails burst license details
// swagger:model BurstLicenseDetails
type BurstLicenseDetails struct {

	// Number of cores.
	Cores *int32 `json:"cores,omitempty"`

	// end_time of BurstLicenseDetails.
	EndTime *string `json:"end_time,omitempty"`

	// se_name of BurstLicenseDetails.
	SeName *string `json:"se_name,omitempty"`

	// Unique object identifier of se.
	SeUUID *string `json:"se_uuid,omitempty"`

	// start_time of BurstLicenseDetails.
	StartTime *string `json:"start_time,omitempty"`
}
