package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NsxtImageDetails nsxt image details
// swagger:model NsxtImageDetails
type NsxtImageDetails struct {

	// Cloud Id. Field introduced in 20.1.1.
	CcID *string `json:"cc_id,omitempty"`

	// Error message. Field introduced in 20.1.1.
	ErrorString *string `json:"error_string,omitempty"`

	// Image version. Field introduced in 20.1.1.
	ImageVersion *string `json:"image_version,omitempty"`

	// VC url. Field introduced in 20.1.1.
	VcURL *string `json:"vc_url,omitempty"`
}
