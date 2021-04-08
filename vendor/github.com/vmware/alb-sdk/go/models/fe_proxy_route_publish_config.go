package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// FeProxyRoutePublishConfig fe proxy route publish config
// swagger:model FeProxyRoutePublishConfig
type FeProxyRoutePublishConfig struct {

	// Publish ECMP route to upstream router for VIP. Enum options - FE_PROXY_ROUTE_PUBLISH_NONE, FE_PROXY_ROUTE_PUBLISH_QUAGGA_WEBAPP.
	Mode *string `json:"mode,omitempty"`

	// Listener port for publisher.
	PublisherPort *int32 `json:"publisher_port,omitempty"`

	// Subnet for publisher.
	Subnet *int32 `json:"subnet,omitempty"`

	// Token for tracking changes.
	Token *string `json:"token,omitempty"`
}
