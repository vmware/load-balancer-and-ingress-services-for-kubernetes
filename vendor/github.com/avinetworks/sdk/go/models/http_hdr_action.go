package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HTTPHdrAction HTTP hdr action
// swagger:model HTTPHdrAction
type HTTPHdrAction struct {

	// ADD  A new header with the new value is added irrespective of the existence of an HTTP header of the given name. REPLACE  A new header with the new value is added if no header of the given name exists, else existing headers with the given name are removed and a new header with the new value is added. REMOVE  All the headers of the given name are removed. Enum options - HTTP_ADD_HDR, HTTP_REMOVE_HDR, HTTP_REPLACE_HDR.
	// Required: true
	Action *string `json:"action"`

	// Cookie information.
	Cookie *HTTPCookieData `json:"cookie,omitempty"`

	// HTTP header information.
	Hdr *HTTPHdrData `json:"hdr,omitempty"`
}
