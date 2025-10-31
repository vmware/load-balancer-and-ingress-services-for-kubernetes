// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeGeoDbDetails se geo db details
// swagger:model SeGeoDbDetails
type SeGeoDbDetails struct {

	// Geo Db file name. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FileName *string `json:"file_name,omitempty"`

	// Name of the Gslb Geo Db Profile. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GeoDbProfileName *string `json:"geo_db_profile_name,omitempty"`

	// UUID of the Gslb Geo Db Profile. It is a reference to an object of type GslbGeoDbProfile. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GeoDbProfileRef *string `json:"geo_db_profile_ref,omitempty"`

	// Reason for Gslb Geo Db failure. Enum options - NO_ERROR, FILE_ERROR, FILE_FORMAT_ERROR, FILE_NO_RESOURCES. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Reason *string `json:"reason,omitempty"`

	// UUID of the SE responsible for this event. It is a reference to an object of type ServiceEngine. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeRef *string `json:"se_ref,omitempty"`

	// VIP id. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VipID *string `json:"vip_id,omitempty"`

	// Virtual Service name. It is a reference to an object of type VirtualService. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VirtualService *string `json:"virtual_service,omitempty"`
}
