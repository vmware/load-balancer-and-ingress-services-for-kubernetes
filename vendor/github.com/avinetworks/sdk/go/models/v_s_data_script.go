package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VSDataScript v s data script
// swagger:model VSDataScript
type VSDataScript struct {

	// Event triggering execution of datascript. Enum options - VS_DATASCRIPT_EVT_HTTP_REQ, VS_DATASCRIPT_EVT_HTTP_RESP, VS_DATASCRIPT_EVT_HTTP_RESP_DATA, VS_DATASCRIPT_EVT_HTTP_LB_FAILED, VS_DATASCRIPT_EVT_HTTP_REQ_DATA, VS_DATASCRIPT_EVT_HTTP_RESP_FAILED, VS_DATASCRIPT_EVT_TCP_CLIENT_ACCEPT, VS_DATASCRIPT_EVT_DNS_REQ, VS_DATASCRIPT_EVT_DNS_RESP, VS_DATASCRIPT_EVT_MAX.
	// Required: true
	Evt *string `json:"evt"`

	// Datascript to execute when the event triggers.
	// Required: true
	Script *string `json:"script"`
}
