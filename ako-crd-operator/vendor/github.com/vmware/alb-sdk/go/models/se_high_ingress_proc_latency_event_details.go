// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeHighIngressProcLatencyEventDetails se high ingress proc latency event details
// swagger:model SeHighIngressProcLatencyEventDetails
type SeHighIngressProcLatencyEventDetails struct {

	// Dispatcher core which received the packet. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DispatcherCore []int64 `json:"dispatcher_core,omitempty,omitempty"`

	// Number of events in a 30 second interval. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EventCount uint64 `json:"event_count,omitempty"`

	// Proxy core which processed the packet. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FlowCore uint32 `json:"flow_core,omitempty"`

	// Proxy dequeue latency. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxDispToProxyQueingDelay uint32 `json:"max_disp_to_proxy_queing_delay,omitempty"`

	// Dispatcher processing latency. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxDispatcherProcTime uint32 `json:"max_dispatcher_proc_time,omitempty"`

	// SE name. It is a reference to an object of type ServiceEngine. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeName *string `json:"se_name,omitempty"`

	// SE UUID. It is a reference to an object of type ServiceEngine. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeRef *string `json:"se_ref,omitempty"`

	// Deprecated in 22.1.1. It is a reference to an object of type VirtualService. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsName *string `json:"vs_name,omitempty"`

	// Deprecated in 22.1.1. It is a reference to an object of type VirtualService. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsRef *string `json:"vs_ref,omitempty"`
}
