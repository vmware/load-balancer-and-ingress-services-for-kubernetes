// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MetricsAPISrvDebugFilter metrics Api srv debug filter
// swagger:model MetricsApiSrvDebugFilter
type MetricsAPISrvDebugFilter struct {

	// uuid of the entity. It is a reference to an object of type Virtualservice. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EntityRef *string `json:"entity_ref,omitempty"`
}
