// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbService gslb service
// swagger:model GslbService
type GslbService struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// The federated application persistence associated with GslbService site persistence functionality. . It is a reference to an object of type ApplicationPersistenceProfile. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ApplicationPersistenceProfileRef *string `json:"application_persistence_profile_ref,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// GS member's overall health status is derived based on a combination of controller and datapath health-status inputs. Note that the datapath status is determined by the association of health monitor profiles. Only the controller provided status is determined through this configuration. . Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ControllerHealthStatusEnabled *bool `json:"controller_health_status_enabled,omitempty"`

	// Creator name. Field introduced in 17.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CreatedBy *string `json:"created_by,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// Fully qualified domain name of the GSLB service. Minimum of 1 items required. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DomainNames []string `json:"domain_names,omitempty"`

	// Response to the client query when the GSLB service is DOWN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DownResponse *GslbServiceDownResponse `json:"down_response,omitempty"`

	// Enable or disable the GSLB service. If the GSLB service is enabled, then the VIPs are sent in the DNS responses based on reachability and configured algorithm. If the GSLB service is disabled, then the VIPs are no longer available in the DNS response. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Enabled *bool `json:"enabled,omitempty"`

	// Select list of pools belonging to this GSLB service. Minimum of 1 items required. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Groups []*GslbPool `json:"groups,omitempty"`

	// Verify VS health by applying one or more health monitors.  Active monitors generate synthetic traffic from DNS Service Engine and to mark a VS up or down based on the response. . It is a reference to an object of type HealthMonitor. Maximum of 6 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HealthMonitorRefs []string `json:"health_monitor_refs,omitempty"`

	// Health monitor probe can be executed for all the members or it can be executed only for third-party members. This operational mode is useful to reduce the number of health monitor probes in case of a hybrid scenario. In such a case, Avi members can have controller derived status while Non-Avi members can be probed by via health monitor probes in dataplane. Enum options - GSLB_SERVICE_HEALTH_MONITOR_ALL_MEMBERS, GSLB_SERVICE_HEALTH_MONITOR_ONLY_NON_AVI_MEMBERS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HealthMonitorScope *string `json:"health_monitor_scope,omitempty"`

	// This field is an internal field and is used in SE. Field introduced in 18.2.2. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	HmOff *bool `json:"hm_off,omitempty"`

	// This field indicates that this object is replicated across GSLB federation. Field introduced in 17.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IsFederated *bool `json:"is_federated,omitempty"`

	// List of labels to be used for granular RBAC. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	Markers []*RoleFilterMatchLabel `json:"markers,omitempty"`

	// The minimum number of members to distribute traffic to. Allowed values are 1-65535. Special values are 0 - Disable. Field introduced in 17.2.4. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MinMembers *uint32 `json:"min_members,omitempty"`

	// Name for the GSLB service. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// Number of IP addresses of this GSLB service to be returned by the DNS Service. Enter 0 to return all IP addresses. Allowed values are 1-20. Special values are 0- Return all IP addresses. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumDNSIP *uint32 `json:"num_dns_ip,omitempty"`

	// PKI profile associated with the Gslb Service. It is a reference to an object of type PKIProfile. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PkiProfileRef *string `json:"pki_profile_ref,omitempty"`

	// The load balancing algorithm will pick a GSLB pool within the GSLB service list of available pools. Enum options - GSLB_SERVICE_ALGORITHM_PRIORITY, GSLB_SERVICE_ALGORITHM_GEO. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PoolAlgorithm *string `json:"pool_algorithm,omitempty"`

	// This field indicates that for a CNAME query, respond with resolved CNAMEs in the additional section with A records. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ResolveCname *bool `json:"resolve_cname,omitempty"`

	// Enable site-persistence for the GslbService. . Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SitePersistenceEnabled *bool `json:"site_persistence_enabled,omitempty"`

	//  It is a reference to an object of type Tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// When enabled, topology policy rules are used for member selection first. If no valid member is found using the topology policy rules, configured GSLB algorithms for pool selection and member selection are used. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TopologyPolicyEnabled *bool `json:"topology_policy_enabled,omitempty"`

	// TTL value (in seconds) for records served for this GSLB service by the DNS Service. Allowed values are 0-86400. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TTL *uint32 `json:"ttl,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Use the client ip subnet from the EDNS option as source IPaddress for client geo-location and consistent hash algorithm. Default is true. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UseEdnsClientSubnet *bool `json:"use_edns_client_subnet,omitempty"`

	// UUID of the GSLB service. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`

	// Enable wild-card match of fqdn  if an exact match is not found in the DNS table, the longest match is chosen by wild-carding the fqdn in the DNS request. Default is false. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	WildcardMatch *bool `json:"wildcard_match,omitempty"`
}
