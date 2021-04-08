package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// BotConfigIPReputation bot config IP reputation
// swagger:model BotConfigIPReputation
type BotConfigIPReputation struct {

	// Whether IP reputation-based Bot detection is enabled. Field introduced in 21.1.1.
	Enabled *bool `json:"enabled,omitempty"`

	// The UUID of the IP reputation DB to use. It is a reference to an object of type IPReputationDB. Field introduced in 21.1.1.
	IPReputationDbRef *string `json:"ip_reputation_db_ref,omitempty"`

	// Map every IPReputationType to a bot type (can be unknown). Field introduced in 21.1.1.
	IPReputationMappings []*IPReputationTypeMapping `json:"ip_reputation_mappings,omitempty"`
}
