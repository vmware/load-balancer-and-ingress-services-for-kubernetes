// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DebugVirtualService debug virtual service
// swagger:model DebugVirtualService
type DebugVirtualService struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Capture *bool `json:"capture,omitempty"`

	// Per packet capture filters for Debug Virtual Service. Applies to both frontend and backend packets. Field introduced in 18.2.7. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CaptureFilters *CaptureFilters `json:"capture_filters,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CaptureParams *DebugVirtualServiceCapture `json:"capture_params,omitempty"`

	// Per packet capture filters for Debug Virtual Service. Applies to both frontend and backend packets. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CapturePktFilter *CapturePacketFilter `json:"capture_pkt_filter,omitempty"`

	//  It is a reference to an object of type Cloud. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudRef *string `json:"cloud_ref,omitempty"`

	// This option controls the capture of Health Monitor flows. Enum options - DEBUG_VS_HM_NONE, DEBUG_VS_HM_ONLY, DEBUG_VS_HM_INCLUDE. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DebugHm *string `json:"debug_hm,omitempty"`

	// Filters all packets of a complete transaction (client and server side), based on client ip. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DebugIP *DebugIPAddr `json:"debug_ip,omitempty"`

	// Dns debug options. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DNSOptions *DebugDNSOptions `json:"dns_options,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Flags []*DebugVsDataplane `json:"flags,omitempty"`

	// Deprecated in 22.1.1. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LatencyAuditFilters *CaptureFilters `json:"latency_audit_filters,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// Object sync debug options. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Objsync *DebugVirtualServiceObjSync `json:"objsync,omitempty"`

	// This option re-synchronizes flows between Active-Standby service engines for all the virtual services placed on them. It should be used with caution because as it can cause a flood between Active-Standby. Field introduced in 18.1.3,18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ResyncFlows *bool `json:"resync_flows,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeParams *DebugVirtualServiceSeParams `json:"se_params,omitempty"`

	//  It is a reference to an object of type Tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}
