// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ServiceEngineRuntimeSummary service engine runtime summary
// swagger:model ServiceEngineRuntimeSummary
type ServiceEngineRuntimeSummary struct {

	//  Enum options - ACTIVE_STANDBY_SE_1, ACTIVE_STANDBY_SE_2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ActiveTags []string `json:"active_tags,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AtCurrVer *bool `json:"at_curr_ver,omitempty"`

	// Indicates if at least 1 BGP peer with advertise_vip is UP and at least 1 BGP peer with advertise_snat_ip is UP if there are such peers configured. Flag will be set to false if the condition above is not true for any of the VRFs configured on the SE. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	BgpPeersUp *bool `json:"bgp_peers_up,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GatewayUp *bool `json:"gateway_up,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HbStatus *SeHbStatus `json:"hb_status,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	InbandMgmt *bool `json:"inband_mgmt,omitempty"`

	// Indicates the License state of the SE. Enum options - LICENSE_STATE_INSUFFICIENT_RESOURCES, LICENSE_STATE_LICENSED, LICENSE_STATE_AWAITING_RESPONSE, LICENSE_STATE_UNDETERMINED. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LicenseState *string `json:"license_state,omitempty"`

	// Number of Service Cores assigned to the SE by License Manager. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LicensedServiceCores *float64 `json:"licensed_service_cores,omitempty"`

	// This state is used to indicate the current state of disable SE process. Enum options - SE_MIGRATE_STATE_IDLE, SE_MIGRATE_STATE_STARTED, SE_MIGRATE_STATE_FINISHED_WITH_FAILURE, SE_MIGRATE_STATE_FINISHED. Field introduced in 17.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MigrateState *string `json:"migrate_state,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OnlineSince *string `json:"online_since,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OperStatus *OperationalStatus `json:"oper_status,omitempty"`

	//  Enum options - SE_POWER_OFF, SE_POWER_ON, SE_SUSPENDED. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PowerState *string `json:"power_state,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeConnected *bool `json:"se_connected,omitempty"`

	// Indicates SE reboot following SE group change is pending. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeGrpRebootPending *bool `json:"se_grp_reboot_pending,omitempty"`

	//  Field introduced in 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SufficientMemory *bool `json:"sufficient_memory,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Version *string `json:"version,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VinfraDiscovered *bool `json:"vinfra_discovered,omitempty"`

	// vSphere HA on cluster enabled. Field introduced in 20.1.7, 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VsphereHaEnabled *bool `json:"vsphere_ha_enabled,omitempty"`

	// This flag is set to true when Cloud Connector has detected an ESX host failure. This flag is set to false when the SE connects back to the controller, or when vSphere HA recovery timeout has occurred. Field introduced in 20.1.7, 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VsphereHaInprogress *bool `json:"vsphere_ha_inprogress,omitempty"`

	// vSphere HA monitor job has been created or is running. Field introduced in 20.1.7, 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VsphereHaJobActive *bool `json:"vsphere_ha_job_active,omitempty"`
}
