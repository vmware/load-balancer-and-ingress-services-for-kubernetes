package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// Webhook webhook
// swagger:model Webhook
type Webhook struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Callback URL for the Webhook. Field introduced in 17.1.1.
	CallbackURL *string `json:"callback_url,omitempty"`

	//  Field introduced in 17.1.1.
	Description *string `json:"description,omitempty"`

	// The name of the webhook profile. Field introduced in 17.1.1.
	// Required: true
	Name *string `json:"name"`

	//  It is a reference to an object of type Tenant. Field introduced in 17.1.1.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the webhook profile. Field introduced in 17.1.1.
	UUID *string `json:"uuid,omitempty"`

	// Verification token sent back with the callback asquery parameters. Field introduced in 17.1.1.
	VerificationToken *string `json:"verification_token,omitempty"`
}
