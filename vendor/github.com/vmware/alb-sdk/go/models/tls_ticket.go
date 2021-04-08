package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// TLSTicket TLS ticket
// swagger:model TLSTicket
type TLSTicket struct {

	// aes_key of TLSTicket.
	// Required: true
	AesKey *string `json:"aes_key"`

	// hmac_key of TLSTicket.
	// Required: true
	HmacKey *string `json:"hmac_key"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`
}
