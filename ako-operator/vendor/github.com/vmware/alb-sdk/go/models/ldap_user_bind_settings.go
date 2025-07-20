// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// LdapUserBindSettings ldap user bind settings
// swagger:model LdapUserBindSettings
type LdapUserBindSettings struct {

	// LDAP user DN pattern is used to bind LDAP user after replacing the user token with real username. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	DnTemplate *string `json:"dn_template"`

	// LDAP token is replaced with real user name in the user DN pattern. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Token *string `json:"token"`

	// LDAP user attributes to fetch on a successful user bind. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UserAttributes []string `json:"user_attributes,omitempty"`

	// LDAP user id attribute is the login attribute that uniquely identifies a single user record. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	UserIDAttribute *string `json:"user_id_attribute"`
}
