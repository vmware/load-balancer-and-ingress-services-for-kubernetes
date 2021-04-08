package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HTTPResponsePolicy HTTP response policy
// swagger:model HTTPResponsePolicy
type HTTPResponsePolicy struct {

	// Add rules to the HTTP response policy.
	Rules []*HTTPResponseRule `json:"rules,omitempty"`
}
