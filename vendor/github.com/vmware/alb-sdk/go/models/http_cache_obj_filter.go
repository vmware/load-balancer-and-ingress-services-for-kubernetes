// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

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
