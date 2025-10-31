// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// Backup backup
// swagger:model Backup
type Backup struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// BackupConfiguration Information. It is a reference to an object of type BackupConfiguration. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	BackupConfigRef *string `json:"backup_config_ref,omitempty"`

	// The file name of backup. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	FileName *string `json:"file_name"`

	// URL to download the backup file. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LocalFileURL *string `json:"local_file_url,omitempty"`

	// URL to download the backup file. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RemoteFileURL *string `json:"remote_file_url,omitempty"`

	// Scheduler Information. It is a reference to an object of type Scheduler. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SchedulerRef *string `json:"scheduler_ref,omitempty"`

	//  It is a reference to an object of type Tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Unix Timestamp of when the backup file is created. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Timestamp *string `json:"timestamp,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}
