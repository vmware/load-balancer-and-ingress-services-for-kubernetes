// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSSrvRdata Dns srv rdata
// swagger:model DnsSrvRdata
type DNSSrvRdata struct {

	// Service port. Allowed values are 0-65535. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Port *uint32 `json:"port"`

	// Priority of the target hosting the service, low value implies higher priority for this service record. Allowed values are 0-65535. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Priority *uint32 `json:"priority,omitempty"`

	// Canonical hostname, of the machine hosting the service, with no trailing period. 'default.host' is valid but not 'default.host.'. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Target *string `json:"target,omitempty"`

	// Relative weight for service records with same priority, high value implies higher preference for this service record. Allowed values are 0-65535. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Weight *uint32 `json:"weight,omitempty"`
}
