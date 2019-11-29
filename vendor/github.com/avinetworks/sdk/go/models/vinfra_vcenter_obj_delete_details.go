package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VinfraVcenterObjDeleteDetails vinfra vcenter obj delete details
// swagger:model VinfraVcenterObjDeleteDetails
type VinfraVcenterObjDeleteDetails struct {

	// obj_name of VinfraVcenterObjDeleteDetails.
	// Required: true
	ObjName *string `json:"obj_name"`

	// vcenter of VinfraVcenterObjDeleteDetails.
	// Required: true
	Vcenter *string `json:"vcenter"`
}
