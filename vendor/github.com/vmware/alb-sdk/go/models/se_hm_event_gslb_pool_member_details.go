package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeHmEventGslbPoolMemberDetails se hm event gslb pool member details
// swagger:model SeHmEventGslbPoolMemberDetails
type SeHmEventGslbPoolMemberDetails struct {

	// Placeholder for description of property app_info of obj type SeHmEventGslbPoolMemberDetails field type str  type object
	AppInfo []*AppInfo `json:"app_info,omitempty"`

	// Domain name used to health monitor this member.
	Domain *string `json:"domain,omitempty"`

	// Gslb health monitor failure code. Enum options - ARP_UNRESOLVED, CONNECTION_REFUSED, CONNECTION_TIMEOUT, RESPONSE_CODE_MISMATCH, PAYLOAD_CONTENT_MISMATCH, SERVER_UNREACHABLE, CONNECTION_RESET, CONNECTION_ERROR, HOST_ERROR, ADDRESS_ERROR, NO_PORT, PAYLOAD_TIMEOUT, NO_RESPONSE, NO_RESOURCES, SSL_ERROR, SSL_CERT_ERROR, PORT_UNREACHABLE, SCRIPT_ERROR, OTHER_ERROR, SERVER_DISABLED...
	FailureCode *string `json:"failure_code,omitempty"`

	// IP address of GslbService member.
	// Required: true
	IP *IPAddr `json:"ip"`

	// Placeholder for description of property shm of obj type SeHmEventGslbPoolMemberDetails field type str  type object
	Shm []*SeHmEventShmDetails `json:"shm,omitempty"`

	//  Enum options - ADF_CLIENT_CONN_SETUP_REFUSED. ADF_SERVER_CONN_SETUP_REFUSED. ADF_CLIENT_CONN_SETUP_TIMEDOUT. ADF_SERVER_CONN_SETUP_TIMEDOUT. ADF_CLIENT_CONN_SETUP_FAILED_INTERNAL. ADF_SERVER_CONN_SETUP_FAILED_INTERNAL. ADF_CLIENT_CONN_SETUP_FAILED_BAD_PACKET. ADF_UDP_CONN_SETUP_FAILED_INTERNAL. ADF_UDP_SERVER_CONN_SETUP_FAILED_INTERNAL. ADF_CLIENT_SENT_RESET. ADF_SERVER_SENT_RESET. ADF_CLIENT_CONN_TIMEDOUT. ADF_SERVER_CONN_TIMEDOUT. ADF_USER_DELETE_OPERATION. ADF_CLIENT_REQUEST_TIMEOUT. ADF_CLIENT_CONN_ABORTED. ADF_CLIENT_SSL_HANDSHAKE_FAILURE. ADF_CLIENT_CONN_FAILED. ADF_SERVER_CERTIFICATE_VERIFICATION_FAILED. ADF_SERVER_SIDE_SSL_HANDSHAKE_FAILED...
	SslErrorCode *string `json:"ssl_error_code,omitempty"`
}
