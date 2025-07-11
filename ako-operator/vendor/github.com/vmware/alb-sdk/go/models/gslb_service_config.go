// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbServiceConfig gslb service config
// swagger:model GslbServiceConfig
type GslbServiceConfig struct {

	// GS member's overall health status is derived based on a combination of controller and datapath health-status inputs. Note that the datapath status is determined by the association of health monitor profiles. Only the controller provided status is determined through this configuration. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ControllerHealthStatusEnabled *bool `json:"controller_health_status_enabled,omitempty"`

	// Fully qualified domain name of the GSLB service. Field introduced in 22.1.1. Minimum of 1 items required. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DomainNames []string `json:"domain_names,omitempty"`

	// Response to the client query when the GSLB service is DOWN. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DownResponse *GslbServiceDownResponse `json:"down_response,omitempty"`

	// Enable or disable the GSLB service. If the GSLB service is enabled, then the VIPs are sent in the DNS responses based on reachability and configured algorithm. If the GSLB service is disabled, then the VIPs are no longer available in the DNS response. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Enabled *bool `json:"enabled,omitempty"`

	// Select list of pools belonging to this GSLB service. Field introduced in 22.1.1. Minimum of 1 items required. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Groups []*GslbPool `json:"groups,omitempty"`

	// Health monitor probe can be executed for all the members or it can be executed only for third-party members. This operational mode is useful to reduce the number of health monitor probes in case of a hybrid scenario. In such a case, Avi members can have controller derived status while Non-Avi members can be probed by via health monitor probes in dataplane. Enum options - GSLB_SERVICE_HEALTH_MONITOR_ALL_MEMBERS, GSLB_SERVICE_HEALTH_MONITOR_ONLY_NON_AVI_MEMBERS. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	HealthMonitorScope *string `json:"health_monitor_scope,omitempty"`

	// This field indicates that this object is replicated across GSLB federation. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IsFederated *bool `json:"is_federated,omitempty"`

	// The minimum number of members to distribute traffic to. Allowed values are 1-65535. Special values are 0 - Disable. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MinMembers *uint32 `json:"min_members,omitempty"`

	// Name of the GSLB Service. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// The load balancing algorithm will pick a GSLB pool within the GSLB service list of available pools. Enum options - GSLB_SERVICE_ALGORITHM_PRIORITY, GSLB_SERVICE_ALGORITHM_GEO. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PoolAlgorithm *string `json:"pool_algorithm,omitempty"`

	// This field indicates that for a CNAME query, respond with resolved CNAMEs in the additional section with A records. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ResolveCname *bool `json:"resolve_cname,omitempty"`

	// Enable site-persistence for the GslbService. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SitePersistenceEnabled *bool `json:"site_persistence_enabled,omitempty"`

	//  It is a reference to an object of type Tenant. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	TenantRef *string `json:"tenant_ref"`

	// URL of the GSLB Service. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	URL *string `json:"url,omitempty"`

	// Use the client ip subnet from the EDNS option as source IPaddress for client geo-location and consistent hash algorithm. Default is true. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UseEdnsClientSubnet *bool `json:"use_edns_client_subnet,omitempty"`

	// UUID of the GSLB Service. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`

	// Enable wild-card match of fqdn  if an exact match is not found in the DNS table, the longest match is chosen by wild-carding the fqdn in the DNS request. Default is false. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	WildcardMatch *bool `json:"wildcard_match,omitempty"`
}
