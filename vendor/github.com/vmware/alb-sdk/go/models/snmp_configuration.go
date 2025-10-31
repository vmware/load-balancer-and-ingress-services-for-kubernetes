// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SnmpConfiguration snmp configuration
// swagger:model SnmpConfiguration
type SnmpConfiguration struct {

	// Community *string for SNMP v2c. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Community *string `json:"community,omitempty"`

	// Support for 4096 bytes trap payload. Field introduced in 17.2.13,18.1.4,18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LargeTrapPayload *bool `json:"large_trap_payload,omitempty"`

	// SNMP version 3 configuration. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SnmpV3Config *SnmpV3Configuration `json:"snmp_v3_config,omitempty"`

	// Sets the sysContact in system MIB. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SysContact *string `json:"sys_contact,omitempty"`

	// Sets the sysLocation in system MIB. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SysLocation *string `json:"sys_location,omitempty"`

	// SNMP version support. V2 or V3. Enum options - SNMP_VER2, SNMP_VER3. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Version *string `json:"version,omitempty"`
}
