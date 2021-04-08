package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbServiceDownResponse gslb service down response
// swagger:model GslbServiceDownResponse
type GslbServiceDownResponse struct {

	// Fallback IP address to use in A response to the client query when the GSLB service is DOWN.
	FallbackIP *IPAddr `json:"fallback_ip,omitempty"`

	// Fallback IPV6 address to use in AAAA response to the client query when the GSLB service is DOWN. Field introduced in 18.2.8.
	FallbackIp6 *IPAddr `json:"fallback_ip6,omitempty"`

	// Response from DNS service towards the client when the GSLB service is DOWN. Enum options - GSLB_SERVICE_DOWN_RESPONSE_NONE, GSLB_SERVICE_DOWN_RESPONSE_ALL_RECORDS, GSLB_SERVICE_DOWN_RESPONSE_FALLBACK_IP, GSLB_SERVICE_DOWN_RESPONSE_EMPTY.
	// Required: true
	Type *string `json:"type"`
}
