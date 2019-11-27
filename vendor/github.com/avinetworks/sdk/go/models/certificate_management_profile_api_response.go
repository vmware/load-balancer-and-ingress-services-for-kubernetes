package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CertificateManagementProfileAPIResponse certificate management profile Api response
// swagger:model CertificateManagementProfileApiResponse
type CertificateManagementProfileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*CertificateManagementProfile `json:"results,omitempty"`
}
