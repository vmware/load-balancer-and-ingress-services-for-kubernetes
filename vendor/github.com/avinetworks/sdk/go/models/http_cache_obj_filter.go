package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HTTPCacheObjFilter Http cache obj filter
// swagger:model HttpCacheObjFilter
type HTTPCacheObjFilter struct {

	// HTTP cache object's exact key.
	Key *string `json:"key,omitempty"`

	// HTTP cache object's exact raw key.
	RawKey *string `json:"raw_key,omitempty"`

	// HTTP cache object's resource name.
	ResourceName *string `json:"resource_name,omitempty"`

	// objects with resource type.
	ResourceType *string `json:"resource_type,omitempty"`

	// HTTP cache object type. Enum options - CO_ALL, CO_IN, CO_OUT.
	Type *string `json:"type,omitempty"`
}
