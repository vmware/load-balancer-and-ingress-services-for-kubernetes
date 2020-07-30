package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IPReputationConfig Ip reputation config
// swagger:model IpReputationConfig
type IPReputationConfig struct {

	// IP reputation db file object expiry duration in days. Allowed values are 1-7. Field introduced in 20.1.1.
	IPReputationFileObjectExpiryDuration *int32 `json:"ip_reputation_file_object_expiry_duration,omitempty"`

	// IP reputation db sync interval in minutes. Allowed values are 2-1440. Field introduced in 20.1.1.
	IPReputationSyncInterval *int32 `json:"ip_reputation_sync_interval,omitempty"`
}
