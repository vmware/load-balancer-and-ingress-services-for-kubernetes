package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// FailActionBackupPool fail action backup pool
// swagger:model FailActionBackupPool
type FailActionBackupPool struct {

	// Specifies the UUID of the Pool acting as backup pool. It is a reference to an object of type Pool.
	// Required: true
	BackupPoolRef *string `json:"backup_pool_ref"`
}
