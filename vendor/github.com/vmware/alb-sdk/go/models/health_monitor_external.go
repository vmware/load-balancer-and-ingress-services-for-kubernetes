// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HealthMonitorExternal health monitor external
// swagger:model HealthMonitorExternal
type HealthMonitorExternal struct {

	// Command script provided inline. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	CommandCode *string `json:"command_code"`

	// Optional arguments to feed into the script. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CommandParameters *string `json:"command_parameters,omitempty"`

	// Path of external health monitor script. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CommandPath *string `json:"command_path,omitempty"`

	// Environment variables to be fed into the script. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CommandVariables *string `json:"command_variables,omitempty"`
}
