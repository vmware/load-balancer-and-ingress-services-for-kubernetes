package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSAttack Dns attack
// swagger:model DnsAttack
type DNSAttack struct {

	// The DNS attack vector. Enum options - DNS_REFLECTION, DNS_NXDOMAIN, DNS_AMPLIFICATION_EGRESS. Field introduced in 18.2.1.
	// Required: true
	AttackVector *string `json:"attack_vector"`

	// Enable or disable the mitigation of the attack vector. Field introduced in 18.2.1.
	Enabled *bool `json:"enabled,omitempty"`

	// Time in minutes after which mitigation will be deactivated. Allowed values are 1-4294967295. Special values are 0- 'blocked for ever'. Field introduced in 18.2.1.
	MaxMitigationAge *int32 `json:"max_mitigation_age,omitempty"`

	// Mitigation action to perform for this DNS attack vector. Field introduced in 18.2.1.
	MitigationAction *AttackMitigationAction `json:"mitigation_action,omitempty"`
}
