package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AlertTestEmailParams alert test email params
// swagger:model AlertTestEmailParams
type AlertTestEmailParams struct {

	// The Subject line of the originating email from  Avi Controller.
	// Required: true
	Subject *string `json:"subject"`

	// The email context.
	// Required: true
	Text *string `json:"text"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
