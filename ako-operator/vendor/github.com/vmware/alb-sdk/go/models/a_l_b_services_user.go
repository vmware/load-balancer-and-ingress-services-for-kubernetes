// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ALBServicesUser a l b services user
// swagger:model ALBServicesUser
type ALBServicesUser struct {

	// ID of primary account of the portal user. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AccountID *string `json:"account_id,omitempty"`

	// Name of primary account of the portal user. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AccountName *string `json:"account_name,omitempty"`

	// Email ID of the portal user. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Email *string `json:"email"`

	// Information about all the accounts managed by user in the customer portal. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ManagedAccounts []*ALBServicesAccount `json:"managed_accounts,omitempty"`

	// Name of the portal user. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`

	// Phone number of the user. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Phone *string `json:"phone,omitempty"`
}
