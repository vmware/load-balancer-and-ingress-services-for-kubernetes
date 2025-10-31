// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// LogMgrUberEventDetails log mgr uber event details
// swagger:model LogMgrUberEventDetails
type LogMgrUberEventDetails struct {

	//  Enum options - X_ENUM_1, X_ENUM_2, X_ENUM_3, X_ENUM_4. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	XEnum *string `json:"x_enum,omitempty"`

	//  Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	XFloat *float32 `json:"x_float,omitempty"`

	//  Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	XHex *uint64 `json:"x_hex,omitempty"`

	//  Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	XInt *uint64 `json:"x_int,omitempty"`

	//  Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	XMsg *UberEnumMessage1 `json:"x_msg,omitempty"`

	//  Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	XRmsg []*UberEnumMessage1 `json:"x_rmsg,omitempty"`

	//  Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	XStr *string `json:"x_str,omitempty"`

	//  Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	XX []int64 `json:"x_x,omitempty,omitempty"`
}
