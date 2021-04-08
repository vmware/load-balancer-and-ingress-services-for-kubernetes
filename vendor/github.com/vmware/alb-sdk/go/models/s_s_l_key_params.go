package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SSLKeyParams s s l key params
// swagger:model SSLKeyParams
type SSLKeyParams struct {

	//  Enum options - SSL_KEY_ALGORITHM_RSA, SSL_KEY_ALGORITHM_EC.
	// Required: true
	Algorithm *string `json:"algorithm"`

	// Placeholder for description of property ec_params of obj type SSLKeyParams field type str  type object
	EcParams *SSLKeyECParams `json:"ec_params,omitempty"`

	// Placeholder for description of property rsa_params of obj type SSLKeyParams field type str  type object
	RsaParams *SSLKeyRSAParams `json:"rsa_params,omitempty"`
}
