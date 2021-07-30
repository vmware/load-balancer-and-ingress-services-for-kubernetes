// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SSLIgnoredDetails s s l ignored details
// swagger:model SSLIgnoredDetails
type SSLIgnoredDetails struct {

	// Name of SSL Certificate.
	Name *string `json:"name,omitempty"`

	// Reason for ignoring certificate.
	Reason *string `json:"reason,omitempty"`
}
