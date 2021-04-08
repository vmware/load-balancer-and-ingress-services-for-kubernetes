package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// PortalFeatureOptIn portal feature opt in
// swagger:model PortalFeatureOptIn
type PortalFeatureOptIn struct {

	// Enable to subscribe to automated Application Signature Rulesets updates. Field introduced in 20.1.4. Allowed in Basic(Allowed values- false) edition, Essentials(Allowed values- false) edition, Enterprise edition.
	EnableAppsignatureSync *bool `json:"enable_appsignature_sync,omitempty"`

	// Enable pro-active support case creation when a service engine failure occurs. Field introduced in 20.1.1. Allowed in Basic(Allowed values- false) edition, Essentials(Allowed values- false) edition, Enterprise edition.
	EnableAutoCaseCreationOnSeFailure *bool `json:"enable_auto_case_creation_on_se_failure,omitempty"`

	// Enable pro-active support case creation when a system failure occurs. Manual download will still be available in the customer portal. Field introduced in 20.1.1. Allowed in Basic(Allowed values- false) edition, Essentials(Allowed values- false) edition, Enterprise edition.
	EnableAutoCaseCreationOnSystemFailure *bool `json:"enable_auto_case_creation_on_system_failure,omitempty"`

	// Enable to automatically download new CRS version to the Controller. Field introduced in 20.1.1. Allowed in Basic(Allowed values- false) edition, Essentials(Allowed values- false) edition, Enterprise edition.
	EnableAutoDownloadWafSignatures *bool `json:"enable_auto_download_waf_signatures,omitempty"`

	// Enable to subscribe to IP reputation updates. This is a requirement for using IP reputation in the product. Field introduced in 20.1.1. Allowed in Basic(Allowed values- false) edition, Essentials(Allowed values- false) edition, Enterprise edition.
	EnableIPReputation *bool `json:"enable_ip_reputation,omitempty"`

	// Enable event notifications when new CRS versions are available. Field introduced in 20.1.1. Allowed in Basic(Allowed values- false) edition, Essentials(Allowed values- false) edition, Enterprise edition. Special default for Basic edition is false, Essentials edition is false, Enterprise is True.
	EnableWafSignaturesNotifications *bool `json:"enable_waf_signatures_notifications,omitempty"`
}
