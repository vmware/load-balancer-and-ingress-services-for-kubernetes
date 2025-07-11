// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// Postsnapshot postsnapshot
// swagger:model postsnapshot
type Postsnapshot struct {

	// FB Gs snapshot data. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Gssnapshot *FbGsInfo `json:"gssnapshot,omitempty"`

	// FB Pool snapshot data. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Poolsnapshot *FbPoolInfo `json:"poolsnapshot,omitempty"`

	// FB SE snapshot data. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Sesnapshot *FbSeInfo `json:"sesnapshot,omitempty"`

	// FB VS snapshot data. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Vssnapshot *FbVsInfo `json:"vssnapshot,omitempty"`
}
