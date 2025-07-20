// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ServiceOAuth service o auth
// swagger:model ServiceOAuth
type ServiceOAuth struct {

	// URL of authorization server. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	AuthorizationEndpoint *string `json:"authorization_endpoint"`

	// Application specific identifier for service auth. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	ClientID *string `json:"client_id"`

	// Organization Id for service OAuth(required for CSP). Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	OrgID *string `json:"org_id,omitempty"`

	// Uuid value of the service(required for CSP). Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ServiceID *string `json:"service_id,omitempty"`

	// Name of the service(required for CSP). Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ServiceName *string `json:"service_name,omitempty"`
}
