// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// Scheduler scheduler
// swagger:model Scheduler
type Scheduler struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Backup Configuration to be executed by this scheduler. It is a reference to an object of type BackupConfiguration. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	BackupConfigRef *string `json:"backup_config_ref,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Enabled *bool `json:"enabled,omitempty"`

	// Scheduler end date and time. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EndDateTime *string `json:"end_date_time,omitempty"`

	// Frequency at which CUSTOM scheduler will run. Allowed values are 0-60. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Frequency *uint32 `json:"frequency,omitempty"`

	// Unit at which CUSTOM scheduler will run. Enum options - SCHEDULER_FREQUENCY_UNIT_MIN, SCHEDULER_FREQUENCY_UNIT_HOUR, SCHEDULER_FREQUENCY_UNIT_DAY, SCHEDULER_FREQUENCY_UNIT_WEEK, SCHEDULER_FREQUENCY_UNIT_MONTH. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FrequencyUnit *string `json:"frequency_unit,omitempty"`

	// Name of scheduler. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// Scheduler Run Mode. Enum options - RUN_MODE_PERIODIC, RUN_MODE_AT, RUN_MODE_NOW. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RunMode *string `json:"run_mode,omitempty"`

	// Control script to be executed by this scheduler. It is a reference to an object of type AlertScriptConfig. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RunScriptRef *string `json:"run_script_ref,omitempty"`

	// Define Scheduler Action. Enum options - SCHEDULER_ACTION_RUN_A_SCRIPT, SCHEDULER_ACTION_BACKUP. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SchedulerAction *string `json:"scheduler_action,omitempty"`

	// Scheduler start date and time. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	StartDateTime *string `json:"start_date_time,omitempty"`

	//  It is a reference to an object of type Tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}
