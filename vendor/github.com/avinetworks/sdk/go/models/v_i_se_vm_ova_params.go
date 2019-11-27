package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VISeVMOvaParams v i se Vm ova params
// swagger:model VISeVmOvaParams
type VISeVMOvaParams struct {

	// Unique object identifier of controller_cluster.
	ControllerClusterUUID *string `json:"controller_cluster_uuid,omitempty"`

	// controller_ip_addr of VISeVmOvaParams.
	// Required: true
	ControllerIPAddr *string `json:"controller_ip_addr"`

	//  Enum options - APIC_MODE, NON_APIC_MODE.
	Mode *string `json:"mode,omitempty"`

	// rm_cookie of VISeVmOvaParams.
	RmCookie *string `json:"rm_cookie,omitempty"`

	// se_auth_token of VISeVmOvaParams.
	SeAuthToken *string `json:"se_auth_token,omitempty"`

	// sevm_name of VISeVmOvaParams.
	// Required: true
	SevmName *string `json:"sevm_name"`

	// Placeholder for description of property single_socket_affinity of obj type VISeVmOvaParams field type str  type boolean
	SingleSocketAffinity *bool `json:"single_socket_affinity,omitempty"`

	// Placeholder for description of property vcenter_cpu_reserv of obj type VISeVmOvaParams field type str  type boolean
	VcenterCPUReserv *bool `json:"vcenter_cpu_reserv,omitempty"`

	// Placeholder for description of property vcenter_ds_include of obj type VISeVmOvaParams field type str  type boolean
	VcenterDsInclude *bool `json:"vcenter_ds_include,omitempty"`

	// Placeholder for description of property vcenter_ds_info of obj type VISeVmOvaParams field type str  type object
	VcenterDsInfo []*VcenterDatastore `json:"vcenter_ds_info,omitempty"`

	//  Enum options - VCENTER_DATASTORE_ANY, VCENTER_DATASTORE_LOCAL, VCENTER_DATASTORE_SHARED.
	VcenterDsMode *string `json:"vcenter_ds_mode,omitempty"`

	// vcenter_host of VISeVmOvaParams.
	VcenterHost *string `json:"vcenter_host,omitempty"`

	// vcenter_internal of VISeVmOvaParams.
	VcenterInternal *string `json:"vcenter_internal,omitempty"`

	// Placeholder for description of property vcenter_mem_reserv of obj type VISeVmOvaParams field type str  type boolean
	VcenterMemReserv *bool `json:"vcenter_mem_reserv,omitempty"`

	// Number of vcenter_num_mem.
	VcenterNumMem *int64 `json:"vcenter_num_mem,omitempty"`

	// Number of vcenter_num_se_cores.
	VcenterNumSeCores *int32 `json:"vcenter_num_se_cores,omitempty"`

	// vcenter_ovf_path of VISeVmOvaParams.
	VcenterOvfPath *string `json:"vcenter_ovf_path,omitempty"`

	// Number of vcenter_se_disk_size_KB.
	VcenterSeDiskSizeKB *int32 `json:"vcenter_se_disk_size_KB,omitempty"`

	// vcenter_se_mgmt_nw of VISeVmOvaParams.
	VcenterSeMgmtNw *string `json:"vcenter_se_mgmt_nw,omitempty"`

	// vcenter_vm_folder of VISeVmOvaParams.
	VcenterVMFolder *string `json:"vcenter_vm_folder,omitempty"`
}
