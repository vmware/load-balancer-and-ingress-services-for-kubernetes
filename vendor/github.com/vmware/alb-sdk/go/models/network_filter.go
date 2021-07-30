// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NetworkFilter network filter
// swagger:model NetworkFilter
type NetworkFilter struct {

	//  It is a reference to an object of type VIMgrNWRuntime.
	// Required: true
	NetworkRef *string `json:"network_ref"`

	// server_filter of NetworkFilter.
	ServerFilter *string `json:"server_filter,omitempty"`
}
