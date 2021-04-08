package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ServiceEngineCloudLimits service engine cloud limits
// swagger:model ServiceEngineCloudLimits
type ServiceEngineCloudLimits struct {

	// Cloud type for this cloud limit. Enum options - CLOUD_NONE, CLOUD_VCENTER, CLOUD_OPENSTACK, CLOUD_AWS, CLOUD_VCA, CLOUD_APIC, CLOUD_MESOS, CLOUD_LINUXSERVER, CLOUD_DOCKER_UCP, CLOUD_RANCHER, CLOUD_OSHIFT_K8S, CLOUD_AZURE, CLOUD_GCP, CLOUD_NSXT. Field introduced in 20.1.1.
	Type *string `json:"type,omitempty"`

	// Maximum number of vrfcontexts per serviceengine. Field introduced in 20.1.1.
	VrfsPerServiceengine *int32 `json:"vrfs_per_serviceengine,omitempty"`
}
