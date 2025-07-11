// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SSLKeyAndCertificate s s l key and certificate
// swagger:model SSLKeyAndCertificate
type SSLKeyAndCertificate struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// CA certificates in certificate chain. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CaCerts []*CertificateAuthority `json:"ca_certs,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Certificate *SSLCertificate `json:"certificate"`

	// States if the certificate is base64 encoded. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CertificateBase64 *bool `json:"certificate_base64,omitempty"`

	//  It is a reference to an object of type CertificateManagementProfile. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CertificateManagementProfileRef *string `json:"certificate_management_profile_ref,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// Creator name. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CreatedBy *string `json:"created_by,omitempty"`

	// Dynamic parameters needed for certificate management profile. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DynamicParams []*CustomParams `json:"dynamic_params,omitempty"`

	// Enables OCSP Stapling. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	EnableOcspStapling *bool `json:"enable_ocsp_stapling,omitempty"`

	// Encrypted private key corresponding to the private key (e.g. those generated by an HSM such as Thales nShield). Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnckeyBase64 *string `json:"enckey_base64,omitempty"`

	// Name of the encrypted private key (e.g. those generated by an HSM such as Thales nShield). Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnckeyName *string `json:"enckey_name,omitempty"`

	// Format of the Key/Certificate file. Enum options - SSL_PEM, SSL_PKCS12. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Format *string `json:"format,omitempty"`

	//  It is a reference to an object of type HardwareSecurityModuleGroup. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	HardwaresecuritymodulegroupRef *string `json:"hardwaresecuritymodulegroup_ref,omitempty"`

	// Flag to enable Private key import to HSM while importing the certificate. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ImportKeyToHsm *bool `json:"import_key_to_hsm,omitempty"`

	// It Specifies whether the object has to be replicated to the GSLB followers. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IsFederated *bool `json:"is_federated,omitempty"`

	// Private key. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Key *string `json:"key,omitempty"`

	// States if the private key is base64 encoded. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	KeyBase64 *bool `json:"key_base64,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	KeyParams *SSLKeyParams `json:"key_params,omitempty"`

	// Passphrase used to encrypt the private key. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	KeyPassphrase *string `json:"key_passphrase,omitempty"`

	// List of labels to be used for granular RBAC. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	Markers []*RoleFilterMatchLabel `json:"markers,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// Configuration related to OCSP. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	OcspConfig *OCSPConfig `json:"ocsp_config,omitempty"`

	// Error reported during OCSP status query. Enum options - OCSP_ERR_CERTSTATUS_GOOD, OCSP_ERR_CERTSTATUS_REVOKED, OCSP_ERR_CERTSTATUS_UNKNOWN, OCSP_ERR_CERTSTATUS_SERVERFAIL_ERR, OCSP_ERR_CERTSTATUS_JOBDB, OCSP_ERR_CERTSTATUS_DISABLED, OCSP_ERR_CERTSTATUS_GETCERT, OCSP_ERR_CERTSTATUS_NONVSCERT, OCSP_ERR_CERTSTATUS_SELFSIGNED, OCSP_ERR_CERTSTATUS_CERTFINISH, OCSP_ERR_CERTSTATUS_CACERT, OCSP_ERR_CERTSTATUS_REQUEST, OCSP_ERR_CERTSTATUS_ISSUER_REVOKED, OCSP_ERR_CERTSTATUS_PARSE_CERT, OCSP_ERR_CERTSTATUS_HTTP_REQ, OCSP_ERR_CERTSTATUS_URL_LIST, OCSP_ERR_CERTSTATUS_HTTP_SEND, OCSP_ERR_CERTSTATUS_HTTP_RECV, OCSP_ERR_CERTSTATUS_HTTP_RESP. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- OCSP_ERR_CERTSTATUS_DISABLED), Basic edition(Allowed values- OCSP_ERR_CERTSTATUS_DISABLED), Enterprise with Cloud Services edition.
	// Read Only: true
	OcspErrorStatus *string `json:"ocsp_error_status,omitempty"`

	// This is an Internal field to store the OCSP Responder URLs contained in the certificate. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Read Only: true
	OcspResponderURLListFromCerts []string `json:"ocsp_responder_url_list_from_certs,omitempty"`

	// Information related to OCSP response. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Read Only: true
	OcspResponseInfo *OCSPResponseInfo `json:"ocsp_response_info,omitempty"`

	//  Enum options - SSL_CERTIFICATE_FINISHED, SSL_CERTIFICATE_PENDING. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Status *string `json:"status,omitempty"`

	//  It is a reference to an object of type Tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	//  Enum options - SSL_CERTIFICATE_TYPE_VIRTUALSERVICE, SSL_CERTIFICATE_TYPE_SYSTEM, SSL_CERTIFICATE_TYPE_CA. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Type *string `json:"type,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}
