package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AttackMetaDataAPIResponse attack meta data Api response
// swagger:model AttackMetaDataApiResponse
type AttackMetaDataAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*AttackMetaData `json:"results,omitempty"`
}
