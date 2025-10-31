// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeVsDelFlowsDisrupted se vs del flows disrupted
// swagger:model SeVsDelFlowsDisrupted
type SeVsDelFlowsDisrupted struct {

	// Name of the VS which was deleted from the SE. It is a reference to an object of type VirtualService. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DeletedVsName *string `json:"deleted_vs_name,omitempty"`

	// Number of VS flows disrupted when VS was deleted from the SE. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumVsFlowsDisrupted *int32 `json:"num_vs_flows_disrupted,omitempty"`

	// Name of the SE reporting this event. It is a reference to an object of type ServiceEngine. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ReportingSeName *string `json:"reporting_se_name,omitempty"`
}
