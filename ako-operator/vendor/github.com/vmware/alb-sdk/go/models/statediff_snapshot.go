// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// StatediffSnapshot statediff snapshot
// swagger:model StatediffSnapshot
type StatediffSnapshot struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Name of GSLB object. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	GslbName *string `json:"gslb_name,omitempty"`

	// Reference to base gslb object. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	GslbUUID *string `json:"gslb_uuid,omitempty"`

	// Name of Statediff operation. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`

	// Name of POOL object. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PoolName *string `json:"pool_name,omitempty"`

	// Reference to base pool object. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PoolUUID *string `json:"pool_uuid,omitempty"`

	// Post-Upgrade snapshot for VS. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PostSnapshot *Postsnapshot `json:"post_snapshot,omitempty"`

	// Pre-Upgrade snapshot for VS. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PreSnapshot *Presnapshot `json:"pre_snapshot,omitempty"`

	// Name of SEG object. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeGroupName *string `json:"se_group_name,omitempty"`

	// Reference to base SEG object. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeGroupUUID *string `json:"se_group_uuid,omitempty"`

	// Name of SEG object. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeName *string `json:"se_name,omitempty"`

	// Reference to base SE object. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeUUID *string `json:"se_uuid,omitempty"`

	// Type of snapshot eg. VS_SNAPSHOT, SE_SNAPSHOT etc. Enum options - FB_VS_SNAPSHOT, FB_SE_SNAPSHOT, FB_GSLB_SNAPSHOT, FB_POOL_SNAPSHOT. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SnapshotType *string `json:"snapshot_type,omitempty"`

	// Statediff Operation uuid for identifying the operation. It is a reference to an object of type StatediffOperation. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	StatediffOperationRef *string `json:"statediff_operation_ref,omitempty"`

	// Tenant that this object belongs to. It is a reference to an object of type Tenant. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// unique identifier for Statediff entry. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`

	// Name of VS object. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VsName *string `json:"vs_name,omitempty"`

	// Reference to base VS object. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VsUUID *string `json:"vs_uuid,omitempty"`
}
