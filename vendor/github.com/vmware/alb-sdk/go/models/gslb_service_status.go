// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbServiceStatus gslb service status
// swagger:model GslbServiceStatus
type GslbServiceStatus struct {

	// details of GslbServiceStatus.
	Details []string `json:"details,omitempty"`

	// Placeholder for description of property gs_runtime of obj type GslbServiceStatus field type str  type object
	GsRuntime *GslbServiceRuntime `json:"gs_runtime,omitempty"`

	// Name of the object.
	Name *string `json:"name,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
