// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AlertTestSyslogSnmpParams alert test syslog snmp params
// swagger:model AlertTestSyslogSnmpParams
type AlertTestSyslogSnmpParams struct {

	// The contents of the Syslog message/SNMP Trap contents.
	// Required: true
	Text *string `json:"text"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
