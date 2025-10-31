// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SnmpV3UserParams snmp v3 user params
// swagger:model SnmpV3UserParams
type SnmpV3UserParams struct {

	// SNMP V3 authentication passphrase. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AuthPassphrase *string `json:"auth_passphrase,omitempty"`

	// SNMP V3 user authentication type. Enum options - SNMP_V3_AUTH_MD5, SNMP_V3_AUTH_SHA, SNMP_V3_AUTH_SHA_224, SNMP_V3_AUTH_SHA_256, SNMP_V3_AUTH_SHA_384, SNMP_V3_AUTH_SHA_512. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AuthType *string `json:"auth_type,omitempty"`

	// SNMP V3 privacy passphrase. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PrivPassphrase *string `json:"priv_passphrase,omitempty"`

	// SNMP V3 privacy setting. Enum options - SNMP_V3_PRIV_DES, SNMP_V3_PRIV_AES. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PrivType *string `json:"priv_type,omitempty"`

	// SNMP username to be used by SNMP clients for performing SNMP walk. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Username *string `json:"username,omitempty"`
}
