// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SubRequestLog sub request log
// swagger:model SubRequestLog
type SubRequestLog struct {

	// Response headers received from the server. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	HeadersReceivedFromServer *string `json:"headers_received_from_server,omitempty"`

	// Request headers sent to the server. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	HeadersSentToServer *string `json:"headers_sent_to_server,omitempty"`

	// The HTTP response code received from the server. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	HTTPResponseCode uint32 `json:"http_response_code,omitempty"`

	// The HTTP version of the sub-request. Enum options - ZERO_NINE, ONE_ZERO, ONE_ONE, TWO_ZERO. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	HTTPVersion *string `json:"http_version,omitempty"`

	// The HTTP method of the sub-request. Enum options - HTTP_METHOD_GET, HTTP_METHOD_HEAD, HTTP_METHOD_PUT, HTTP_METHOD_DELETE, HTTP_METHOD_POST, HTTP_METHOD_OPTIONS, HTTP_METHOD_TRACE, HTTP_METHOD_CONNECT, HTTP_METHOD_PATCH, HTTP_METHOD_PROPFIND, HTTP_METHOD_PROPPATCH, HTTP_METHOD_MKCOL, HTTP_METHOD_COPY, HTTP_METHOD_MOVE, HTTP_METHOD_LOCK, HTTP_METHOD_UNLOCK. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Method *string `json:"method,omitempty"`

	// The name of the pool that was used for the sub-request. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PoolName *string `json:"pool_name,omitempty"`

	// The uuid of the pool that was used for the sub-request. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PoolUUID *string `json:"pool_uuid,omitempty"`

	// Length of the request sent in bytes. Field introduced in 21.1.3. Unit is BYTES. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RequestLength uint64 `json:"request_length,omitempty"`

	// Length of the response received in bytes. Field introduced in 21.1.3. Unit is BYTES. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ResponseLength uint64 `json:"response_length,omitempty"`

	// The IP of the server that was used for the sub-request. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ServerIP uint32 `json:"server_ip,omitempty"`

	// The name of the server that was used for the sub-request. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ServerName *string `json:"server_name,omitempty"`

	// The port of the server that was used for the sub-request. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ServerPort uint32 `json:"server_port,omitempty"`

	// The source port for this request. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SourcePort uint32 `json:"source_port,omitempty"`

	// Total time taken to process the Oauth Subrequest. This is the time taken from the 1st byte of the request sent to the last byte of the response received. Field introduced in 21.1.3. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TotalTime uint64 `json:"total_time,omitempty"`

	// The URI path of the sub-request. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	URIPath *string `json:"uri_path,omitempty"`

	// The URI query of the sub-request. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	URIQuery *string `json:"uri_query,omitempty"`
}
