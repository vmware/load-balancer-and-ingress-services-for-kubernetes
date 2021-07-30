// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ALBServicesAccount a l b services account
// swagger:model ALBServicesAccount
type ALBServicesAccount struct {

	// ID of an Account in the customer portal. Field introduced in 20.1.1.
	ID *string `json:"id,omitempty"`

	// Account to which the customer portal user belongs. Field introduced in 20.1.1.
	Name *string `json:"name,omitempty"`

	// Information about users within the account in the customer portal. Field introduced in 20.1.1.
	Users []*ALBServicesAccountUser `json:"users,omitempty"`
}
