package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// BotDetectionPolicy bot detection policy
// swagger:model BotDetectionPolicy
type BotDetectionPolicy struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Allow the user to skip BotManagement for selected requests. Field introduced in 21.1.1.
	AllowList *BotAllowList `json:"allow_list,omitempty"`

	// System- and User-defined rules for classification. Field introduced in 21.1.1.
	BotMappingUuids []string `json:"bot_mapping_uuids,omitempty"`

	// The installation provides an updated ruleset for consolidating the results of different decider phases. It is a reference to an object of type BotConfigConsolidator. Field introduced in 21.1.1.
	ConsolidatorRef *string `json:"consolidator_ref,omitempty"`

	// Human-readable description of this Bot Detection Policy. Field introduced in 21.1.1.
	Description *string `json:"description,omitempty"`

	// The IP location configuration used in this policy. Field introduced in 21.1.1.
	// Required: true
	IPLocationDetector *BotConfigIPLocation `json:"ip_location_detector"`

	// The IP reputation configuration used in this policy. Field introduced in 21.1.1.
	// Required: true
	IPReputationDetector *BotConfigIPReputation `json:"ip_reputation_detector"`

	// The name of this Bot Detection Policy. Field introduced in 21.1.1.
	// Required: true
	Name *string `json:"name"`

	// The unique identifier of the tenant to which this policy belongs. It is a reference to an object of type Tenant. Field introduced in 21.1.1.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// The User-Agent configuration used in this policy. Field introduced in 21.1.1.
	// Required: true
	UserAgentDetector *BotConfigUserAgent `json:"user_agent_detector"`

	// A unique identifier to this Bot Detection Policy. Field introduced in 21.1.1.
	UUID *string `json:"uuid,omitempty"`
}
