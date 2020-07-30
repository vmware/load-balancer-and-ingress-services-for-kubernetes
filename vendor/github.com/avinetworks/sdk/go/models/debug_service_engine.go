package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DebugServiceEngine debug service engine
// swagger:model DebugServiceEngine
type DebugServiceEngine struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Enable/disable packet capture. Field introduced in 18.2.2.
	Capture *bool `json:"capture,omitempty"`

	// Per packet capture filters for Debug Service Engine. Not applicable for DOS pcap capture. . Field introduced in 18.2.5.
	CaptureFilters *CaptureFilters `json:"capture_filters,omitempty"`

	// Params for SE pcap. Field introduced in 17.2.14,18.1.5,18.2.1.
	CaptureParams *DebugVirtualServiceCapture `json:"capture_params,omitempty"`

	// Placeholder for description of property cpu_shares of obj type DebugServiceEngine field type str  type object
	CPUShares []*DebugSeCPUShares `json:"cpu_shares,omitempty"`

	// Per packet IP filter for Service Engine PCAP. Matches with source and destination address. Field introduced in 17.2.14,18.1.5,18.2.1.
	DebugIP *DebugIPAddr `json:"debug_ip,omitempty"`

	// Enables the use of kdump on SE. Requires SE Reboot. Applicable only in case of VM Based deployments. Field introduced in 18.2.10, 20.1.1.
	EnableKdump *bool `json:"enable_kdump,omitempty"`

	// Params for SE fault injection. Field introduced in 18.1.2.
	Fault *DebugSeFault `json:"fault,omitempty"`

	// Placeholder for description of property flags of obj type DebugServiceEngine field type str  type object
	Flags []*DebugSeDataplane `json:"flags,omitempty"`

	// Name of the object.
	Name *string `json:"name,omitempty"`

	// Placeholder for description of property seagent_debug of obj type DebugServiceEngine field type str  type object
	SeagentDebug []*DebugSeAgent `json:"seagent_debug,omitempty"`

	// Debug knob for se_log_agent process. Field introduced in 20.1.1.
	SelogagentDebug *DebugSeAgent `json:"selogagent_debug,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
