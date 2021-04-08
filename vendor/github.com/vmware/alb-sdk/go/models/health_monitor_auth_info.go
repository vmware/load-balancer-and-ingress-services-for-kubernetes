package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HealthMonitorAuthInfo health monitor auth info
// swagger:model HealthMonitorAuthInfo
type HealthMonitorAuthInfo struct {

	// Password for server authentication. Field introduced in 20.1.1.
	// Required: true
	Password *string `json:"password"`

	// Username for server authentication. Field introduced in 20.1.1.
	// Required: true
	Username *string `json:"username"`
}
