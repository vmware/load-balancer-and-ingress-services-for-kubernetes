// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AuthenticationMatch authentication match
// swagger:model AuthenticationMatch
type AuthenticationMatch struct {

	// Configure client ip addresses. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClientIP *IPAddrMatch `json:"client_ip,omitempty"`

	// Configure the host header. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HostHdr *HostHdrMatch `json:"host_hdr,omitempty"`

	// Configure request paths. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Path *PathMatch `json:"path,omitempty"`
}
