package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ApplicationPersistenceProfile application persistence profile
// swagger:model ApplicationPersistenceProfile
type ApplicationPersistenceProfile struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Specifies the Application Cookie Persistence profile parameters.
	AppCookiePersistenceProfile *AppCookiePersistenceProfile `json:"app_cookie_persistence_profile,omitempty"`

	// User defined description for the object.
	Description *string `json:"description,omitempty"`

	// Specifies the custom HTTP Header Persistence profile parameters.
	HdrPersistenceProfile *HdrPersistenceProfile `json:"hdr_persistence_profile,omitempty"`

	// Specifies the HTTP Cookie Persistence profile parameters.
	HTTPCookiePersistenceProfile *HTTPCookiePersistenceProfile `json:"http_cookie_persistence_profile,omitempty"`

	// Specifies the Client IP Persistence profile parameters.
	IPPersistenceProfile *IPPersistenceProfile `json:"ip_persistence_profile,omitempty"`

	// This field describes the object's replication scope. If the field is set to false, then the object is visible within the controller-cluster and its associated service-engines.  If the field is set to true, then the object is replicated across the federation.  . Field introduced in 17.1.3.
	IsFederated *bool `json:"is_federated,omitempty"`

	// A user-friendly name for the persistence profile.
	// Required: true
	Name *string `json:"name"`

	// Method used to persist clients to the same server for a duration of time or a session. Enum options - PERSISTENCE_TYPE_CLIENT_IP_ADDRESS, PERSISTENCE_TYPE_HTTP_COOKIE, PERSISTENCE_TYPE_TLS, PERSISTENCE_TYPE_CLIENT_IPV6_ADDRESS, PERSISTENCE_TYPE_CUSTOM_HTTP_HEADER, PERSISTENCE_TYPE_APP_COOKIE, PERSISTENCE_TYPE_GSLB_SITE.
	// Required: true
	PersistenceType *string `json:"persistence_type"`

	// Specifies behavior when a persistent server has been marked down by a health monitor. Enum options - HM_DOWN_PICK_NEW_SERVER, HM_DOWN_ABORT_CONNECTION, HM_DOWN_CONTINUE_PERSISTENT_SERVER.
	ServerHmDownRecovery *string `json:"server_hm_down_recovery,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the persistence profile.
	UUID *string `json:"uuid,omitempty"`
}
