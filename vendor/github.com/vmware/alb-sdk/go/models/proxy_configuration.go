// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ProxyConfiguration proxy configuration
// swagger:model ProxyConfiguration
type ProxyConfiguration struct {

	// Proxy hostname or IP address.
	// Required: true
	Host *string `json:"host"`

	// Password for proxy.
	Password *string `json:"password,omitempty"`

	// Proxy port.
	// Required: true
	Port *int32 `json:"port"`

	// Username for proxy.
	Username *string `json:"username,omitempty"`
}
