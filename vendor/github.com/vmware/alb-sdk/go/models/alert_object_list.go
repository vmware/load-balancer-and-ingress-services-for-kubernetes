package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AlertObjectList alert object list
// swagger:model AlertObjectList
type AlertObjectList struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	//  Enum options - VIRTUALSERVICE. POOL. HEALTHMONITOR. NETWORKPROFILE. APPLICATIONPROFILE. HTTPPOLICYSET. DNSPOLICY. SECURITYPOLICY. IPADDRGROUP. STRINGGROUP. SSLPROFILE. SSLKEYANDCERTIFICATE. NETWORKSECURITYPOLICY. APPLICATIONPERSISTENCEPROFILE. ANALYTICSPROFILE. VSDATASCRIPTSET. TENANT. PKIPROFILE. AUTHPROFILE. CLOUD...
	Objects []string `json:"objects,omitempty"`

	//  Enum options - CONN_LOGS, APP_LOGS, EVENT_LOGS, METRICS.
	// Required: true
	Source *string `json:"source"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
