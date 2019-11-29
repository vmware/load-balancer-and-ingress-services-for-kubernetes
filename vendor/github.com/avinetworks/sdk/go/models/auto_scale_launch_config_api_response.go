package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AutoScaleLaunchConfigAPIResponse auto scale launch config Api response
// swagger:model AutoScaleLaunchConfigApiResponse
type AutoScaleLaunchConfigAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*AutoScaleLaunchConfig `json:"results,omitempty"`
}
