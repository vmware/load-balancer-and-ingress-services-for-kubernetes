// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ConfigVersionStatus config version status
// swagger:model ConfigVersionStatus
type ConfigVersionStatus struct {

	// Type of replication event. Enum options - DNSVS, OBJECT_CONFIG_VERSION. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EventType *string `json:"event_type,omitempty"`

	// Name of config object. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ObjName *string `json:"obj_name,omitempty"`

	// UUID of config object. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ObjUUID *string `json:"obj_uuid,omitempty"`
}
