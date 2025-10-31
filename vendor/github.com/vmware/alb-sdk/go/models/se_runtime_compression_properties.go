// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeRuntimeCompressionProperties se runtime compression properties
// swagger:model SeRuntimeCompressionProperties
type SeRuntimeCompressionProperties struct {

	// If client RTT is higher than this threshold, enable normal compression on the response. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxLowRtt *int32 `json:"max_low_rtt,omitempty"`

	// If client RTT is higher than this threshold, enable aggressive compression on the response. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MinHighRtt *int32 `json:"min_high_rtt,omitempty"`

	// Minimum response content length to enable compression. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MinLength *int32 `json:"min_length,omitempty"`

	// Values that identify mobile browsers in order to enable aggressive compression. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MobileStr []string `json:"mobile_str,omitempty"`
}
