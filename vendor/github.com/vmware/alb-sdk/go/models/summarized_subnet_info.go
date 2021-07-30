// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SummarizedSubnetInfo summarized subnet info
// swagger:model SummarizedSubnetInfo
type SummarizedSubnetInfo struct {

	// cidr_prefix of SummarizedSubnetInfo.
	// Required: true
	CidrPrefix *string `json:"cidr_prefix"`

	// network of SummarizedSubnetInfo.
	// Required: true
	Network *string `json:"network"`
}
