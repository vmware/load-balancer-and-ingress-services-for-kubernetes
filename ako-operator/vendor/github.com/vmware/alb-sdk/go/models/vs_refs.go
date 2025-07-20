// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VsRefs vs refs
// swagger:model VsRefs
type VsRefs struct {

	// UUID of the Virtual Service. It is a reference to an object of type VirtualService. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Ref *string `json:"ref,omitempty"`
}
