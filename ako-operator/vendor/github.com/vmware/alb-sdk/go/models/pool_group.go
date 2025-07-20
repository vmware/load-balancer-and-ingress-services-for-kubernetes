// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PoolGroup pool group
// swagger:model PoolGroup
type PoolGroup struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Checksum of cloud configuration for PoolGroup. Internally set by cloud connector. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudConfigCksum *string `json:"cloud_config_cksum,omitempty"`

	//  It is a reference to an object of type Cloud. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudRef *string `json:"cloud_ref,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// Name of the user who created the object. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CreatedBy *string `json:"created_by,omitempty"`

	// Deactivate primary pool for selection when down until it is activated by user via clear poolgroup command. Field introduced in 20.1.7, 21.1.2, 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DeactivatePrimaryPoolOnDown *bool `json:"deactivate_primary_pool_on_down,omitempty"`

	// When setup autoscale manager will automatically promote new pools into production when deployment goals are met. It is a reference to an object of type PoolGroupDeploymentPolicy. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DeploymentPolicyRef *string `json:"deployment_policy_ref,omitempty"`

	// Description of Pool Group. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// Enable HTTP/2 for traffic from VirtualService to all the backend servers in all the pools configured under this PoolGroup. Field deprecated in 30.2.1. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableHttp2 *bool `json:"enable_http2,omitempty"`

	// Enable an action - Close Connection, HTTP Redirect, or Local HTTP Response - when a pool group failure happens. By default, a connection will be closed, in case the pool group experiences a failure. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FailAction *FailAction `json:"fail_action,omitempty"`

	// Whether an implicit set of priority labels is generated. Field introduced in 17.1.9,17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ImplicitPriorityLabels *bool `json:"implicit_priority_labels,omitempty"`

	// List of labels to be used for granular RBAC. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	Markers []*RoleFilterMatchLabel `json:"markers,omitempty"`

	// List of pool group members object of type PoolGroupMember. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Members []*PoolGroupMember `json:"members,omitempty"`

	// The minimum number of servers to distribute traffic to. Allowed values are 1-65535. Special values are 0 - Disable. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 0), Basic edition(Allowed values- 0), Enterprise with Cloud Services edition.
	MinServers *uint32 `json:"min_servers,omitempty"`

	// The name of the pool group. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// UUID of the priority labels. If not provided, pool group member priority label will be interpreted as a number with a larger number considered higher priority. It is a reference to an object of type PriorityLabels. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PriorityLabelsRef *string `json:"priority_labels_ref,omitempty"`

	// Metadata pertaining to the service provided by this PoolGroup. In Openshift/Kubernetes environments, app metadata info is stored. Any user input to this field will be overwritten by Avi Vantage. Field introduced in 17.2.14,18.1.5,18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServiceMetadata *string `json:"service_metadata,omitempty"`

	//  It is a reference to an object of type Tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the pool group. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}
