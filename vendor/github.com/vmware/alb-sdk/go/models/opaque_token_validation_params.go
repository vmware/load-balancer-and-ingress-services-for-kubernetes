// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// OpaqueTokenValidationParams opaque token validation params
// swagger:model OpaqueTokenValidationParams
type OpaqueTokenValidationParams struct {

	// Resource server specific identifier used to validate against introspection endpoint when access token is opaque. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	ServerID *string `json:"server_id"`

	// Resource server specific password/secret. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	ServerSecret *string `json:"server_secret"`
}
