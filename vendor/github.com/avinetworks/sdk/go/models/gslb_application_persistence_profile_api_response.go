package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbApplicationPersistenceProfileAPIResponse gslb application persistence profile Api response
// swagger:model GslbApplicationPersistenceProfileApiResponse
type GslbApplicationPersistenceProfileAPIResponse struct {

	// count
	// Required: true
	Count int32 `json:"count"`

	// results
	// Required: true
	Results []*GslbApplicationPersistenceProfile `json:"results,omitempty"`
}
