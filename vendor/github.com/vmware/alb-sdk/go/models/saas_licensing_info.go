// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SaasLicensingInfo saas licensing info
// swagger:model SaasLicensingInfo
type SaasLicensingInfo struct {

	// Maximum service units limit for controller. Allowed values are 0-100000. Special values are 0 - infinite. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MaxServiceUnits *float64 `json:"max_service_units,omitempty"`

	// Minimum service units that always remain reserved on controller. Allowed values are 0-1000. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ReserveServiceUnits *float64 `json:"reserve_service_units,omitempty"`
}
