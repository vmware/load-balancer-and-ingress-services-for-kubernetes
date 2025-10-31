// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeDebugModeEventDetail se debug mode event detail
// swagger:model SeDebugModeEventDetail
type SeDebugModeEventDetail struct {

	// Description of the event. Field introduced in 22.1.6. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// Name of the SE, reporting this event. It is a reference to an object of type ServiceEngine. Field introduced in 22.1.6. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeName *string `json:"se_name,omitempty"`

	// UUID of the SE, responsible for this event. It is a reference to an object of type ServiceEngine. Field introduced in 22.1.6. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeRef *string `json:"se_ref,omitempty"`
}
