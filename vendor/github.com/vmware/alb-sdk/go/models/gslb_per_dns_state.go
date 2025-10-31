// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbPerDNSState gslb per Dns state
// swagger:model GslbPerDnsState
type GslbPerDNSState struct {

	// This field describes the GeoDbProfile download status to the dns-vs. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GeoDownload *GslbDownloadStatus `json:"geo_download,omitempty"`

	// This field describes the Gslb, GslbService, HealthMonitor download status to the dns-vs. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GslbDownload *GslbDownloadStatus `json:"gslb_download,omitempty"`

	// Configured dns-vs-name at the site. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OperStatus *OperationalStatus `json:"oper_status,omitempty"`

	// This field describes the SubDomain placement rules for this DNS-VS. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PlacementRules []*GslbSubDomainPlacementRuntime `json:"placement_rules,omitempty"`

	// The service engines associated with the DNS-VS. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeList []string `json:"se_list,omitempty"`

	// Configured dns-vs-uuid at the site. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`

	// This field indicates that the local VS is configured to be a DNS service. The services, network profile and application profile are configured in Virtual Service for DNS operations. . Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ValidDNSVs *bool `json:"valid_dns_vs,omitempty"`
}
