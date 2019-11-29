package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SecurityPolicy security policy
// swagger:model SecurityPolicy
type SecurityPolicy struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Security policy is used to specify various configuration information used to perform Distributed Denial of Service (DDoS) attacks detection and mitigation. Field introduced in 18.2.1.
	Description *string `json:"description,omitempty"`

	// Attacks utilizing the DNS protocol operations. Field introduced in 18.2.1.
	DNSAttacks *DNSAttacks `json:"dns_attacks,omitempty"`

	// Index of the dns policy to use for the mitigation rules applied to the dns attacks. Field introduced in 18.2.1.
	// Required: true
	DNSPolicyIndex *int32 `json:"dns_policy_index"`

	// The name of the security policy. Field introduced in 18.2.1.
	// Required: true
	Name *string `json:"name"`

	// Index of the network security policy to use for the mitigation rules applied to the attacks. Field introduced in 18.2.1.
	// Required: true
	NetworkSecurityPolicyIndex *int32 `json:"network_security_policy_index"`

	// Mode of dealing with the attacks - perform detection only, or detect and mitigate the attacks. Enum options - DETECTION, MITIGATION. Field introduced in 18.2.1.
	OperMode *string `json:"oper_mode,omitempty"`

	// Attacks utilizing the TCP protocol operations. Field introduced in 18.2.1.
	TCPAttacks TCPAttacks `json:"tcp_attacks,omitempty"`

	// Tenancy of the security policy. It is a reference to an object of type Tenant. Field introduced in 18.2.1.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Attacks utilizing the UDP protocol operations. Field introduced in 18.2.1.
	UDPAttacks UDPAttacks `json:"udp_attacks,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// The UUID of the security policy. Field introduced in 18.2.1.
	UUID *string `json:"uuid,omitempty"`
}
