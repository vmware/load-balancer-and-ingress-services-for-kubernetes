// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SessionHeaders session headers
// swagger:model SessionHeaders
type SessionHeaders struct {

	// Header Name. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`

	// Header Value. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Value *string `json:"value,omitempty"`
}
