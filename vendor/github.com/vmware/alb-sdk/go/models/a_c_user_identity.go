package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ACUserIdentity a c user identity
// swagger:model ACUserIdentity
type ACUserIdentity struct {

	// User identity type for audit event (e.g. username, organization, component). Field introduced in 20.1.3.
	// Required: true
	Type *string `json:"type"`

	// User identity value for audit event (e.g. SomeCompany, Jane Doe, Secure-shell). Field introduced in 20.1.3.
	// Required: true
	Value *string `json:"value"`
}
