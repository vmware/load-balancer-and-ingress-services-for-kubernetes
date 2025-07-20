// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// FailActionBackupPool fail action backup pool
// swagger:model FailActionBackupPool
type FailActionBackupPool struct {

	// Specifies the UUID of the Pool acting as backup pool. It is a reference to an object of type Pool. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	BackupPoolRef *string `json:"backup_pool_ref"`
}
