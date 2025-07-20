// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbSiteRuntimeCfg gslb site runtime cfg
// swagger:model GslbSiteRuntimeCfg
type GslbSiteRuntimeCfg struct {

	// Gslb GeoDb files published for a site. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FdInfo *ConfigInfo `json:"fd_info,omitempty"`

	// Gslb Application Persistence info published for a site. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GapInfo *ConfigInfo `json:"gap_info,omitempty"`

	// Gslb GeoDb info published for a site. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GeoInfo *ConfigInfo `json:"geo_info,omitempty"`

	// GHM info published for a site. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GhmInfo *ConfigInfo `json:"ghm_info,omitempty"`

	// Gslb JWTProfile info published for a site. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	GjwtInfo *ConfigInfo `json:"gjwt_info,omitempty"`

	// Gslb info published for a site. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GlbInfo *ConfigInfo `json:"glb_info,omitempty"`

	// Gslb PKI info published for a site. Field introduced in 17.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GpkiInfo *ConfigInfo `json:"gpki_info,omitempty"`

	// GS info published for a site. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GsInfo *ConfigInfo `json:"gs_info,omitempty"`

	// Maintenance mode info published for a site. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MmInfo *ConfigInfo `json:"mm_info,omitempty"`

	// The replication queue for all object-types for a site. Field introduced in 17.2.7. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ReplQueue *ConfigInfo `json:"repl_queue,omitempty"`

	// Configuration sync-info of the site . Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SyncInfo *GslbSiteCfgSyncInfo `json:"sync_info,omitempty"`
}
