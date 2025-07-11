// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbSiteRuntimeInfo gslb site runtime info
// swagger:model GslbSiteRuntimeInfo
type GslbSiteRuntimeInfo struct {

	// The Leader-IP/VIP/FQDN of the site-cluster. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClusterLeader *string `json:"cluster_leader,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClusterUUID *string `json:"cluster_uuid,omitempty"`

	// operational dns state at the site. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DNSInfo *GslbDNSInfo `json:"dns_info,omitempty"`

	// Enable/disable state retrieved from the cfg . Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Enabled *bool `json:"enabled,omitempty"`

	// event-cache used for event throttling. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EventCache *EventCache `json:"event_cache,omitempty"`

	// Health-status monitoring enable or disable. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HsState *bool `json:"hs_state,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LastChangedTime *TimeStamp `json:"last_changed_time,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`

	// Number of retry attempts to reach the remote site. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumOfRetries *int32 `json:"num_of_retries,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OperStatus *OperationalStatus `json:"oper_status,omitempty"`

	// Site Role  Leader or Follower. Enum options - GSLB_LEADER, GSLB_MEMBER, GSLB_NOT_A_MEMBER. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Role *string `json:"role,omitempty"`

	// Current outstanding request-response token of the message to this site. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Rrtoken []string `json:"rrtoken,omitempty"`

	// Indicates if it is Avi Site or third-party. Enum options - GSLB_AVI_SITE, GSLB_THIRD_PARTY_SITE. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SiteType *string `json:"site_type,omitempty"`

	//  Enum options - SITE_STATE_NULL, SITE_STATE_JOIN_IN_PROGRESS, SITE_STATE_LEAVE_IN_PROGRESS, SITE_STATE_INIT, SITE_STATE_UNREACHABLE, SITE_STATE_MMODE, SITE_STATE_DISABLE_IN_PROGRESS, SITE_STATE_DISABLED. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	State *string `json:"state,omitempty"`

	// State - Reason. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	StateReason *string `json:"state_reason,omitempty"`

	// Current Software version of the site. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SwVersion *string `json:"sw_version,omitempty"`
}
