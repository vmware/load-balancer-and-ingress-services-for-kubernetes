// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ConnectionEventDetails connection event details
// swagger:model ConnectionEventDetails
type ConnectionEventDetails struct {

	// Destinaton host name to be connected. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Host *string `json:"host,omitempty"`

	// Connection status information. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Info *string `json:"info,omitempty"`

	// Destinaton port to be connected. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Port *uint32 `json:"port,omitempty"`
}
