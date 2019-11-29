package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DebugVirtualService debug virtual service
// swagger:model DebugVirtualService
type DebugVirtualService struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Placeholder for description of property capture of obj type DebugVirtualService field type str  type boolean
	Capture *bool `json:"capture,omitempty"`

	// Placeholder for description of property capture_params of obj type DebugVirtualService field type str  type object
	CaptureParams *DebugVirtualServiceCapture `json:"capture_params,omitempty"`

	//  It is a reference to an object of type Cloud.
	CloudRef *string `json:"cloud_ref,omitempty"`

	// This option controls the capture of Health Monitor flows. Enum options - DEBUG_VS_HM_NONE, DEBUG_VS_HM_ONLY, DEBUG_VS_HM_INCLUDE.
	DebugHm *string `json:"debug_hm,omitempty"`

	// Placeholder for description of property debug_ip of obj type DebugVirtualService field type str  type object
	DebugIP *DebugIPAddr `json:"debug_ip,omitempty"`

	// Dns debug options. Field introduced in 18.2.1.
	DNSOptions *DebugDNSOptions `json:"dns_options,omitempty"`

	// Placeholder for description of property flags of obj type DebugVirtualService field type str  type object
	Flags []*DebugVsDataplane `json:"flags,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// This option re-synchronizes flows between Active-Standby service engines for all the virtual services placed on them. It should be used with caution because as it can cause a flood between Active-Standby. Field introduced in 18.1.3,18.2.1.
	ResyncFlows *bool `json:"resync_flows,omitempty"`

	// Placeholder for description of property se_params of obj type DebugVirtualService field type str  type object
	SeParams *DebugVirtualServiceSeParams `json:"se_params,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
