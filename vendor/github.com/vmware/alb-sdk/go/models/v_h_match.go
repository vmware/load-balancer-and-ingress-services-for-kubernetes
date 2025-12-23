// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VHMatch v h match
// swagger:model VHMatch
type VHMatch struct {

	// Host/domain name match configuration. Must be configured along with at least one path match criteria. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Host *string `json:"host"`

	// Resource/uri path match configuration. Must be configured along with Host match criteria. Field introduced in 20.1.3. Minimum of 1 items required. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Path []*PathMatch `json:"path,omitempty"`
}
