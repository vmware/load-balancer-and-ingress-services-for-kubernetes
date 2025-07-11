// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// Http2ApplicationProfile http2 application profile
// swagger:model HTTP2ApplicationProfile
type Http2ApplicationProfile struct {

	// Enables automatic conversion of preload links specified in the 'Link' response header fields into Server push requests. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EnableHttp2ServerPush *bool `json:"enable_http2_server_push,omitempty"`

	// The initial flow control window size in KB for HTTP/2 streams. Allowed values are 64-32768. Field introduced in 18.2.10, 20.1.1. Unit is KB. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Http2InitialWindowSize *uint32 `json:"http2_initial_window_size,omitempty"`

	// Maximum number of concurrent push streams over a client side HTTP/2 connection. Allowed values are 1-256. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MaxHttp2ConcurrentPushesPerConnection *uint32 `json:"max_http2_concurrent_pushes_per_connection,omitempty"`

	// Maximum number of concurrent streams over a client side HTTP/2 connection. Allowed values are 1-256. Field introduced in 18.2.10, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxHttp2ConcurrentStreamsPerConnection *uint32 `json:"max_http2_concurrent_streams_per_connection,omitempty"`

	// Maximum number of control frames that client can send over an HTTP/2 connection. '0' means unlimited. Allowed values are 0-10000. Special values are 0- Unlimited control frames on a client side HTTP/2 connection. Field introduced in 18.2.10, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxHttp2ControlFramesPerConnection *uint32 `json:"max_http2_control_frames_per_connection,omitempty"`

	// Maximum number of empty data frames that client can send over an HTTP/2 connection. '0' means unlimited. Allowed values are 0-10000. Special values are 0- Unlimited empty data frames over a client side HTTP/2 connection. Field introduced in 18.2.10, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxHttp2EmptyDataFramesPerConnection *uint32 `json:"max_http2_empty_data_frames_per_connection,omitempty"`

	// Maximum size in bytes of the compressed request header field. The limit applies equally to both name and value. Allowed values are 1-8192. Field introduced in 18.2.10, 20.1.1. Unit is BYTES. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxHttp2HeaderFieldSize *uint32 `json:"max_http2_header_field_size,omitempty"`

	// Maximum number of frames that can be queued waiting to be sent over a client side HTTP/2 connection at any given time. '0' means unlimited. Allowed values are 0-10000. Special values are 0- Unlimited frames can be queued on a client side HTTP/2 connection. Field introduced in 18.2.10, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxHttp2QueuedFramesToClientPerConnection *uint32 `json:"max_http2_queued_frames_to_client_per_connection,omitempty"`

	// Maximum number of requests over a client side HTTP/2 connection. Allowed values are 0-10000. Special values are 0- Unlimited requests on a client side HTTP/2 connection. Field introduced in 18.2.10, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxHttp2RequestsPerConnection *uint32 `json:"max_http2_requests_per_connection,omitempty"`
}
