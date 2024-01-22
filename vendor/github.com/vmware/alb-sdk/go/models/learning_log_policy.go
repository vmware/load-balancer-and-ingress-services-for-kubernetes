// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// LearningLogPolicy learning log policy
// swagger:model LearningLogPolicy
type LearningLogPolicy struct {

	// Determine whether app learning logging is enabled. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Enabled *bool `json:"enabled,omitempty"`

	// Host name where learning logs will be sent to. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Host *string `json:"host,omitempty"`

	// Port number for the service listening for learning logs. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Port uint32 `json:"port,omitempty"`
}
