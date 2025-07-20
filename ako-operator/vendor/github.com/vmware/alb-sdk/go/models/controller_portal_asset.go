// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ControllerPortalAsset controller portal asset
// swagger:model ControllerPortalAsset
type ControllerPortalAsset struct {

	// Asset ID corresponding to this Controller Cluster, returned on a successful registration. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AssetID *string `json:"asset_id,omitempty"`
}
