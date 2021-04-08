package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// BotConfigIPLocation bot config IP location
// swagger:model BotConfigIPLocation
type BotConfigIPLocation struct {

	// If this is enabled, IP location information is used to determine if a client is a known search engine bot, comes from the cloud, etc. Field introduced in 21.1.1.
	Enabled *bool `json:"enabled,omitempty"`

	// The UUID of the Geo-IP databse to use. It is a reference to an object of type GeoDB. Field introduced in 21.1.1.
	IPLocationDbRef *string `json:"ip_location_db_ref,omitempty"`
}
