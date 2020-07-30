package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ALBServicesConfig a l b services config
// swagger:model ALBServicesConfig
type ALBServicesConfig struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Information about the default contact for this controller cluster. Field introduced in 20.1.1.
	AssetContact *ALBServicesUser `json:"asset_contact,omitempty"`

	// Information about the portal features opted in for controller. Field introduced in 20.1.1.
	// Required: true
	FeatureOptInStatus *PortalFeatureOptIn `json:"feature_opt_in_status"`

	// Default values to be used for IP Reputation sync. Field introduced in 20.1.1.
	// Required: true
	IPReputationConfig *IPReputationConfig `json:"ip_reputation_config"`

	// Time interval in minutes. Allowed values are 5-60. Field introduced in 18.2.6.
	PollingInterval *int32 `json:"polling_interval,omitempty"`

	// The FQDN or IP address of the customer portal. Field introduced in 18.2.6.
	// Required: true
	PortalURL *string `json:"portal_url"`

	// Default values to be used during proactive case creation and techsupport attachment. Field introduced in 20.1.1.
	// Required: true
	ProactiveSupportDefaults *ProactiveSupportDefaults `json:"proactive_support_defaults"`

	// Split proxy configuration to connect external pulse services. Field introduced in 20.1.1.
	// Required: true
	SplitProxyConfiguration *ProxyConfiguration `json:"split_proxy_configuration"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// By default, use system proxy configuration.If true, use split proxy configuration. Field introduced in 20.1.1.
	UseSplitProxy *bool `json:"use_split_proxy,omitempty"`

	//  Field introduced in 18.2.6.
	UUID *string `json:"uuid,omitempty"`
}
