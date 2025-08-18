// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DSRequestLog d s request log
// swagger:model DSRequestLog
type DSRequestLog struct {

	// Name of the DataScript where this request was called. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DsName *string `json:"ds_name,omitempty"`

	// DataScript event where out-of-band request was sent. Enum options - VS_DATASCRIPT_EVT_HTTP_REQ, VS_DATASCRIPT_EVT_HTTP_RESP, VS_DATASCRIPT_EVT_HTTP_RESP_DATA, VS_DATASCRIPT_EVT_HTTP_LB_FAILED, VS_DATASCRIPT_EVT_HTTP_REQ_DATA, VS_DATASCRIPT_EVT_HTTP_RESP_FAILED, VS_DATASCRIPT_EVT_HTTP_LB_DONE, VS_DATASCRIPT_EVT_HTTP_AUTH, VS_DATASCRIPT_EVT_HTTP_POST_AUTH, VS_DATASCRIPT_EVT_TCP_CLIENT_ACCEPT, VS_DATASCRIPT_EVT_SSL_HANDSHAKE_DONE, VS_DATASCRIPT_EVT_CLIENT_SSL_PRE_CONNECT, VS_DATASCRIPT_EVT_CLIENT_SSL_CLIENT_HELLO, VS_DATASCRIPT_EVT_DNS_REQ, VS_DATASCRIPT_EVT_DNS_RESP, VS_DATASCRIPT_EVT_L4_REQUEST, VS_DATASCRIPT_EVT_L4_RESPONSE, VS_DATASCRIPT_EVT_MAX. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Event *string `json:"event,omitempty"`

	// Response headers received from the server. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	HeadersReceivedFromServer *string `json:"headers_received_from_server,omitempty"`

	// Request headers sent to the server. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	HeadersSentToServer *string `json:"headers_sent_to_server,omitempty"`

	// The HTTP response code received from the external server. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	HTTPResponseCode *uint32 `json:"http_response_code,omitempty"`

	// The HTTP version of the out-of-band request. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	HTTPVersion *string `json:"http_version,omitempty"`

	// The HTTP method of the out-of-band request. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Method *string `json:"method,omitempty"`

	// The name of the pool that was used for the request. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PoolName *string `json:"pool_name,omitempty"`

	// The uuid of the pool that was used for the request. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PoolUUID *string `json:"pool_uuid,omitempty"`

	// Length of the request sent in bytes. Field introduced in 20.1.3. Unit is BYTES. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RequestLength *uint64 `json:"request_length,omitempty"`

	// Length of the response received in bytes. Field introduced in 20.1.3. Unit is BYTES. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ResponseLength *uint64 `json:"response_length,omitempty"`

	// The IP of the server that was used for the request. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ServerIP *uint32 `json:"server_ip,omitempty"`

	// The name of the server that was used for the request. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ServerName *string `json:"server_name,omitempty"`

	// The port of the server that was used for the request. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ServerPort *uint32 `json:"server_port,omitempty"`

	// Number of servers tried during server reselect before the response is sent back. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ServersTried *uint32 `json:"servers_tried,omitempty"`

	// The source port for this request. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SourcePort *uint32 `json:"source_port,omitempty"`

	// Total time taken to process the Out-of-Band request. This is the time taken from the 1st byte of the request sent to the last byte of the response received. Field introduced in 20.1.3. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TotalTime *uint64 `json:"total_time,omitempty"`

	// The URI path of the out-of-band request. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	URIPath *string `json:"uri_path,omitempty"`

	// The URI query of the out-of-band request. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	URIQuery *string `json:"uri_query,omitempty"`
}
