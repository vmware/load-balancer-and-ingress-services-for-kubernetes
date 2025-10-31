// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSRuleActionGslbSiteSelection Dns rule action gslb site selection
// swagger:model DnsRuleActionGslbSiteSelection
type DNSRuleActionGslbSiteSelection struct {

	// GSLB fallback sites to use in case the desired site is down. Field introduced in 17.2.5. Maximum of 64 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FallbackSiteNames []string `json:"fallback_site_names,omitempty"`

	// When set to true, GSLB site is a preferred site. This setting comes into play when the site is down, as well as no configured fallback site is available (all fallback sites are also down), then any one available site is selected based on the default algorithm for GSLB pool member selection. Field introduced in 17.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IsSitePreferred *bool `json:"is_site_preferred,omitempty"`

	// GSLB site name. Field introduced in 17.1.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	SiteName *string `json:"site_name"`
}
