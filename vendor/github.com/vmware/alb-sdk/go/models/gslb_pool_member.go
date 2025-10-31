// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbPoolMember gslb pool member
// swagger:model GslbPoolMember
type GslbPoolMember struct {

	// The Cloud UUID of the Site. Field introduced in 17.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudUUID *string `json:"cloud_uuid,omitempty"`

	// The Cluster UUID of the Site. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClusterUUID *string `json:"cluster_uuid,omitempty"`

	// User provided information that records member details such as application owner name, contact, etc. Field introduced in 17.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// Enable or Disable member to decide if this address should be provided in DNS responses. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Enabled *bool `json:"enabled,omitempty"`

	// The pool member is configured with a fully qualified domain name.  The FQDN is resolved to an IP address by the controller. DNS service shall health monitor the resolved IP address while it will return the fqdn(cname) in the DNS response.If the user has configured an IP address (in addition to the FQDN), then the IP address will get overwritten whenever periodic FQDN refresh is done by the controller. . Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Fqdn *string `json:"fqdn,omitempty"`

	// Hostname to be used as host header for http health monitors and as TLS server name for https health monitors.(By default, the fqdn of the GSLB pool member or GSLB service is used.) Note  this field is not used as http host header when exact_http_request is set in the health monitor. . Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Hostname *string `json:"hostname,omitempty"`

	// IP address of the pool member. If this IP address is hosted via an AVI virtual service, then the user should configure the cluster uuid and virtual service uuid. If this IP address is hosted on a third-party device and the device is tagged/tethered to a third-party site, then user can configure the third-party site uuid.  User may configure the IP address without the cluster uuid or the virtual service uuid.  In this option, some advanced site related features cannot be enabled. If the user has configured a fqdn for the pool member, then it takes precedence and will overwrite the configured IP address. . Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IP *IPAddr `json:"ip,omitempty"`

	// Geographic location of the pool member. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Location *GslbGeoLocation `json:"location,omitempty"`

	// Preference order of this member in the group. The DNS Service chooses the member with the lowest preference that is operationally up. Allowed values are 1-128. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PreferenceOrder *uint32 `json:"preference_order,omitempty"`

	// Alternate IP addresses of the pool member. In usual deployments, the VIP in the virtual service is a private IP address. This gets configured in the 'ip' field of the GSLB service. This field is used to host the public IP address for the VIP, which gets NATed to the private IP by a firewall. Client DNS requests coming in from within the intranet should have the private IP served in the A record, and requests from outside this should be served the public IP address. Field introduced in 17.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PublicIP *GslbIPAddr `json:"public_ip,omitempty"`

	// Overrides the default ratio of 1.  Reduces the percentage the LB algorithm would pick the server in relation to its peers.  Range is 1-20. Allowed values are 1-20. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Ratio *uint32 `json:"ratio,omitempty"`

	// This field indicates if the fqdn should be resolved to a v6 or a v4 address family. . Field introduced in 18.2.8, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ResolveFqdnToV6 *bool `json:"resolve_fqdn_to_v6,omitempty"`

	// Select local virtual service in the specified controller cluster belonging to this GSLB service. The virtual service may have multiple IP addresses and FQDNs.  User will have to choose IP address or FQDN and configure it in the respective field. . Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsUUID *string `json:"vs_uuid,omitempty"`
}
