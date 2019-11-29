package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HealthMonitorDNS health monitor DNS
// swagger:model HealthMonitorDNS
type HealthMonitorDNS struct {

	//   Query_Type  Response has atleast one answer of which      the resource record type matches the query type   Any_Type  Response should contain atleast one answer  AnyThing  An empty answer is enough. Enum options - DNS_QUERY_TYPE, DNS_ANY_TYPE, DNS_ANY_THING.
	Qtype *string `json:"qtype,omitempty"`

	// The DNS monitor will query the DNS server for the fully qualified name in this field.
	// Required: true
	QueryName *string `json:"query_name"`

	// When No Error is selected, a DNS query will be marked failed is any error code is returned by the server.  With Any selected, the monitor ignores error code in the responses. Enum options - RCODE_NO_ERROR, RCODE_ANYTHING.
	Rcode *string `json:"rcode,omitempty"`

	// The resource record of the queried DNS server's response for the Request Name must include the IP address defined in this field. .
	ResponseString *string `json:"response_string,omitempty"`
}
