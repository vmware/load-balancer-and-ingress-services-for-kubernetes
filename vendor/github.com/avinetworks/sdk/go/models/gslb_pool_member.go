package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbPoolMember gslb pool member
// swagger:model GslbPoolMember
type GslbPoolMember struct {

	// The Cloud UUID of the Site. Field introduced in 17.1.2.
	CloudUUID *string `json:"cloud_uuid,omitempty"`

	// The Cluster UUID of the Site.
	ClusterUUID *string `json:"cluster_uuid,omitempty"`

	// User provided information that records member details such as application owner name, contact, etc. Field introduced in 17.1.3.
	Description *string `json:"description,omitempty"`

	// Enable or Disable member to decide if this address should be provided in DNS responses.
	Enabled *bool `json:"enabled,omitempty"`

	// The pool member is configured with a fully qualified domain name.  The FQDN is resolved to an IP address by the controller. DNS service shall health monitor the resolved IP address while it will return the fqdn(cname) in the DNS response.If the user has configured an IP address (in addition to the FQDN), then the IP address will get overwritten whenever periodic FQDN refresh is done by the controller. .
	Fqdn *string `json:"fqdn,omitempty"`

	// Internal generated system-field. Field deprecated in 18.2.2. Field introduced in 17.1.1.
	HmProxies []*GslbHealthMonitorProxy `json:"hm_proxies,omitempty"`

	// IP address of the pool member. If this IP address is hosted via an AVI virtual service, then the user should configure the cluster uuid and virtual service uuid. If this IP address is hosted on a third-party device and the device is tagged/tethered to a third-party site, then user can configure the third-party site uuid.  User may configure the IP address without the cluster uuid or the virtual service uuid.  In this option, some advanced site related features cannot be enabled. If the user has configured a fqdn for the pool member, then it takes precedence and will overwrite the configured IP address. .
	IP *IPAddr `json:"ip,omitempty"`

	// Geographic location of the pool member. Field introduced in 17.1.1.
	Location *GslbGeoLocation `json:"location,omitempty"`

	// Alternate IP addresses of the pool member. In usual deployments, the VIP in the virtual service is a private IP address. This gets configured in the 'ip' field of the GSLB service. This field is used to host the public IP address for the VIP, which gets NATed to the private IP by a firewall. Client DNS requests coming in from within the intranet should have the private IP served in the A record, and requests from outside this should be served the public IP address. Field introduced in 17.1.2.
	PublicIP *GslbIPAddr `json:"public_ip,omitempty"`

	// Overrides the default ratio of 1.  Reduces the percentage the LB algorithm would pick the server in relation to its peers.  Range is 1-20. Allowed values are 1-20.
	Ratio *int32 `json:"ratio,omitempty"`

	// Select local virtual service in the specified controller cluster belonging to this GSLB service. The virtual service may have multiple IP addresses and FQDNs.  User will have to choose IP address or FQDN and configure it in the respective field. .
	VsUUID *string `json:"vs_uuid,omitempty"`
}
