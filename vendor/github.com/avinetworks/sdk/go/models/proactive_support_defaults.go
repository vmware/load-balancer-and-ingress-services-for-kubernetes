package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ProactiveSupportDefaults proactive support defaults
// swagger:model ProactiveSupportDefaults
type ProactiveSupportDefaults struct {

	// Opt-in to attach core dump with support case. Field introduced in 20.1.1. Allowed in Basic(Allowed values- false) edition, Essentials(Allowed values- false) edition, Enterprise edition.
	AttachCoreDump *bool `json:"attach_core_dump,omitempty"`

	// Opt-in to attach tech support with support case. Field introduced in 20.1.1. Allowed in Basic(Allowed values- false) edition, Essentials(Allowed values- false) edition, Enterprise edition. Special default for Basic edition is false, Essentials edition is false, Enterprise is True.
	AttachTechSupport *bool `json:"attach_tech_support,omitempty"`

	// Case severity to be used for proactive support case creation. Field introduced in 20.1.1.
	CaseSeverity *string `json:"case_severity,omitempty"`
}
