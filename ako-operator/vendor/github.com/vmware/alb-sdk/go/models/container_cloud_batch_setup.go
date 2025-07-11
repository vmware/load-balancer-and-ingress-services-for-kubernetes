// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ContainerCloudBatchSetup container cloud batch setup
// swagger:model ContainerCloudBatchSetup
type ContainerCloudBatchSetup struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Ccs []*ContainerCloudSetup `json:"ccs,omitempty"`
}
