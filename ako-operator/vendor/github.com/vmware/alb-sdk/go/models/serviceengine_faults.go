// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ServiceengineFaults serviceengine faults
// swagger:model ServiceengineFaults
type ServiceengineFaults struct {

	// Enable debug faults. Field introduced in 20.1.6. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DebugFaults *bool `json:"debug_faults,omitempty"`
}
