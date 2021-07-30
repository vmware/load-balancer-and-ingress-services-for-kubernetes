// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// RmRebootSeEventDetails rm reboot se event details
// swagger:model RmRebootSeEventDetails
type RmRebootSeEventDetails struct {

	// reason of RmRebootSeEventDetails.
	Reason *string `json:"reason,omitempty"`

	// se_name of RmRebootSeEventDetails.
	SeName *string `json:"se_name,omitempty"`
}
