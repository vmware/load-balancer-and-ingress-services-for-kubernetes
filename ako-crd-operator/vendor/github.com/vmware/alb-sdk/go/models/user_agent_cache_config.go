// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// UserAgentCacheConfig user agent cache config
// swagger:model UserAgentCacheConfig
type UserAgentCacheConfig struct {

	// How many unknown User-Agents to batch up before querying Controller - unless max_wait_time is reached first. Allowed values are 1-500. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	BatchSize *uint32 `json:"batch_size,omitempty"`

	// The number of User-Agent entries to cache on the Controller. Allowed values are 500-10000000. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ControllerCacheSize *uint32 `json:"controller_cache_size,omitempty"`

	// Time interval in seconds after which an existing entry is refreshed from upstream if it has been accessed during max_last_hit_time. Allowed values are 60-604800. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MaxAge *uint32 `json:"max_age,omitempty"`

	// Time interval in seconds backwards from now during which an existing entry must have been hit for refresh from upstream. Entries that have last been accessed further in the past than max_last_hit time are not included in upstream refresh requests even if they are older than 'max_age'. Allowed values are 60-604800. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MaxLastHitTime *uint32 `json:"max_last_hit_time,omitempty"`

	// How often at most to query controller for a given User-Agent. Allowed values are 2-100. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MaxUpstreamQueries *uint32 `json:"max_upstream_queries,omitempty"`

	// The time interval in seconds after which to make a request to the Controller, even if the 'batch_size' hasn't been reached yet. Allowed values are 20-100000. Field introduced in 21.1.1. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MaxWaitTime *uint32 `json:"max_wait_time,omitempty"`

	// How many BotUACacheResult elements to include in an upstream update message. Allowed values are 1-10000. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumEntriesUpstreamUpdate *uint32 `json:"num_entries_upstream_update,omitempty"`

	// How much space to reserve in percent for known bad bots. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PercentReservedForBadBots *uint32 `json:"percent_reserved_for_bad_bots,omitempty"`

	// How much space to reserve in percent for browsers. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PercentReservedForBrowsers *uint32 `json:"percent_reserved_for_browsers,omitempty"`

	// How much space to reserve in percent for known good bots. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PercentReservedForGoodBots *uint32 `json:"percent_reserved_for_good_bots,omitempty"`

	// How much space to reserve in percent for outstanding upstream requests. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PercentReservedForOutstanding *uint32 `json:"percent_reserved_for_outstanding,omitempty"`

	// The number of User-Agent entries to cache on each Service Engine. Allowed values are 500-10000000. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeCacheSize *uint32 `json:"se_cache_size,omitempty"`

	// How often in seconds to send updates about User-Agent cache entries to the next upstream cache. Field introduced in 21.1.1. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UpstreamUpdateInterval *uint32 `json:"upstream_update_interval,omitempty"`
}
