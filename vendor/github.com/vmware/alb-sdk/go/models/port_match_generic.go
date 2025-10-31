// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PortMatchGeneric port match generic
// swagger:model PortMatchGeneric
type PortMatchGeneric struct {

	// Criterion to use for src/dest port in a TCP/UDP packet. Enum options - IS_IN, IS_NOT_IN. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	MatchCriteria *string `json:"match_criteria"`

	// Listening TCP port(s). Allowed values are 1-65535. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Ports []int64 `json:"ports,omitempty,omitempty"`

	// A port range defined by a start and end, including both. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Ranges []*PortRange `json:"ranges,omitempty"`
}
