// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

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
