package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HardwareSecurityModuleGroupAPIResponse hardware security module group Api response
// swagger:model HardwareSecurityModuleGroupApiResponse
type HardwareSecurityModuleGroupAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*HardwareSecurityModuleGroup `json:"results,omitempty"`
}
