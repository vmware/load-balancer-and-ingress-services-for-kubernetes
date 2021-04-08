package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// OpenStackVipNetwork open stack vip network
// swagger:model OpenStackVipNetwork
type OpenStackVipNetwork struct {

	// Neutron network UUID. Field introduced in 18.1.2.
	OsNetworkUUID *string `json:"os_network_uuid,omitempty"`

	// UUIDs of OpenStack tenants that should be allowed to use the specified Neutron network for VIPs. Use '*' to make this network available to all tenants. Field introduced in 18.1.2.
	OsTenantUuids []string `json:"os_tenant_uuids,omitempty"`
}
