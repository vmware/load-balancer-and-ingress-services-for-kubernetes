// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ConfigUserPasswordChangeRequest config user password change request
// swagger:model ConfigUserPasswordChangeRequest
type ConfigUserPasswordChangeRequest struct {

	// client ip. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClientIP *string `json:"client_ip,omitempty"`

	// Password link is sent or rejected. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Status *string `json:"status,omitempty"`

	// Matched username of email address. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	User *string `json:"user,omitempty"`

	// Email address of user. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UserEmail *string `json:"user_email,omitempty"`
}
