// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ApplicationProfile application profile
// swagger:model ApplicationProfile
type ApplicationProfile struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Specifies app service type for an application. Enum options - APP_SERVICE_TYPE_L7_HORIZON, APP_SERVICE_TYPE_L4_BLAST, APP_SERVICE_TYPE_L4_PCOIP, APP_SERVICE_TYPE_L4_FTP. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AppServiceType *string `json:"app_service_type,omitempty"`

	// Checksum of application profiles. Internally set by cloud connector. Field introduced in 17.2.14, 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudConfigCksum *string `json:"cloud_config_cksum,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// Name of the application profile creator. Field introduced in 17.2.14, 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CreatedBy *string `json:"created_by,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// Specifies various DNS service related controls for virtual service. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DNSServiceProfile *DNSServiceApplicationProfile `json:"dns_service_profile,omitempty"`

	// Specifies various security related controls for virtual service. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DosRlProfile *DosRateLimitProfile `json:"dos_rl_profile,omitempty"`

	// Specifies the HTTP application proxy profile parameters. Allowed in Enterprise edition with any value, Basic, Enterprise with Cloud Services edition.
	HTTPProfile *HTTPApplicationProfile `json:"http_profile,omitempty"`

	// Specifies various L4 SSL service related controls for virtual service. Field introduced in 22.1.2. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	L4SslProfile *L4SSlapplicationProfile `json:"l4_ssl_profile,omitempty"`

	// List of labels to be used for granular RBAC. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	Markers []*RoleFilterMatchLabel `json:"markers,omitempty"`

	// The name of the application profile. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// Specifies if client IP needs to be preserved for backend connection. Not compatible with Connection Multiplexing. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PreserveClientIP *bool `json:"preserve_client_ip,omitempty"`

	// Specifies if we need to preserve client port while preserving client IP for backend connections. Field introduced in 17.2.7. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PreserveClientPort *bool `json:"preserve_client_port,omitempty"`

	// Specifies if destination IP and port needs to be preserved for backend connection. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	PreserveDestIPPort *bool `json:"preserve_dest_ip_port,omitempty"`

	// Specifies various SIP service related controls for virtual service. Field introduced in 17.2.8, 18.1.3, 18.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SipServiceProfile *SipServiceApplicationProfile `json:"sip_service_profile,omitempty"`

	// Specifies the TCP application proxy profile parameters. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TCPAppProfile *TCPApplicationProfile `json:"tcp_app_profile,omitempty"`

	//  It is a reference to an object of type Tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Specifies which application layer proxy is enabled for the virtual service. Enum options - APPLICATION_PROFILE_TYPE_L4, APPLICATION_PROFILE_TYPE_HTTP, APPLICATION_PROFILE_TYPE_SYSLOG, APPLICATION_PROFILE_TYPE_DNS, APPLICATION_PROFILE_TYPE_SSL, APPLICATION_PROFILE_TYPE_SIP. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- APPLICATION_PROFILE_TYPE_L4), Basic edition(Allowed values- APPLICATION_PROFILE_TYPE_L4,APPLICATION_PROFILE_TYPE_HTTP), Enterprise with Cloud Services edition.
	// Required: true
	Type *string `json:"type"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the application profile. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}
