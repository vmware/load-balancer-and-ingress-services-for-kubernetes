package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VIMgrClusterRuntime v i mgr cluster runtime
// swagger:model VIMgrClusterRuntime
type VIMgrClusterRuntime struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	//  It is a reference to an object of type Cloud.
	CloudRef *string `json:"cloud_ref,omitempty"`

	// datacenter_managed_object_id of VIMgrClusterRuntime.
	DatacenterManagedObjectID *string `json:"datacenter_managed_object_id,omitempty"`

	// Unique object identifier of datacenter.
	DatacenterUUID *string `json:"datacenter_uuid,omitempty"`

	//  It is a reference to an object of type VIMgrHostRuntime.
	HostRefs []string `json:"host_refs,omitempty"`

	// managed_object_id of VIMgrClusterRuntime.
	// Required: true
	ManagedObjectID *string `json:"managed_object_id"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	//  Enum options - CLOUD_NONE, CLOUD_VCENTER, CLOUD_OPENSTACK, CLOUD_AWS, CLOUD_VCA, CLOUD_APIC, CLOUD_MESOS, CLOUD_LINUXSERVER, CLOUD_DOCKER_UCP, CLOUD_RANCHER, CLOUD_OSHIFT_K8S, CLOUD_AZURE, CLOUD_GCP.
	// Required: true
	Type *string `json:"type"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
