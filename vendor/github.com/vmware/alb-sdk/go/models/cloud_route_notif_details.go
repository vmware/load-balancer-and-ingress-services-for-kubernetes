// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CloudRouteNotifDetails cloud route notif details
// swagger:model CloudRouteNotifDetails
type CloudRouteNotifDetails struct {

	// Cloud id. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CcID *string `json:"cc_id,omitempty"`

	// Detailed reason for the route update notification. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Reason *string `json:"reason,omitempty"`

	// Name of route table for which update was performed. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RouteTable *string `json:"route_table,omitempty"`

	// Names of routes for which update was performed. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Routes []string `json:"routes,omitempty"`
}
