// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// L4SSlapplicationProfile l4 s slapplication profile
// swagger:model L4SSLApplicationProfile
type L4SSlapplicationProfile struct {

	// L4 stream idle connection timeout in seconds. Allowed values are 60-86400. Field introduced in 22.1.2. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SslStreamIDLETimeout *uint32 `json:"ssl_stream_idle_timeout,omitempty"`
}
