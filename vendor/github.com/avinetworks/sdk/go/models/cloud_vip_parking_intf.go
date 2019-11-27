package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CloudVipParkingIntf cloud vip parking intf
// swagger:model CloudVipParkingIntf
type CloudVipParkingIntf struct {

	// cc_id of CloudVipParkingIntf.
	CcID *string `json:"cc_id,omitempty"`

	// error_string of CloudVipParkingIntf.
	ErrorString *string `json:"error_string,omitempty"`

	// intf_id of CloudVipParkingIntf.
	IntfID *string `json:"intf_id,omitempty"`

	// subnet_id of CloudVipParkingIntf.
	// Required: true
	SubnetID *string `json:"subnet_id"`

	//  Enum options - CLOUD_NONE, CLOUD_VCENTER, CLOUD_OPENSTACK, CLOUD_AWS, CLOUD_VCA, CLOUD_APIC, CLOUD_MESOS, CLOUD_LINUXSERVER, CLOUD_DOCKER_UCP, CLOUD_RANCHER, CLOUD_OSHIFT_K8S, CLOUD_AZURE, CLOUD_GCP.
	Vtype *string `json:"vtype,omitempty"`
}
