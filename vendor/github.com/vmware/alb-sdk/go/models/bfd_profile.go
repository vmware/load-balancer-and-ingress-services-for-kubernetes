// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// BfdProfile bfd profile
// swagger:model BfdProfile
type BfdProfile struct {

	// Default required minimum receive interval (in ms) used in BFD. Allowed values are 500-4000000. Field introduced in 20.1.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Minrx *uint32 `json:"minrx,omitempty"`

	// Default desired minimum transmit interval (in ms) used in BFD. Allowed values are 500-4000000. Field introduced in 20.1.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Mintx *uint32 `json:"mintx,omitempty"`

	// Default detection multiplier used in BFD. Allowed values are 3-255. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Multi *uint32 `json:"multi,omitempty"`
}
