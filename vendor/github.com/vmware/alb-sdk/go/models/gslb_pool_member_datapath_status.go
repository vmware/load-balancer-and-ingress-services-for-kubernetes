// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbPoolMemberDatapathStatus gslb pool member datapath status
// swagger:model GslbPoolMemberDatapathStatus
type GslbPoolMemberDatapathStatus struct {

	//  Field introduced in 17.1.1.
	Location *GeoLocation `json:"location,omitempty"`

	// Placeholder for description of property oper_status of obj type GslbPoolMemberDatapathStatus field type str  type object
	OperStatus *OperationalStatus `json:"oper_status,omitempty"`

	// Unique object identifier of site.
	SiteUUID *string `json:"site_uuid,omitempty"`
}
