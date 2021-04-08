package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HSMThalesNetHsm h s m thales net hsm
// swagger:model HSMThalesNetHsm
type HSMThalesNetHsm struct {

	// Electronic serial number of the netHSM device. Use Thales anonkneti utility to find the netHSM ESN.
	// Required: true
	Esn *string `json:"esn"`

	// Hash of the key that netHSM device uses to authenticate itself. Use Thales anonkneti utility to find the netHSM keyhash.
	// Required: true
	Keyhash *string `json:"keyhash"`

	// Local module id of the netHSM device.
	ModuleID *int32 `json:"module_id,omitempty"`

	// Priority class of the nethsm in an high availability setup. 1 is the highest priority and 100 is the lowest priority. Allowed values are 1-100.
	// Required: true
	Priority *int32 `json:"priority"`

	// IP address of the netHSM device.
	// Required: true
	RemoteIP *IPAddr `json:"remote_ip"`

	// Port at which the netHSM device accepts the connection. Allowed values are 1-65535.
	RemotePort *int32 `json:"remote_port,omitempty"`
}
