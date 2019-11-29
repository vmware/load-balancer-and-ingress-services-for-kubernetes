package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HTTPRequestPolicy HTTP request policy
// swagger:model HTTPRequestPolicy
type HTTPRequestPolicy struct {

	// Add rules to the HTTP request policy.
	Rules []*HTTPRequestRule `json:"rules,omitempty"`
}
