package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ServerConfig server config
// swagger:model ServerConfig
type ServerConfig struct {

	// Placeholder for description of property def_port of obj type ServerConfig field type str  type boolean
	DefPort *bool `json:"def_port,omitempty"`

	// hostname of ServerConfig.
	Hostname *string `json:"hostname,omitempty"`

	// Placeholder for description of property ip_addr of obj type ServerConfig field type str  type object
	// Required: true
	IPAddr *IPAddr `json:"ip_addr"`

	// Placeholder for description of property is_enabled of obj type ServerConfig field type str  type boolean
	// Required: true
	IsEnabled *bool `json:"is_enabled"`

	//  Enum options - OPER_UP, OPER_DOWN, OPER_CREATING, OPER_RESOURCES, OPER_INACTIVE, OPER_DISABLED, OPER_UNUSED, OPER_UNKNOWN, OPER_PROCESSING, OPER_INITIALIZING, OPER_ERROR_DISABLED, OPER_AWAIT_MANUAL_PLACEMENT, OPER_UPGRADING, OPER_SE_PROCESSING, OPER_PARTITIONED, OPER_DISABLING, OPER_FAILED, OPER_UNAVAIL.
	LastState *string `json:"last_state,omitempty"`

	// VirtualService member in case this server is a member of GS group, and Geo Location available.
	Location *GeoLocation `json:"location,omitempty"`

	// Placeholder for description of property oper_status of obj type ServerConfig field type str  type object
	OperStatus *OperationalStatus `json:"oper_status,omitempty"`

	// Number of port.
	// Required: true
	Port *int32 `json:"port"`

	// Placeholder for description of property timer_exists of obj type ServerConfig field type str  type boolean
	TimerExists *bool `json:"timer_exists,omitempty"`
}
