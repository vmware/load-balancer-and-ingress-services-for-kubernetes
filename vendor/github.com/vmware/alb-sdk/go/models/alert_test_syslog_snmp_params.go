package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AlertTestSyslogSnmpParams alert test syslog snmp params
// swagger:model AlertTestSyslogSnmpParams
type AlertTestSyslogSnmpParams struct {

	// The contents of the Syslog message/SNMP Trap contents.
	// Required: true
	Text *string `json:"text"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
