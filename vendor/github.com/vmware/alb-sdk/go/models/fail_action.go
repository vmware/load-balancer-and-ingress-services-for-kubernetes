// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// FailAction fail action
// swagger:model FailAction
type FailAction struct {

	// Local response to HTTP requests when pool experiences a failure. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LocalRsp *FailActionHTTPLocalResponse `json:"local_rsp,omitempty"`

	// URL to redirect HTTP requests to when pool experiences a failure. Allowed in Enterprise edition with any value, Basic, Enterprise with Cloud Services edition.
	Redirect *FailActionHTTPRedirect `json:"redirect,omitempty"`

	// Enables a response to client when pool experiences a failure. By default TCP connection is closed. Enum options - FAIL_ACTION_HTTP_REDIRECT, FAIL_ACTION_HTTP_LOCAL_RSP, FAIL_ACTION_CLOSE_CONN, FAIL_ACTION_BACKUP_POOL. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- FAIL_ACTION_CLOSE_CONN), Basic edition(Allowed values- FAIL_ACTION_CLOSE_CONN,FAIL_ACTION_HTTP_REDIRECT), Enterprise with Cloud Services edition.
	// Required: true
	Type *string `json:"type"`
}
