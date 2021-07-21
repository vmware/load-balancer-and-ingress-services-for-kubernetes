// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeVersionCheckFailedEvent se version check failed event
// swagger:model SeVersionCheckFailedEvent
type SeVersionCheckFailedEvent struct {

	// Software version on the controller.
	ControllerVersion *string `json:"controller_version,omitempty"`

	// UUID of the SE.
	SeUUID *string `json:"se_uuid,omitempty"`

	// Software version on the SE.
	SeVersion *string `json:"se_version,omitempty"`
}
