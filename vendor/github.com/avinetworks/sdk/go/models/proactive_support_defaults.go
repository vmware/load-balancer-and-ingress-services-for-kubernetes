package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ProactiveSupportDefaults proactive support defaults
// swagger:model ProactiveSupportDefaults
type ProactiveSupportDefaults struct {

	// Opt-in to attach core dump with support case. Field introduced in 20.1.1.
	AttachCoreDump *bool `json:"attach_core_dump,omitempty"`

	// Opt-in to attach tech support with support case. Field introduced in 20.1.1.
	AttachTechSupport *bool `json:"attach_tech_support,omitempty"`

	// Case severity to be used for proactive support case creation. Field introduced in 20.1.1.
	CaseSeverity *string `json:"case_severity,omitempty"`
}
