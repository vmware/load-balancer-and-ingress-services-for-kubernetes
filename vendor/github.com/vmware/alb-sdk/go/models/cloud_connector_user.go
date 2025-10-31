// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CloudConnectorUser cloud connector user
// swagger:model CloudConnectorUser
type CloudConnectorUser struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	//  Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AzureServiceprincipal *AzureServicePrincipalCredentials `json:"azure_serviceprincipal,omitempty"`

	//  Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AzureUserpass *AzureUserPassCredentials `json:"azure_userpass,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// Credentials for Google Cloud Platform. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	GcpCredentials *GCPCredentials `json:"gcp_credentials,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// Credentials to talk to NSX-T manager. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Basic, Enterprise with Cloud Services edition.
	NsxtCredentials *NsxtCredentials `json:"nsxt_credentials,omitempty"`

	// Credentials for Oracle Cloud Infrastructure. Field introduced in 18.2.1,18.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	OciCredentials *OCICredentials `json:"oci_credentials,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Password *string `json:"password,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PrivateKey *string `json:"private_key,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PublicKey *string `json:"public_key,omitempty"`

	//  It is a reference to an object of type Tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Credentials for Tencent Cloud. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TencentCredentials *TencentCredentials `json:"tencent_credentials,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`

	// Credentials to talk to VCenter. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterCredentials *VCenterCredentials `json:"vcenter_credentials,omitempty"`
}
