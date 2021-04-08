package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VsResyncParams vs resync params
// swagger:model VsResyncParams
type VsResyncParams struct {

	//  It is a reference to an object of type ServiceEngine.
	SeRef []string `json:"se_ref,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
