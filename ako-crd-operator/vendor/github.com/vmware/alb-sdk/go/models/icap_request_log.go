// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// IcapRequestLog icap request log
// swagger:model IcapRequestLog
type IcapRequestLog struct {

	// Denotes whether the content was processed by ICAP server and an action was taken. Enum options - ICAP_DISABLED, ICAP_PASSED, ICAP_MODIFIED, ICAP_BLOCKED, ICAP_FAILED. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Action *string `json:"action,omitempty"`

	// Complete request body from client was sent to The ICAP server. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CompleteBodySent *bool `json:"complete_body_sent,omitempty"`

	// The HTTP method of the request. Enum options - HTTP_METHOD_GET, HTTP_METHOD_HEAD, HTTP_METHOD_PUT, HTTP_METHOD_DELETE, HTTP_METHOD_POST, HTTP_METHOD_OPTIONS, HTTP_METHOD_TRACE, HTTP_METHOD_CONNECT, HTTP_METHOD_PATCH, HTTP_METHOD_PROPFIND, HTTP_METHOD_PROPPATCH, HTTP_METHOD_MKCOL, HTTP_METHOD_COPY, HTTP_METHOD_MOVE, HTTP_METHOD_LOCK, HTTP_METHOD_UNLOCK. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HTTPMethod *string `json:"http_method,omitempty"`

	// The HTTP response code received from the ICAP server. HTTP response code is only available if content is blocked. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HTTPResponseCode uint32 `json:"http_response_code,omitempty"`

	// The absolute ICAP uri of the request. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IcapAbsoluteURI *string `json:"icap_absolute_uri,omitempty"`

	// ICAP response headers received from ICAP server. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IcapHeadersReceivedFromServer *string `json:"icap_headers_received_from_server,omitempty"`

	// ICAP request headers sent to ICAP server. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IcapHeadersSentToServer *string `json:"icap_headers_sent_to_server,omitempty"`

	// The ICAP method of the request. Enum options - ICAP_METHOD_REQMOD, ICAP_METHOD_RESPMOD. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IcapMethod *string `json:"icap_method,omitempty"`

	// The response code received from the ICAP server. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IcapResponseCode uint32 `json:"icap_response_code,omitempty"`

	// ICAP server IP for this connection. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IcapServerIP uint32 `json:"icap_server_ip,omitempty"`

	// ICAP server port for this connection. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IcapServerPort uint32 `json:"icap_server_port,omitempty"`

	// Latency added due to ICAP processing. This is the time taken from 1st byte of ICAP request sent to last byte of ICAP response received. Field introduced in 20.1.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Latency uint64 `json:"latency,omitempty"`

	// Content-Length of the modified content from ICAP server. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ModifiedContentLength uint32 `json:"modified_content_length,omitempty"`

	// ICAP log specific to NSX Defender. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NsxDefenderLog *IcapNSXDefenderLog `json:"nsx_defender_log,omitempty"`

	// ICAP log specific to OPSWAT. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	OpswatLog *IcapOPSWATLog `json:"opswat_log,omitempty"`

	// The name of the pool that was used for the request. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PoolName *string `json:"pool_name,omitempty"`

	// The uuid of the pool that was used for the request. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PoolUUID *string `json:"pool_uuid,omitempty"`

	// Source port for this connection. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SourcePort uint32 `json:"source_port,omitempty"`

	// Selected ICAP vendor for the request. Enum options - ICAP_VENDOR_GENERIC, ICAP_VENDOR_OPSWAT, ICAP_VENDOR_LASTLINE. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Vendor *string `json:"vendor,omitempty"`
}
