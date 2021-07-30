// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VinfraPoolServerDeleteDetails vinfra pool server delete details
// swagger:model VinfraPoolServerDeleteDetails
type VinfraPoolServerDeleteDetails struct {

	// pool_name of VinfraPoolServerDeleteDetails.
	// Required: true
	PoolName *string `json:"pool_name"`

	// server_ip of VinfraPoolServerDeleteDetails.
	ServerIP []string `json:"server_ip,omitempty"`
}
