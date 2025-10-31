// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// IPPersistenceProfile IP persistence profile
// swagger:model IPPersistenceProfile
type IPPersistenceProfile struct {

	// Mask to be applied on client IP. This may be used to persist clients from a subnet to the same server. When set to 0, all requests are sent to the same server. Allowed values are 0-128. Field introduced in 18.2.7. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IPMask *uint32 `json:"ip_mask,omitempty"`

	// The length of time after a client's connections have closed before expiring the client's persistence to a server. Allowed values are 1-720. Unit is MIN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IPPersistentTimeout *int32 `json:"ip_persistent_timeout,omitempty"`
}
