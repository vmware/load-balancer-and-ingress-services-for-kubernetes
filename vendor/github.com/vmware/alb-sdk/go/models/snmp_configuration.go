package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SnmpConfiguration snmp configuration
// swagger:model SnmpConfiguration
type SnmpConfiguration struct {

	// Community *string for SNMP v2c.
	Community *string `json:"community,omitempty"`

	// Support for 4096 bytes trap payload. Field introduced in 17.2.13,18.1.4,18.2.1.
	LargeTrapPayload *bool `json:"large_trap_payload,omitempty"`

	// SNMP version 3 configuration. Field introduced in 17.2.3.
	SnmpV3Config *SnmpV3Configuration `json:"snmp_v3_config,omitempty"`

	// Sets the sysContact in system MIB.
	SysContact *string `json:"sys_contact,omitempty"`

	// Sets the sysLocation in system MIB.
	SysLocation *string `json:"sys_location,omitempty"`

	// SNMP version support. V2 or V3. Enum options - SNMP_VER2, SNMP_VER3. Field introduced in 17.2.3.
	Version *string `json:"version,omitempty"`
}
