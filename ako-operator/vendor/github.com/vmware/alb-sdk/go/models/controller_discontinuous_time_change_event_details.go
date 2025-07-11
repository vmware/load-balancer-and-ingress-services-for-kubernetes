// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ControllerDiscontinuousTimeChangeEventDetails controller discontinuous time change event details
// swagger:model ControllerDiscontinuousTimeChangeEventDetails
type ControllerDiscontinuousTimeChangeEventDetails struct {

	// Time stamp before the discontinuous jump in time. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FromTime *string `json:"from_time,omitempty"`

	// Name of the Controller responsible for this event. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NodeName *string `json:"node_name,omitempty"`

	// System Peer and Candidate NTP Servers active at the point of time jump. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NtpServers *string `json:"ntp_servers,omitempty"`

	// Time stamp to which the time has discontinuously jumped. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ToTime *string `json:"to_time,omitempty"`
}
