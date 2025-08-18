// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GCPVIPILB g c p v IP i l b
// swagger:model GCPVIPILB
type GCPVIPILB struct {

	// Google Cloud Router Names to advertise BYOIP. Field introduced in 18.2.9, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudRouterNames []string `json:"cloud_router_names,omitempty"`
}
