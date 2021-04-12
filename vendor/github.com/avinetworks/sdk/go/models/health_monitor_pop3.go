package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HealthMonitorPop3 health monitor pop3
// swagger:model HealthMonitorPop3
type HealthMonitorPop3 struct {

	// SSL attributes for POP3S monitor. Field introduced in 20.1.5.
	SslAttributes *HealthMonitorSSlattributes `json:"ssl_attributes,omitempty"`
}
