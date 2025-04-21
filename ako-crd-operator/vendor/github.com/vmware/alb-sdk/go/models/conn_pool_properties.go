// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ConnPoolProperties conn pool properties
// swagger:model ConnPoolProperties
type ConnPoolProperties struct {

	// Connection idle timeout. Allowed values are 0-86400000. Special values are 0- Infinite idle time.. Field introduced in 18.2.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 60000), Basic edition(Allowed values- 60000), Enterprise with Cloud Services edition.
	UpstreamConnpoolConnIDLETmo *uint32 `json:"upstream_connpool_conn_idle_tmo,omitempty"`

	// Connection life timeout. Allowed values are 0-86400000. Special values are 0- Infinite life time.. Field introduced in 18.2.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 600000), Basic edition(Allowed values- 600000), Enterprise with Cloud Services edition.
	UpstreamConnpoolConnLifeTmo *uint32 `json:"upstream_connpool_conn_life_tmo,omitempty"`

	// Maximum number of times a connection can be reused. Special values are 0- unlimited. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 0), Basic edition(Allowed values- 0), Enterprise with Cloud Services edition.
	UpstreamConnpoolConnMaxReuse uint32 `json:"upstream_connpool_conn_max_reuse,omitempty"`

	// Maximum number of connections a server can cache. Special values are 0- unlimited. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UpstreamConnpoolServerMaxCache uint32 `json:"upstream_connpool_server_max_cache,omitempty"`
}
