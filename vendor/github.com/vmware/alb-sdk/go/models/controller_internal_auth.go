// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ControllerInternalAuth controller internal auth
// swagger:model ControllerInternalAuth
type ControllerInternalAuth struct {

	// Symmetric keys used for signing/validating the JWT, only allowed with profile_type CONTROLLER_INTERNAL_AUTH. Field introduced in 20.1.6. Minimum of 1 items required. Maximum of 1 items allowed. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SymmetricJwksKeys []*JWSKey `json:"symmetric_jwks_keys,omitempty"`
}
