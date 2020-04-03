package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AuthenticationMatch authentication match
// swagger:model AuthenticationMatch
type AuthenticationMatch struct {

	// Configure client ip addresses. Field introduced in 18.2.5.
	ClientIP *IPAddrMatch `json:"client_ip,omitempty"`

	// Configure the host header. Field introduced in 18.2.5.
	HostHdr *HostHdrMatch `json:"host_hdr,omitempty"`

	// Configure request paths. Field introduced in 18.2.5.
	Path *PathMatch `json:"path,omitempty"`
}
