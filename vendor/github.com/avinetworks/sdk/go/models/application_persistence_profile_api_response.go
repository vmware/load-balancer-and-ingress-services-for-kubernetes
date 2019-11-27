package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ApplicationPersistenceProfileAPIResponse application persistence profile Api response
// swagger:model ApplicationPersistenceProfileApiResponse
type ApplicationPersistenceProfileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*ApplicationPersistenceProfile `json:"results,omitempty"`
}
