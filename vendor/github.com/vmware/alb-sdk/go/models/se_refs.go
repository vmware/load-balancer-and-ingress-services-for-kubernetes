// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeRefs se refs
// swagger:model SeRefs
type SeRefs struct {

	// UUID of the Service Engine. It is a reference to an object of type ServiceEngine. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Ref *string `json:"ref,omitempty"`
}
