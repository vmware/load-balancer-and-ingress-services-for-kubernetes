// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbSite gslb site
// swagger:model GslbSite
type GslbSite struct {

	// IP Address or a DNS resolvable, fully qualified domain name of the Site Controller Cluster. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Address *string `json:"address,omitempty"`

	// UUID of the 'Cluster' object of the Controller Cluster in this site. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ClusterUUID *string `json:"cluster_uuid"`

	// This field identifies the DNS VS and the subdomains it hosts for Gslb services. . Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DNSVses []*GslbSiteDNSVs `json:"dns_vses,omitempty"`

	// Enable or disable the Site.  This is useful in maintenance scenarios such as upgrade and routine maintenance.  A disabled site's configuration shall be retained but it will not get any new configuration updates.  It shall not participate in Health-Status monitoring.  VIPs of the Virtual Services on the disabled site shall not be sent in DNS response.  When a site transitions from disabled to enabled, it is treated similar to the addition of a new site. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Enabled *bool `json:"enabled,omitempty"`

	// User can designate certain Avi sites to run health monitor probes for VIPs/VS(es) for this site. This is useful in network deployments where the VIPs/VS(es) are reachable only from certain sites. A typical scenario is a firewall between two GSLB sites. User may want to run health monitor probes from sites on either side of the firewall so that each designated site can derive a datapath view of the reachable members. If the health monitor proxies are not configured, then the default behavior is to run health monitor probes from all the active sites. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HmProxies []*GslbHealthMonitorProxy `json:"hm_proxies,omitempty"`

	// This field enables the health monitor shard functionality on a site-basis. Field introduced in 18.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HmShardEnabled *bool `json:"hm_shard_enabled,omitempty"`

	// IP Address(es) of the Site's Cluster. For a 3-node cluster, either the cluster vIP is provided, or the list of controller IPs in the cluster are provided. Maximum of 3 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IPAddresses []*IPAddr `json:"ip_addresses,omitempty"`

	// Geographic location of the site. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Location *GslbGeoLocation `json:"location,omitempty"`

	// The site's member type  A leader is set to ACTIVE while allmembers are set to passive. . Enum options - GSLB_ACTIVE_MEMBER, GSLB_PASSIVE_MEMBER. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MemberType *string `json:"member_type,omitempty"`

	// Name for the Site Controller Cluster. After any changes to site name, references to GSLB site name should be updated manually. Ex  Site name used in DNS policies or Topology policies should be updated to use the new site name. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// The password used when authenticating with the Site. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Password *string `json:"password"`

	// The Site Controller Cluster's REST API port number. Allowed values are 1-65535. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Port *uint32 `json:"port,omitempty"`

	// User can overide the individual GslbPoolMember ratio for all the VIPs/VS(es) of this site. If this field is not  configured then the GslbPoolMember ratio gets applied. . Allowed values are 1-20. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Ratio *uint32 `json:"ratio,omitempty"`

	// This modes applies to follower sites. When an active site is in suspend mode, the site does not receive any further federated objects. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SuspendMode *bool `json:"suspend_mode,omitempty"`

	// The username used when authenticating with the Site. . Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Username *string `json:"username"`

	// This field is used as a key in the datastore for the GslbSite table to encapsulate site-related info. . Field introduced in 17.2.5. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}
