package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SnmpTrapServer snmp trap server
// swagger:model SnmpTrapServer
type SnmpTrapServer struct {

	// The community *string to communicate with the trap server.
	Community *string `json:"community,omitempty"`

	// IP Address of the SNMP trap destination.
	// Required: true
	IPAddr *IPAddr `json:"ip_addr"`

	// The UDP port of the trap server. Field introduced in 16.5.4,17.2.5.
	Port *int32 `json:"port,omitempty"`

	// SNMP version 3 configuration. Field introduced in 17.2.3.
	User *SnmpV3UserParams `json:"user,omitempty"`

	// SNMP version support. V2 or V3. Enum options - SNMP_VER2, SNMP_VER3. Field introduced in 17.2.3.
	Version *string `json:"version,omitempty"`
}
