package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbPerDNSState gslb per Dns state
// swagger:model GslbPerDnsState
type GslbPerDNSState struct {

	// This field describes the GeoDbProfile download status to the dns-vs. Field introduced in 17.1.1.
	GeoDownload *GslbDownloadStatus `json:"geo_download,omitempty"`

	// This field describes the Gslb, GslbService, HealthMonitor download status to the dns-vs. Field introduced in 17.1.1.
	GslbDownload *GslbDownloadStatus `json:"gslb_download,omitempty"`

	// Configured dns-vs-name at the site.
	Name *string `json:"name,omitempty"`

	// Placeholder for description of property oper_status of obj type GslbPerDnsState field type str  type object
	OperStatus *OperationalStatus `json:"oper_status,omitempty"`

	// This field describes the SubDomain placement rules for this DNS-VS. Field introduced in 17.2.3.
	PlacementRules []*GslbSubDomainPlacementRuntime `json:"placement_rules,omitempty"`

	// The service engines associated with the DNS-VS. Field introduced in 17.1.1.
	SeList []string `json:"se_list,omitempty"`

	// Configured dns-vs-uuid at the site.
	UUID *string `json:"uuid,omitempty"`

	// This field indicates that the local VS is configured to be a DNS service. The services, network profile and application profile are configured in Virtual Service for DNS operations. .
	ValidDNSVs *bool `json:"valid_dns_vs,omitempty"`
}
