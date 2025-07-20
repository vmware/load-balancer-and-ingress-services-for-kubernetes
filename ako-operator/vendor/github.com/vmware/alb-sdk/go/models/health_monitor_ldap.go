// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HealthMonitorLdap health monitor ldap
// swagger:model HealthMonitorLdap
type HealthMonitorLdap struct {

	// Attributes which will be retrieved. commas can be used to delimit more than one attributes (example- cn,address,email). Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Attributes *string `json:"attributes,omitempty"`

	// DN(Distinguished Name) of a directory entry. which will be starting point of the search. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	BaseDn *string `json:"base_dn"`

	// Filter to search entries in specified scope. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Filter *string `json:"filter,omitempty"`

	// Search scope which can be base, one, sub. Enum options - LDAP_BASE_MODE, LDAP_ONE_MODE, LDAP_SUB_MODE. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Scope *string `json:"scope,omitempty"`

	// SSL attributes for LDAPS monitor. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SslAttributes *HealthMonitorSSlattributes `json:"ssl_attributes,omitempty"`
}
