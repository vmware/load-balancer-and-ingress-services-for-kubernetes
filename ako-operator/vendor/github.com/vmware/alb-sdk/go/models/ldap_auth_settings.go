// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// LdapAuthSettings ldap auth settings
// swagger:model LdapAuthSettings
type LdapAuthSettings struct {

	// The LDAP base DN.  For example, avinetworks.com would be DC=avinetworks,DC=com. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	BaseDn *string `json:"base_dn,omitempty"`

	// LDAP administrator credentials are used to search for users and group memberships. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	BindAsAdministrator *bool `json:"bind_as_administrator,omitempty"`

	// LDAP attribute that refers to user email. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EmailAttribute *string `json:"email_attribute,omitempty"`

	// LDAP attribute that refers to user's full name. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FullNameAttribute *string `json:"full_name_attribute,omitempty"`

	// Query the LDAP servers on this port. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Port *uint32 `json:"port,omitempty"`

	// LDAP connection security mode. Enum options - AUTH_LDAP_SECURE_NONE, AUTH_LDAP_SECURE_USE_LDAPS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	SecurityMode *string `json:"security_mode"`

	// LDAP server IP(v4/v6) address or FQDN. Use IP address if an auth profile is used to configure Virtual Service. Minimum of 1 items required. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Server []string `json:"server,omitempty"`

	// LDAP full directory configuration with administrator credentials. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Settings *LdapDirectorySettings `json:"settings,omitempty"`

	// LDAP anonymous bind configuration. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UserBind *LdapUserBindSettings `json:"user_bind,omitempty"`
}
