package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbDNSGeoUpdate gslb Dns geo update
// swagger:model GslbDnsGeoUpdate
type GslbDNSGeoUpdate struct {

	// GslbGeoDbProfile object that is pushed on on a per Dns basis. Field deprecated in 18.1.5, 18.2.1. Field introduced in 17.1.1.
	ObjInfo []*GslbObjectInfo `json:"obj_info,omitempty"`

	//  Enum options - GSLB_NONE, GSLB_CREATE, GSLB_UPDATE, GSLB_DELETE, GSLB_PURGE, GSLB_DECL. Field deprecated in 18.1.5, 18.2.1. Field introduced in 17.1.1.
	Ops *string `json:"ops,omitempty"`

	//  Field deprecated in 18.1.5, 18.2.1. Field introduced in 17.1.1.
	SeList []string `json:"se_list,omitempty"`
}
