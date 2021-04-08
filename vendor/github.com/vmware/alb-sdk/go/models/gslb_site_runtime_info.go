package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbSiteRuntimeInfo gslb site runtime info
// swagger:model GslbSiteRuntimeInfo
type GslbSiteRuntimeInfo struct {

	// The Leader-IP/VIP/FQDN of the site-cluster.
	ClusterLeader *string `json:"cluster_leader,omitempty"`

	// Unique object identifier of cluster.
	ClusterUUID *string `json:"cluster_uuid,omitempty"`

	// operational dns state at the site.
	DNSInfo *GslbDNSInfo `json:"dns_info,omitempty"`

	// Enable/disable state retrieved from the cfg .
	Enabled *bool `json:"enabled,omitempty"`

	// event-cache used for event throttling.
	EventCache *EventCache `json:"event_cache,omitempty"`

	// Health-status monitoring enable or disable.
	HsState *bool `json:"hs_state,omitempty"`

	// Placeholder for description of property last_changed_time of obj type GslbSiteRuntimeInfo field type str  type object
	LastChangedTime *TimeStamp `json:"last_changed_time,omitempty"`

	// Name of the object.
	Name *string `json:"name,omitempty"`

	// Number of retry attempts to reach the remote site.
	NumOfRetries *int32 `json:"num_of_retries,omitempty"`

	// Placeholder for description of property oper_status of obj type GslbSiteRuntimeInfo field type str  type object
	OperStatus *OperationalStatus `json:"oper_status,omitempty"`

	// Site Role  Leader or Follower. Enum options - GSLB_LEADER, GSLB_MEMBER, GSLB_NOT_A_MEMBER.
	Role *string `json:"role,omitempty"`

	// Current outstanding request-response token of the message to this site.
	Rrtoken []string `json:"rrtoken,omitempty"`

	// Indicates if it is Avi Site or third-party. Enum options - GSLB_AVI_SITE, GSLB_THIRD_PARTY_SITE. Field introduced in 17.1.1.
	SiteType *string `json:"site_type,omitempty"`

	//  Enum options - SITE_STATE_NULL, SITE_STATE_JOIN_IN_PROGRESS, SITE_STATE_LEAVE_IN_PROGRESS, SITE_STATE_INIT, SITE_STATE_UNREACHABLE, SITE_STATE_MMODE, SITE_STATE_DISABLE_IN_PROGRESS, SITE_STATE_DISABLED.
	State *string `json:"state,omitempty"`

	// State - Reason.
	StateReason *string `json:"state_reason,omitempty"`

	// Current Software version of the site.
	SwVersion *string `json:"sw_version,omitempty"`
}
