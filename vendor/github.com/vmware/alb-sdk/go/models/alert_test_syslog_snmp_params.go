// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AlertTestSyslogSnmpParams alert test syslog snmp params
// swagger:model AlertTestSyslogSnmpParams
type AlertTestSyslogSnmpParams struct {

	// The contents of the Syslog message/SNMP Trap contents. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Text *string `json:"text"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}
