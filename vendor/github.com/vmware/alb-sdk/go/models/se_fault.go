// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeFault se fault
// swagger:model SeFault
type SeFault struct {

	// Optional 64 bit unsigned integer that can be used within the enabled fault. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Arg *uint64 `json:"arg,omitempty"`

	// The name of the target fault. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	FaultName *string `json:"fault_name"`

	// The name of the function that contains the target fault. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FunctionName *string `json:"function_name,omitempty"`

	// Number of times the fault should be executed. Allowed values are 1-4294967295. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumExecutions *uint32 `json:"num_executions,omitempty"`

	// Number of times the fault should be skipped before executing. Field introduced in 18.2.9. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumSkips *uint32 `json:"num_skips,omitempty"`
}
