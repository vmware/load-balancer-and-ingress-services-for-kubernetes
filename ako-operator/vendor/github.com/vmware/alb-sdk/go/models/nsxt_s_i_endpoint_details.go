// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NsxtSIEndpointDetails nsxt s i endpoint details
// swagger:model NsxtSIEndpointDetails
type NsxtSIEndpointDetails struct {

	// VirtualEndpoint Path. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Endpoint *string `json:"endpoint,omitempty"`

	// Error message. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ErrorString *string `json:"error_string,omitempty"`

	// ServiceEngineGroup name. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Segroup *string `json:"segroup,omitempty"`

	// Services where endpoint refers. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Services []string `json:"services,omitempty"`

	// Endpoint Target IPs. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TargetIps []string `json:"targetIps,omitempty"`

	// Tier1 path. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Tier1 *string `json:"tier1,omitempty"`
}
