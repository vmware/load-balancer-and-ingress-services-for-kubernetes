// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AlertSummary alert summary
// swagger:model AlertSummary
type AlertSummary struct {

	// Resolved Alert Type. Enum options - ALERT_LOW, ALERT_MEDIUM, ALERT_HIGH. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Level *string `json:"level,omitempty"`
}
