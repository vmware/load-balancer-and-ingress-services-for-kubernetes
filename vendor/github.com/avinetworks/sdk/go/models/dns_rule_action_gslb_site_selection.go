package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSRuleActionGslbSiteSelection Dns rule action gslb site selection
// swagger:model DnsRuleActionGslbSiteSelection
type DNSRuleActionGslbSiteSelection struct {

	// GSLB fallback sites to use in case the desired site is down. Field introduced in 17.2.5.
	FallbackSiteNames []string `json:"fallback_site_names,omitempty"`

	// When set to true, GSLB site is a preferred site. This setting comes into play when the site is down, as well as no configured fallback site is available (all fallback sites are also down), then any one available site is selected based on the default algorithm for GSLB pool member selection. Field introduced in 17.2.5.
	IsSitePreferred *bool `json:"is_site_preferred,omitempty"`

	// GSLB site name. Field introduced in 17.1.5.
	// Required: true
	SiteName *string `json:"site_name"`
}
