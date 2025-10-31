// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DebugVrf debug vrf
// swagger:model DebugVrf
type DebugVrf struct {

	//  Enum options - DEBUG_VRF_BGP, DEBUG_VRF_QUAGGA, DEBUG_VRF_ALL, DEBUG_VRF_NONE. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Flag *string `json:"flag"`
}
