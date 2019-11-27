package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// Scheduler scheduler
// swagger:model Scheduler
type Scheduler struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Backup Configuration to be executed by this scheduler. It is a reference to an object of type BackupConfiguration.
	BackupConfigRef *string `json:"backup_config_ref,omitempty"`

	// Placeholder for description of property enabled of obj type Scheduler field type str  type boolean
	Enabled *bool `json:"enabled,omitempty"`

	// Scheduler end date and time.
	EndDateTime *string `json:"end_date_time,omitempty"`

	// Frequency at which CUSTOM scheduler will run. Allowed values are 0-60.
	Frequency *int32 `json:"frequency,omitempty"`

	// Unit at which CUSTOM scheduler will run. Enum options - SCHEDULER_FREQUENCY_UNIT_MIN, SCHEDULER_FREQUENCY_UNIT_HOUR, SCHEDULER_FREQUENCY_UNIT_DAY, SCHEDULER_FREQUENCY_UNIT_WEEK, SCHEDULER_FREQUENCY_UNIT_MONTH.
	FrequencyUnit *string `json:"frequency_unit,omitempty"`

	// Name of scheduler.
	// Required: true
	Name *string `json:"name"`

	// Scheduler Run Mode. Enum options - RUN_MODE_PERIODIC, RUN_MODE_AT, RUN_MODE_NOW.
	RunMode *string `json:"run_mode,omitempty"`

	// Control script to be executed by this scheduler. It is a reference to an object of type AlertScriptConfig.
	RunScriptRef *string `json:"run_script_ref,omitempty"`

	// Define Scheduler Action. Enum options - SCHEDULER_ACTION_RUN_A_SCRIPT, SCHEDULER_ACTION_BACKUP.
	SchedulerAction *string `json:"scheduler_action,omitempty"`

	// Scheduler start date and time.
	StartDateTime *string `json:"start_date_time,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
