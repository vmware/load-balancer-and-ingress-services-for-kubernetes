// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// FTPProfile f t p profile
// swagger:model FTPProfile
type FTPProfile struct {

	// Deactivate active FTP mode. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DeactivateActive *bool `json:"deactivate_active,omitempty"`

	// Deactivate passive FTP mode. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DeactivatePassive *bool `json:"deactivate_passive,omitempty"`
}
