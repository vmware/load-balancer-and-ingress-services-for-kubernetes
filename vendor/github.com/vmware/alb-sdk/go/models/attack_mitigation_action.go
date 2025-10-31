// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AttackMitigationAction attack mitigation action
// swagger:model AttackMitigationAction
type AttackMitigationAction struct {

	// Deny the attack packets further processing and drop them. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Deny *bool `json:"deny,omitempty"`
}
