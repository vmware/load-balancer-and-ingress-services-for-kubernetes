// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// FeProxyRoutePublishConfig fe proxy route publish config
// swagger:model FeProxyRoutePublishConfig
type FeProxyRoutePublishConfig struct {

	// Publish ECMP route to upstream router for VIP. Enum options - FE_PROXY_ROUTE_PUBLISH_NONE, FE_PROXY_ROUTE_PUBLISH_QUAGGA_WEBAPP. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Mode *string `json:"mode,omitempty"`

	// Listener port for publisher. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PublisherPort *uint32 `json:"publisher_port,omitempty"`

	// Subnet for publisher. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Subnet *uint32 `json:"subnet,omitempty"`

	// Token for tracking changes. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Token *string `json:"token,omitempty"`
}
