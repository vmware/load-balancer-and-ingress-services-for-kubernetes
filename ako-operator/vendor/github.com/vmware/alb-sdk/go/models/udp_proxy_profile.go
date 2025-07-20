// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// UDPProxyProfile UDP proxy profile
// swagger:model UDPProxyProfile
type UDPProxyProfile struct {

	// The amount of time (in sec) for which a flow needs to be idle before it is deleted. Allowed values are 2-3600. Field introduced in 17.2.8, 18.1.3, 18.2.1. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SessionIDLETimeout *int32 `json:"session_idle_timeout,omitempty"`
}
