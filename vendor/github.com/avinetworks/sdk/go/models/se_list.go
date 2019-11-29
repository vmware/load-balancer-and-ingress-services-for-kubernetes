package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeList se list
// swagger:model SeList
type SeList struct {

	// This flag is set when scaling in an SE in admin down mode.
	AdminDownRequested *bool `json:"admin_down_requested,omitempty"`

	// Indicates if an SE is at the current version. This state will now be derived from SE Group runtime. Field deprecated in 18.1.5, 18.2.1.
	AtCurrVer *bool `json:"at_curr_ver,omitempty"`

	// This field indicates the status of programming network reachability to the Virtual Service IP in the cloud. Field introduced in 17.2.3.
	AttachIPStatus *string `json:"attach_ip_status,omitempty"`

	// This flag indicates if network reachability to the Virtual Service IP in the cloud has been successfully programmed. Field introduced in 17.2.3.
	AttachIPSuccess *bool `json:"attach_ip_success,omitempty"`

	// This flag is set when an SE is admin down or scaling in.
	DeleteInProgress *bool `json:"delete_in_progress,omitempty"`

	// This field is not needed with the current implementation of Update RPCs to SEs. Field deprecated in 18.1.5, 18.2.1.
	DownloadSelistOnly *bool `json:"download_selist_only,omitempty"`

	// Placeholder for description of property floating_intf_ip of obj type SeList field type str  type object
	FloatingIntfIP []*IPAddr `json:"floating_intf_ip,omitempty"`

	// This flag indicates whether the geo-files have been pushed to the DNS-VS's SE. No longer used, replaced by SE DataStore. Field deprecated in 18.1.5, 18.2.1. Field introduced in 17.1.1.
	GeoDownload *bool `json:"geo_download,omitempty"`

	// This flag indicates whether the geodb object has been pushed to the DNS-VS's SE. No longer used, replaced by SE DataStore. Field deprecated in 18.1.5, 18.2.1. Field introduced in 17.1.2.
	GeodbDownload *bool `json:"geodb_download,omitempty"`

	// This flag indicates whether the gslb, ghm, gs objects have been pushed to the DNS-VS's SE. No longer used, replaced by SE DataStore. Field deprecated in 18.1.5, 18.2.1. Field introduced in 17.1.1.
	GslbDownload *bool `json:"gslb_download,omitempty"`

	// Updated whenever this entry is created. When the sees this has changed, it means that the SE should disrupt, since there was a delete then create, not an update. Field introduced in 18.1.5,18.2.1.
	Incarnation *string `json:"incarnation,omitempty"`

	// This flag was used to display the SE connected state. This state will now be derived from SE Group runtime. Field deprecated in 18.1.5, 18.2.1.
	IsConnected *bool `json:"is_connected,omitempty"`

	// Placeholder for description of property is_portchannel of obj type SeList field type str  type boolean
	IsPortchannel *bool `json:"is_portchannel,omitempty"`

	// Placeholder for description of property is_primary of obj type SeList field type str  type boolean
	IsPrimary *bool `json:"is_primary,omitempty"`

	// Placeholder for description of property is_standby of obj type SeList field type str  type boolean
	IsStandby *bool `json:"is_standby,omitempty"`

	// Number of memory.
	Memory *int32 `json:"memory,omitempty"`

	// This field is not needed with the current implementation of Update RPCs to SEs. Field deprecated in 18.1.5, 18.2.1.
	PendingDownload *bool `json:"pending_download,omitempty"`

	// SE scaling in status is determined by delete_in_progress. Field deprecated in 18.1.5, 18.2.1.
	ScaleinInProgress *bool `json:"scalein_in_progress,omitempty"`

	// This flag is set when a VS is actively scaling out to this SE. Field introduced in 18.1.5, 18.2.1.
	ScaleoutInProgress *bool `json:"scaleout_in_progress,omitempty"`

	//  It is a reference to an object of type ServiceEngine.
	// Required: true
	SeRef *string `json:"se_ref"`

	// Number of sec_idx.
	SecIdx *int32 `json:"sec_idx,omitempty"`

	// Placeholder for description of property snat_ip of obj type SeList field type str  type object
	SnatIP *IPAddr `json:"snat_ip,omitempty"`

	// Number of vcpus.
	Vcpus *int32 `json:"vcpus,omitempty"`

	// Version of the SE. This state will now be derived from SE Group runtime. Field deprecated in 18.1.5, 18.2.1.
	Version *string `json:"version,omitempty"`

	//  Field introduced in 18.1.1.
	Vip6SubnetMask *int32 `json:"vip6_subnet_mask,omitempty"`

	// Placeholder for description of property vip_intf_ip of obj type SeList field type str  type object
	VipIntfIP *IPAddr `json:"vip_intf_ip,omitempty"`

	// Placeholder for description of property vip_intf_list of obj type SeList field type str  type object
	VipIntfList []*SeVipInterfaceList `json:"vip_intf_list,omitempty"`

	// vip_intf_mac of SeList.
	VipIntfMac *string `json:"vip_intf_mac,omitempty"`

	// Number of vip_subnet_mask.
	VipSubnetMask *int32 `json:"vip_subnet_mask,omitempty"`

	// Number of vlan_id.
	VlanID *int32 `json:"vlan_id,omitempty"`

	// Placeholder for description of property vnic of obj type SeList field type str  type object
	Vnic []*VsSeVnic `json:"vnic,omitempty"`
}
