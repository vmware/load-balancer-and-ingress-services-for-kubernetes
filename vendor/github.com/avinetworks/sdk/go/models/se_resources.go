package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeResources se resources
// swagger:model SeResources
type SeResources struct {

	// Number of cores_per_socket.
	CoresPerSocket *int32 `json:"cores_per_socket,omitempty"`

	// Number of disk.
	// Required: true
	Disk *int32 `json:"disk"`

	// Placeholder for description of property hyper_threading of obj type SeResources field type str  type boolean
	HyperThreading *bool `json:"hyper_threading,omitempty"`

	// Number of memory.
	// Required: true
	Memory *int32 `json:"memory"`

	// Number of num_vcpus.
	// Required: true
	NumVcpus *int32 `json:"num_vcpus"`

	// Number of sockets.
	Sockets *int32 `json:"sockets,omitempty"`
}
