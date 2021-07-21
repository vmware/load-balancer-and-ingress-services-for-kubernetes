// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VsSwitchoverParams vs switchover params
// swagger:model VsSwitchoverParams
type VsSwitchoverParams struct {

	// Unique object identifier of se.
	SeUUID *string `json:"se_uuid,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`

	//  Field introduced in 17.1.1.
	// Required: true
	VipID *string `json:"vip_id"`
}
