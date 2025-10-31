// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PoolGroupDeploymentPolicy pool group deployment policy
// swagger:model PoolGroupDeploymentPolicy
type PoolGroupDeploymentPolicy struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// It will automatically disable old production pools once there is a new production candidate. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AutoDisableOldProdPools *bool `json:"auto_disable_old_prod_pools,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// Duration of evaluation period for automatic deployment. Allowed values are 60-86400. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EvaluationDuration *uint32 `json:"evaluation_duration,omitempty"`

	// List of labels to be used for granular RBAC. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	Markers []*RoleFilterMatchLabel `json:"markers,omitempty"`

	// The name of the pool group deployment policy. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Rules []*PGDeploymentRule `json:"rules,omitempty"`

	// deployment scheme. Enum options - BLUE_GREEN, CANARY. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Scheme *string `json:"scheme,omitempty"`

	// Target traffic ratio before pool is made production. Allowed values are 1-100. Unit is RATIO. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TargetTestTrafficRatio *uint32 `json:"target_test_traffic_ratio,omitempty"`

	//  It is a reference to an object of type Tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Ratio of the traffic that is sent to the pool under test. test ratio of 100 means blue green. Allowed values are 1-100. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TestTrafficRatioRampup *uint32 `json:"test_traffic_ratio_rampup,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the pool group deployment policy. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`

	// Webhook configured with URL that Avi controller will pass back information about pool group, old and new pool information and current deployment rule results. It is a reference to an object of type Webhook. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	WebhookRef *string `json:"webhook_ref,omitempty"`
}
