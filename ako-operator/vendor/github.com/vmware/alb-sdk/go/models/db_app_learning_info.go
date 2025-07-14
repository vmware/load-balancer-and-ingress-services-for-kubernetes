// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DbAppLearningInfo db app learning info
// swagger:model DbAppLearningInfo
type DbAppLearningInfo struct {

	// Application UUID. Combination of Virtualservice UUID and WAF Policy UUID. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AppID *string `json:"app_id,omitempty"`

	// Information about various URIs under a application. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	URIInfo []*URIInfo `json:"uri_info,omitempty"`

	// Virtualserivce UUID. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsUUID *string `json:"vs_uuid,omitempty"`
}
