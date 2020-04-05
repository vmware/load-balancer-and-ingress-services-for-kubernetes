package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// LdapAuthSettings ldap auth settings
// swagger:model LdapAuthSettings
type LdapAuthSettings struct {

	// The LDAP base DN.  For example, avinetworks.com would be DC=avinetworks,DC=com.
	BaseDn *string `json:"base_dn,omitempty"`

	// LDAP administrator credentials are used to search for users and group memberships.
	BindAsAdministrator *bool `json:"bind_as_administrator,omitempty"`

	// LDAP attribute that refers to user email.
	EmailAttribute *string `json:"email_attribute,omitempty"`

	// LDAP attribute that refers to user's full name.
	FullNameAttribute *string `json:"full_name_attribute,omitempty"`

	// Query the LDAP servers on this port.
	Port *int32 `json:"port,omitempty"`

	// LDAP connection security mode. Enum options - AUTH_LDAP_SECURE_NONE, AUTH_LDAP_SECURE_USE_LDAPS.
	SecurityMode *string `json:"security_mode,omitempty"`

	// LDAP server IP address or Hostname. Use IP address if an auth profile is used to configure Virtual Service.
	Server []string `json:"server,omitempty"`

	// LDAP full directory configuration with administrator credentials.
	Settings *LdapDirectorySettings `json:"settings,omitempty"`

	// LDAP anonymous bind configuration.
	UserBind *LdapUserBindSettings `json:"user_bind,omitempty"`
}
