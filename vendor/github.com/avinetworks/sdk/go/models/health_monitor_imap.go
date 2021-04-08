package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HealthMonitorImap health monitor imap
// swagger:model HealthMonitorImap
type HealthMonitorImap struct {

	// Folder to access. Field introduced in 20.1.5.
	Folder *string `json:"folder,omitempty"`

	// SSL attributes for IMAPS monitor. Field introduced in 20.1.5.
	SslAttributes *HealthMonitorSSlattributes `json:"ssl_attributes,omitempty"`
}
