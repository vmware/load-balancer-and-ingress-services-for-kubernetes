// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbSiteDNSVs gslb site Dns vs
// swagger:model GslbSiteDnsVs
type GslbSiteDNSVs struct {

	// This field identifies the DNS VS uuid for this site. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	DNSVsUUID *string `json:"dns_vs_uuid"`

	// This field identifies the subdomains that are hosted on the DNS VS. GslbService(s) whose FQDNs map to one of the subdomains will be hosted on this DNS VS. If no subdomains are configured, then the default behavior is to host all the GslbServices on this DNS VS. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DomainNames []string `json:"domain_names,omitempty"`
}
