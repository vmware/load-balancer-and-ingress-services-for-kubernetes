// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ServerAutoScalePolicy server auto scale policy
// swagger:model ServerAutoScalePolicy
type ServerAutoScalePolicy struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// Delay in minutes after which a down server will be removed from Pool. Value 0 disables this functionality. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DelayForServerGarbageCollection *uint32 `json:"delay_for_server_garbage_collection,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// Use Avi intelligent autoscale algorithm where autoscale is performed by comparing load on the pool against estimated capacity of all the servers. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IntelligentAutoscale *bool `json:"intelligent_autoscale,omitempty"`

	// Maximum extra capacity as percentage of load used by the intelligent scheme. Scale-in is triggered when available capacity is more than this margin. Allowed values are 1-99. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IntelligentScaleinMargin *uint32 `json:"intelligent_scalein_margin,omitempty"`

	// Minimum extra capacity as percentage of load used by the intelligent scheme. Scale-out is triggered when available capacity is less than this margin. Allowed values are 1-99. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IntelligentScaleoutMargin *uint32 `json:"intelligent_scaleout_margin,omitempty"`

	// List of labels to be used for granular RBAC. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	Markers []*RoleFilterMatchLabel `json:"markers,omitempty"`

	// Maximum number of servers to scale-in simultaneously. The actual number of servers to scale-in is chosen such that target number of servers is always more than or equal to the min_size. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxScaleinAdjustmentStep *uint32 `json:"max_scalein_adjustment_step,omitempty"`

	// Maximum number of servers to scale-out simultaneously. The actual number of servers to scale-out is chosen such that target number of servers is always less than or equal to the max_size. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxScaleoutAdjustmentStep *uint32 `json:"max_scaleout_adjustment_step,omitempty"`

	// Maximum number of servers after scale-out. Allowed values are 0-400. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxSize *uint32 `json:"max_size,omitempty"`

	// No scale-in happens once number of operationally up servers reach min_servers. Allowed values are 0-400. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MinSize *uint32 `json:"min_size,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// Trigger scale-in when alerts due to any of these Alert configurations are raised. It is a reference to an object of type AlertConfig. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ScaleinAlertconfigRefs []string `json:"scalein_alertconfig_refs,omitempty"`

	// Cooldown period during which no new scale-in is triggered to allow previous scale-in to successfully complete. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ScaleinCooldown *uint32 `json:"scalein_cooldown,omitempty"`

	// Trigger scale-out when alerts due to any of these Alert configurations are raised. It is a reference to an object of type AlertConfig. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ScaleoutAlertconfigRefs []string `json:"scaleout_alertconfig_refs,omitempty"`

	// Cooldown period during which no new scale-out is triggered to allow previous scale-out to successfully complete. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ScaleoutCooldown *uint32 `json:"scaleout_cooldown,omitempty"`

	// Scheduled-based scale-in/out policy. During scheduled intervals, metrics based autoscale is not enabled and number of servers will be solely derived from ScheduleScale policy. Field introduced in 21.1.1. Maximum of 1 items allowed. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ScheduledScalings []*ScheduledScaling `json:"scheduled_scalings,omitempty"`

	//  It is a reference to an object of type Tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Use predicted load rather than current load. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UsePredictedLoad *bool `json:"use_predicted_load,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}
