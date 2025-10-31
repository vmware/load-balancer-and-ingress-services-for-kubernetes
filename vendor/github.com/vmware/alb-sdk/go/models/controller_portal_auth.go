// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ControllerPortalAuth controller portal auth
// swagger:model ControllerPortalAuth
type ControllerPortalAuth struct {

	// Access Token to authenticate Customer Portal REST calls. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AccessToken *string `json:"access_token,omitempty"`

	// Grant type of the JWT token. Enum options - REFRESH_TOKEN, CLIENT_CREDENTIALS. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	GrantType *string `json:"grant_type,omitempty"`

	// Cloud services instance URL. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	InstanceURL *string `json:"instance_url,omitempty"`

	// Signed JWT to refresh the access token. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	JwtToken *string `json:"jwt_token,omitempty"`

	// Tenant information for which cloud services authentication information is persisted. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Tenant *string `json:"tenant,omitempty"`
}
