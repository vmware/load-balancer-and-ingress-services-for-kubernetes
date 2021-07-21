// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PlacementScopeConfig placement scope config
// swagger:model PlacementScopeConfig
type PlacementScopeConfig struct {

	// List of transport node clusters include or exclude. Field introduced in 20.1.6.
	NsxtClusters *NsxtClusters `json:"nsxt_clusters,omitempty"`

	// List of shared datastores to include or exclude. Field introduced in 20.1.2. Allowed in Basic edition, Enterprise edition.
	NsxtDatastores *NsxtDatastores `json:"nsxt_datastores,omitempty"`

	// List of transport nodes include or exclude. Field introduced in 20.1.1.
	NsxtHosts *NsxtHosts `json:"nsxt_hosts,omitempty"`

	// Folder to place all the Service Engine virtual machines in vCenter. Field introduced in 20.1.1.
	VcenterFolder *string `json:"vcenter_folder,omitempty"`

	// VCenter server configuration. It is a reference to an object of type VCenterServer. Field introduced in 20.1.1.
	// Required: true
	VcenterRef *string `json:"vcenter_ref"`
}
