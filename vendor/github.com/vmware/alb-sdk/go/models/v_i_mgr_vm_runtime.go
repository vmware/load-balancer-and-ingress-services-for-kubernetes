// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VIMgrVMRuntime v i mgr VM runtime
// swagger:model VIMgrVMRuntime
type VIMgrVMRuntime struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvailabilityZone *string `json:"availability_zone,omitempty"`

	//  It is a reference to an object of type Cloud. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudRef *string `json:"cloud_ref,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ConnectionState *string `json:"connection_state,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ControllerClusterUUID *string `json:"controller_cluster_uuid,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ControllerIPAddr *string `json:"controller_ip_addr,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ControllerVM *bool `json:"controller_vm,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CPUReservation *int64 `json:"cpu_reservation,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CPUShares *int32 `json:"cpu_shares,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CreationInProgress *bool `json:"creation_in_progress,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GuestNic []*VIMgrGuestNicRuntime `json:"guest_nic,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Host *string `json:"host,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	InitVnics *int32 `json:"init_vnics,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ManagedObjectID *string `json:"managed_object_id"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MemShares *int32 `json:"mem_shares,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Memory *int64 `json:"memory,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MemoryReservation *int64 `json:"memory_reservation,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumCPU *int32 `json:"num_cpu,omitempty"`

	//  Field introduced in 17.1.1,17.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OvfAvisetypeField *string `json:"ovf_avisetype_field,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Powerstate *string `json:"powerstate,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeVer *int32 `json:"se_ver,omitempty"`

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

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterRmCookie *string `json:"vcenter_rm_cookie,omitempty"`

	//  Enum options - VIMGR_SE_NETWORK_ADMIN, VIMGR_SE_UNIFIED_ADMIN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterSeType *string `json:"vcenter_se_type,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterTemplateVM *bool `json:"vcenter_template_vm,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterVAppName *string `json:"vcenter_vAppName,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterVAppVendor *string `json:"vcenter_vAppVendor,omitempty"`

	//  Enum options - VMTYPE_SE_VM, VMTYPE_POOL_SRVR. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterVMType *string `json:"vcenter_vm_type,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterVnicDiscovered *bool `json:"vcenter_vnic_discovered,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VMLbWeight *int32 `json:"vm_lb_weight,omitempty"`
}
