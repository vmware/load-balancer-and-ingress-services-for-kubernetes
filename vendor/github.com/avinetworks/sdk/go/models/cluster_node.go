package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ClusterNode cluster node
// swagger:model ClusterNode
type ClusterNode struct {

	// Optional service categories that a node can be assigned (e.g. SYSTEM, INFRASTRUCTURE or ANALYTICS). Field introduced in 18.1.1.
	Categories []string `json:"categories,omitempty"`

	// IP address of controller VM.
	// Required: true
	IP *IPAddr `json:"ip"`

	// Name of the object.
	Name *string `json:"name,omitempty"`

	// Public IP address or hostname of the controller VM. Field introduced in 17.2.3.
	PublicIPOrName *IPAddr `json:"public_ip_or_name,omitempty"`

	// Hostname assigned to this controller VM.
	VMHostname *string `json:"vm_hostname,omitempty"`

	// Managed object reference of this controller VM.
	VMMor *string `json:"vm_mor,omitempty"`

	// Name of the controller VM.
	VMName *string `json:"vm_name,omitempty"`

	// UUID on the controller VM.
	VMUUID *string `json:"vm_uuid,omitempty"`
}
