package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// PortalFeatureOptIn portal feature opt in
// swagger:model PortalFeatureOptIn
type PortalFeatureOptIn struct {

	// Flag to check if the user has opted in for proactive case creation on service engine failure. Field introduced in 20.1.1.
	EnableAutoCaseCreationOnSeFailure *bool `json:"enable_auto_case_creation_on_se_failure,omitempty"`

	// Flag to check if the user has opted in for proactive case creation on system failure. Field introduced in 20.1.1.
	EnableAutoCaseCreationOnSystemFailure *bool `json:"enable_auto_case_creation_on_system_failure,omitempty"`

	// Flag to check if the user has opted in for auto deployment of CRS data on controller. Field introduced in 20.1.1.
	EnableAutoDownloadWafSignatures *bool `json:"enable_auto_download_waf_signatures,omitempty"`

	// Flag to check if the user has opted in for automated IP reputation db sync. Field introduced in 20.1.1.
	EnableIPReputation *bool `json:"enable_ip_reputation,omitempty"`

	// Flag to check if the user has opted in for notifications about the availability of new CRS data. Field introduced in 20.1.1.
	EnableWafSignaturesNotifications *bool `json:"enable_waf_signatures_notifications,omitempty"`
}
