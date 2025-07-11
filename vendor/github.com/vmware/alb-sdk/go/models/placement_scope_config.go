// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PlacementScopeConfig placement scope config
// swagger:model PlacementScopeConfig
type PlacementScopeConfig struct {

	// Cluster vSphere HA configuration. Field introduced in 20.1.7, 21.1.3. Allowed in Enterprise edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	Clusters []*ClusterHAConfig `json:"clusters,omitempty"`

	// List of transport node clusters include or exclude. Field introduced in 20.1.6. Allowed in Enterprise edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	NsxtClusters *NsxtClusters `json:"nsxt_clusters,omitempty"`

	// List of shared datastores to include or exclude. Field introduced in 20.1.2. Allowed in Enterprise edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	NsxtDatastores *NsxtDatastores `json:"nsxt_datastores,omitempty"`

	// List of transport nodes include or exclude. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NsxtHosts *NsxtHosts `json:"nsxt_hosts,omitempty"`

	// Folder to place all the Service Engine virtual machines in vCenter. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterFolder *string `json:"vcenter_folder,omitempty"`

	// VCenter server configuration. It is a reference to an object of type VCenterServer. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	VcenterRef *string `json:"vcenter_ref"`
}
