// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbObj gslb obj
// swagger:model GslbObj
type GslbObj struct {

	//  Field introduced in 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GslbGeoDbProfileUUID *string `json:"gslb_geo_db_profile_uuid,omitempty"`

	//  Field introduced in 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GslbServiceUUID *string `json:"gslb_service_uuid,omitempty"`

	//  Field introduced in 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GslbUUID *string `json:"gslb_uuid,omitempty"`
}
