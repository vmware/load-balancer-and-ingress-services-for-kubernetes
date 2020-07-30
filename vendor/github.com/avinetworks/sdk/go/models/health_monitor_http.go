package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HealthMonitorHTTP health monitor Http
// swagger:model HealthMonitorHttp
type HealthMonitorHTTP struct {

	// Type of the authentication method. Enum options - AUTH_BASIC, AUTH_NTLM. Field introduced in 20.1.1.
	AuthType *string `json:"auth_type,omitempty"`

	// Use the exact http_request *string as specified by user, without any automatic insert of headers like Host header. Field introduced in 17.1.6,17.2.2.
	ExactHTTPRequest *bool `json:"exact_http_request,omitempty"`

	// Send an HTTP request to the server.  The default GET / HTTP/1.0 may be extended with additional headers or information.  For instance, GET /index.htm HTTP/1.1 Host  www.site.com Connection  Close.
	HTTPRequest *string `json:"http_request,omitempty"`

	// HTTP request body. Field introduced in 20.1.1.
	HTTPRequestBody *string `json:"http_request_body,omitempty"`

	// Match for a keyword in the first 2Kb of the server header and body response.
	HTTPResponse *string `json:"http_response,omitempty"`

	// List of HTTP response codes to match as successful.  Default is 2xx. Enum options - HTTP_ANY, HTTP_1XX, HTTP_2XX, HTTP_3XX, HTTP_4XX, HTTP_5XX.
	HTTPResponseCode []string `json:"http_response_code,omitempty"`

	// Match or look for this HTTP response code indicating server maintenance.  A successful match results in the server being marked down. Allowed values are 101-599.
	MaintenanceCode []int64 `json:"maintenance_code,omitempty,omitempty"`

	// Match or look for this keyword in the first 2KB of server header and body response indicating server maintenance.  A successful match results in the server being marked down.
	MaintenanceResponse *string `json:"maintenance_response,omitempty"`

	// Expected http/https response page size. Allowed values are 2048-16384. Field introduced in 20.1.1.
	ResponseSize *int32 `json:"response_size,omitempty"`

	// SSL attributes for HTTPS health monitor. Field introduced in 17.1.1.
	SslAttributes *HealthMonitorSSlattributes `json:"ssl_attributes,omitempty"`
}
