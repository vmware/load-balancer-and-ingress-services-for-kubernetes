// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PoolGroupConfig pool group config
// swagger:model PoolGroupConfig
type PoolGroupConfig struct {

	//  It is a reference to an object of type Cloud. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CloudRef *string `json:"cloud_ref,omitempty"`

	// Deactivate primary pool for selection when down until it is activated by user via clear poolgroup command. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DeactivatePrimaryPoolOnDown *bool `json:"deactivate_primary_pool_on_down,omitempty"`

	// When setup autoscale manager will automatically promote new pools into production when deployment goals are met. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DeploymentPolicyRef *string `json:"deployment_policy_ref,omitempty"`

	// Enable HTTP/2 for traffic from VirtualService to all the backend servers in all the pools configured under this PoolGroup. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EnableHttp2 *bool `json:"enable_http2,omitempty"`

	// Enable an action - Close Connection, HTTP Redirect, or Local HTTP Response - when a pool group failure happens. By default, a connection will be closed, in case the pool group experiences a failure. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FailAction *FailAction `json:"fail_action,omitempty"`

	// Whether an implicit set of priority labels is generated. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ImplicitPriorityLabels *bool `json:"implicit_priority_labels,omitempty"`

	// List of labels to be used for granular RBAC. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Markers []*RoleFilterMatchLabel `json:"markers,omitempty"`

	// List of pool group members object of type PoolGroupMember. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Members []*PoolGroupMember `json:"members,omitempty"`

	// The minimum number of servers to distribute traffic to. Allowed values are 1-65535. Special values are 0 - Disable. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MinServers uint32 `json:"min_servers,omitempty"`

	// The name of the pool group. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	//  It is a reference to an object of type Tenant. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// URL of the pool grop. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	URL *string `json:"url,omitempty"`

	// UUID of the pool group. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}
