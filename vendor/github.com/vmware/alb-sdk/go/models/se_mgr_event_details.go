// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeMgrEventDetails se mgr event details
// swagger:model SeMgrEventDetails
type SeMgrEventDetails struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudName *string `json:"cloud_name,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudUUID *string `json:"cloud_uuid,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableState *string `json:"enable_state,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GcpInfo *GcpInfo `json:"gcp_info,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HostName *string `json:"host_name,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HostUUID *string `json:"host_uuid,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Memory *int32 `json:"memory,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MigrateState *string `json:"migrate_state,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	//  Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NewMgmtIP *string `json:"new_mgmt_ip,omitempty"`

	//  Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NewMgmtIp6 *string `json:"new_mgmt_ip6,omitempty"`

	//  Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	OldMgmtIP *string `json:"old_mgmt_ip,omitempty"`

	//  Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	OldMgmtIp6 *string `json:"old_mgmt_ip6,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Reason *string `json:"reason,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeGrpName *string `json:"se_grp_name,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeGrpUUID *string `json:"se_grp_uuid,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Vcpus *int32 `json:"vcpus,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsName []string `json:"vs_name,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsUUID []string `json:"vs_uuid,omitempty"`

	// vSphere HA on cluster enabled. Field introduced in 20.1.7, 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VsphereHaEnabled *bool `json:"vsphere_ha_enabled,omitempty"`

	// This flag is set to true when Cloud Connector has detected an ESX host failure. This flag is set to false when the SE connects back to the controller, or when vSphere HA recovery timeout has occurred. Field introduced in 20.1.7, 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VsphereHaInprogress *bool `json:"vsphere_ha_inprogress,omitempty"`
}
