// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ControllerPortalAuth controller portal auth
// swagger:model ControllerPortalAuth
type ControllerPortalAuth struct {

	// Access Token to authenticate Customer Portal REST calls. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AccessToken *string `json:"access_token,omitempty"`

	// Salesforce instance URL. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	InstanceURL *string `json:"instance_url,omitempty"`

	// Signed JWT to refresh the access token. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	JwtToken *string `json:"jwt_token,omitempty"`
}
