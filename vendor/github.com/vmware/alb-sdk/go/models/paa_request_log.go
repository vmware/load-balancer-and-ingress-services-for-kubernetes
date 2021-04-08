package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// PaaRequestLog paa request log
// swagger:model PaaRequestLog
type PaaRequestLog struct {

	// Response headers received from PingAccess server. Field introduced in 18.2.3.
	HeadersReceivedFromServer *string `json:"headers_received_from_server,omitempty"`

	// Request headers sent to PingAccess server. Field introduced in 18.2.3.
	HeadersSentToServer *string `json:"headers_sent_to_server,omitempty"`

	// The http version of the request. Field introduced in 18.2.3.
	HTTPVersion *string `json:"http_version,omitempty"`

	// The http method of the request. Field introduced in 18.2.3.
	Method *string `json:"method,omitempty"`

	// The name of the pool that was used for the request. Field introduced in 18.2.3.
	PoolName *string `json:"pool_name,omitempty"`

	// The response code received from the PingAccess server. Field introduced in 18.2.3.
	ResponseCode *int32 `json:"response_code,omitempty"`

	// The IP of the server that was sent the request. Field introduced in 18.2.3.
	ServerIP *int32 `json:"server_ip,omitempty"`

	// Number of servers tried during server reselect before the response is sent back. Field introduced in 18.2.3.
	ServersTried *int32 `json:"servers_tried,omitempty"`

	// The uri of the request. Field introduced in 18.2.3.
	URIPath *string `json:"uri_path,omitempty"`
}
