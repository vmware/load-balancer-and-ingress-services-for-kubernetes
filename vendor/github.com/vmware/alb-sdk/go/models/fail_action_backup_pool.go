// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// FailActionBackupPool fail action backup pool
// swagger:model FailActionBackupPool
type FailActionBackupPool struct {

	// Specifies the UUID of the Pool acting as backup pool. It is a reference to an object of type Pool.
	// Required: true
	BackupPoolRef *string `json:"backup_pool_ref"`
}
