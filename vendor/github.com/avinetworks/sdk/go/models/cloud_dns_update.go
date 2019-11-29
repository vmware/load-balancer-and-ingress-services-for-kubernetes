package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CloudDNSUpdate cloud Dns update
// swagger:model CloudDnsUpdate
type CloudDNSUpdate struct {

	// cc_id of CloudDnsUpdate.
	CcID *string `json:"cc_id,omitempty"`

	// dns_fqdn of CloudDnsUpdate.
	DNSFqdn *string `json:"dns_fqdn,omitempty"`

	// error_string of CloudDnsUpdate.
	ErrorString *string `json:"error_string,omitempty"`

	// Placeholder for description of property fip of obj type CloudDnsUpdate field type str  type object
	Fip *IPAddr `json:"fip,omitempty"`

	// Placeholder for description of property vip of obj type CloudDnsUpdate field type str  type object
	Vip *IPAddr `json:"vip,omitempty"`

	// Unique object identifier of vs.
	VsUUID *string `json:"vs_uuid,omitempty"`

	//  Enum options - CLOUD_NONE, CLOUD_VCENTER, CLOUD_OPENSTACK, CLOUD_AWS, CLOUD_VCA, CLOUD_APIC, CLOUD_MESOS, CLOUD_LINUXSERVER, CLOUD_DOCKER_UCP, CLOUD_RANCHER, CLOUD_OSHIFT_K8S, CLOUD_AZURE, CLOUD_GCP.
	Vtype *string `json:"vtype,omitempty"`
}
