// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NetworkSubnetClash network subnet clash
// swagger:model NetworkSubnetClash
type NetworkSubnetClash struct {

	// ip_nw of NetworkSubnetClash.
	// Required: true
	IPNw *string `json:"ip_nw"`

	// networks of NetworkSubnetClash.
	Networks []string `json:"networks,omitempty"`
}
