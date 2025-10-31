// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CapturePacketFilter capture packet filter
// swagger:model CapturePacketFilter
type CapturePacketFilter struct {

	// TCP Params filter. And'ed internally and Or'ed amongst each other. . Field introduced in 30.2.1. Maximum of 20 items allowed. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CaptureTCPFilters []*CaptureTCPFilter `json:"capture_tcp_filters,omitempty"`
}
