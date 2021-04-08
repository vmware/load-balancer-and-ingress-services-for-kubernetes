package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// UpgradeOpsEntry upgrade ops entry
// swagger:model UpgradeOpsEntry
type UpgradeOpsEntry struct {

	// Name of the system such as cluster name, se group name and se name. Field introduced in 18.2.6.
	Name *string `json:"name,omitempty"`

	// Describes the system  controller or se-group or se. Enum options - NODE_CONTROLLER_CLUSTER, NODE_SE_GROUP, NODE_SE_TYPE. Field introduced in 18.2.6.
	NodeType *string `json:"node_type,omitempty"`

	// Cloud that this object belongs to. It is a reference to an object of type Cloud. Field introduced in 18.2.6.
	ObjCloudRef *string `json:"obj_cloud_ref,omitempty"`

	// Parameters associated with the upgrade ops request. Field introduced in 18.2.6.
	Params *UpgradeOpsParam `json:"params,omitempty"`

	// Tenant that this object belongs to. It is a reference to an object of type Tenant. Field introduced in 18.2.6.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Current Upgrade Status info for this node. Field introduced in 18.2.6.
	UpgradeInfo *UpgradeStatusInfo `json:"upgrade_info,omitempty"`

	// Identifies the upgrade operations. Enum options - UPGRADE, PATCH, ROLLBACK, ROLLBACKPATCH, SEGROUP_RESUME. Field introduced in 18.2.6.
	UpgradeOps *string `json:"upgrade_ops,omitempty"`

	// Uuid identifier for the system such as cluster, se group and se. Field introduced in 18.2.6.
	UUID *string `json:"uuid,omitempty"`
}
