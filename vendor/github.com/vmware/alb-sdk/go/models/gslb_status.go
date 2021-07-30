// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbStatus gslb status
// swagger:model GslbStatus
type GslbStatus struct {

	// details of GslbStatus.
	Details []string `json:"details,omitempty"`

	// Placeholder for description of property gslb_runtime of obj type GslbStatus field type str  type object
	GslbRuntime *GslbRuntime `json:"gslb_runtime,omitempty"`

	// Name of the object.
	Name *string `json:"name,omitempty"`

	//  Field introduced in 17.2.5.
	Site *GslbSiteRuntime `json:"site,omitempty"`

	//  Field introduced in 17.2.5.
	ThirdPartySite *GslbThirdPartySiteRuntime `json:"third_party_site,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
