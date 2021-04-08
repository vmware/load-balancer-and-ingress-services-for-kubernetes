package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VirtualServiceResource virtual service resource
// swagger:model VirtualServiceResource
type VirtualServiceResource struct {

	// This field is not being used. Field deprecated in 18.1.5, 18.2.1.
	IsExclusive *bool `json:"is_exclusive,omitempty"`

	// Number of memory.
	Memory *int32 `json:"memory,omitempty"`

	// Number of num_se.
	NumSe *int32 `json:"num_se,omitempty"`

	// Number of num_standby_se.
	NumStandbySe *int32 `json:"num_standby_se,omitempty"`

	// Number of num_vcpus.
	NumVcpus *int32 `json:"num_vcpus,omitempty"`

	// Indicates if the primary SE is being scaled in. This state is now derived from the Virtual Service runtime. Field deprecated in 18.1.5, 18.2.1.
	ScaleinPrimary *bool `json:"scalein_primary,omitempty"`

	// Indicates which SE is being scaled in. This information is now derived from the Virtual Service runtime. Field deprecated in 18.1.5, 18.2.1.
	ScaleinSeUUID *string `json:"scalein_se_uuid,omitempty"`
}
