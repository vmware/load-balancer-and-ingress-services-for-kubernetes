// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbConfig gslb config
// swagger:model GslbConfig
type GslbConfig struct {

	// Frequency with which messages are propagated to Vs Mgr. Value of 0 disables async behavior and RPC are sent inline. Allowed values are 0-5. Field introduced in 22.1.1. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AsyncInterval uint32 `json:"async_interval,omitempty"`

	// Max retries after which the remote site is treated as a fresh start. In fresh start all the configs are downloaded. Allowed values are 1-1024. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ClearOnMaxRetries *uint32 `json:"clear_on_max_retries,omitempty"`

	// Allows enable/disable of GslbService pool groups and pool members from the gslb follower members. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EnableConfigByMembers *bool `json:"enable_config_by_members,omitempty"`

	// Frequency with which errored messages are resynced to follower sites. Value of 0 disables resync behavior. Allowed values are 60-3600. Special values are 0 - Disable. Field introduced in 22.1.1. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ErrorResyncInterval *uint32 `json:"error_resync_interval,omitempty"`

	// This field indicates that this object is replicated across GSLB federation. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IsFederated *bool `json:"is_federated,omitempty"`

	// Mark this Site as leader of GSLB configuration. This site is the one among the Avi sites. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	LeaderClusterUUID *string `json:"leader_cluster_uuid"`

	// This field disables the configuration operations on the leader for all federated objects.  CUD operations on Gslb, GslbService, GslbGeoDbProfile and other federated objects will be rejected. The rest-api disabling helps in upgrade scenarios where we don't want configuration sync operations to the Gslb member when the member is being upgraded.  This configuration programmatically blocks the leader from accepting new Gslb configuration when member sites are undergoing upgrade. . Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MaintenanceMode *bool `json:"maintenance_mode,omitempty"`

	// Name of the GSLB. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// Policy for replicating configuration to the active follower sites. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ReplicationPolicy *ReplicationPolicy `json:"replication_policy,omitempty"`

	// Frequency with which group members communicate. Allowed values are 1-3600. Field introduced in 22.1.1. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SendInterval *uint32 `json:"send_interval,omitempty"`

	// Select Avi site member belonging to this Gslb. Field introduced in 22.1.1. Minimum of 1 items required. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Sites []*GslbSite `json:"sites,omitempty"`

	//  It is a reference to an object of type Tenant. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// This field indicates tenant visibility for GS pool member selection across the Gslb federated objects. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TenantScoped *bool `json:"tenant_scoped,omitempty"`

	// URL of the GSLB. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	URL *string `json:"url,omitempty"`

	// UUID of the GSLB. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`

	// The view-id is used in change-leader mode to differentiate partitioned groups while they have the same GSLB namespace. Each partitioned group will be able to operate independently by using the view-id. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ViewID uint64 `json:"view_id,omitempty"`
}
