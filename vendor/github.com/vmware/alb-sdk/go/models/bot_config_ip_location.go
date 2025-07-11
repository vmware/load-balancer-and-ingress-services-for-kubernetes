// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// BotConfigIPLocation bot config IP location
// swagger:model BotConfigIPLocation
type BotConfigIPLocation struct {

	// If this is enabled, IP location information is used to determine if a client is a known search engine bot, comes from the cloud, etc. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Enabled *bool `json:"enabled,omitempty"`

	// The UUID of the Geo-IP database to use. It is a reference to an object of type GeoDB. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IPLocationDbRef *string `json:"ip_location_db_ref,omitempty"`

	// The system-defined cloud providers. It is a reference to an object of type StringGroup. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SystemCloudProvidersRef *string `json:"system_cloud_providers_ref,omitempty"`

	// The system-defined search engines. It is a reference to an object of type StringGroup. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SystemSearchEnginesRef *string `json:"system_search_engines_ref,omitempty"`
}
