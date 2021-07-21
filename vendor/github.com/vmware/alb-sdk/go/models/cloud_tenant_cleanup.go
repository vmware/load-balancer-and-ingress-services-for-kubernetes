// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CloudTenantCleanup cloud tenant cleanup
// swagger:model CloudTenantCleanup
type CloudTenantCleanup struct {

	// id of CloudTenantCleanup.
	ID *string `json:"id,omitempty"`

	// Name of the object.
	Name *string `json:"name,omitempty"`

	// Number of num_ports.
	NumPorts *int32 `json:"num_ports,omitempty"`

	// Number of num_se.
	NumSe *int32 `json:"num_se,omitempty"`

	// Number of num_secgrp.
	NumSecgrp *int32 `json:"num_secgrp,omitempty"`

	// Number of num_svrgrp.
	NumSvrgrp *int32 `json:"num_svrgrp,omitempty"`
}
