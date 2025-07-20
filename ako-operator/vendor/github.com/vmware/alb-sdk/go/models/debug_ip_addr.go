// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DebugIPAddr debug Ip addr
// swagger:model DebugIpAddr
type DebugIPAddr struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Addrs []*IPAddr `json:"addrs,omitempty"`

	// Match criteria. Enum options - IS_IN, IS_NOT_IN. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MatchOperation *string `json:"match_operation,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Prefixes []*IPAddrPrefix `json:"prefixes,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Ranges []*IPAddrRange `json:"ranges,omitempty"`
}
