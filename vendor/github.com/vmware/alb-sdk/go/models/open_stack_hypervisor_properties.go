package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// OpenStackHypervisorProperties open stack hypervisor properties
// swagger:model OpenStackHypervisorProperties
type OpenStackHypervisorProperties struct {

	// Hypervisor type. Enum options - DEFAULT, VMWARE_ESX, KVM, VMWARE_VSAN, XEN. Field introduced in 17.2.1.
	// Required: true
	Hypervisor *string `json:"hypervisor"`

	// Custom properties to be associated with the SE image in Glance for this hypervisor type. Field introduced in 17.2.1.
	ImageProperties []*Property `json:"image_properties,omitempty"`
}
