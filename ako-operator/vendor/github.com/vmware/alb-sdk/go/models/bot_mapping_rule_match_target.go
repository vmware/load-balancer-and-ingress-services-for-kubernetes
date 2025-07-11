// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// BotMappingRuleMatchTarget bot mapping rule match target
// swagger:model BotMappingRuleMatchTarget
type BotMappingRuleMatchTarget struct {

	// How to match the BotClientClass. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ClassMatcher *BotClassMatcher `json:"class_matcher,omitempty"`

	// Configure client ip addresses. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ClientIP *IPAddrMatch `json:"client_ip,omitempty"`

	// The component for which this mapping is used. Enum options - BOT_DECIDER_CONSOLIDATION, BOT_DECIDER_USER_AGENT, BOT_DECIDER_IP_REPUTATION, BOT_DECIDER_IP_NETWORK_LOCATION, BOT_DECIDER_CLIENT_BEHAVIOR. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ComponentMatcher *string `json:"component_matcher,omitempty"`

	// Configure HTTP header(s). All configured headers must match. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Hdrs []*HdrMatch `json:"hdrs,omitempty"`

	// Configure the host header. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	HostHdr *HostHdrMatch `json:"host_hdr,omitempty"`

	// The list of bot identifier names and how they're matched. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IdentifierMatcher *StringMatch `json:"identifier_matcher,omitempty"`

	// Configure HTTP methods. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Method *MethodMatch `json:"method,omitempty"`

	// Configure request paths. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Path *PathMatch `json:"path,omitempty"`

	// How to match the BotClientType. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TypeMatcher *BotTypeMatcher `json:"type_matcher,omitempty"`
}
