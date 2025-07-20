// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// Action action
// swagger:model Action
type Action struct {

	// A description of the change to this object. This field is opaque to the caller, it should not be interpreted or modified. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Data *string `json:"data"`

	// The referenced object on which this action will be applied. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	URLRef *string `json:"url_ref"`
}
