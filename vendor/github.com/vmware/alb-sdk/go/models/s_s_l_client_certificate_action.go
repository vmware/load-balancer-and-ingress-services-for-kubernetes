package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SSLClientCertificateAction s s l client certificate action
// swagger:model SSLClientCertificateAction
type SSLClientCertificateAction struct {

	// Placeholder for description of property close_connection of obj type SSLClientCertificateAction field type str  type boolean
	CloseConnection *bool `json:"close_connection,omitempty"`

	// Placeholder for description of property headers of obj type SSLClientCertificateAction field type str  type object
	Headers []*SSLClientRequestHeader `json:"headers,omitempty"`
}
