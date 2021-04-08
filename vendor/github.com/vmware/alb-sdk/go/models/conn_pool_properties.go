package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ConnPoolProperties conn pool properties
// swagger:model ConnPoolProperties
type ConnPoolProperties struct {

	// Connection idle timeout. Field introduced in 18.2.1. Allowed in Basic(Allowed values- 60000) edition, Essentials(Allowed values- 60000) edition, Enterprise edition.
	UpstreamConnpoolConnIDLETmo *int32 `json:"upstream_connpool_conn_idle_tmo,omitempty"`

	// Connection life timeout. Field introduced in 18.2.1. Allowed in Basic(Allowed values- 600000) edition, Essentials(Allowed values- 600000) edition, Enterprise edition.
	UpstreamConnpoolConnLifeTmo *int32 `json:"upstream_connpool_conn_life_tmo,omitempty"`

	// Maximum number of times a connection can be reused. Special values are 0- 'unlimited'. Field introduced in 18.2.1. Allowed in Basic(Allowed values- 0) edition, Essentials(Allowed values- 0) edition, Enterprise edition.
	UpstreamConnpoolConnMaxReuse *int32 `json:"upstream_connpool_conn_max_reuse,omitempty"`

	// Maximum number of connections a server can cache. Special values are 0- 'unlimited'. Field introduced in 18.2.1.
	UpstreamConnpoolServerMaxCache *int32 `json:"upstream_connpool_server_max_cache,omitempty"`
}
