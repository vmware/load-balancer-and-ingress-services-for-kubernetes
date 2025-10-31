// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HTTPCacheConfig Http cache config
// swagger:model HttpCacheConfig
type HTTPCacheConfig struct {

	// Add an Age header to content served from cache, which indicates to the client the number of seconds the object has been in the cache. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AgeHeader *bool `json:"age_header,omitempty"`

	// Enable/disable caching objects without Cache-Control headers. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Aggressive *bool `json:"aggressive,omitempty"`

	// If a Date header was not added by the server, add a Date header to the object served from cache.  This indicates to the client when the object was originally sent by the server to the cache. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DateHeader *bool `json:"date_header,omitempty"`

	// Default expiration time of cache objects received from the server without a Cache-Control expiration header.  This value may be overwritten by the Heuristic Expire setting. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DefaultExpire *uint32 `json:"default_expire,omitempty"`

	// Enable/disable HTTP object caching.When enabling caching for the first time, SE Group app_cache_percent must be set to allocate shared memory required for caching (A service engine restart is needed after setting/resetting the SE group value). Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Enabled *bool `json:"enabled,omitempty"`

	// If a response object from the server does not include the Cache-Control header, but does include a Last-Modified header, the system will use this time to calculate the Cache-Control expiration.  If unable to solicit an Last-Modified header, then the system will fall back to the Cache Expire Time value. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HeuristicExpire *bool `json:"heuristic_expire,omitempty"`

	// Ignore client's cache control headers when fetching or storing from and to the cache. Field introduced in 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IgnoreRequestCacheControl *bool `json:"ignore_request_cache_control,omitempty"`

	// Max size, in bytes, of the cache.  The default, zero, indicates auto configuration. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxCacheSize *uint64 `json:"max_cache_size,omitempty"`

	// Maximum size of an object to store in the cache. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxObjectSize *uint32 `json:"max_object_size,omitempty"`

	// Blocklist *string group of non-cacheable mime types. It is a reference to an object of type StringGroup. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MimeTypesBlockGroupRefs []string `json:"mime_types_block_group_refs,omitempty"`

	// Blocklist of non-cacheable mime types. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MimeTypesBlockLists []string `json:"mime_types_block_lists,omitempty"`

	// Allowlist *string group of cacheable mime types. If both Cacheable Mime Types *string list and *string group are empty, this defaults to */*. It is a reference to an object of type StringGroup. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MimeTypesGroupRefs []string `json:"mime_types_group_refs,omitempty"`

	// Allowlist of cacheable mime types. If both Cacheable Mime Types *string list and *string group are empty, this defaults to */*. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MimeTypesList []string `json:"mime_types_list,omitempty"`

	// Minimum size of an object to store in the cache. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MinObjectSize *uint32 `json:"min_object_size,omitempty"`

	// Allow caching of objects whose URI included a query argument.  When disabled, these objects are not cached.  When enabled, the request must match the URI query to be considered a hit. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	QueryCacheable *bool `json:"query_cacheable,omitempty"`

	// Non-cacheable URI configuration with match criteria. Field introduced in 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	URINonCacheable *PathMatch `json:"uri_non_cacheable,omitempty"`

	// Add an X-Cache header to content served from cache, which indicates to the client that the object was served from an intermediate cache. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	XcacheHeader *bool `json:"xcache_header,omitempty"`
}
