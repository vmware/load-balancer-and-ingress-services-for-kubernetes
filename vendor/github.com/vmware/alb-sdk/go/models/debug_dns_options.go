// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DebugDNSOptions debug Dns options
// swagger:model DebugDnsOptions
type DebugDNSOptions struct {

	// This field filters the FQDN for Dns debug. Field introduced in 18.2.1. Maximum of 1 items allowed.
	DomainName []string `json:"domain_name,omitempty"`

	// This field filters the Gslb service for Dns debug. Field introduced in 18.2.1. Maximum of 1 items allowed.
	GslbServiceName []string `json:"gslb_service_name,omitempty"`
}
