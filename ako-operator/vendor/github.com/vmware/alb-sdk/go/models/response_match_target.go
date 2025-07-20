// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ResponseMatchTarget response match target
// swagger:model ResponseMatchTarget
type ResponseMatchTarget struct {

	// Configure client ip addresses. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClientIP *IPAddrMatch `json:"client_ip,omitempty"`

	// Configure HTTP cookie(s). Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Cookie *CookieMatch `json:"cookie,omitempty"`

	// Configure HTTP headers. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Hdrs []*HdrMatch `json:"hdrs,omitempty"`

	// Configure the host header. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HostHdr *HostHdrMatch `json:"host_hdr,omitempty"`

	// Configure the location header. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LocHdr *LocationHdrMatch `json:"loc_hdr,omitempty"`

	// Configure HTTP methods. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Method *MethodMatch `json:"method,omitempty"`

	// Configure request paths. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Path *PathMatch `json:"path,omitempty"`

	// Configure the type of HTTP protocol. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Protocol *ProtocolMatch `json:"protocol,omitempty"`

	// Configure request query. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Query *QueryMatch `json:"query,omitempty"`

	// Configure the HTTP headers in response. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RspHdrs []*HdrMatch `json:"rsp_hdrs,omitempty"`

	// Configure source ip addresses. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SourceIP *IPAddrMatch `json:"source_ip,omitempty"`

	// Configure the HTTP status code(s). Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Status *HttpstatusMatch `json:"status,omitempty"`

	// Configure versions of the HTTP protocol. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Version *HTTPVersionMatch `json:"version,omitempty"`

	// Configure virtual service ports. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsPort *PortMatch `json:"vs_port,omitempty"`
}
