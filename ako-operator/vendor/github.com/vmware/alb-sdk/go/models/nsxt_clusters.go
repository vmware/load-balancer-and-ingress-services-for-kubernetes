// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NsxtClusters nsxt clusters
// swagger:model NsxtClusters
type NsxtClusters struct {

	// List of transport node clusters. Field introduced in 20.1.6. Allowed in Enterprise edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ClusterIds []string `json:"cluster_ids,omitempty"`

	// Include or Exclude. Field introduced in 20.1.6. Allowed in Enterprise edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	Include *bool `json:"include,omitempty"`
}
