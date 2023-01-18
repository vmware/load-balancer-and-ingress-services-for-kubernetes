// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbHealthMonitorProxy gslb health monitor proxy
// swagger:model GslbHealthMonitorProxy
type GslbHealthMonitorProxy struct {

	// This field identifies the health monitor proxy behavior. The designated site for health monitor proxy can monitor public or private or all the members of a given site. . Enum options - GSLB_HEALTH_MONITOR_PROXY_ALL_MEMBERS, GSLB_HEALTH_MONITOR_PROXY_PRIVATE_MEMBERS. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ProxyType *string `json:"proxy_type,omitempty"`

	// This field identifies the site that will health monitor on behalf of the current site. i.e. it will be a health monitor proxy and monitor members of the current site. . Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SiteUUID *string `json:"site_uuid,omitempty"`
}
