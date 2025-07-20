// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbServiceDownResponse gslb service down response
// swagger:model GslbServiceDownResponse
type GslbServiceDownResponse struct {

	// Fallback CNAME to use in response to the client query when the GSLB service is DOWN. Field introduced in 30.2.1, 22.1.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FallbackCname *string `json:"fallback_cname,omitempty"`

	// Fallback IP address to use in A response to the client query when the GSLB service is DOWN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FallbackIP *IPAddr `json:"fallback_ip,omitempty"`

	// Fallback IPV6 address to use in AAAA response to the client query when the GSLB service is DOWN. Field introduced in 18.2.8. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FallbackIp6 *IPAddr `json:"fallback_ip6,omitempty"`

	// Response from DNS service towards the client when the GSLB service is DOWN. Enum options - GSLB_SERVICE_DOWN_RESPONSE_NONE, GSLB_SERVICE_DOWN_RESPONSE_ALL_RECORDS, GSLB_SERVICE_DOWN_RESPONSE_FALLBACK_IP, GSLB_SERVICE_DOWN_RESPONSE_EMPTY, GSLB_SERVICE_DOWN_RESPONSE_CNAME. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Type *string `json:"type"`
}
