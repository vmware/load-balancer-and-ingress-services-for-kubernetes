// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeDiscontinuousTimeChangeEventDetails se discontinuous time change event details
// swagger:model SeDiscontinuousTimeChangeEventDetails
type SeDiscontinuousTimeChangeEventDetails struct {

	// Relative time drift between SE and controller in terms of microseconds. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DriftTime *int64 `json:"drift_time,omitempty"`

	// Time stamp before the discontinuous jump in time. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FromTime *string `json:"from_time,omitempty"`

	// System Peer and Candidate NTP Servers active at the point of time jump. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NtpServers *string `json:"ntp_servers,omitempty"`

	// Name of the SE responsible for this event. It is a reference to an object of type ServiceEngine. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeName *string `json:"se_name,omitempty"`

	// UUID of the SE responsible for this event. It is a reference to an object of type ServiceEngine. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeRef *string `json:"se_ref,omitempty"`

	// Time stamp to which the time has discontinuously jumped. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ToTime *string `json:"to_time,omitempty"`
}
