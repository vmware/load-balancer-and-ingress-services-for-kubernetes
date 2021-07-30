// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HostUnavailEventDetails host unavail event details
// swagger:model HostUnavailEventDetails
type HostUnavailEventDetails struct {

	// Name of the object.
	Name *string `json:"name,omitempty"`

	// reasons of HostUnavailEventDetails.
	Reasons []string `json:"reasons,omitempty"`

	// vs_name of HostUnavailEventDetails.
	VsName *string `json:"vs_name,omitempty"`
}
