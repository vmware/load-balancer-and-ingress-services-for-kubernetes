package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HTTPCacheConfig Http cache config
// swagger:model HttpCacheConfig
type HTTPCacheConfig struct {

	// Add an Age header to content served from cache, which indicates to the client the number of seconds the object has been in the cache.
	AgeHeader *bool `json:"age_header,omitempty"`

	// Enable/disable caching objects without Cache-Control headers.
	Aggressive *bool `json:"aggressive,omitempty"`

	// If a Date header was not added by the server, add a Date header to the object served from cache.  This indicates to the client when the object was originally sent by the server to the cache.
	DateHeader *bool `json:"date_header,omitempty"`

	// Default expiration time of cache objects received from the server without a Cache-Control expiration header.  This value may be overwritten by the Heuristic Expire setting.
	DefaultExpire *int32 `json:"default_expire,omitempty"`

	// Enable/disable HTTP object caching.When enabling caching for the first time, SE Group app_cache_percent must beset to allocate shared memory required for caching (A service engine restart is needed after setting/resetting the SE group value).
	Enabled *bool `json:"enabled,omitempty"`

	// If a response object from the server does not include the Cache-Control header, but does include a Last-Modified header, the system will use this time to calculate the Cache-Control expiration.  If unable to solicit an Last-Modified header, then the system will fall back to the Cache Expire Time value.
	HeuristicExpire *bool `json:"heuristic_expire,omitempty"`

	// Ignore client's cache control headers when fetching or storing from and to the cache. Field introduced in 18.1.2.
	IgnoreRequestCacheControl *bool `json:"ignore_request_cache_control,omitempty"`

	// Max size, in bytes, of the cache.  The default, zero, indicates auto configuration.
	MaxCacheSize *int64 `json:"max_cache_size,omitempty"`

	// Maximum size of an object to store in the cache.
	MaxObjectSize *int32 `json:"max_object_size,omitempty"`

	// Blacklist *string group of non-cacheable mime types. It is a reference to an object of type StringGroup.
	MimeTypesBlackGroupRefs []string `json:"mime_types_black_group_refs,omitempty"`

	// Blacklist of non-cacheable mime types.
	MimeTypesBlackList []string `json:"mime_types_black_list,omitempty"`

	// Whitelist *string group of cacheable mime types. If both Cacheable Mime Types *string list and *string group are empty, this defaults to */*. It is a reference to an object of type StringGroup.
	MimeTypesGroupRefs []string `json:"mime_types_group_refs,omitempty"`

	// Whitelist of cacheable mime types. If both Cacheable Mime Types *string list and *string group are empty, this defaults to */*.
	MimeTypesList []string `json:"mime_types_list,omitempty"`

	// Minimum size of an object to store in the cache.
	MinObjectSize *int32 `json:"min_object_size,omitempty"`

	// Allow caching of objects whose URI included a query argument.  When disabled, these objects are not cached.  When enabled, the request must match the URI query to be considered a hit.
	QueryCacheable *bool `json:"query_cacheable,omitempty"`

	// Non-cacheable URI configuration with match criteria. Field introduced in 18.1.2.
	URINonCacheable *PathMatch `json:"uri_non_cacheable,omitempty"`

	// Add an X-Cache header to content served from cache, which indicates to the client that the object was served from an intermediate cache.
	XcacheHeader *bool `json:"xcache_header,omitempty"`
}
