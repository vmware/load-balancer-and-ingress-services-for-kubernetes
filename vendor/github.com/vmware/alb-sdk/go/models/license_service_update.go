// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// LicenseServiceUpdate license service update
// swagger:model LicenseServiceUpdate
type LicenseServiceUpdate struct {

	// Name. Field introduced in 21.1.4. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`

	// Organization id. Field introduced in 21.1.4. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ServiceUnits *OrgServiceUnits `json:"service_units,omitempty"`
}
