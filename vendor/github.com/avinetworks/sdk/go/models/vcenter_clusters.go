package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VcenterClusters vcenter clusters
// swagger:model VcenterClusters
type VcenterClusters struct {

	//  It is a reference to an object of type VIMgrClusterRuntime.
	ClusterRefs []string `json:"cluster_refs,omitempty"`

	// Placeholder for description of property include of obj type VcenterClusters field type str  type boolean
	Include *bool `json:"include,omitempty"`
}
