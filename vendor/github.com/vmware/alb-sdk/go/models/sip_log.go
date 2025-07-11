// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SipLog sip log
// swagger:model SipLog
type SipLog struct {

	// Server connection protocol type. Enum options - PROTOCOL_ICMP, PROTOCOL_TCP, PROTOCOL_UDP, PROTOCOL_SCTP. Field introduced in 17.2.12, 18.1.3, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerProtocol *string `json:"server_protocol,omitempty"`

	// SIP CallId header. Field introduced in 17.2.12, 18.1.3, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SipCallidHdr *string `json:"sip_callid_hdr,omitempty"`

	// Client's SIP Contact header. Field introduced in 17.2.12, 18.1.3, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SipContactHdr *string `json:"sip_contact_hdr,omitempty"`

	// SIP From header. Field introduced in 17.2.12, 18.1.3, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SipFromHdr *string `json:"sip_from_hdr,omitempty"`

	// SIP Messages. Field introduced in 17.2.12, 18.1.3, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SipMessages []*SipMessage `json:"sip_messages,omitempty"`

	// SIP To header. Field introduced in 17.2.12, 18.1.3, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SipToHdr *string `json:"sip_to_hdr,omitempty"`
}
