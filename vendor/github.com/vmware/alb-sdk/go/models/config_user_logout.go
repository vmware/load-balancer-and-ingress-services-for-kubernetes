// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ConfigUserLogout config user logout
// swagger:model ConfigUserLogout
type ConfigUserLogout struct {

	// client ip.
	ClientIP *string `json:"client_ip,omitempty"`

	// error message if logging out failed.
	ErrorMessage *string `json:"error_message,omitempty"`

	// Local user. Field introduced in 17.1.1.
	Local *bool `json:"local,omitempty"`

	// Status.
	Status *string `json:"status,omitempty"`

	// Request user.
	User *string `json:"user,omitempty"`
}
