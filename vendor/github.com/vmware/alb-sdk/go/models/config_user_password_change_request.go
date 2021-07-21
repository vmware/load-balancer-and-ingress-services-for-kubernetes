// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ConfigUserPasswordChangeRequest config user password change request
// swagger:model ConfigUserPasswordChangeRequest
type ConfigUserPasswordChangeRequest struct {

	// client ip.
	ClientIP *string `json:"client_ip,omitempty"`

	// Password link is sent or rejected.
	Status *string `json:"status,omitempty"`

	// Matched username of email address.
	User *string `json:"user,omitempty"`

	// Email address of user.
	UserEmail *string `json:"user_email,omitempty"`
}
