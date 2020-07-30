package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// OCSPResponseInfo o c s p response info
// swagger:model OCSPResponseInfo
type OCSPResponseInfo struct {

	// Revocation status of the certificate. Enum options - OCSP_CERTSTATUS_GOOD, OCSP_CERTSTATUS_REVOKED, OCSP_CERTSTATUS_UNKNOWN. Field introduced in 20.1.1.
	// Required: true
	// Read Only: true
	CertStatus *string `json:"cert_status"`

	// The time at or before which newer information will be available about the status of the certificate. Field introduced in 20.1.1.
	// Read Only: true
	NextUpdate *string `json:"next_update,omitempty"`

	// The OCSP Responder URL from which the response is received. Field introduced in 20.1.1.
	// Required: true
	// Read Only: true
	OcspRespFromResponderURL *string `json:"ocsp_resp_from_responder_url"`

	// Signed OCSP response received from the CA's OCSP Responder. Field introduced in 20.1.1.
	// Required: true
	// Read Only: true
	OcspResponse *string `json:"ocsp_response"`

	// The reason for the revocation of the certificate. Enum options - OCSP_REVOCATION_REASON_UNSPECIFIED, OCSP_REVOCATION_REASON_KEY_COMPROMISE, OCSP_REVOCATION_REASON_CA_COMPROMISE, OCSP_REVOCATION_REASON_AFFILIATION_CHANGED, OCSP_REVOCATION_REASON_SUPERSEDED, OCSP_REVOCATION_REASON_CESSATION_OF_OPERATION, OCSP_REVOCATION_REASON_CERTIFICATE_HOLD, OCSP_REVOCATION_REASON_REMOVE_FROM_CRL, OCSP_REVOCATION_REASON_PRIVILEGE_WITHDRAWN, OCSP_REVOCATION_REASON_AA_COMPROMISE. Field introduced in 20.1.1.
	// Read Only: true
	RevocationReason *string `json:"revocation_reason,omitempty"`

	// ISO 8601 compatible timestamp at which the certificate was revoked or placed on hold. Field introduced in 20.1.1.
	// Read Only: true
	RevocationTime *string `json:"revocation_time,omitempty"`

	// The most recent time at which the status being indicated is known by the OCSP Responder to have been correct. Field introduced in 20.1.1.
	// Read Only: true
	ThisUpdate *string `json:"this_update,omitempty"`
}
