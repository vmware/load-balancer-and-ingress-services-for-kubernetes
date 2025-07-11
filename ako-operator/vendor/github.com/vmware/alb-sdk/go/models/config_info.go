// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ConfigInfo config info
// swagger:model ConfigInfo
type ConfigInfo struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Queue []*VersionInfo `json:"queue,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ReaderCount *uint32 `json:"reader_count,omitempty"`

	//  Enum options - REPL_NONE, REPL_ENABLED, REPL_DISABLED. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	State *string `json:"state,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	WriterCount *uint32 `json:"writer_count,omitempty"`
}
