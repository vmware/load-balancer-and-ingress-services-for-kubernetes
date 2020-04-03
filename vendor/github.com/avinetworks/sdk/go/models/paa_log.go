package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// PaaLog paa log
// swagger:model PaaLog
type PaaLog struct {

	// PingAccess Agent cache was used for authentication. Field introduced in 18.2.3.
	CacheHit *bool `json:"cache_hit,omitempty"`

	// The PingAccess server required the client request body for authentication. Field introduced in 18.2.3.
	ClientRequestBodySent *bool `json:"client_request_body_sent,omitempty"`

	// Logs for each request sent to PA server to completeauthentication for the initial request. Field introduced in 18.2.3.
	RequestLogs []*PaaRequestLog `json:"request_logs,omitempty"`
}
