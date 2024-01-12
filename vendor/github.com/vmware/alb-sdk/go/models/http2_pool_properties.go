// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// Http2PoolProperties http2 pool properties
// swagger:model HTTP2PoolProperties
type Http2PoolProperties struct {

	// The max number of control frames that server can send over an HTTP/2 connection. '0' means unlimited. Allowed values are 0-10000. Special values are 0- Unlimited control frames on a server side HTTP/2 connection. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MaxHttp2ControlFramesPerConnection uint32 `json:"max_http2_control_frames_per_connection,omitempty"`

	// The maximum size in bytes of the compressed request header field. The limit applies equally to both name and value. Allowed values are 1-8192. Field introduced in 21.1.1. Unit is BYTES. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MaxHttp2HeaderFieldSize *uint32 `json:"max_http2_header_field_size,omitempty"`
}
