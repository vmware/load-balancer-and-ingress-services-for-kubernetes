// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SecurityMgrRuntime security mgr runtime
// swagger:model SecurityMgrRuntime
type SecurityMgrRuntime struct {

	//  Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Thresholds []*SecMgrThreshold `json:"thresholds,omitempty"`
}
