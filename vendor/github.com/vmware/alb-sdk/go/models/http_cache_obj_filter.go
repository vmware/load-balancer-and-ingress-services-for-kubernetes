// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HTTPCacheObjFilter Http cache obj filter
// swagger:model HttpCacheObjFilter
type HTTPCacheObjFilter struct {

	// HTTP cache object's exact key. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Key *string `json:"key,omitempty"`

	// HTTP cache object's exact raw key. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RawKey *string `json:"raw_key,omitempty"`

	// HTTP cache object's resource name. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ResourceName *string `json:"resource_name,omitempty"`

	// objects with resource type. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ResourceType *string `json:"resource_type,omitempty"`

	// HTTP cache object type. Enum options - CO_ALL, CO_IN, CO_OUT. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Type *string `json:"type,omitempty"`
}
