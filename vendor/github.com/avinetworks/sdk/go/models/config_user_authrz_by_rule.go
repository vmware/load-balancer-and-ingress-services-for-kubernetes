package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ConfigUserAuthrzByRule config user authrz by rule
// swagger:model ConfigUserAuthrzByRule
type ConfigUserAuthrzByRule struct {

	// assigned roles.
	Roles *string `json:"roles,omitempty"`

	// matching rule string.
	Rule *string `json:"rule,omitempty"`

	// assigned tenants.
	Tenants *string `json:"tenants,omitempty"`

	// Request user.
	User *string `json:"user,omitempty"`
}
