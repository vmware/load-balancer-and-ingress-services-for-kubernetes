package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CloudInfo cloud info
// swagger:model CloudInfo
type CloudInfo struct {

	// CloudConnectorAgent properties specific to this cloud type.
	CcaProps *CCAgentProperties `json:"cca_props,omitempty"`

	// Controller properties specific to this cloud type.
	ControllerProps *ControllerProperties `json:"controller_props,omitempty"`

	// Flavor properties specific to this cloud type.
	FlavorProps []*CloudFlavor `json:"flavor_props,omitempty"`

	// flavor_regex_filter of CloudInfo.
	FlavorRegexFilter *string `json:"flavor_regex_filter,omitempty"`

	// Supported hypervisors. Enum options - DEFAULT, VMWARE_ESX, KVM, VMWARE_VSAN, XEN.
	Htypes []string `json:"htypes,omitempty"`

	// Cloud type. Enum options - CLOUD_NONE, CLOUD_VCENTER, CLOUD_OPENSTACK, CLOUD_AWS, CLOUD_VCA, CLOUD_APIC, CLOUD_MESOS, CLOUD_LINUXSERVER, CLOUD_DOCKER_UCP, CLOUD_RANCHER, CLOUD_OSHIFT_K8S, CLOUD_AZURE, CLOUD_GCP.
	// Required: true
	Vtype *string `json:"vtype"`
}
