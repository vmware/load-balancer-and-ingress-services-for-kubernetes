// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AuthorizationAction authorization action
// swagger:model AuthorizationAction
type AuthorizationAction struct {

	// HTTP status code to use for local response when an policy rule is matched. Enum options - HTTP_RESPONSE_STATUS_CODE_401, HTTP_RESPONSE_STATUS_CODE_403. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	StatusCode *string `json:"status_code,omitempty"`

	// Defines the action taken when an authorization policy rule is matched. By default, access is allowed to the requested resource. Enum options - ALLOW_ACCESS, CLOSE_CONNECTION, HTTP_LOCAL_RESPONSE. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Type *string `json:"type,omitempty"`
}
