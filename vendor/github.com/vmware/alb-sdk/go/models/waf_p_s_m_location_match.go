// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// WafPSMLocationMatch waf p s m location match
// swagger:model WafPSMLocationMatch
type WafPSMLocationMatch struct {

	// Apply the rules only to requests that match the specified Host header. If this is not set, the host header will not be checked. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Host *HostHdrMatch `json:"host,omitempty"`

	// Apply the rules only to requests that have the specified methods. If this is not set, the method will not be checked. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Methods *MethodMatch `json:"methods,omitempty"`

	// Apply the rules only to requests that match the specified URI. If this is not set, the path will not be checked. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Path *PathMatch `json:"path,omitempty"`
}
