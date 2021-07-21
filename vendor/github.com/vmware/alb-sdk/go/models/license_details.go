// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// LicenseDetails license details
// swagger:model LicenseDetails
type LicenseDetails struct {

	// Number of backend_servers.
	BackendServers *int32 `json:"backend_servers,omitempty"`

	// expiry_at of LicenseDetails.
	ExpiryAt *string `json:"expiry_at,omitempty"`

	// license_id of LicenseDetails.
	LicenseID *string `json:"license_id,omitempty"`

	// license_type of LicenseDetails.
	LicenseType *string `json:"license_type,omitempty"`

	// Name of the object.
	Name *string `json:"name,omitempty"`
}
