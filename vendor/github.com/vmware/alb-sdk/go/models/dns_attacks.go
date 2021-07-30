// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSAttacks Dns attacks
// swagger:model DnsAttacks
type DNSAttacks struct {

	// Mode of dealing with the attacks - perform detection only, or detect and mitigate the attacks. Field introduced in 18.2.1.
	Attacks []*DNSAttack `json:"attacks,omitempty"`

	// Mode of dealing with the attacks - perform detection only, or detect and mitigate the attacks. Enum options - DETECTION, MITIGATION. Field introduced in 18.2.1.
	OperMode *string `json:"oper_mode,omitempty"`
}
