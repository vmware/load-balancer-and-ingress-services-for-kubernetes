package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SnmpTrapProfile snmp trap profile
// swagger:model SnmpTrapProfile
type SnmpTrapProfile struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// A user-friendly name of the SNMP trap configuration.
	// Required: true
	Name *string `json:"name"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// The IP address or hostname of the SNMP trap destination server.
	TrapServers []*SnmpTrapServer `json:"trap_servers,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the SNMP trap profile object.
	UUID *string `json:"uuid,omitempty"`
}
