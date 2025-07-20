// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// IPAMDNSOCIprofile ipam Dns o c iprofile
// swagger:model IpamDnsOCIProfile
type IPAMDNSOCIprofile struct {

	// Credentials to access oracle cloud. It is a reference to an object of type CloudConnectorUser. Field introduced in 18.2.1,18.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudCredentialsRef *string `json:"cloud_credentials_ref,omitempty"`

	// Region in which Oracle cloud resource resides. Field introduced in 18.2.1,18.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Region *string `json:"region,omitempty"`

	// Oracle Cloud Id for tenant aka root compartment. Field introduced in 18.2.1,18.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Tenancy *string `json:"tenancy,omitempty"`

	// Oracle cloud compartment id in which VCN resides. Field introduced in 18.2.1,18.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcnCompartmentID *string `json:"vcn_compartment_id,omitempty"`

	// Virtual Cloud network id where virtual ip will belong. Field introduced in 18.2.1,18.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcnID *string `json:"vcn_id,omitempty"`
}
