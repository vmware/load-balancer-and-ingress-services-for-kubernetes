// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AutoScaleMesosSettings auto scale mesos settings
// swagger:model AutoScaleMesosSettings
type AutoScaleMesosSettings struct {

	// Apply scale-out even when there are deployments inprogress. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Force *bool `json:"force,omitempty"`
}
