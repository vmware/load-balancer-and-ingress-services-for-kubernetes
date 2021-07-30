// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// RmSeBootupFailEventDetails rm se bootup fail event details
// swagger:model RmSeBootupFailEventDetails
type RmSeBootupFailEventDetails struct {

	// host_name of RmSeBootupFailEventDetails.
	HostName *string `json:"host_name,omitempty"`

	// reason of RmSeBootupFailEventDetails.
	Reason *string `json:"reason,omitempty"`

	// se_name of RmSeBootupFailEventDetails.
	SeName *string `json:"se_name,omitempty"`
}
