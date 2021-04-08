package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

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
