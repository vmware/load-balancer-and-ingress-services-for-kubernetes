// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// KniPortRange kni port range
// swagger:model KniPortRange
type KniPortRange struct {

	// Protocol associated with port range. Enum options - KNI_PROTO_TCP, KNI_PROTO_UDP. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Protocol *string `json:"protocol"`

	// Port range to be allowed to KNI. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Range *PortRange `json:"range"`
}
