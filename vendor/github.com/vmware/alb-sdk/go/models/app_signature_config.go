package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AppSignatureConfig app signature config
// swagger:model AppSignatureConfig
type AppSignatureConfig struct {

	// Application Signature db sync interval in minutes. Allowed values are 60-10080. Field introduced in 20.1.4. Unit is MIN. Allowed in Basic(Allowed values- 60) edition, Essentials(Allowed values- 60) edition, Enterprise edition. Special default for Basic edition is 60, Essentials edition is 60, Enterprise is 1440.
	AppSignatureSyncInterval *int32 `json:"app_signature_sync_interval,omitempty"`
}
