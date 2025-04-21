// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// IcapNSXDefenderLog icap n s x defender log
// swagger:model IcapNSXDefenderLog
type IcapNSXDefenderLog struct {

	// Score associated with the uploaded file, if known, value is in between 0 and 100. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Score uint32 `json:"score,omitempty"`

	// URL to get details from NSXDefender for the request. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	StatusURL *string `json:"status_url,omitempty"`

	// The NSX Defender task UUID associated with the analysis of the file. It is possible to use this UUID in order to access the analysis details from the NSX Defender Portal/Manager Web UI. URL to access this information is https //user.lastline.com/portal#/analyst/task/<uuid>/overview. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TaskUUID *string `json:"task_uuid,omitempty"`
}
