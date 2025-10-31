// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SecMgrUAEventDetails sec mgr u a event details
// swagger:model SecMgrUAEventDetails
type SecMgrUAEventDetails struct {

	// Error descibing UA cache status in controller. Field introduced in 21.1.2. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Error *string `json:"error,omitempty"`
}
