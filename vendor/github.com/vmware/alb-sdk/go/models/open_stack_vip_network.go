// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// OpenStackVipNetwork open stack vip network
// swagger:model OpenStackVipNetwork
type OpenStackVipNetwork struct {

	// Neutron network UUID. Field introduced in 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OsNetworkUUID *string `json:"os_network_uuid,omitempty"`

	// UUIDs of OpenStack tenants that should be allowed to use the specified Neutron network for VIPs. Use '*' to make this network available to all tenants. Field introduced in 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OsTenantUuids []string `json:"os_tenant_uuids,omitempty"`
}
