package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// OCSPConfig o c s p config
// swagger:model OCSPConfig
type OCSPConfig struct {

	// Describes the Time Interval after which the next OCSP job needs to be scheduled in case of the OCSP job failures. Allowed values are 15-MAX. Field introduced in 20.1.1.
	FailedOcspJobsRetryInterval *int32 `json:"failed_ocsp_jobs_retry_interval,omitempty"`

	// Interval between the OCSP queries. Allowed values are 30-MAX. Field introduced in 20.1.1.
	OcspReqInterval *int32 `json:"ocsp_req_interval,omitempty"`

	// Time in seconds that the system waits for a reply from the OCSP responder before dropping the connection. Field introduced in 20.1.1.
	OcspRespTimeout *int32 `json:"ocsp_resp_timeout,omitempty"`

	// List of Responder URLs configured by user to do failover/override the AIA extension contained in the OCSP responder's SSL/TLS certificate. Field introduced in 20.1.1.
	ResponderURLLists []string `json:"responder_url_lists,omitempty"`

	// Describes the type of action to take with the Responder URLs. Enum options - OCSP_RESPONDER_URL_FAILOVER, OCSP_RESPONDER_URL_OVERRIDE. Field introduced in 20.1.1.
	URLAction *string `json:"url_action,omitempty"`
}
