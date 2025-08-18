// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ClusterHAConfig cluster h a config
// swagger:model ClusterHAConfig
type ClusterHAConfig struct {

	// Transport node cluster. Avi derives vSphere HA property from vCenter cluster.If vSphere HA enabled on vCenter cluster, vSphere will handle HA of ServiceEngine VMs in case of underlying ESX failure.Ex MOB  domain-c23. Field introduced in 20.1.7, 21.1.3. Allowed in Enterprise edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ClusterID *string `json:"cluster_id,omitempty"`

	// If this flag set to True, Avi handles ServiceEngine failure irrespective of vSphere HA enabled on vCenter cluster or not. Field introduced in 20.1.7, 21.1.3. Allowed in Enterprise edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	OverrideVsphereHa *bool `json:"override_vsphere_ha,omitempty"`

	// Cluster VM Group name.VM Group name is unique inside cluster. Field introduced in 20.1.7, 21.1.3. Allowed in Enterprise edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	VmgName *string `json:"vmg_name,omitempty"`
}
