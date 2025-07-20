// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbDownloadStatus gslb download status
// swagger:model GslbDownloadStatus
type GslbDownloadStatus struct {

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LastChangedTime *TimeStamp `json:"last_changed_time,omitempty"`

	// This field indicates the download state to a dns-vs(es) or a VS or a SE depending on the usage context. . Enum options - GSLB_DOWNLOAD_NONE, GSLB_DOWNLOAD_DONE, GSLB_DOWNLOAD_PENDING, GSLB_DOWNLOAD_ERROR. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	State *string `json:"state,omitempty"`
}
