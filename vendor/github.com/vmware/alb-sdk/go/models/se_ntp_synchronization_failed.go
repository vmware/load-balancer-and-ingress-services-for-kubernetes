// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeNtpSynchronizationFailed se ntp synchronization failed
// swagger:model SeNtpSynchronizationFailed
type SeNtpSynchronizationFailed struct {

	// List of NTP Servers. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NtpServers *string `json:"ntp_servers,omitempty"`

	// Name of the SE reporting this event. It is a reference to an object of type ServiceEngine. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeName *string `json:"se_name,omitempty"`

	// UUID of the SE responsible for this event. It is a reference to an object of type ServiceEngine. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeRef *string `json:"se_ref,omitempty"`
}
