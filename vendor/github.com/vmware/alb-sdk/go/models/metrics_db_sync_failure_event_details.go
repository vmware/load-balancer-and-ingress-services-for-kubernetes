// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MetricsDbSyncFailureEventDetails metrics db sync failure event details
// swagger:model MetricsDbSyncFailureEventDetails
type MetricsDbSyncFailureEventDetails struct {

	// Name of the node responsible for this event. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NodeName *string `json:"node_name,omitempty"`

	// Name of the process responsible for this event. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ProcessName *string `json:"process_name,omitempty"`

	// Timestamp at which this event occurred. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Timestamp *string `json:"timestamp,omitempty"`
}
