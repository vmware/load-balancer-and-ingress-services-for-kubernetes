// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// UserAccountProfile user account profile
// swagger:model UserAccountProfile
type UserAccountProfile struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Lock timeout period (in minutes). Default is 30 minutes. Unit is MIN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AccountLockTimeout *uint32 `json:"account_lock_timeout,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// The time period after which credentials expire. Default is 180 days. Unit is DAYS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CredentialsTimeoutThreshold *uint32 `json:"credentials_timeout_threshold,omitempty"`

	// The configurable time window beyond which we need to pop all the login failure timestamps from the login_failure_timestamps. Special values are 0 - Do not reset login_failure_counts on the basis of time.. Field introduced in 22.1.1. Unit is MIN. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LoginFailureCountExpiryWindow *uint32 `json:"login_failure_count_expiry_window,omitempty"`

	// Maximum number of concurrent sessions allowed. There are unlimited sessions by default. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxConcurrentSessions *uint32 `json:"max_concurrent_sessions,omitempty"`

	// Number of login attempts before lockout. Default is 3 attempts. Allowed values are 3-20. Special values are 0- Unlimited login attempts allowed.. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxLoginFailureCount *uint32 `json:"max_login_failure_count,omitempty"`

	// Maximum number of passwords to be maintained in the password history. Default is 4 passwords. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxPasswordHistoryCount *uint32 `json:"max_password_history_count,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}
