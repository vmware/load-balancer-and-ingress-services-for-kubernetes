// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PaaRequestLog paa request log
// swagger:model PaaRequestLog
type PaaRequestLog struct {

	// Response headers received from PingAccess server. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HeadersReceivedFromServer *string `json:"headers_received_from_server,omitempty"`

	// Request headers sent to PingAccess server. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HeadersSentToServer *string `json:"headers_sent_to_server,omitempty"`

	// The http version of the request. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HTTPVersion *string `json:"http_version,omitempty"`

	// The http method of the request. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Method *string `json:"method,omitempty"`

	// The name of the pool that was used for the request. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PoolName *string `json:"pool_name,omitempty"`

	// The response code received from the PingAccess server. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ResponseCode uint32 `json:"response_code,omitempty"`

	// The IP of the server that was sent the request. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerIP uint32 `json:"server_ip,omitempty"`

	// Number of servers tried during server reselect before the response is sent back. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServersTried uint32 `json:"servers_tried,omitempty"`

	// The uri of the request. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	URIPath *string `json:"uri_path,omitempty"`
}
