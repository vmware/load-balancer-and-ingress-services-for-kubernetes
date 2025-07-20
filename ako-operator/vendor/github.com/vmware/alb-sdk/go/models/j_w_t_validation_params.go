// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// JWTValidationParams j w t validation params
// swagger:model JWTValidationParams
type JWTValidationParams struct {

	// Audience parameter used for validation using JWT token. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Audience *string `json:"audience"`
}
