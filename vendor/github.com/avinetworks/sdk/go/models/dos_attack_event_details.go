package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DosAttackEventDetails dos attack event details
// swagger:model DosAttackEventDetails
type DosAttackEventDetails struct {

	//  Enum options - LAND. SMURF. ICMP_PING_FLOOD. UNKOWN_PROTOCOL. TEARDROP. IP_FRAG_OVERRUN. IP_FRAG_TOOSMALL. IP_FRAG_FULL. IP_FRAG_INCOMPLETE. PORT_SCAN. TCP_NON_SYN_FLOOD_OLD. SYN_FLOOD. BAD_RST_FLOOD. MALFORMED_FLOOD. FAKE_SESSION. ZERO_WINDOW_STRESS. SMALL_WINDOW_STRESS. DOS_HTTP_TIMEOUT. DOS_HTTP_ERROR. DOS_HTTP_ABORT...
	// Required: true
	Attack *string `json:"attack"`

	// Placeholder for description of property attack_count of obj type DosAttackEventDetails field type str  type number
	// Required: true
	AttackCount *float64 `json:"attack_count"`

	// Unique object identifiers of ipgroups.
	IpgroupUuids []string `json:"ipgroup_uuids,omitempty"`

	// Placeholder for description of property meta_data of obj type DosAttackEventDetails field type str  type object
	MetaData []*AttackMetaData `json:"meta_data,omitempty"`

	// src_ips of DosAttackEventDetails.
	SrcIps []string `json:"src_ips,omitempty"`

	// urls of DosAttackEventDetails.
	Urls []string `json:"urls,omitempty"`
}
