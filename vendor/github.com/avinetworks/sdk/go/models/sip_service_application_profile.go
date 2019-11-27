package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SipServiceApplicationProfile sip service application profile
// swagger:model SipServiceApplicationProfile
type SipServiceApplicationProfile struct {

	// SIP transaction timeout in seconds. Allowed values are 2-512. Field introduced in 17.2.8, 18.1.3, 18.2.1.
	TransactionTimeout *int32 `json:"transaction_timeout,omitempty"`
}
