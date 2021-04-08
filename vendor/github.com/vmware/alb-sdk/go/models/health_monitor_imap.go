package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HealthMonitorImap health monitor imap
// swagger:model HealthMonitorImap
type HealthMonitorImap struct {

	// Folder to access. Field introduced in 21.1.1.
	Folder *string `json:"folder,omitempty"`

	// SSL attributes for IMAPS monitor. Field introduced in 21.1.1.
	SslAttributes *HealthMonitorSSlattributes `json:"ssl_attributes,omitempty"`
}
