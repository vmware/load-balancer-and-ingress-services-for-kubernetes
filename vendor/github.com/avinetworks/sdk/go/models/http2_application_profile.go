package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// Http2ApplicationProfile http2 application profile
// swagger:model HTTP2ApplicationProfile
type Http2ApplicationProfile struct {

	// The initial flow control window size in KB for HTTP/2 streams. Allowed values are 64-32768. Field introduced in 18.2.10, 20.1.1.
	Http2InitialWindowSize *int32 `json:"http2_initial_window_size,omitempty"`

	// The max number of concurrent streams over a client side HTTP/2 connection. Allowed values are 1-256. Field introduced in 18.2.10, 20.1.1.
	MaxHttp2ConcurrentStreamsPerConnection *int32 `json:"max_http2_concurrent_streams_per_connection,omitempty"`

	// The max number of control frames that client can send over an HTTP/2 connection. '0' means unlimited. Allowed values are 0-10000. Special values are 0- 'Unlimited control frames on a client side HTTP/2 connection'. Field introduced in 18.2.10, 20.1.1.
	MaxHttp2ControlFramesPerConnection *int32 `json:"max_http2_control_frames_per_connection,omitempty"`

	// The max number of empty data frames that client can send over an HTTP/2 connection. '0' means unlimited. Allowed values are 0-10000. Special values are 0- 'Unlimited empty data frames over a client side HTTP/2 connection'. Field introduced in 18.2.10, 20.1.1.
	MaxHttp2EmptyDataFramesPerConnection *int32 `json:"max_http2_empty_data_frames_per_connection,omitempty"`

	// The maximum size in bytes of the compressed request header field. The limit applies equally to both name and value. Allowed values are 1-8192. Field introduced in 18.2.10, 20.1.1.
	MaxHttp2HeaderFieldSize *int32 `json:"max_http2_header_field_size,omitempty"`

	// The max number of frames that can be queued waiting to be sent over a client side HTTP/2 connection at any given time. '0' means unlimited. Allowed values are 0-10000. Special values are 0- 'Unlimited frames can be queued on a client side HTTP/2 connection'. Field introduced in 18.2.10, 20.1.1.
	MaxHttp2QueuedFramesToClientPerConnection *int32 `json:"max_http2_queued_frames_to_client_per_connection,omitempty"`

	// The maximum number of requests over a client side HTTP/2 connection. Allowed values are 0-10000. Special values are 0- 'Unlimited requests on a client side HTTP/2 connection'. Field introduced in 18.2.10, 20.1.1.
	MaxHttp2RequestsPerConnection *int32 `json:"max_http2_requests_per_connection,omitempty"`
}
