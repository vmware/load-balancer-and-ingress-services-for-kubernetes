// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VcenterQuarantinedHostParams vcenter quarantined host params
// swagger:model VcenterQuarantinedHostParams
type VcenterQuarantinedHostParams struct {

	// Vcenter cloud id. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CloudUUID *string `json:"cloud_uuid,omitempty"`
}
