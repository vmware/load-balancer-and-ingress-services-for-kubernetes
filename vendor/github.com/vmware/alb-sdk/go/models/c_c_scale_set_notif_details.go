// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CCScaleSetNotifDetails c c scale set notif details
// swagger:model CCScaleSetNotifDetails
type CCScaleSetNotifDetails struct {

	// Cloud id. Field introduced in 18.2.5.
	CcID *string `json:"cc_id,omitempty"`

	// Detailed reason for the scale set notification. Field introduced in 18.2.5.
	Reason *string `json:"reason,omitempty"`

	// Names of scale sets for which polling failed. Field introduced in 18.2.5.
	ScalesetNames []string `json:"scaleset_names,omitempty"`
}
