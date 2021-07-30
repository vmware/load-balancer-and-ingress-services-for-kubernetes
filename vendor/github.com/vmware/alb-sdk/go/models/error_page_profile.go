// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ErrorPageProfile error page profile
// swagger:model ErrorPageProfile
type ErrorPageProfile struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Name of the Virtual Service which generated the error page. Field deprecated in 18.1.1. Field introduced in 17.2.4.
	AppName *string `json:"app_name,omitempty"`

	// Name of the company to show in error page. Field deprecated in 18.1.1. Field introduced in 17.2.4.
	CompanyName *string `json:"company_name,omitempty"`

	// Defined Error Pages for HTTP status codes. Field introduced in 17.2.4.
	ErrorPages []*ErrorPage `json:"error_pages,omitempty"`

	// Fully qualified domain name for which the error page is generated. Field deprecated in 18.1.1. Field introduced in 17.2.4.
	HostName *string `json:"host_name,omitempty"`

	// Key value pairs for granular object access control. Also allows for classification and tagging of similar objects. Field deprecated in 20.1.5. Field introduced in 20.1.3. Maximum of 4 items allowed.
	Labels []*KeyValue `json:"labels,omitempty"`

	// List of labels to be used for granular RBAC. Field introduced in 20.1.5. Allowed in Basic edition, Essentials edition, Enterprise edition.
	Markers []*RoleFilterMatchLabel `json:"markers,omitempty"`

	//  Field introduced in 17.2.4.
	// Required: true
	Name *string `json:"name"`

	//  It is a reference to an object of type Tenant. Field introduced in 17.2.4.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Field introduced in 17.2.4.
	UUID *string `json:"uuid,omitempty"`
}
