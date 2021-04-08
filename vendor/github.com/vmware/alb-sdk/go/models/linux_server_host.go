package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// LinuxServerHost linux server host
// swagger:model LinuxServerHost
type LinuxServerHost struct {

	// Placeholder for description of property host_attr of obj type LinuxServerHost field type str  type object
	HostAttr []*HostAttributes `json:"host_attr,omitempty"`

	// Placeholder for description of property host_ip of obj type LinuxServerHost field type str  type object
	// Required: true
	HostIP *IPAddr `json:"host_ip"`

	// Node's availability zone. ServiceEngines belonging to the availability zone will be rebooted during a manual DR failover.
	NodeAvailabilityZone *string `json:"node_availability_zone,omitempty"`

	// The SE Group association for the SE. If None, then 'Default-Group' SEGroup is associated with the SE. It is a reference to an object of type ServiceEngineGroup. Field introduced in 17.2.1.
	SeGroupRef *string `json:"se_group_ref,omitempty"`
}
