package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CloudTenantsDeleted cloud tenants deleted
// swagger:model CloudTenantsDeleted
type CloudTenantsDeleted struct {

	// cc_id of CloudTenantsDeleted.
	CcID *string `json:"cc_id,omitempty"`

	// Placeholder for description of property tenants of obj type CloudTenantsDeleted field type str  type object
	Tenants []*CloudTenantCleanup `json:"tenants,omitempty"`

	//  Enum options - CLOUD_NONE, CLOUD_VCENTER, CLOUD_OPENSTACK, CLOUD_AWS, CLOUD_VCA, CLOUD_APIC, CLOUD_MESOS, CLOUD_LINUXSERVER, CLOUD_DOCKER_UCP, CLOUD_RANCHER, CLOUD_OSHIFT_K8S, CLOUD_AZURE, CLOUD_GCP.
	Vtype *string `json:"vtype,omitempty"`
}
