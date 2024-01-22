// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSAttack Dns attack
// swagger:model DnsAttack
type DNSAttack struct {

	// The DNS attack vector. Enum options - DNS_REFLECTION, DNS_NXDOMAIN, DNS_AMPLIFICATION_EGRESS. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	AttackVector *string `json:"attack_vector"`

	// Enable or disable the mitigation of the attack vector. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Enabled *bool `json:"enabled,omitempty"`

	// Time in minutes after which mitigation will be deactivated. Allowed values are 1-4294967295. Special values are 0- blocked for ever. Field introduced in 18.2.1. Unit is MIN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxMitigationAge *uint32 `json:"max_mitigation_age,omitempty"`

	// Mitigation action to perform for this DNS attack vector. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MitigationAction *AttackMitigationAction `json:"mitigation_action,omitempty"`

	// Threshold, in terms of DNS packet per second, for the DNS attack vector. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Threshold uint64 `json:"threshold,omitempty"`
}
