// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SSLProfileSelector s s l profile selector
// swagger:model SSLProfileSelector
type SSLProfileSelector struct {

	// Configure client IP address groups. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ClientIPList *IPAddrMatch `json:"client_ip_list"`

	// SSL profile for the client IP addresses listed. It is a reference to an object of type SSLProfile. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	SslProfileRef *string `json:"ssl_profile_ref"`
}
