package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSRegisterInfo DNS register info
// swagger:model DNSRegisterInfo
type DNSRegisterInfo struct {

	// Placeholder for description of property dns_info of obj type DNSRegisterInfo field type str  type object
	DNSInfo []*DNSInfo `json:"dns_info,omitempty"`

	// error of DNSRegisterInfo.
	Error *string `json:"error,omitempty"`

	// Placeholder for description of property fip of obj type DNSRegisterInfo field type str  type object
	Fip *IPAddr `json:"fip,omitempty"`

	// Number of total_records.
	TotalRecords *int32 `json:"total_records,omitempty"`

	// Placeholder for description of property vip of obj type DNSRegisterInfo field type str  type object
	Vip *IPAddr `json:"vip,omitempty"`

	// vip_id of DNSRegisterInfo.
	VipID *string `json:"vip_id,omitempty"`

	// vs_names of DNSRegisterInfo.
	VsNames []string `json:"vs_names,omitempty"`

	// Unique object identifiers of vss.
	VsUuids []string `json:"vs_uuids,omitempty"`
}
