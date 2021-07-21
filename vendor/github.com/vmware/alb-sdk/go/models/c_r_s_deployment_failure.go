// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CRSDeploymentFailure c r s deployment failure
// swagger:model CRSDeploymentFailure
type CRSDeploymentFailure struct {

	// List of all CRS updates that failed to install. Field introduced in 20.1.1.
	CrsInfo []*CRSDetails `json:"crs_info,omitempty"`

	// Error message to be conveyed to controller UI. Field introduced in 20.1.1.
	Message *string `json:"message,omitempty"`
}
