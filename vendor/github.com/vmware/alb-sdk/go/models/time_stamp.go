// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// TimeStamp time stamp
// swagger:model TimeStamp
type TimeStamp struct {

	// Number of secs.
	// Required: true
	Secs *int64 `json:"secs"`

	// Number of usecs.
	// Required: true
	Usecs *int64 `json:"usecs"`
}
