// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ACUserIdentity a c user identity
// swagger:model ACUserIdentity
type ACUserIdentity struct {

	// User identity type for audit event (e.g. username, organization, component). Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Type *string `json:"type"`

	// User identity value for audit event (e.g. SomeCompany, Jane Doe, Secure-shell). Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Value *string `json:"value"`
}
