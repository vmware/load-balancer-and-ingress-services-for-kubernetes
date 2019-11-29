package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ConnPoolProperties conn pool properties
// swagger:model ConnPoolProperties
type ConnPoolProperties struct {

	// Connection idle timeout. Field introduced in 18.2.1.
	UpstreamConnpoolConnIDLETmo *int32 `json:"upstream_connpool_conn_idle_tmo,omitempty"`

	// Connection life timeout. Field introduced in 18.2.1.
	UpstreamConnpoolConnLifeTmo *int32 `json:"upstream_connpool_conn_life_tmo,omitempty"`

	// Connection max used times. Field introduced in 18.2.1.
	UpstreamConnpoolConnMaxReuse *int32 `json:"upstream_connpool_conn_max_reuse,omitempty"`

	// Max cache connections per server. Field introduced in 18.2.1.
	UpstreamConnpoolServerMaxCache *int32 `json:"upstream_connpool_server_max_cache,omitempty"`
}
