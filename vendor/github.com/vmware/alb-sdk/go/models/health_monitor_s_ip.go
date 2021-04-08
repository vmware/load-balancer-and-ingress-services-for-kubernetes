package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HealthMonitorSIP health monitor s IP
// swagger:model HealthMonitorSIP
type HealthMonitorSIP struct {

	// Specify the transport protocol TCP or UDP, to be used for SIP health monitor. The default transport is UDP. Enum options - SIP_UDP_PROTO, SIP_TCP_PROTO. Field introduced in 17.2.14, 18.1.5, 18.2.1.
	SipMonitorTransport *string `json:"sip_monitor_transport,omitempty"`

	// Specify the SIP request to be sent to the server. By default, SIP OPTIONS request will be sent. Enum options - SIP_OPTIONS. Field introduced in 17.2.8, 18.1.3, 18.2.1.
	SipRequestCode *string `json:"sip_request_code,omitempty"`

	// Match for a keyword in the first 2KB of the server header and body response. By default, it matches for SIP/2.0. Field introduced in 17.2.8, 18.1.3, 18.2.1.
	SipResponse *string `json:"sip_response,omitempty"`
}
