// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HTTPHdrData HTTP hdr data
// swagger:model HTTPHdrData
type HTTPHdrData struct {

	// HTTP header name. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`

	// HTTP header value. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Value *HTTPHdrValue `json:"value,omitempty"`
}
