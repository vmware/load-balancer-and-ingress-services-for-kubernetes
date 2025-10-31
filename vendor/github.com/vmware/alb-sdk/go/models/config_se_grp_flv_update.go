// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ConfigSeGrpFlvUpdate config se grp flv update
// swagger:model ConfigSeGrpFlvUpdate
type ConfigSeGrpFlvUpdate struct {

	// New Flavor Name. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NewFlv *string `json:"new_flv,omitempty"`

	// Old Flavor Name. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OldFlv *string `json:"old_flv,omitempty"`

	// SE Group Name. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeGroupName *string `json:"se_group_name,omitempty"`

	// SE Group UUID. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeGroupUUID *string `json:"se_group_uuid,omitempty"`

	// Tenant Name. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantName *string `json:"tenant_name,omitempty"`

	// Tenant UUID. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantUUID *string `json:"tenant_uuid,omitempty"`
}
