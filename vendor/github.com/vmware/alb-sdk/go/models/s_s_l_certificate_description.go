package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SSLCertificateDescription s s l certificate description
// swagger:model SSLCertificateDescription
type SSLCertificateDescription struct {

	// common_name of SSLCertificateDescription.
	CommonName *string `json:"common_name,omitempty"`

	// country of SSLCertificateDescription.
	Country *string `json:"country,omitempty"`

	// distinguished_name of SSLCertificateDescription.
	DistinguishedName *string `json:"distinguished_name,omitempty"`

	// email_address of SSLCertificateDescription.
	EmailAddress *string `json:"email_address,omitempty"`

	// locality of SSLCertificateDescription.
	Locality *string `json:"locality,omitempty"`

	// organization of SSLCertificateDescription.
	Organization *string `json:"organization,omitempty"`

	// organization_unit of SSLCertificateDescription.
	OrganizationUnit *string `json:"organization_unit,omitempty"`

	// state of SSLCertificateDescription.
	State *string `json:"state,omitempty"`
}
