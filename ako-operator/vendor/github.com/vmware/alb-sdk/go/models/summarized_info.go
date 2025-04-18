// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SummarizedInfo summarized info
// swagger:model SummarizedInfo
type SummarizedInfo struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SubnetInfo []*SummarizedSubnetInfo `json:"subnet_info,omitempty"`
}
