package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ResponseMatchTarget response match target
// swagger:model ResponseMatchTarget
type ResponseMatchTarget struct {

	// Configure client ip addresses.
	ClientIP *IPAddrMatch `json:"client_ip,omitempty"`

	// Configure HTTP cookie(s).
	Cookie *CookieMatch `json:"cookie,omitempty"`

	// Configure HTTP headers.
	Hdrs []*HdrMatch `json:"hdrs,omitempty"`

	// Configure the host header.
	HostHdr *HostHdrMatch `json:"host_hdr,omitempty"`

	// Configure the location header.
	LocHdr *LocationHdrMatch `json:"loc_hdr,omitempty"`

	// Configure HTTP methods.
	Method *MethodMatch `json:"method,omitempty"`

	// Configure request paths.
	Path *PathMatch `json:"path,omitempty"`

	// Configure the type of HTTP protocol.
	Protocol *ProtocolMatch `json:"protocol,omitempty"`

	// Configure request query.
	Query *QueryMatch `json:"query,omitempty"`

	// Configure the HTTP headers in response.
	RspHdrs []*HdrMatch `json:"rsp_hdrs,omitempty"`

	// Configure the HTTP status code(s).
	Status *HttpstatusMatch `json:"status,omitempty"`

	// Configure versions of the HTTP protocol.
	Version *HTTPVersionMatch `json:"version,omitempty"`

	// Configure virtual service ports.
	VsPort *PortMatch `json:"vs_port,omitempty"`
}
