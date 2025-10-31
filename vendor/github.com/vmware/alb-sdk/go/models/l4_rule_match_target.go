// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// L4RuleMatchTarget l4 rule match target
// swagger:model L4RuleMatchTarget
type L4RuleMatchTarget struct {

	// IP addresses to match against client IP. Field introduced in 17.2.7. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ClientIP *IPAddrMatch `json:"client_ip,omitempty"`

	// Port number to match against Virtual Service listner port. Field introduced in 17.2.7. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Port *L4RulePortMatch `json:"port,omitempty"`

	// TCP/UDP/ICMP protocol to match against transport protocol. Field introduced in 17.2.7. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Protocol *L4RuleProtocolMatch `json:"protocol,omitempty"`
}
