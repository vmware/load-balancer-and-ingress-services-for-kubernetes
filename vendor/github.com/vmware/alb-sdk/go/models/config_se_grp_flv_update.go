package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ConfigSeGrpFlvUpdate config se grp flv update
// swagger:model ConfigSeGrpFlvUpdate
type ConfigSeGrpFlvUpdate struct {

	// New Flavor Name.
	NewFlv *string `json:"new_flv,omitempty"`

	// Old Flavor Name.
	OldFlv *string `json:"old_flv,omitempty"`

	// SE Group Name.
	SeGroupName *string `json:"se_group_name,omitempty"`

	// SE Group UUID.
	SeGroupUUID *string `json:"se_group_uuid,omitempty"`

	// Tenant Name.
	TenantName *string `json:"tenant_name,omitempty"`

	// Tenant UUID.
	TenantUUID *string `json:"tenant_uuid,omitempty"`
}
