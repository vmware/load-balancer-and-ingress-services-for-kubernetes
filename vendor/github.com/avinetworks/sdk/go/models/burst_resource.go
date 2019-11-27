package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// BurstResource burst resource
// swagger:model BurstResource
type BurstResource struct {

	// License ID against which this burst has been accounted. Field introduced in 17.2.5.
	AccountedLicenseID *string `json:"accounted_license_id,omitempty"`

	// Time UTC of the last alert created for this burst resource. Field introduced in 17.2.5.
	LastAlertTime *string `json:"last_alert_time,omitempty"`

	//  Enum options - ENTERPRISE_16, ENTERPRISE_18. Field introduced in 17.2.5.
	LicenseTier *string `json:"license_tier,omitempty"`

	//  Field introduced in 17.2.5.
	SeCookie *string `json:"se_cookie,omitempty"`

	// Service Engine which triggered the burst license usage. Field introduced in 17.2.5.
	SeUUID *string `json:"se_uuid,omitempty"`

	// Time UTC when the burst license was put in use. Field introduced in 17.2.5.
	StartTime *string `json:"start_time,omitempty"`
}
