// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CentralLicenseRefreshDetails central license refresh details
// swagger:model CentralLicenseRefreshDetails
type CentralLicenseRefreshDetails struct {

	// Message. Field introduced in 21.1.4. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Message *string `json:"message,omitempty"`

	// Service units. Field introduced in 21.1.4. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ServiceUnits *float64 `json:"service_units,omitempty"`

	// Tenant uuid. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TenantUUID *string `json:"tenant_uuid,omitempty"`
}
