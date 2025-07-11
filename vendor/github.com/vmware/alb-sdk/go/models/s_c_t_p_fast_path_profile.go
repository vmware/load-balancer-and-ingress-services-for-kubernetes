// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SCTPFastPathProfile s c t p fast path profile
// swagger:model SCTPFastPathProfile
type SCTPFastPathProfile struct {

	// When enabled, Avi will complete the 4-way handshake with the client before forwarding any packets to the server.  This will protect the server from INIT chunks flood and half open connections. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EnableInitChunkProtection *bool `json:"enable_init_chunk_protection,omitempty"`

	// SCTP autoclose timeout. 0 means autoclose deactivated. Allowed values are 0-247483647. Field introduced in 22.1.3. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IDLETimeout *int32 `json:"idle_timeout,omitempty"`
}
