package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IPReputationTypeMatch IP reputation type match
// swagger:model IPReputationTypeMatch
type IPReputationTypeMatch struct {

	// Match criteria. Enum options - IS_IN, IS_NOT_IN. Field introduced in 20.1.1.
	// Required: true
	MatchOperation *string `json:"match_operation"`

	// IP reputation type. Enum options - IP_REPUTATION_TYPE_SPAM_SOURCE, IP_REPUTATION_TYPE_WINDOWS_EXPLOITS, IP_REPUTATION_TYPE_WEB_ATTACKS, IP_REPUTATION_TYPE_BOTNETS, IP_REPUTATION_TYPE_SCANNERS, IP_REPUTATION_TYPE_DOS, IP_REPUTATION_TYPE_REPUTATION, IP_REPUTATION_TYPE_PHISHING, IP_REPUTATION_TYPE_PROXY, IP_REPUTATION_TYPE_CLOUD, IP_REPUTATION_TYPE_MOBILE_THREATS, IP_REPUTATION_TYPE_TOR, IP_REPUTATION_TYPE_ALL. Field introduced in 20.1.1. Minimum of 1 items required.
	ReputationTypes []string `json:"reputation_types,omitempty"`
}
