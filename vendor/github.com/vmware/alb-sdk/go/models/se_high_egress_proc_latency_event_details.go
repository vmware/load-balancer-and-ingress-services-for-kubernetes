// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeHighEgressProcLatencyEventDetails se high egress proc latency event details
// swagger:model SeHighEgressProcLatencyEventDetails
type SeHighEgressProcLatencyEventDetails struct {

	// Dispatcher core which received the packet. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DispatcherCore *uint32 `json:"dispatcher_core,omitempty"`

	// Number of events in a 30 second interval. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EventCount *uint64 `json:"event_count,omitempty"`

	// Proxy core which processed the packet. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FlowCore []int64 `json:"flow_core,omitempty,omitempty"`

	// Proxy dequeue latency. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxProxyToDispQueingDelay *uint32 `json:"max_proxy_to_disp_queing_delay,omitempty"`

	// SE name. It is a reference to an object of type ServiceEngine. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeName *string `json:"se_name,omitempty"`

	// SE UUID. It is a reference to an object of type ServiceEngine. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeRef *string `json:"se_ref,omitempty"`
}
