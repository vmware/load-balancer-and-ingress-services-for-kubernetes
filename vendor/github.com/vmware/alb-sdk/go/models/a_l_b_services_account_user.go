// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ALBServicesAccountUser a l b services account user
// swagger:model ALBServicesAccountUser
type ALBServicesAccountUser struct {

	//  Field introduced in 20.1.1.
	// Required: true
	Email *string `json:"email"`

	//  Field introduced in 20.1.1.
	// Required: true
	Name *string `json:"name"`

	//  Field introduced in 20.1.1.
	// Required: true
	Phone *string `json:"phone"`
}
