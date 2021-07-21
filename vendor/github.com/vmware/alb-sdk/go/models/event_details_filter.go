// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// EventDetailsFilter event details filter
// swagger:model EventDetailsFilter
type EventDetailsFilter struct {

	//  Enum options - ALERT_OP_LT, ALERT_OP_LE, ALERT_OP_EQ, ALERT_OP_NE, ALERT_OP_GE, ALERT_OP_GT.
	// Required: true
	Comparator *string `json:"comparator"`

	// event_details_key of EventDetailsFilter.
	// Required: true
	EventDetailsKey *string `json:"event_details_key"`

	// event_details_value of EventDetailsFilter.
	// Required: true
	EventDetailsValue *string `json:"event_details_value"`
}
