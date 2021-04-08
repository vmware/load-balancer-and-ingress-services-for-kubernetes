package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

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
