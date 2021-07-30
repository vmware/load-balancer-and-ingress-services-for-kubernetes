// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

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
