package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// PoolGroupDeploymentPolicyAPIResponse pool group deployment policy Api response
// swagger:model PoolGroupDeploymentPolicyApiResponse
type PoolGroupDeploymentPolicyAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*PoolGroupDeploymentPolicy `json:"results,omitempty"`
}
