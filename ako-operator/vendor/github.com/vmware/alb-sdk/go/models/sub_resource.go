// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SubResource sub resource
// swagger:model SubResource
type SubResource struct {

	// Allows modification of all fields except for the specified subresources. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ExcludeSubresources *bool `json:"exclude_subresources,omitempty"`

	// Subresources user can modify. Each subresource specifies and individual field. I.e. SUBRESOURCE_POOL_ENABLED allows modification of the enabled field in the Pool object. Enum options - SUBRESOURCE_POOL_ENABLED, SUBRESOURCE_POOL_SERVERS, SUBRESOURCE_POOL_SERVER_ENABLED, SUBRESOURCE_VIRTUALSERVICE_ENABLED, SUBRESOURCE_VIRTUALSERVICE_AUTO_ALLOCATE_FLOATING_IP, SUBRESOURCE_GSLBSERVICE_ENABLED, SUBRESOURCE_GSLBSERVICE_GROUPS, SUBRESOURCE_GSLBSERVICE_GROUP_ENABLED, SUBRESOURCE_GSLBSERVICE_GROUP_MEMBERS, SUBRESOURCE_GSLBSERVICE_GROUP_MEMBER_ENABLED. Field introduced in 20.1.5. Minimum of 1 items required. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Subresources []string `json:"subresources,omitempty"`
}
