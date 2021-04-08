package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IPAMDNSOCIprofile ipam Dns o c iprofile
// swagger:model IpamDnsOCIProfile
type IPAMDNSOCIprofile struct {

	// Credentials to access oracle cloud. It is a reference to an object of type CloudConnectorUser. Field introduced in 18.2.1,18.1.3.
	CloudCredentialsRef *string `json:"cloud_credentials_ref,omitempty"`

	// Region in which Oracle cloud resource resides. Field introduced in 18.2.1,18.1.3.
	Region *string `json:"region,omitempty"`

	// Oracle Cloud Id for tenant aka root compartment. Field introduced in 18.2.1,18.1.3.
	Tenancy *string `json:"tenancy,omitempty"`

	// Oracle cloud compartment id in which VCN resides. Field introduced in 18.2.1,18.1.3.
	VcnCompartmentID *string `json:"vcn_compartment_id,omitempty"`

	// Virtual Cloud network id where virtual ip will belong. Field introduced in 18.2.1,18.1.3.
	VcnID *string `json:"vcn_id,omitempty"`
}
