package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SnmpV3UserParams snmp v3 user params
// swagger:model SnmpV3UserParams
type SnmpV3UserParams struct {

	// SNMP V3 authentication passphrase. Field introduced in 17.2.3.
	AuthPassphrase *string `json:"auth_passphrase,omitempty"`

	// SNMP V3 user authentication type. Enum options - SNMP_V3_AUTH_MD5, SNMP_V3_AUTH_SHA. Field introduced in 17.2.3.
	AuthType *string `json:"auth_type,omitempty"`

	// SNMP V3 privacy passphrase. Field introduced in 17.2.3.
	PrivPassphrase *string `json:"priv_passphrase,omitempty"`

	// SNMP V3 privacy setting. Enum options - SNMP_V3_PRIV_DES, SNMP_V3_PRIV_AES. Field introduced in 17.2.3.
	PrivType *string `json:"priv_type,omitempty"`

	// SNMP username to be used by SNMP clients for performing SNMP walk. Field introduced in 17.2.3.
	Username *string `json:"username,omitempty"`
}
