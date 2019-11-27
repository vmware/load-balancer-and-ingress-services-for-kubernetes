package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// TenantConfiguration tenant configuration
// swagger:model TenantConfiguration
type TenantConfiguration struct {

	// Controls the ownership of ServiceEngines. Service Engines can either be exclusively owned by each tenant or owned by the administrator and shared by all tenants. When ServiceEngines are owned by the administrator, each tenant can have either read access or no access to their Service Engines.
	SeInProviderContext *bool `json:"se_in_provider_context,omitempty"`

	// Placeholder for description of property tenant_access_to_provider_se of obj type TenantConfiguration field type str  type boolean
	TenantAccessToProviderSe *bool `json:"tenant_access_to_provider_se,omitempty"`

	// When 'Per Tenant IP Domain' is selected, each tenant gets its own routing domain that is not shared with any other tenant. When 'Share IP Domain across all tenants' is selected, all tenants share the same routing domain.
	TenantVrf *bool `json:"tenant_vrf,omitempty"`
}
