// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MesosSeResources mesos se resources
// swagger:model MesosSeResources
type MesosSeResources struct {

	// Attribute (Fleet or Mesos) key of Hosts. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	AttributeKey *string `json:"attribute_key"`

	// Attribute (Fleet or Mesos) value of Hosts. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	AttributeValue *string `json:"attribute_value"`

	// Obsolete - ignored. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CPU *float32 `json:"cpu,omitempty"`

	// Obsolete - ignored. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Memory *uint32 `json:"memory,omitempty"`
}
