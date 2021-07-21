// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AutoScaleMesosSettings auto scale mesos settings
// swagger:model AutoScaleMesosSettings
type AutoScaleMesosSettings struct {

	// Apply scaleout even when there are deployments inprogress.
	Force *bool `json:"force,omitempty"`
}
