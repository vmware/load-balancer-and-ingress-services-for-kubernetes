// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HealthMonitorExternal health monitor external
// swagger:model HealthMonitorExternal
type HealthMonitorExternal struct {

	// Command script provided inline.
	// Required: true
	CommandCode *string `json:"command_code"`

	// Optional arguments to feed into the script.
	CommandParameters *string `json:"command_parameters,omitempty"`

	// Path of external health monitor script.
	CommandPath *string `json:"command_path,omitempty"`

	// Environment variables to be fed into the script.
	CommandVariables *string `json:"command_variables,omitempty"`
}
