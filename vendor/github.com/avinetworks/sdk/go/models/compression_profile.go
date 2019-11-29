package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CompressionProfile compression profile
// swagger:model CompressionProfile
type CompressionProfile struct {

	// Compress only content types listed in this *string group. Content types not present in this list are not compressed. It is a reference to an object of type StringGroup.
	CompressibleContentRef *string `json:"compressible_content_ref,omitempty"`

	// Compress HTTP response content if it wasn't already compressed.
	// Required: true
	Compression *bool `json:"compression"`

	// Custom filters used when auto compression is not selected.
	Filter []*CompressionFilter `json:"filter,omitempty"`

	// Offload compression from the servers to AVI. Saves compute cycles on the servers.
	// Required: true
	RemoveAcceptEncodingHeader *bool `json:"remove_accept_encoding_header"`

	// Compress content automatically or add custom filters to define compressible content and compression levels. Enum options - AUTO_COMPRESSION, CUSTOM_COMPRESSION.
	// Required: true
	Type *string `json:"type"`
}
