// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AutoScaleOpenStackSettings auto scale open stack settings
// swagger:model AutoScaleOpenStackSettings
type AutoScaleOpenStackSettings struct {

	// Avi Controller will use this URL to scale downthe pool. Cloud connector will automatically update the membership. This is an alpha feature. Field introduced in 17.1.1.
	HeatScaleDownURL *string `json:"heat_scale_down_url,omitempty"`

	// Avi Controller will use this URL to scale upthe pool. Cloud connector will automatically update the membership. This is an alpha feature. Field introduced in 17.1.1.
	HeatScaleUpURL *string `json:"heat_scale_up_url,omitempty"`
}
