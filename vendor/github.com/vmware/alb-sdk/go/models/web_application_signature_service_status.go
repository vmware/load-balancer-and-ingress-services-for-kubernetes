// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// WebApplicationSignatureServiceStatus web application signature service status
// swagger:model WebApplicationSignatureServiceStatus
type WebApplicationSignatureServiceStatus struct {

	// If the last attempted update failed, this is a more detailed error message. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Error *string `json:"error,omitempty"`

	// The time when the Application Signature service last successfull attemped to update this object. It will be not update, if an error occurs during an update attempt. In this case, the error will be set. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LastSuccessfulUpdateCheck *TimeStamp `json:"last_successful_update_check,omitempty"`

	// A timestamp field. It is used by the Application Signature Sync service to keep track of the current version. Field introduced in 21.1.1, 20.1.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UpstreamSyncTimestamp *TimeStamp `json:"upstream_sync_timestamp,omitempty"`
}
