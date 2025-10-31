// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HorizonProfile horizon profile
// swagger:model HorizonProfile
type HorizonProfile struct {

	// Horizon blast port of the UAG server. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	BlastPort *uint32 `json:"blast_port,omitempty"`

	// Horizon pcoip port of the UAG server. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PcoipPort *uint32 `json:"pcoip_port,omitempty"`
}
