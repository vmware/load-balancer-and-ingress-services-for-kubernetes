package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DSRequestLog d s request log
// swagger:model DSRequestLog
type DSRequestLog struct {

	// Name of the DataScript where this request was called. Field introduced in 20.1.3.
	DsName *string `json:"ds_name,omitempty"`

	// DataScript event where out-of-band request was sent. Enum options - VS_DATASCRIPT_EVT_HTTP_REQ, VS_DATASCRIPT_EVT_HTTP_RESP, VS_DATASCRIPT_EVT_HTTP_RESP_DATA, VS_DATASCRIPT_EVT_HTTP_LB_FAILED, VS_DATASCRIPT_EVT_HTTP_REQ_DATA, VS_DATASCRIPT_EVT_HTTP_RESP_FAILED, VS_DATASCRIPT_EVT_HTTP_LB_DONE, VS_DATASCRIPT_EVT_HTTP_AUTH, VS_DATASCRIPT_EVT_HTTP_POST_AUTH, VS_DATASCRIPT_EVT_TCP_CLIENT_ACCEPT, VS_DATASCRIPT_EVT_SSL_HANDSHAKE_DONE, VS_DATASCRIPT_EVT_DNS_REQ, VS_DATASCRIPT_EVT_DNS_RESP, VS_DATASCRIPT_EVT_L4_REQUEST, VS_DATASCRIPT_EVT_L4_RESPONSE, VS_DATASCRIPT_EVT_MAX. Field introduced in 20.1.3.
	Event *string `json:"event,omitempty"`

	// Response headers received from the server. Field introduced in 20.1.3.
	HeadersReceivedFromServer *string `json:"headers_received_from_server,omitempty"`

	// Request headers sent to the server. Field introduced in 20.1.3.
	HeadersSentToServer *string `json:"headers_sent_to_server,omitempty"`

	// The HTTP response code received from the external server. Field introduced in 20.1.3.
	HTTPResponseCode *int32 `json:"http_response_code,omitempty"`

	// The HTTP version of the out-of-band request. Field introduced in 20.1.3.
	HTTPVersion *string `json:"http_version,omitempty"`

	// The HTTP method of the out-of-band request. Field introduced in 20.1.3.
	Method *string `json:"method,omitempty"`

	// The name of the pool that was used for the request. Field introduced in 20.1.3.
	PoolName *string `json:"pool_name,omitempty"`

	// The uuid of the pool that was used for the request. Field introduced in 20.1.3.
	PoolUUID *string `json:"pool_uuid,omitempty"`

	// Length of the request sent in bytes. Field introduced in 20.1.3. Unit is BYTES.
	RequestLength *int64 `json:"request_length,omitempty"`

	// Length of the response received in bytes. Field introduced in 20.1.3. Unit is BYTES.
	ResponseLength *int64 `json:"response_length,omitempty"`

	// The IP of the server that was used for the request. Field introduced in 20.1.3.
	ServerIP *int32 `json:"server_ip,omitempty"`

	// The name of the server that was used for the request. Field introduced in 20.1.3.
	ServerName *string `json:"server_name,omitempty"`

	// The port of the server that was used for the request. Field introduced in 20.1.3.
	ServerPort *int32 `json:"server_port,omitempty"`

	// Number of servers tried during server reselect before the response is sent back. Field introduced in 20.1.3.
	ServersTried *int32 `json:"servers_tried,omitempty"`

	// The source port for this request. Field introduced in 20.1.3.
	SourcePort *int32 `json:"source_port,omitempty"`

	// Total time taken to process the Out-of-Band request. This is the time taken from the 1st byte of the request sent to the last byte of the response received. Field introduced in 20.1.3. Unit is MILLISECONDS.
	TotalTime *int64 `json:"total_time,omitempty"`

	// The URI path of the out-of-band request. Field introduced in 20.1.3.
	URIPath *string `json:"uri_path,omitempty"`

	// The URI query of the out-of-band request. Field introduced in 20.1.3.
	URIQuery *string `json:"uri_query,omitempty"`
}
