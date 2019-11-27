package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CloudConnectorDebugFilter cloud connector debug filter
// swagger:model CloudConnectorDebugFilter
type CloudConnectorDebugFilter struct {

	// filter debugs for an app.
	AppID *string `json:"app_id,omitempty"`

	// Disable SE reboot via cloud connector on HB miss.
	DisableSeReboot *bool `json:"disable_se_reboot,omitempty"`

	// filter debugs for a SE.
	SeID *string `json:"se_id,omitempty"`
}
