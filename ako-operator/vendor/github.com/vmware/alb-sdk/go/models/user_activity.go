// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// UserActivity user activity
// swagger:model UserActivity
type UserActivity struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Number of concurrent user sessions open. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ConcurrentSessions *uint32 `json:"concurrent_sessions,omitempty"`

	// Number of failed login attempts before a successful login. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FailedLoginAttempts *uint32 `json:"failed_login_attempts,omitempty"`

	// IP of the machine the user was last logged in from. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LastLoginIP *string `json:"last_login_ip,omitempty"`

	// Timestamp of last login. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LastLoginTimestamp *string `json:"last_login_timestamp,omitempty"`

	// Timestamp of last password update. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LastPasswordUpdate *string `json:"last_password_update,omitempty"`

	// Indicates whether the user is logged in or not. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LoggedIn *bool `json:"logged_in,omitempty"`

	// Its a queue that store the timestamps for past login_failures. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LoginFailureTimestamps []string `json:"login_failure_timestamps,omitempty"`

	// Name of the user this object refers to. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`

	// Stores the previous n passwords  where n is ControllerProperties.max_password_history_count. . Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PreviousPassword []string `json:"previous_password,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}
