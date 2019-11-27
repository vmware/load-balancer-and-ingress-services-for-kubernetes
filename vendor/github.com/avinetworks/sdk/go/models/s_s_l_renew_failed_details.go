package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SSLRenewFailedDetails s s l renew failed details
// swagger:model SSLRenewFailedDetails
type SSLRenewFailedDetails struct {

	// Error when renewing certificate.
	Error *string `json:"error,omitempty"`

	// Name of SSL Certificate.
	Name *string `json:"name,omitempty"`
}
