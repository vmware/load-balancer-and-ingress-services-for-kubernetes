package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CloudProperties cloud properties
// swagger:model CloudProperties
type CloudProperties struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// CloudConnector properties.
	CcProps *CCProperties `json:"cc_props,omitempty"`

	// Cloud types supported by CloudConnector. Enum options - CLOUD_NONE, CLOUD_VCENTER, CLOUD_OPENSTACK, CLOUD_AWS, CLOUD_VCA, CLOUD_APIC, CLOUD_MESOS, CLOUD_LINUXSERVER, CLOUD_DOCKER_UCP, CLOUD_RANCHER, CLOUD_OSHIFT_K8S, CLOUD_AZURE, CLOUD_GCP.
	CcVtypes []string `json:"cc_vtypes,omitempty"`

	// Hypervisor properties.
	HypProps []*HypervisorProperties `json:"hyp_props,omitempty"`

	// Properties specific to a cloud type.
	Info []*CloudInfo `json:"info,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
