// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSClientPortMatch Dns client port match
// swagger:model DnsClientPortMatch
type DNSClientPortMatch struct {

	// Port number to match against client port number in request. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	ClientPorts *PortMatchGeneric `json:"client_ports"`
}
