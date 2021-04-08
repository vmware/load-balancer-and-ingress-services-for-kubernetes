package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CertificateAuthority certificate authority
// swagger:model CertificateAuthority
type CertificateAuthority struct {

	//  It is a reference to an object of type SSLKeyAndCertificate.
	CaRef *string `json:"ca_ref,omitempty"`

	// Name of the object.
	Name *string `json:"name,omitempty"`
}
