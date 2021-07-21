// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SystemLimits system limits
// swagger:model SystemLimits
type SystemLimits struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// System limits for the entire controller cluster. Field introduced in 20.1.1.
	ControllerLimits *ControllerLimits `json:"controller_limits,omitempty"`

	// Possible controller sizes. Field introduced in 20.1.1.
	ControllerSizes []*ControllerSize `json:"controller_sizes,omitempty"`

	// System limits that apply to a serviceengine. Field introduced in 20.1.1.
	ServiceengineLimits *ServiceEngineLimits `json:"serviceengine_limits,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID for the system limits object. Field introduced in 20.1.1.
	UUID *string `json:"uuid,omitempty"`
}
