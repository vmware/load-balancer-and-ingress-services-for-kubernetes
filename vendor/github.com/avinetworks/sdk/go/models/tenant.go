package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// Tenant tenant
// swagger:model Tenant
type Tenant struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Placeholder for description of property config_settings of obj type Tenant field type str  type object
	ConfigSettings *TenantConfiguration `json:"config_settings,omitempty"`

	// Creator of this tenant.
	CreatedBy *string `json:"created_by,omitempty"`

	// User defined description for the object.
	Description *string `json:"description,omitempty"`

	// Placeholder for description of property local of obj type Tenant field type str  type boolean
	Local *bool `json:"local,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// Suggestive pool of key value pairs for recommending assignment of labels to objects in the User Interface. Every entry is unique in both key and value. Field introduced in 20.1.3.
	SuggestedObjectLabels []*TenantLabel `json:"suggested_object_labels,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
