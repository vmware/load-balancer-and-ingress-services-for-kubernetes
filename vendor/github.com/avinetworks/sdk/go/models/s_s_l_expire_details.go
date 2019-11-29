package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SSLExpireDetails s s l expire details
// swagger:model SSLExpireDetails
type SSLExpireDetails struct {

	// Number of days until certificate is expired.
	DaysLeft *int32 `json:"days_left,omitempty"`

	// Name of SSL Certificate.
	Name *string `json:"name,omitempty"`
}
