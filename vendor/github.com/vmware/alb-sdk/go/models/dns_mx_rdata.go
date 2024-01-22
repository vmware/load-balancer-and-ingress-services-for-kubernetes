// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSMxRdata Dns mx rdata
// swagger:model DnsMxRdata
type DNSMxRdata struct {

	// Fully qualified domain name of a mailserver . The host name maps directly to one or more address records in the DNS table, and must not point to any CNAME records (RFC 2181). Field introduced in 18.2.9, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Host *string `json:"host"`

	// The priority field identifies which mail server should be preferred. Allowed values are 0-65535. Field introduced in 18.2.9, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Priority *uint32 `json:"priority"`
}
