// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ALBServicesJobParam a l b services job param
// swagger:model ALBServicesJobParam
type ALBServicesJobParam struct {

	// Parameter name. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Key *string `json:"key,omitempty"`

	// Parameter value. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Value *string `json:"value,omitempty"`
}
