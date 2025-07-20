// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HSMThalesRFS h s m thales r f s
// swagger:model HSMThalesRFS
type HSMThalesRFS struct {

	// IP address of the RFS server from where to sync the Thales encrypted private key. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	IP *IPAddr `json:"ip"`

	// Port at which the RFS server accepts the sync request from clients for Thales encrypted private key. Allowed values are 1-65535. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Port *uint32 `json:"port,omitempty"`
}
