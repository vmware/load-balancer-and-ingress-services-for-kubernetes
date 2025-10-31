// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AuthenticationAction authentication action
// swagger:model AuthenticationAction
type AuthenticationAction struct {

	// Authentication Action to be taken for a matched Rule. Enum options - SKIP_AUTHENTICATION, USE_DEFAULT_AUTHENTICATION. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Type *string `json:"type"`
}
