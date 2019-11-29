package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// PKIprofile p k iprofile
// swagger:model PKIProfile
type PKIprofile struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// List of Certificate Authorities (Root and Intermediate) trusted that is used for certificate validation.
	CaCerts []*SSLCertificate `json:"ca_certs,omitempty"`

	// Creator name.
	CreatedBy *string `json:"created_by,omitempty"`

	// When enabled, Avi will verify via CRL checks that certificates in the trust chain have not been revoked.
	CrlCheck *bool `json:"crl_check,omitempty"`

	// Certificate Revocation Lists.
	Crls []*CRL `json:"crls,omitempty"`

	// When enabled, Avi will not trust Intermediate and Root certs presented by a client.  Instead, only the chain certs configured in the Certificate Authority section will be used to verify trust of the client's cert.
	IgnorePeerChain *bool `json:"ignore_peer_chain,omitempty"`

	// This field describes the object's replication scope. If the field is set to false, then the object is visible within the controller-cluster and its associated service-engines.  If the field is set to true, then the object is replicated across the federation.  . Field introduced in 17.1.3.
	IsFederated *bool `json:"is_federated,omitempty"`

	// Name of the PKI Profile.
	// Required: true
	Name *string `json:"name"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`

	// When enabled, Avi will only validate the revocation status of the leaf certificate using CRL. To enable validation for the entire chain, disable this option and provide all the relevant CRLs.
	ValidateOnlyLeafCrl *bool `json:"validate_only_leaf_crl,omitempty"`
}
