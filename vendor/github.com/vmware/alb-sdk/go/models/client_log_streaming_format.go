package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ClientLogStreamingFormat client log streaming format
// swagger:model ClientLogStreamingFormat
type ClientLogStreamingFormat struct {

	// Format for the streamed logs. Enum options - LOG_STREAMING_FORMAT_JSON_FULL, LOG_STREAMING_FORMAT_JSON_SELECTED. Field introduced in 18.2.5.
	// Required: true
	Format *string `json:"format"`

	// List of log fields to be streamed, when selective fields (LOG_STREAMING_FORMAT_JSON_SELECTED) option is chosen. Only top-level fields in application or connection logs are supported. Field introduced in 18.2.5.
	IncludedFields []string `json:"included_fields,omitempty"`
}
