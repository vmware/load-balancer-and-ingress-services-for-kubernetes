package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// L4RuleMatchTarget l4 rule match target
// swagger:model L4RuleMatchTarget
type L4RuleMatchTarget struct {

	// IP addresses to match against client IP. Field introduced in 17.2.7. Allowed in Basic edition, Essentials edition, Enterprise edition.
	ClientIP *IPAddrMatch `json:"client_ip,omitempty"`

	// Port number to match against Virtual Service listner port. Field introduced in 17.2.7.
	Port *L4RulePortMatch `json:"port,omitempty"`

	// TCP/UDP/ICMP protocol to match against transport protocol. Field introduced in 17.2.7. Allowed in Basic edition, Essentials edition, Enterprise edition.
	Protocol *L4RuleProtocolMatch `json:"protocol,omitempty"`
}
