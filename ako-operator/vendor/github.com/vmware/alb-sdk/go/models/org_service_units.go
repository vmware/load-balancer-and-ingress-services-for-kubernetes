// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// OrgServiceUnits org service units
// swagger:model OrgServiceUnits
type OrgServiceUnits struct {

	// Available service units on pulse portal. Field introduced in 21.1.4. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AvailableServiceUnits *float64 `json:"available_service_units,omitempty"`

	// Organization id. Field introduced in 21.1.4. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	OrgID *string `json:"org_id,omitempty"`

	// Used service units on pulse portal. Field introduced in 21.1.4. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UsedServiceUnits *float64 `json:"used_service_units,omitempty"`
}
