package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ErrorPage error page
// swagger:model ErrorPage
type ErrorPage struct {

	// Enable or disable the error page. Field introduced in 17.2.4.
	Enable *bool `json:"enable,omitempty"`

	// Custom error page body used to sent to the client. It is a reference to an object of type ErrorPageBody. Field introduced in 17.2.4.
	ErrorPageBodyRef *string `json:"error_page_body_ref,omitempty"`

	// Redirect sent to client when match. Field introduced in 17.2.4.
	ErrorRedirect *string `json:"error_redirect,omitempty"`

	// Index of the error page. Field introduced in 17.2.4.
	Index *int32 `json:"index,omitempty"`

	// Add match criteria for http status codes to the error page. Field introduced in 17.2.4.
	Match *HttpstatusMatch `json:"match,omitempty"`
}
