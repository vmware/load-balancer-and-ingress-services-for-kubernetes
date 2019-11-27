package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// Cluster cluster
// swagger:model Cluster
type Cluster struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// Placeholder for description of property nodes of obj type Cluster field type str  type object
	Nodes []*ClusterNode `json:"nodes,omitempty"`

	// Re-join cluster nodes automatically in the event one of the node is reset to factory.
	RejoinNodesAutomatically *bool `json:"rejoin_nodes_automatically,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`

	// A virtual IP address. This IP address will be dynamically reconfigured so that it always is the IP of the cluster leader.
	VirtualIP *IPAddr `json:"virtual_ip,omitempty"`
}
