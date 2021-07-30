// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DosThresholdProfile dos threshold profile
// swagger:model DosThresholdProfile
type DosThresholdProfile struct {

	// Attack type, min and max values for DoS attack detection.
	ThreshInfo []*DosThreshold `json:"thresh_info,omitempty"`

	// Timer value in seconds to collect DoS attack metrics based on threshold on the Service Engine for this Virtual Service. Unit is SEC.
	// Required: true
	ThreshPeriod *int32 `json:"thresh_period"`
}
