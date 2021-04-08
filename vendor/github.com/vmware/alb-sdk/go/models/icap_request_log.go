package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IcapRequestLog icap request log
// swagger:model IcapRequestLog
type IcapRequestLog struct {

	// Denotes whether the content was processed by ICAP server and an action was taken. Enum options - ICAP_DISABLED, ICAP_PASSED, ICAP_MODIFIED, ICAP_BLOCKED, ICAP_FAILED. Field introduced in 20.1.1.
	Action *string `json:"action,omitempty"`

	// Complete request body from client was sent to The ICAP server. Field introduced in 20.1.1.
	CompleteBodySent *bool `json:"complete_body_sent,omitempty"`

	// The HTTP method of the request. Enum options - HTTP_METHOD_GET, HTTP_METHOD_HEAD, HTTP_METHOD_PUT, HTTP_METHOD_DELETE, HTTP_METHOD_POST, HTTP_METHOD_OPTIONS, HTTP_METHOD_TRACE, HTTP_METHOD_CONNECT, HTTP_METHOD_PATCH, HTTP_METHOD_PROPFIND, HTTP_METHOD_PROPPATCH, HTTP_METHOD_MKCOL, HTTP_METHOD_COPY, HTTP_METHOD_MOVE, HTTP_METHOD_LOCK, HTTP_METHOD_UNLOCK. Field introduced in 20.1.1.
	HTTPMethod *string `json:"http_method,omitempty"`

	// The HTTP response code received from the ICAP server. HTTP response code is only available if content is blocked. Field introduced in 20.1.1.
	HTTPResponseCode *int32 `json:"http_response_code,omitempty"`

	// The absolute ICAP uri of the request. Field introduced in 20.1.1.
	IcapAbsoluteURI *string `json:"icap_absolute_uri,omitempty"`

	// ICAP response headers received from ICAP server. Field introduced in 20.1.1.
	IcapHeadersReceivedFromServer *string `json:"icap_headers_received_from_server,omitempty"`

	// ICAP request headers sent to ICAP server. Field introduced in 20.1.1.
	IcapHeadersSentToServer *string `json:"icap_headers_sent_to_server,omitempty"`

	// The ICAP method of the request. Enum options - ICAP_METHOD_REQMOD, ICAP_METHOD_RESPMOD. Field introduced in 20.1.1.
	IcapMethod *string `json:"icap_method,omitempty"`

	// The response code received from the ICAP server. Field introduced in 20.1.1.
	IcapResponseCode *int32 `json:"icap_response_code,omitempty"`

	// ICAP server IP for this connection. Field introduced in 20.1.3.
	IcapServerIP *int32 `json:"icap_server_ip,omitempty"`

	// ICAP server port for this connection. Field introduced in 20.1.3.
	IcapServerPort *int32 `json:"icap_server_port,omitempty"`

	// Latency added due to ICAP processing. This is the time taken from 1st byte of ICAP request sent to last byte of ICAP response received. Field introduced in 20.1.1. Unit is MILLISECONDS.
	Latency *int64 `json:"latency,omitempty"`

	// Content-Length of the modified content from ICAP server. Field introduced in 20.1.1.
	ModifiedContentLength *int32 `json:"modified_content_length,omitempty"`

	// The name of the pool that was used for the request. Field introduced in 20.1.1.
	PoolName *string `json:"pool_name,omitempty"`

	// The uuid of the pool that was used for the request. Field introduced in 20.1.1.
	PoolUUID *string `json:"pool_uuid,omitempty"`

	// Blocking reason for the content. It is available only if content was scanned by ICAP server and some violations were found. Field introduced in 20.1.1.
	Reason *string `json:"reason,omitempty"`

	// ICAP server IP for this connection. Field deprecated in 20.1.3. Field introduced in 20.1.1.
	ServerIP *IPAddr `json:"server_ip,omitempty"`

	// Source port for this connection. Field introduced in 20.1.1.
	SourcePort *int32 `json:"source_port,omitempty"`

	// Detailed description of the threat found in the content. Available only if request was scanned by ICAP server and some violations were found. Field deprecated in 20.1.3. Field introduced in 20.1.1.
	ThreatDescription *string `json:"threat_description,omitempty"`

	// Short description of the threat found in the content. Available only if content was scanned by ICAP server and some violations were found. Field introduced in 20.1.1.
	ThreatID *string `json:"threat_id,omitempty"`

	// Threat found in the content.  Available only if content was scanned by ICAP server and some violations were found. Field introduced in 20.1.3.
	Violations []*IcapViolation `json:"violations,omitempty"`
}
