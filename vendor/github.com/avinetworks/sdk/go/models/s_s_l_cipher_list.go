package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SSLCipherList s s l cipher list
// swagger:model SSLCipherList
type SSLCipherList struct {

	// List of ciphers from the client's SSL cipher list that could be identified. The ciphers are represented by their RFC name. Enum options - AVI_TLS_NULL_WITH_NULL_NULL, AVI_TLS_RSA_WITH_NULL_MD5, AVI_TLS_RSA_WITH_NULL_SHA, AVI_TLS_RSA_EXPORT_WITH_RC4_40_MD5, AVI_TLS_RSA_WITH_RC4_128_MD5, AVI_TLS_RSA_WITH_RC4_128_SHA, AVI_TLS_RSA_EXPORT_WITH_RC2_CBC_40_MD5, AVI_TLS_RSA_WITH_IDEA_CBC_SHA, AVI_TLS_RSA_EXPORT_WITH_DES40_CBC_SHA, AVI_TLS_RSA_WITH_DES_CBC_SHA, AVI_TLS_RSA_WITH_3DES_EDE_CBC_SHA, AVI_TLS_DH_DSS_EXPORT_WITH_DES40_CBC_SHA, AVI_TLS_DH_DSS_WITH_DES_CBC_SHA, AVI_TLS_DH_DSS_WITH_3DES_EDE_CBC_SHA, AVI_TLS_DH_RSA_EXPORT_WITH_DES40_CBC_SHA, AVI_TLS_DH_RSA_WITH_DES_CBC_SHA, AVI_TLS_DH_RSA_WITH_3DES_EDE_CBC_SHA, AVI_TLS_DHE_DSS_EXPORT_WITH_DES40_CBC_SHA, AVI_TLS_DHE_DSS_WITH_DES_CBC_SHA, AVI_TLS_DHE_DSS_WITH_3DES_EDE_CBC_SHA.... Field introduced in 18.1.4, 18.2.1.
	IdentifiedCiphers []string `json:"identified_ciphers,omitempty"`

	// List of ciphers from the client's SSL cipher list, that could not be identified. The ciphers are represented by their RFC 2 byte hex value. Field introduced in 18.1.4, 18.2.1.
	UnidentifiedCiphers []string `json:"unidentified_ciphers,omitempty"`
}
