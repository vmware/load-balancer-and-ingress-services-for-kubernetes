package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VIMgrHostRuntime v i mgr host runtime
// swagger:model VIMgrHostRuntime
type VIMgrHostRuntime struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	//  It is a reference to an object of type Cloud.
	CloudRef *string `json:"cloud_ref,omitempty"`

	// cluster_name of VIMgrHostRuntime.
	ClusterName *string `json:"cluster_name,omitempty"`

	// Unique object identifier of cluster.
	ClusterUUID *string `json:"cluster_uuid,omitempty"`

	// Placeholder for description of property cntlr_accessible of obj type VIMgrHostRuntime field type str  type boolean
	CntlrAccessible *bool `json:"cntlr_accessible,omitempty"`

	// connection_state of VIMgrHostRuntime.
	ConnectionState *string `json:"connection_state,omitempty"`

	// Number of cpu_hz.
	CPUHz *int64 `json:"cpu_hz,omitempty"`

	// Placeholder for description of property maintenance_mode of obj type VIMgrHostRuntime field type str  type boolean
	MaintenanceMode *bool `json:"maintenance_mode,omitempty"`

	// managed_object_id of VIMgrHostRuntime.
	// Required: true
	ManagedObjectID *string `json:"managed_object_id"`

	// Number of mem.
	Mem *int64 `json:"mem,omitempty"`

	// mgmt_portgroup of VIMgrHostRuntime.
	MgmtPortgroup *string `json:"mgmt_portgroup,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// Unique object identifiers of networks.
	NetworkUuids []string `json:"network_uuids,omitempty"`

	// Number of num_cpu_cores.
	NumCPUCores *int32 `json:"num_cpu_cores,omitempty"`

	// Number of num_cpu_packages.
	NumCPUPackages *int32 `json:"num_cpu_packages,omitempty"`

	// Number of num_cpu_threads.
	NumCPUThreads *int32 `json:"num_cpu_threads,omitempty"`

	// Placeholder for description of property pnics of obj type VIMgrHostRuntime field type str  type object
	Pnics []*CdpLldpInfo `json:"pnics,omitempty"`

	// powerstate of VIMgrHostRuntime.
	Powerstate *string `json:"powerstate,omitempty"`

	// quarantine_start_ts of VIMgrHostRuntime.
	QuarantineStartTs *string `json:"quarantine_start_ts,omitempty"`

	// Placeholder for description of property quarantined of obj type VIMgrHostRuntime field type str  type boolean
	Quarantined *bool `json:"quarantined,omitempty"`

	// Number of quarantined_periods.
	QuarantinedPeriods *int32 `json:"quarantined_periods,omitempty"`

	// Number of se_fail_cnt.
	SeFailCnt *int32 `json:"se_fail_cnt,omitempty"`

	// Number of se_success_cnt.
	SeSuccessCnt *int32 `json:"se_success_cnt,omitempty"`

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

	//  It is a reference to an object of type VIMgrVMRuntime.
	VMRefs []string `json:"vm_refs,omitempty"`
}
