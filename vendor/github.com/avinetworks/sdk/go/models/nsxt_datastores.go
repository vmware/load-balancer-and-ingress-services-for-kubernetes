package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NsxtDatastores nsxt datastores
// swagger:model NsxtDatastores
type NsxtDatastores struct {

	// List of shared datastores. Field introduced in 20.1.2.
	DsIds []string `json:"ds_ids,omitempty"`

	// Include or Exclude. Field introduced in 20.1.2.
	Include *bool `json:"include,omitempty"`
}
