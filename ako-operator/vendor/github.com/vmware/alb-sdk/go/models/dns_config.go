// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSConfig DNS config
// swagger:model DNSConfig
type DNSConfig struct {

	// GSLB subdomain used for GSLB service FQDN match and placement. . Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	DomainName *string `json:"domain_name"`
}
