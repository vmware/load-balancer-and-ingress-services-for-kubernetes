// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SnmpV3Configuration snmp v3 configuration
// swagger:model SnmpV3Configuration
type SnmpV3Configuration struct {

	// Engine Id of the Avi Controller SNMP. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EngineID *string `json:"engine_id,omitempty"`

	// SNMP ver 3 user definition. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	User *SnmpV3UserParams `json:"user,omitempty"`
}
