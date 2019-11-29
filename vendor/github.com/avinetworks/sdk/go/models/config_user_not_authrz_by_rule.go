package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ConfigUserNotAuthrzByRule config user not authrz by rule
// swagger:model ConfigUserNotAuthrzByRule
type ConfigUserNotAuthrzByRule struct {

	// assigned roles.
	Roles *string `json:"roles,omitempty"`

	// assigned tenants.
	Tenants *string `json:"tenants,omitempty"`

	// Request user.
	User *string `json:"user,omitempty"`
}
