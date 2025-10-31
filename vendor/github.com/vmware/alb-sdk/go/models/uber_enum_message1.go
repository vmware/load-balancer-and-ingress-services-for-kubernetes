// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// UberEnumMessage1 uber enum message1
// swagger:model UberEnumMessage1
type UberEnumMessage1 struct {

	//  Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Rm []*UberEnumMessage2 `json:"rm,omitempty"`

	//  Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Rv []int64 `json:"rv,omitempty,omitempty"`

	//  Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	V *uint64 `json:"v,omitempty"`
}
