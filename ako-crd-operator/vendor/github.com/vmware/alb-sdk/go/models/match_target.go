// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MatchTarget match target
// swagger:model MatchTarget
type MatchTarget struct {

	// Configure the bot classification result. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	BotDetectionResult *BotDetectionMatch `json:"bot_detection_result,omitempty"`

	// Configure client ip addresses. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClientIP *IPAddrMatch `json:"client_ip,omitempty"`

	// Configure HTTP cookie(s). Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Cookie *CookieMatch `json:"cookie,omitempty"`

	// Configure the geo information. Field introduced in 21.1.1. Maximum of 1 items allowed. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	GeoMatches []*GeoMatch `json:"geo_matches,omitempty"`

	// Configure HTTP header(s). All configured headers must match. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Hdrs []*HdrMatch `json:"hdrs,omitempty"`

	// Configure the host header. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HostHdr *HostHdrMatch `json:"host_hdr,omitempty"`

	// Configure IP reputation. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IPReputationType *IPReputationTypeMatch `json:"ip_reputation_type,omitempty"`

	// Configure HTTP methods. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Method *MethodMatch `json:"method,omitempty"`

	// Configure request paths. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Path *PathMatch `json:"path,omitempty"`

	// Configure the type of HTTP protocol. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Protocol *ProtocolMatch `json:"protocol,omitempty"`

	// Configure request query. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Query *QueryMatch `json:"query,omitempty"`

	// Configure source ip addresses. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SourceIP *IPAddrMatch `json:"source_ip,omitempty"`

	// Configure the TLS fingerprint. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TLSFingerprintMatch *TLSFingerprintMatch `json:"tls_fingerprint_match,omitempty"`

	// Configure versions of the HTTP protocol. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Version *HTTPVersionMatch `json:"version,omitempty"`

	// Configure virtual service ports. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsPort *PortMatch `json:"vs_port,omitempty"`
}
