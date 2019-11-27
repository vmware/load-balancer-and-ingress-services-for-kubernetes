package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// LdapUserBindSettings ldap user bind settings
// swagger:model LdapUserBindSettings
type LdapUserBindSettings struct {

	// LDAP user DN pattern is used to bind LDAP user after replacing the user token with real username.
	DnTemplate *string `json:"dn_template,omitempty"`

	// LDAP token is replaced with real user name in the user DN pattern.
	Token *string `json:"token,omitempty"`

	// LDAP user attributes to fetch on a successful user bind.
	UserAttributes []string `json:"user_attributes,omitempty"`

	// LDAP user id attribute is the login attribute that uniquely identifies a single user record.
	UserIDAttribute *string `json:"user_id_attribute,omitempty"`
}
