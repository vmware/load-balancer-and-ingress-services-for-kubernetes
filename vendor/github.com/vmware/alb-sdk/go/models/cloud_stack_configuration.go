package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CloudStackConfiguration cloud stack configuration
// swagger:model CloudStackConfiguration
type CloudStackConfiguration struct {

	// CloudStack API Key.
	// Required: true
	AccessKeyID *string `json:"access_key_id"`

	// CloudStack API URL.
	// Required: true
	APIURL *string `json:"api_url"`

	// If controller's management IP is in a private network, a publicly accessible IP to reach the controller.
	CntrPublicIP *string `json:"cntr_public_ip,omitempty"`

	// Default hypervisor type. Enum options - DEFAULT, VMWARE_ESX, KVM, VMWARE_VSAN, XEN.
	Hypervisor *string `json:"hypervisor,omitempty"`

	// Avi Management network name.
	// Required: true
	MgmtNetworkName *string `json:"mgmt_network_name"`

	// Avi Management network name.
	MgmtNetworkUUID *string `json:"mgmt_network_uuid,omitempty"`

	// CloudStack Secret Key.
	// Required: true
	SecretAccessKey *string `json:"secret_access_key"`
}
