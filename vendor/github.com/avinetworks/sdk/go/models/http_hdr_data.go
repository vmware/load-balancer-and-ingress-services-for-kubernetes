package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HTTPHdrData HTTP hdr data
// swagger:model HTTPHdrData
type HTTPHdrData struct {

	// HTTP header name.
	Name *string `json:"name,omitempty"`

	// HTTP header value.
	Value *HTTPHdrValue `json:"value,omitempty"`
}
