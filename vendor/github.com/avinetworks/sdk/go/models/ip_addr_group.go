package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IPAddrGroup Ip addr group
// swagger:model IpAddrGroup
type IPAddrGroup struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Configure IP address(es).
	Addrs []*IPAddr `json:"addrs,omitempty"`

	// Populate IP addresses from members of this Cisco APIC EPG.
	ApicEpgName *string `json:"apic_epg_name,omitempty"`

	// Populate the IP address ranges from the geo database for this country.
	CountryCodes []string `json:"country_codes,omitempty"`

	// User defined description for the object.
	Description *string `json:"description,omitempty"`

	// Configure (IP address, port) tuple(s).
	IPPorts []*IPAddrPort `json:"ip_ports,omitempty"`

	// Populate IP addresses from tasks of this Marathon app.
	MarathonAppName *string `json:"marathon_app_name,omitempty"`

	// Task port associated with marathon service port. If Marathon app has multiple service ports, this is required. Else, the first task port is used.
	MarathonServicePort *int32 `json:"marathon_service_port,omitempty"`

	// Name of the IP address group.
	// Required: true
	Name *string `json:"name"`

	// Configure IP address prefix(es).
	Prefixes []*IPAddrPrefix `json:"prefixes,omitempty"`

	// Configure IP address range(s).
	Ranges []*IPAddrRange `json:"ranges,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the IP address group.
	UUID *string `json:"uuid,omitempty"`
}
