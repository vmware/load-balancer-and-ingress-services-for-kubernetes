// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSConfig DNS config
// swagger:model DNSConfig
type DNSConfig struct {

	// GSLB subdomain used for GSLB service FQDN match and placement. .
	// Required: true
	DomainName *string `json:"domain_name"`
}
