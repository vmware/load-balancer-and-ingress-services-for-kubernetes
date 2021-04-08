package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// Backup backup
// swagger:model Backup
type Backup struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// BackupConfiguration Information. It is a reference to an object of type BackupConfiguration.
	BackupConfigRef *string `json:"backup_config_ref,omitempty"`

	// The file name of backup.
	// Required: true
	FileName *string `json:"file_name"`

	// URL to download the backup file.
	LocalFileURL *string `json:"local_file_url,omitempty"`

	// URL to download the backup file.
	RemoteFileURL *string `json:"remote_file_url,omitempty"`

	// Scheduler Information. It is a reference to an object of type Scheduler.
	SchedulerRef *string `json:"scheduler_ref,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Unix Timestamp of when the backup file is created.
	Timestamp *string `json:"timestamp,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
