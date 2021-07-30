// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HealthMonitorPop3 health monitor pop3
// swagger:model HealthMonitorPop3
type HealthMonitorPop3 struct {

	// SSL attributes for POP3S monitor. Field introduced in 20.1.5.
	SslAttributes *HealthMonitorSSlattributes `json:"ssl_attributes,omitempty"`
}
