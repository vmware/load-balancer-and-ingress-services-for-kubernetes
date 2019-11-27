package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbSite gslb site
// swagger:model GslbSite
type GslbSite struct {

	// IP Address or a DNS resolvable, fully qualified domain name of the Site Controller Cluster.
	Address *string `json:"address,omitempty"`

	// UUID of the 'Cluster' object of the Controller Cluster in this site.
	// Required: true
	ClusterUUID *string `json:"cluster_uuid"`

	// The DNS VSes on which the GslbServices shall be placed. The site has to be an ACTIVE member.  This field is deprecated in 17.2.3 and replaced by 'dns_vses' field. . Field deprecated in 17.2.3.
	DNSVsUuids []string `json:"dns_vs_uuids,omitempty"`

	// This field identifies the DNS VS and the subdomains it hosts for Gslb services. . Field introduced in 17.2.3.
	DNSVses []*GslbSiteDNSVs `json:"dns_vses,omitempty"`

	// Enable or disable the Site.  This is useful in maintenance scenarios such as upgrade and routine maintenance.  A disabled site's configuration shall be retained but it will not get any new configuration updates.  It shall not participate in Health-Status monitoring.  VIPs of the Virtual Services on the disabled site shall not be sent in DNS response.  When a site transitions from disabled to enabled, it is treated similar to the addition of a new site.
	Enabled *bool `json:"enabled,omitempty"`

	// User can designate certain Avi sites to run health monitor probes for VIPs/VS(es) for this site. This is useful in network deployments where the VIPs/VS(es) are reachable only from certain sites. A typical scenario is a firewall between two GSLB sites. User may want to run health monitor probes from sites on either side of the firewall so that each designated site can derive a datapath view of the reachable members. If the health monitor proxies are not configured, then the default behavior is to run health monitor probes from all the active sites. Field introduced in 17.1.1.
	HmProxies []*GslbHealthMonitorProxy `json:"hm_proxies,omitempty"`

	// This field enables the health monitor shard functionality on a site-basis. Field introduced in 18.2.2.
	HmShardEnabled *bool `json:"hm_shard_enabled,omitempty"`

	// IP Address(es) of the Site's Cluster. For a 3-node cluster, either the cluster vIP is provided, or the list of controller IPs in the cluster are provided.
	IPAddresses []*IPAddr `json:"ip_addresses,omitempty"`

	// Geographic location of the site. Field introduced in 17.1.1.
	Location *GslbGeoLocation `json:"location,omitempty"`

	// The site's member type  A leader is set to ACTIVE while allmembers are set to passive. . Enum options - GSLB_ACTIVE_MEMBER, GSLB_PASSIVE_MEMBER.
	MemberType *string `json:"member_type,omitempty"`

	// Name for the Site Controller Cluster.
	// Required: true
	Name *string `json:"name"`

	// The password used when authenticating with the Site.
	// Required: true
	Password *string `json:"password"`

	// The Site Controller Cluster's REST API port number. Allowed values are 1-65535.
	Port *int32 `json:"port,omitempty"`

	// User can overide the individual GslbPoolMember ratio for all the VIPs/VS(es) of this site. If this field is not  configured then the GslbPoolMember ratio gets applied. . Allowed values are 1-20. Field introduced in 17.1.1.
	Ratio *int32 `json:"ratio,omitempty"`

	// The username used when authenticating with the Site. .
	// Required: true
	Username *string `json:"username"`

	// This field is used as a key in the datastore for the GslbSite table to encapsulate site-related info. . Field introduced in 17.2.5.
	UUID *string `json:"uuid,omitempty"`
}
