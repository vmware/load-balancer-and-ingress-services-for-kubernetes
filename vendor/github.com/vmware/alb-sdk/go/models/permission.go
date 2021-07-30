// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// Permission permission
// swagger:model Permission
type Permission struct {

	//  Enum options - PERMISSION_CONTROLLER. PERMISSION_INTERNAL. PERMISSION_VIRTUALSERVICE. PERMISSION_POOL. PERMISSION_HEALTHMONITOR. PERMISSION_NETWORKPROFILE. PERMISSION_APPLICATIONPROFILE. PERMISSION_HTTPPOLICYSET. PERMISSION_IPADDRGROUP. PERMISSION_STRINGGROUP. PERMISSION_SSLPROFILE. PERMISSION_SSLKEYANDCERTIFICATE. PERMISSION_NETWORKSECURITYPOLICY. PERMISSION_APPLICATIONPERSISTENCEPROFILE. PERMISSION_ANALYTICSPROFILE. PERMISSION_VSDATASCRIPTSET. PERMISSION_TENANT. PERMISSION_PKIPROFILE. PERMISSION_AUTHPROFILE. PERMISSION_CLOUD...
	// Required: true
	Resource *string `json:"resource"`

	// Limits the scope of Write Access on the parent resource to modification of only the specified subresources. Field introduced in 20.1.5.
	Subresource *SubResource `json:"subresource,omitempty"`

	//  Enum options - NO_ACCESS, READ_ACCESS, WRITE_ACCESS.
	// Required: true
	Type *string `json:"type"`
}
