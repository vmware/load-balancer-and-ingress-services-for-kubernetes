package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SSLKeyAndCertificateAPIResponse s s l key and certificate Api response
// swagger:model SSLKeyAndCertificateApiResponse
type SSLKeyAndCertificateAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*SSLKeyAndCertificate `json:"results,omitempty"`
}
