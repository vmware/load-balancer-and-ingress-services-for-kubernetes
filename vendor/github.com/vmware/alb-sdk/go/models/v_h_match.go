// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VHMatch v h match
// swagger:model VHMatch
type VHMatch struct {

	// Host/domain name match configuration. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Host *string `json:"host"`

	// Add rules for selecting the virtual service. At least one rule must be configured. Field introduced in 22.1.3. Minimum of 1 items required. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Rules []*VHMatchRule `json:"rules,omitempty"`
}
