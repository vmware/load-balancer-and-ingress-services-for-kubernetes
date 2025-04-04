// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// LogMgrCleanupEventDetails log mgr cleanup event details
// swagger:model LogMgrCleanupEventDetails
type LogMgrCleanupEventDetails struct {

	//  Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CleanupCount *uint32 `json:"cleanup_count,omitempty"`

	//  Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Controller *string `json:"controller,omitempty"`

	//  Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CurrSize *uint64 `json:"curr_size,omitempty"`

	//  Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FromTime *string `json:"from_time,omitempty"`

	//  Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SizeLimit *uint64 `json:"size_limit,omitempty"`
}
