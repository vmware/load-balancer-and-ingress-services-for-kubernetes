package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeProperties se properties
// swagger:model SeProperties
type SeProperties struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Placeholder for description of property se_agent_properties of obj type SeProperties field type str  type object
	SeAgentProperties *SeAgentProperties `json:"se_agent_properties,omitempty"`

	// Placeholder for description of property se_bootup_properties of obj type SeProperties field type str  type object
	SeBootupProperties *SeBootupProperties `json:"se_bootup_properties,omitempty"`

	// Placeholder for description of property se_runtime_properties of obj type SeProperties field type str  type object
	SeRuntimeProperties *SeRuntimeProperties `json:"se_runtime_properties,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
