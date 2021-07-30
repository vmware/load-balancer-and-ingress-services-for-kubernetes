// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VcenterClusters vcenter clusters
// swagger:model VcenterClusters
type VcenterClusters struct {

	//  It is a reference to an object of type VIMgrClusterRuntime.
	ClusterRefs []string `json:"cluster_refs,omitempty"`

	// Placeholder for description of property include of obj type VcenterClusters field type str  type boolean
	Include *bool `json:"include,omitempty"`
}
