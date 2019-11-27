package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// PoolGroup pool group
// swagger:model PoolGroup
type PoolGroup struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Checksum of cloud configuration for PoolGroup. Internally set by cloud connector.
	CloudConfigCksum *string `json:"cloud_config_cksum,omitempty"`

	//  It is a reference to an object of type Cloud.
	CloudRef *string `json:"cloud_ref,omitempty"`

	// Name of the user who created the object.
	CreatedBy *string `json:"created_by,omitempty"`

	// When setup autoscale manager will automatically promote new pools into production when deployment goals are met. It is a reference to an object of type PoolGroupDeploymentPolicy.
	DeploymentPolicyRef *string `json:"deployment_policy_ref,omitempty"`

	// Description of Pool Group.
	Description *string `json:"description,omitempty"`

	// Enable an action - Close Connection, HTTP Redirect, or Local HTTP Response - when a pool group failure happens. By default, a connection will be closed, in case the pool group experiences a failure.
	FailAction *FailAction `json:"fail_action,omitempty"`

	// Whether an implicit set of priority labels is generated. Field introduced in 17.1.9,17.2.3.
	ImplicitPriorityLabels *bool `json:"implicit_priority_labels,omitempty"`

	// List of pool group members object of type PoolGroupMember.
	Members []*PoolGroupMember `json:"members,omitempty"`

	// The minimum number of servers to distribute traffic to. Allowed values are 1-65535. Special values are 0 - 'Disable'.
	MinServers *int32 `json:"min_servers,omitempty"`

	// The name of the pool group.
	// Required: true
	Name *string `json:"name"`

	// UUID of the priority labels. If not provided, pool group member priority label will be interpreted as a number with a larger number considered higher priority. It is a reference to an object of type PriorityLabels.
	PriorityLabelsRef *string `json:"priority_labels_ref,omitempty"`

	// Metadata pertaining to the service provided by this PoolGroup. In Openshift/Kubernetes environments, app metadata info is stored. Any user input to this field will be overwritten by Avi Vantage. Field introduced in 17.2.14,18.1.5,18.2.1.
	ServiceMetadata *string `json:"service_metadata,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the pool group.
	UUID *string `json:"uuid,omitempty"`
}
