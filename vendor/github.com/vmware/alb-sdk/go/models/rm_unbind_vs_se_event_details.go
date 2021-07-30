// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// RmUnbindVsSeEventDetails rm unbind vs se event details
// swagger:model RmUnbindVsSeEventDetails
type RmUnbindVsSeEventDetails struct {

	// ip of RmUnbindVsSeEventDetails.
	IP *string `json:"ip,omitempty"`

	// ip6 of RmUnbindVsSeEventDetails.
	Ip6 *string `json:"ip6,omitempty"`

	// reason of RmUnbindVsSeEventDetails.
	Reason *string `json:"reason,omitempty"`

	// se_name of RmUnbindVsSeEventDetails.
	SeName *string `json:"se_name,omitempty"`

	// vs_name of RmUnbindVsSeEventDetails.
	VsName *string `json:"vs_name,omitempty"`

	// Unique object identifier of vs.
	VsUUID *string `json:"vs_uuid,omitempty"`
}
