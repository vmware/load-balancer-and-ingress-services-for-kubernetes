// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HttpsecurityPolicy httpsecurity policy
// swagger:model HTTPSecurityPolicy
type HttpsecurityPolicy struct {

	// Add rules to the HTTP security policy. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Rules []*HttpsecurityRule `json:"rules,omitempty"`
}
