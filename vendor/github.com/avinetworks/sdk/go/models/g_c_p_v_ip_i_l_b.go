package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GCPVIPILB g c p v IP i l b
// swagger:model GCPVIPILB
type GCPVIPILB struct {

	// Google Cloud Router Names to advertise BYOIP. Field introduced in 18.2.9, 20.1.1.
	CloudRouterNames []string `json:"cloud_router_names,omitempty"`
}
