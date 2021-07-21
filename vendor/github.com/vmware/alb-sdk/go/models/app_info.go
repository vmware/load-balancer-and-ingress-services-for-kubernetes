// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AppInfo app info
// swagger:model AppInfo
type AppInfo struct {

	// app_hdr_name of AppInfo.
	// Required: true
	AppHdrName *string `json:"app_hdr_name"`

	// app_hdr_value of AppInfo.
	// Required: true
	AppHdrValue *string `json:"app_hdr_value"`
}
