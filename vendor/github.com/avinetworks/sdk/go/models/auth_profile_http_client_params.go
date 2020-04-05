package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AuthProfileHTTPClientParams auth profile HTTP client params
// swagger:model AuthProfileHTTPClientParams
type AuthProfileHTTPClientParams struct {

	// The max allowed length of time a clients authentication is cached. Allowed values are 1-30.
	CacheExpirationTime *int32 `json:"cache_expiration_time,omitempty"`

	// Group member entries contain full DNs instead of just user id attribute values. This should now be configured using the LdapDirectorySettings field instead. Field deprecated in 18.2.1.
	GroupMemberIsFullDn *bool `json:"group_member_is_full_dn,omitempty"`

	// Insert an HTTP header.  This field is used to define the header name.  The value of the header is set to the client's HTTP Auth user ID.
	RequestHeader *string `json:"request_header,omitempty"`

	// A user should be a member of these groups.  Each group is defined by the DN.  For example, CN=testgroup,OU=groups,dc=example,dc=avinetworks,DC=com.
	RequireUserGroups []string `json:"require_user_groups,omitempty"`
}
