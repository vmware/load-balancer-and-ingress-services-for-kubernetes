// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ServerScaleInParams server scale in params
// swagger:model ServerScaleInParams
type ServerScaleInParams struct {

	// Reason for the manual scale-in. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Reason *string `json:"reason,omitempty"`

	// List of server IDs that should be scaled in. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Servers []*ServerID `json:"servers,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}
