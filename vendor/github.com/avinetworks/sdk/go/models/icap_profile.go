package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IcapProfile icap profile
// swagger:model IcapProfile
type IcapProfile struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// The maximum buffer size for the HTTP request body. If the request body exceeds this size, the request will not be checked by the ICAP server. In this case, the configured action will be executed and a significant log entry will be generated. Allowed values are 1-51200. Field introduced in 20.1.1. Unit is KB.
	BufferSize *int32 `json:"buffer_size,omitempty"`

	// Decide what should happen if the request body size exceeds the configured buffer size. If this is set to Fail Open, the request will not be checked by the ICAP server. If this is set to Fail Closed, the request will be rejected with 413 status code. Enum options - ICAP_FAIL_OPEN, ICAP_FAIL_CLOSED. Field introduced in 20.1.1.
	BufferSizeExceedAction *string `json:"buffer_size_exceed_action,omitempty"`

	// The cloud where this object belongs to. This must match the cloud referenced in the pool group below. It is a reference to an object of type Cloud. Field introduced in 20.1.1.
	CloudRef *string `json:"cloud_ref,omitempty"`

	// A description for this ICAP profile. Field introduced in 20.1.1.
	Description *string `json:"description,omitempty"`

	// Use the ICAP preview feature as described in RFC 3507 section 4.5. Field introduced in 20.1.1.
	EnablePreview *bool `json:"enable_preview,omitempty"`

	// Decide what should happen if there is a problem with the ICAP server like communication timeout, protocol error, pool error, etc. If this is set to Fail Open, the request will continue, but will create a significant log entry. If this is set to Fail Closed, the request will be rejected with a 503 status code. Enum options - ICAP_FAIL_OPEN, ICAP_FAIL_CLOSED. Field introduced in 20.1.1.
	FailAction *string `json:"fail_action,omitempty"`

	// Name of the ICAP profile. Field introduced in 20.1.1.
	// Required: true
	Name *string `json:"name"`

	// The pool group which is used to connect to ICAP servers. It is a reference to an object of type PoolGroup. Field introduced in 20.1.1.
	// Required: true
	PoolGroupRef *string `json:"pool_group_ref"`

	// The ICAP preview size as described in RFC 3507 section 4.5. This should not exceed the size supported by the ICAP server. If this is set to 0, only the HTTP header will be sent to the ICAP server as a preview. To disable preview completely, set the enable-preview option to false. Allowed values are 0-5000. Field introduced in 20.1.1. Unit is BYTES.
	PreviewSize *int32 `json:"preview_size,omitempty"`

	// How long do we wait for a request to the ICAP server to finish. If this timeout is exceeded, the request to the ICAP server will be aborted and the configured fail action is executed. Allowed values are 50-3600000. Field introduced in 20.1.1. Unit is MILLISECONDS.
	ResponseTimeout *int32 `json:"response_timeout,omitempty"`

	// The path and query component of the ICAP URL. Host name and port will be taken from the pool. Field introduced in 20.1.1.
	// Required: true
	ServiceURI *string `json:"service_uri"`

	// If the ICAP request takes longer than this value, this request will generate a significant log entry. Allowed values are 50-3600000. Field introduced in 20.1.1. Unit is MILLISECONDS.
	SlowResponseWarningThreshold *int32 `json:"slow_response_warning_threshold,omitempty"`

	// Tenant which this object belongs to. It is a reference to an object of type Tenant. Field introduced in 20.1.1.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the ICAP profile. Field introduced in 20.1.1.
	UUID *string `json:"uuid,omitempty"`

	// The vendor of the ICAP server. Enum options - ICAP_VENDOR_GENERIC, ICAP_VENDOR_OPSWAT. Field introduced in 20.1.1.
	Vendor *string `json:"vendor,omitempty"`
}
