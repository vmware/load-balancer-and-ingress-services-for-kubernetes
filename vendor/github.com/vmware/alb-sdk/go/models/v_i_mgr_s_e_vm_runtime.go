// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VIMgrSEVMRuntime v i mgr s e VM runtime
// swagger:model VIMgrSEVMRuntime
type VIMgrSEVMRuntime struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvailabilityZone *string `json:"availability_zone,omitempty"`

	//  Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AzureInfo *AzureInfo `json:"azure_info,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudName *string `json:"cloud_name,omitempty"`

	//  It is a reference to an object of type Cloud. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudRef *string `json:"cloud_ref,omitempty"`

	// ServiceEngine deployed on cluster.Ex MOB  domain-c23. Field introduced in 20.1.7, 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ClusterID *string `json:"cluster_id,omitempty"`

	// ServiceEngine added to cluster vmgroup. Field introduced in 20.1.7, 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ClusterVmgroup *string `json:"cluster_vmgroup,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ConnectionState *string `json:"connection_state,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ControllerClusterUUID *string `json:"controller_cluster_uuid,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ControllerIPAddr *string `json:"controller_ip_addr,omitempty"`

	// Service Engine Cookie set by the resource manager. Field introduced in 18.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Cookie *string `json:"cookie,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CreationInProgress *bool `json:"creation_in_progress,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DeletionInProgress *bool `json:"deletion_in_progress,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DiscoveryResponse *string `json:"discovery_response,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DiscoveryStatus uint32 `json:"discovery_status,omitempty"`

	// Disk space in GB for each service engine VM. Field introduced in 18.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DiskGb uint32 `json:"disk_gb,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Flavor *string `json:"flavor,omitempty"`

	// GCP Project ID in which SE is created. This field is applicable for GCP cloud type only. Field introduced in 20.1.7, 21.1.2, 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	GcpSeProjectID *string `json:"gcp_se_project_id,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GuestNic []*VIMgrGuestNicRuntime `json:"guest_nic,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Host *string `json:"host,omitempty"`

	//  It is a reference to an object of type VIMgrHostRuntime. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HostRef *string `json:"host_ref,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Hostid *string `json:"hostid,omitempty"`

	//  Enum options - DEFAULT, VMWARE_ESX, KVM, VMWARE_VSAN, XEN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Hypervisor *string `json:"hypervisor,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	InitVnics *int32 `json:"init_vnics,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LastDiscovery uint32 `json:"last_discovery,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ManagedObjectID *string `json:"managed_object_id"`

	// Memory in MB for each service engine VM. Field introduced in 18.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MemoryMb uint32 `json:"memory_mb,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Powerstate *string `json:"powerstate,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SecurityGroupUUID *string `json:"security_group_uuid,omitempty"`

	//  It is a reference to an object of type ServiceEngineGroup. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SegroupRef *string `json:"segroup_ref,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerGroupUUID *string `json:"server_group_uuid,omitempty"`

	//  It is a reference to an object of type Tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	//  Enum options - CLOUD_NONE, CLOUD_VCENTER, CLOUD_OPENSTACK, CLOUD_AWS, CLOUD_VCA, CLOUD_APIC, CLOUD_MESOS, CLOUD_LINUXSERVER, CLOUD_DOCKER_UCP, CLOUD_RANCHER, CLOUD_OSHIFT_K8S, CLOUD_AZURE, CLOUD_GCP, CLOUD_NSXT. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Type *string `json:"type"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterDatacenterUUID *string `json:"vcenter_datacenter_uuid,omitempty"`

	// ServiceEngine host connection state in vCenter. Field introduced in 20.1.7, 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VcenterHostConnectionState *string `json:"vcenter_host_connection_state,omitempty"`

	// VCenter Host HA state.Ex  election, fdmUnreachable, hostDown, initializationError, networkIsolated, uninitializationError, uninitialized. Field introduced in 20.1.7, 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VcenterHostHaState *string `json:"vcenter_host_ha_state,omitempty"`

	// ServiceEngine instance uuid from vCenter. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterInstanceUUID *string `json:"vcenter_instance_uuid,omitempty"`

	// ServiceEngine belongs to VCenter. It is a reference to an object of type VCenterServer. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterRef *string `json:"vcenter_ref,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterRmCookie *string `json:"vcenter_rm_cookie,omitempty"`

	//  Enum options - VIMGR_SE_NETWORK_ADMIN, VIMGR_SE_UNIFIED_ADMIN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterSeType *string `json:"vcenter_se_type,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterTemplateVM *bool `json:"vcenter_template_vm,omitempty"`

	// Service Engine deployed in vcenter. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterURL *string `json:"vcenter_url,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterVAppName *string `json:"vcenter_vAppName,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterVAppVendor *string `json:"vcenter_vAppVendor,omitempty"`

	//  Enum options - VMTYPE_SE_VM, VMTYPE_POOL_SRVR. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterVMType *string `json:"vcenter_vm_type,omitempty"`

	// Count of vcpus for each service engine VM. Field introduced in 18.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Vcpus uint32 `json:"vcpus,omitempty"`

	// VSphere HA on cluster enabled or not. Field introduced in 20.1.7, 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VsphereHaEnabled *bool `json:"vsphere_ha_enabled,omitempty"`

	// If this flag is set to True, vCenter vSphere HA handles ServiceEngine failure. This flag is set dynamiclly based on underlying ESX HA state(connected, hostDown..etc). Field introduced in 20.1.7, 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VsphereHaInprogress *bool `json:"vsphere_ha_inprogress,omitempty"`
}
