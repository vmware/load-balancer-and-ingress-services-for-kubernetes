package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// MatchTarget match target
// swagger:model MatchTarget
type MatchTarget struct {

	// Configure client ip addresses.
	ClientIP *IPAddrMatch `json:"client_ip,omitempty"`

	// Configure HTTP cookie(s).
	Cookie *CookieMatch `json:"cookie,omitempty"`

	// Configure HTTP header(s).
	Hdrs []*HdrMatch `json:"hdrs,omitempty"`

	// Configure the host header.
	HostHdr *HostHdrMatch `json:"host_hdr,omitempty"`

	// Configure HTTP methods.
	Method *MethodMatch `json:"method,omitempty"`

	// Configure request paths.
	Path *PathMatch `json:"path,omitempty"`

	// Configure the type of HTTP protocol.
	Protocol *ProtocolMatch `json:"protocol,omitempty"`

	// Configure request query.
	Query *QueryMatch `json:"query,omitempty"`

	// Configure versions of the HTTP protocol.
	Version *HTTPVersionMatch `json:"version,omitempty"`

	// Configure virtual service ports.
	VsPort *PortMatch `json:"vs_port,omitempty"`
}
