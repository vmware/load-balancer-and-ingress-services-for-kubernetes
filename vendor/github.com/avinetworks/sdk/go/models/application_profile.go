package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ApplicationProfile application profile
// swagger:model ApplicationProfile
type ApplicationProfile struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Checksum of application profiles. Internally set by cloud connector. Field introduced in 17.2.14, 18.1.5, 18.2.1.
	CloudConfigCksum *string `json:"cloud_config_cksum,omitempty"`

	// Name of the application profile creator. Field introduced in 17.2.14, 18.1.5, 18.2.1.
	CreatedBy *string `json:"created_by,omitempty"`

	// User defined description for the object.
	Description *string `json:"description,omitempty"`

	// Specifies various DNS service related controls for virtual service.
	DNSServiceProfile *DNSServiceApplicationProfile `json:"dns_service_profile,omitempty"`

	// Specifies various security related controls for virtual service.
	DosRlProfile *DosRateLimitProfile `json:"dos_rl_profile,omitempty"`

	// Specifies the HTTP application proxy profile parameters.
	HTTPProfile *HTTPApplicationProfile `json:"http_profile,omitempty"`

	// The name of the application profile.
	// Required: true
	Name *string `json:"name"`

	// Specifies if client IP needs to be preserved for backend connection. Not compatible with Connection Multiplexing.
	PreserveClientIP *bool `json:"preserve_client_ip,omitempty"`

	// Specifies if we need to preserve client port while preserving client IP for backend connections. Field introduced in 17.2.7.
	PreserveClientPort *bool `json:"preserve_client_port,omitempty"`

	// Specifies various SIP service related controls for virtual service. Field introduced in 17.2.8, 18.1.3, 18.2.1.
	SipServiceProfile *SipServiceApplicationProfile `json:"sip_service_profile,omitempty"`

	// Specifies the TCP application proxy profile parameters.
	TCPAppProfile *TCPApplicationProfile `json:"tcp_app_profile,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Specifies which application layer proxy is enabled for the virtual service. Enum options - APPLICATION_PROFILE_TYPE_L4, APPLICATION_PROFILE_TYPE_HTTP, APPLICATION_PROFILE_TYPE_SYSLOG, APPLICATION_PROFILE_TYPE_DNS, APPLICATION_PROFILE_TYPE_SSL, APPLICATION_PROFILE_TYPE_SIP.
	// Required: true
	Type *string `json:"type"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the application profile.
	UUID *string `json:"uuid,omitempty"`
}
