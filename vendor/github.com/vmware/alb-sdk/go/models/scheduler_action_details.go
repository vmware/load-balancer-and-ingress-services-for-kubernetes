// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SchedulerActionDetails scheduler action details
// swagger:model SchedulerActionDetails
type SchedulerActionDetails struct {

	// backup_uri of SchedulerActionDetails.
	BackupURI []string `json:"backup_uri,omitempty"`

	// control_script_output of SchedulerActionDetails.
	ControlScriptOutput *string `json:"control_script_output,omitempty"`

	// execution_datestamp of SchedulerActionDetails.
	ExecutionDatestamp *string `json:"execution_datestamp,omitempty"`

	// Unique object identifier of scheduler.
	SchedulerUUID *string `json:"scheduler_uuid,omitempty"`

	// status of SchedulerActionDetails.
	Status *string `json:"status,omitempty"`
}
