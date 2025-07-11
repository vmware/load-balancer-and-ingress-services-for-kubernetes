// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PoolRuntimeSummary pool runtime summary
// swagger:model PoolRuntimeSummary
type PoolRuntimeSummary struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	NumServers *int64 `json:"num_servers"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	NumServersEnabled *int64 `json:"num_servers_enabled"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	NumServersUp *int64 `json:"num_servers_up"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	OperStatus *OperationalStatus `json:"oper_status"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PercentServersUpEnabled *int32 `json:"percent_servers_up_enabled,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PercentServersUpTotal *int32 `json:"percent_servers_up_total,omitempty"`
}
