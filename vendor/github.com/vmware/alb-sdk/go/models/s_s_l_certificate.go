package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SSLCertificate s s l certificate
// swagger:model SSLCertificate
type SSLCertificate struct {

	// certificate of SSLCertificate.
	Certificate *string `json:"certificate,omitempty"`

	// certificate_signing_request of SSLCertificate.
	CertificateSigningRequest *string `json:"certificate_signing_request,omitempty"`

	// Placeholder for description of property chain_verified of obj type SSLCertificate field type str  type boolean
	ChainVerified *bool `json:"chain_verified,omitempty"`

	// Number of days_until_expire.
	DaysUntilExpire *int32 `json:"days_until_expire,omitempty"`

	//  Enum options - SSL_CERTIFICATE_GOOD, SSL_CERTIFICATE_EXPIRY_WARNING, SSL_CERTIFICATE_EXPIRED.
	ExpiryStatus *string `json:"expiry_status,omitempty"`

	// fingerprint of SSLCertificate.
	Fingerprint *string `json:"fingerprint,omitempty"`

	// Placeholder for description of property issuer of obj type SSLCertificate field type str  type object
	Issuer *SSLCertificateDescription `json:"issuer,omitempty"`

	// Placeholder for description of property key_params of obj type SSLCertificate field type str  type object
	KeyParams *SSLKeyParams `json:"key_params,omitempty"`

	// not_after of SSLCertificate.
	NotAfter *string `json:"not_after,omitempty"`

	// not_before of SSLCertificate.
	NotBefore *string `json:"not_before,omitempty"`

	// public_key of SSLCertificate.
	PublicKey *string `json:"public_key,omitempty"`

	// Placeholder for description of property self_signed of obj type SSLCertificate field type str  type boolean
	SelfSigned *bool `json:"self_signed,omitempty"`

	// serial_number of SSLCertificate.
	SerialNumber *string `json:"serial_number,omitempty"`

	// signature of SSLCertificate.
	Signature *string `json:"signature,omitempty"`

	// signature_algorithm of SSLCertificate.
	SignatureAlgorithm *string `json:"signature_algorithm,omitempty"`

	// Placeholder for description of property subject of obj type SSLCertificate field type str  type object
	Subject *SSLCertificateDescription `json:"subject,omitempty"`

	// subjectAltName that provides additional subject identities.
	SubjectAltNames []string `json:"subject_alt_names,omitempty"`

	// text of SSLCertificate.
	Text *string `json:"text,omitempty"`

	// version of SSLCertificate.
	Version *string `json:"version,omitempty"`
}
