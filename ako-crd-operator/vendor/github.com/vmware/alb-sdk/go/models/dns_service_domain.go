// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSServiceDomain Dns service domain
// swagger:model DnsServiceDomain
type DNSServiceDomain struct {

	// Service domain *string used for FQDN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	DomainName *string `json:"domain_name"`

	// Third-party Authoritative domain requests are delegated toDNS VirtualService's pool of nameservers. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PassThrough *bool `json:"pass_through,omitempty"`

	// TTL value for DNS records. Allowed values are 1-604800. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RecordTTL uint32 `json:"record_ttl,omitempty"`
}
