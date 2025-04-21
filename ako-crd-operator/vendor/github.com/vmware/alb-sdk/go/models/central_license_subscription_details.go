// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CentralLicenseSubscriptionDetails central license subscription details
// swagger:model CentralLicenseSubscriptionDetails
type CentralLicenseSubscriptionDetails struct {

	// Message. Field introduced in 21.1.4. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Message *string `json:"message,omitempty"`

	// Tenant uuid. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TenantUUID *string `json:"tenant_uuid,omitempty"`
}
