// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ParamsInURI params in URI
// swagger:model ParamsInURI
type ParamsInURI struct {

	// Params info in hitted signature rule which has ARGS match element. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ParamInfo []*ParamInURI `json:"param_info,omitempty"`
}
