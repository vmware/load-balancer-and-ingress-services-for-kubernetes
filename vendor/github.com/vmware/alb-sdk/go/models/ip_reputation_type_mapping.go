package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IPReputationTypeMapping IP reputation type mapping
// swagger:model IPReputationTypeMapping
type IPReputationTypeMapping struct {

	// The Bot Identification to which the IP reputation type is mapped. Field introduced in 21.1.1.
	// Required: true
	BotIdentification *BotIdentification `json:"bot_identification"`

	// The type of IP reputation that is mapped to a Bot Identification. Enum options - IP_REPUTATION_TYPE_SPAM_SOURCE, IP_REPUTATION_TYPE_WINDOWS_EXPLOITS, IP_REPUTATION_TYPE_WEB_ATTACKS, IP_REPUTATION_TYPE_BOTNETS, IP_REPUTATION_TYPE_SCANNERS, IP_REPUTATION_TYPE_DOS, IP_REPUTATION_TYPE_REPUTATION, IP_REPUTATION_TYPE_PHISHING, IP_REPUTATION_TYPE_PROXY, IP_REPUTATION_TYPE_CLOUD, IP_REPUTATION_TYPE_MOBILE_THREATS, IP_REPUTATION_TYPE_TOR, IP_REPUTATION_TYPE_ALL. Field introduced in 21.1.1.
	// Required: true
	IPReputationType *string `json:"ip_reputation_type"`
}
