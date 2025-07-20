// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AuthProfileHTTPClientParams auth profile HTTP client params
// swagger:model AuthProfileHTTPClientParams
type AuthProfileHTTPClientParams struct {

	// The max allowed length of time a clients authentication is cached. Allowed values are 1-30. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CacheExpirationTime *int32 `json:"cache_expiration_time,omitempty"`

	// Insert an HTTP header.  This field is used to define the header name.  The value of the header is set to the client's HTTP Auth user ID. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RequestHeader *string `json:"request_header,omitempty"`

	// A user should be a member of these groups.  Each group is defined by the DN.  For example, CN=testgroup,OU=groups,dc=example,dc=avinetworks,DC=com. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RequireUserGroups []string `json:"require_user_groups,omitempty"`
}
