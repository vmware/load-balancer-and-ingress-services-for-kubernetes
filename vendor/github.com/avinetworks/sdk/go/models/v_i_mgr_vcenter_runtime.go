package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VIMgrVcenterRuntime v i mgr vcenter runtime
// swagger:model VIMgrVcenterRuntime
type VIMgrVcenterRuntime struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// api_version of VIMgrVcenterRuntime.
	APIVersion *string `json:"api_version,omitempty"`

	// Placeholder for description of property apic_mode of obj type VIMgrVcenterRuntime field type str  type boolean
	ApicMode *bool `json:"apic_mode,omitempty"`

	//  It is a reference to an object of type Cloud.
	CloudRef *string `json:"cloud_ref,omitempty"`

	//  It is a reference to an object of type VIMgrDCRuntime.
	DatacenterRefs []string `json:"datacenter_refs,omitempty"`

	// disc_end_time of VIMgrVcenterRuntime.
	DiscEndTime *string `json:"disc_end_time,omitempty"`

	// disc_start_time of VIMgrVcenterRuntime.
	DiscStartTime *string `json:"disc_start_time,omitempty"`

	// discovered_datacenter of VIMgrVcenterRuntime.
	DiscoveredDatacenter *string `json:"discovered_datacenter,omitempty"`

	// inventory_progress of VIMgrVcenterRuntime.
	InventoryProgress *string `json:"inventory_progress,omitempty"`

	//  Enum options - VCENTER_DISCOVERY_BAD_CREDENTIALS, VCENTER_DISCOVERY_RETRIEVING_DC, VCENTER_DISCOVERY_WAITING_DC, VCENTER_DISCOVERY_RETRIEVING_NW, VCENTER_DISCOVERY_ONGOING, VCENTER_DISCOVERY_RESYNCING, VCENTER_DISCOVERY_COMPLETE, VCENTER_DISCOVERY_DELETING_VCENTER, VCENTER_DISCOVERY_FAILURE, VCENTER_DISCOVERY_COMPLETE_NO_MGMT_NW, VCENTER_DISCOVERY_COMPLETE_PER_TENANT_IP_ROUTE, VCENTER_DISCOVERY_MAKING_SE_OVA, VCENTER_DISCOVERY_RESYNC_FAILED, VCENTER_DISCOVERY_OBJECT_LIMIT_REACHED.
	InventoryState *string `json:"inventory_state,omitempty"`

	// management_network of VIMgrVcenterRuntime.
	ManagementNetwork *string `json:"management_network,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// Number of num_clusters.
	NumClusters *int64 `json:"num_clusters,omitempty"`

	// Number of num_dcs.
	NumDcs *int64 `json:"num_dcs,omitempty"`

	// Number of num_hosts.
	NumHosts *int64 `json:"num_hosts,omitempty"`

	// Number of num_nws.
	NumNws *int64 `json:"num_nws,omitempty"`

	// Number of num_vcenter_req_pending.
	NumVcenterReqPending *int64 `json:"num_vcenter_req_pending,omitempty"`

	// Number of num_vms.
	NumVms *int64 `json:"num_vms,omitempty"`

	// password of VIMgrVcenterRuntime.
	// Required: true
	Password *string `json:"password"`

	//  Enum options - NO_ACCESS, READ_ACCESS, WRITE_ACCESS.
	Privilege *string `json:"privilege,omitempty"`

	// Number of progress.
	Progress *int64 `json:"progress,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	//  Enum options - CLOUD_NONE, CLOUD_VCENTER, CLOUD_OPENSTACK, CLOUD_AWS, CLOUD_VCA, CLOUD_APIC, CLOUD_MESOS, CLOUD_LINUXSERVER, CLOUD_DOCKER_UCP, CLOUD_RANCHER, CLOUD_OSHIFT_K8S, CLOUD_AZURE, CLOUD_GCP.
	// Required: true
	Type *string `json:"type"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// username of VIMgrVcenterRuntime.
	// Required: true
	Username *string `json:"username"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`

	// Placeholder for description of property vcenter_connected of obj type VIMgrVcenterRuntime field type str  type boolean
	VcenterConnected *bool `json:"vcenter_connected,omitempty"`

	// vcenter_fullname of VIMgrVcenterRuntime.
	VcenterFullname *string `json:"vcenter_fullname,omitempty"`

	// vcenter_template_se_location of VIMgrVcenterRuntime.
	VcenterTemplateSeLocation *string `json:"vcenter_template_se_location,omitempty"`

	// vcenter_url of VIMgrVcenterRuntime.
	// Required: true
	VcenterURL *string `json:"vcenter_url"`
}
