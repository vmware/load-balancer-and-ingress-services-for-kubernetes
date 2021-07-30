// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AttackMetaData attack meta data
// swagger:model AttackMetaData
type AttackMetaData struct {

	// ip of AttackMetaData.
	IP *string `json:"ip,omitempty"`

	// Number of max_resp_time.
	MaxRespTime *int32 `json:"max_resp_time,omitempty"`

	// url of AttackMetaData.
	URL *string `json:"url,omitempty"`
}
