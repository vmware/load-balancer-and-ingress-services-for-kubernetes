package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbHealthMonitorProxy gslb health monitor proxy
// swagger:model GslbHealthMonitorProxy
type GslbHealthMonitorProxy struct {

	// This field identifies the health monitor proxy behavior. The designated site for health monitor proxy can monitor public or private or all the members of a given site. . Enum options - GSLB_HEALTH_MONITOR_PROXY_ALL_MEMBERS, GSLB_HEALTH_MONITOR_PROXY_PRIVATE_MEMBERS. Field introduced in 17.1.1.
	ProxyType *string `json:"proxy_type,omitempty"`

	// This field identifies the site that will health monitor on behalf of the current site. i.e. it will be a health monitor proxy and monitor members of the current site. . Field introduced in 17.1.1.
	SiteUUID *string `json:"site_uuid,omitempty"`
}
