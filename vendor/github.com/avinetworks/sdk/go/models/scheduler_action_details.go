package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

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
