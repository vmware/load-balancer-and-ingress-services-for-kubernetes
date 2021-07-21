// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbDNSUpdate gslb Dns update
// swagger:model GslbDnsUpdate
type GslbDNSUpdate struct {

	// Number of clear_on_max_retries.
	ClearOnMaxRetries *int32 `json:"clear_on_max_retries,omitempty"`

	// List of Geo DB Profiles associated with this DNS VS. Field introduced in 18.2.3.
	GslbGeoDbProfileUuids []string `json:"gslb_geo_db_profile_uuids,omitempty"`

	// List of Gslb Services associated with the DNS VS. Field introduced in 18.2.3.
	GslbServiceUuids []string `json:"gslb_service_uuids,omitempty"`

	// Gslb object associated with the DNS VS. Field introduced in 18.2.3.
	GslbUuids []string `json:"gslb_uuids,omitempty"`

	// Gslb, GslbService objects that is pushed on a per Dns basis. Field introduced in 17.1.1.
	ObjInfo []*GslbObjectInfo `json:"obj_info,omitempty"`

	// Number of send_interval.
	SendInterval *int32 `json:"send_interval,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
