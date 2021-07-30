// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NsxtCredentials nsxt credentials
// swagger:model NsxtCredentials
type NsxtCredentials struct {

	// Password to talk to Nsx-t manager. Field introduced in 20.1.1.
	Password *string `json:"password,omitempty"`

	// Username to talk to Nsx-t manager. Field introduced in 20.1.1.
	Username *string `json:"username,omitempty"`
}
