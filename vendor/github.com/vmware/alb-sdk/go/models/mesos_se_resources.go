// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MesosSeResources mesos se resources
// swagger:model MesosSeResources
type MesosSeResources struct {

	// Attribute (Fleet or Mesos) key of Hosts.
	// Required: true
	AttributeKey *string `json:"attribute_key"`

	// Attribute (Fleet or Mesos) value of Hosts.
	// Required: true
	AttributeValue *string `json:"attribute_value"`

	// Obsolete - ignored.
	CPU *float32 `json:"cpu,omitempty"`

	// Obsolete - ignored.
	Memory *int32 `json:"memory,omitempty"`
}
