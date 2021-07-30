// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ControllerFaults controller faults
// swagger:model ControllerFaults
type ControllerFaults struct {

	// Enable backup scheduler faults. Field introduced in 20.1.6.
	BackupSchedulerFaults *bool `json:"backup_scheduler_faults,omitempty"`

	// Enable cluster faults. Field introduced in 20.1.6.
	ClusterFaults *bool `json:"cluster_faults,omitempty"`

	// Enable deprecated api version faults. Field introduced in 20.1.6.
	DeprecatedAPIVersionFaults *bool `json:"deprecated_api_version_faults,omitempty"`

	// Enable license faults. Field introduced in 20.1.6.
	LicenseFaults *bool `json:"license_faults,omitempty"`

	// Enable DB migration faults. Field introduced in 20.1.6.
	MigrationFaults *bool `json:"migration_faults,omitempty"`

	// Enable SSL Profile faults. Field introduced in 20.1.6.
	SslprofileFaults *bool `json:"sslprofile_faults,omitempty"`
}
