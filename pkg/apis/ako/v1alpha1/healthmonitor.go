package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HealthMonitorType defines the type of health monitor.
// +kubebuilder:validation:Enum=HEALTH_MONITOR_TCP;HEALTH_MONITOR_PING;HEALTH_MONITOR_HTTP;HEALTH_MONITOR_HTTPS
type HealthMonitorType string

const (
	// types of health monitor
	HealthMonitorTCP   HealthMonitorType = "HEALTH_MONITOR_TCP"
	HealthMonitorPing  HealthMonitorType = "HEALTH_MONITOR_PING"
	HealthMonitorHTTP  HealthMonitorType = "HEALTH_MONITOR_HTTP"
	HealthMonitorHTTPS HealthMonitorType = "HEALTH_MONITOR_HTTPS"
)

// HealthMonitorAuth defines the type of authentication for HTTP/HTTPS monitors.
// +kubebuilder:validation:Enum=Basic;NTLM
type HealthMonitorAuth string

const (
	HealthMonitorBasicAuth HealthMonitorAuth = "Basic"
	HealthMonitorNTLM      HealthMonitorAuth = "NTLM"
)

// HealthMonitorSpec defines the desired state of HealthMonitor
type HealthMonitorSpec struct {
	// SendInterval is the frequency, in seconds, that pings are sent.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=3600
	SendInterval int32 `json:"send_interval,omitempty"`

	// ReceiveTimeout is the timeout for receiving a ping response, in seconds.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=2400
	ReceiveTimeout int32 `json:"receive_timeout,omitempty"`

	// SuccessfulChecks is the number of successful pings before marking up.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=50
	SuccessfulChecks int32 `json:"successful_checks,omitempty"`

	// FailedChecks is the number of failed pings before marking down.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=50
	FailedChecks int32 `json:"failed_checks,omitempty"`

	// TenantRef is the reference to the tenant in Avi
	TenantRef string `json:"tenant_ref,omitempty"`

	// CloudRef is the reference to the cloud in Avi
	CloudRef string `json:"cloud_ref,omitempty"`

	// Type is the type of health monitor.
	Type HealthMonitorType `json:"type,omitempty"`

	// MonitorPort is the port to use for the health check.
	MonitorPort int32 `json:"monitor_port,omitempty"`

	// Authentication defines the authentication information for HTTP/HTTPS monitors.
	// +optional
	Authentication *HealthMonitorInfo `json:"authentication,omitempty"`

	// TCP defines the TCP monitor configuration.
	// +optional
	TCP *TCPMonitor `json:"tcp,omitempty"`

	// HTTP defines the HTTP monitor configuration.
	// +optional
	HTTP *HTTPMonitor `json:"http,omitempty"`

	// HTTPS defines the HTTPS monitor configuration.
	// +optional
	HTTPS *HTTPMonitor `json:"https,omitempty"`
}

// HealthMonitorInfo defines authentication information for HTTP/HTTPS monitors.
type HealthMonitorInfo struct {
	// Username for server authentication.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=128
	Username string `json:"username"`
	// Password for server authentication.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=128
	Password string `json:"password"`
}

// TCPMonitor defines the TCP monitor configuration.
type TCPMonitor struct {
	// MonitorPort is the port to use for the TCP health check.
	MonitorPort int32 `json:"monitor_port,omitempty"`
	// TcpRequest is the request to send for the TCP health check.
	TcpRequest string `json:"tcp_request,omitempty"`
	// TcpResponse is the expected response for the TCP health check.
	TcpResponse string `json:"tcp_response,omitempty"`
	// MaintenanceResponse is the response to send when in maintenance mode.
	MaintenanceResponse string `json:"maintenance_response,omitempty"`
	// TcpHalfOpen is a boolean to check if the tcp monitor is in half open mode
	TcpHalfOpen bool `json:"tcp_half_open,omitempty"`
	// Add other common fields
}

// HTTPMonitor defines the HTTP monitor configuration.
type HTTPMonitor struct {
	// HTTPRequest is the HTTP request to send.
	HTTPRequest string `json:"http_request,omitempty"`
	// HTTPResponseCode is the list of expected HTTP response codes.
	HTTPResponseCode []string `json:"http_response_code,omitempty"` // Use string array for enum values
	// HTTPResponse is a keyword to match in the response body.
	HTTPResponse string `json:"http_response,omitempty"`
	// MaintenanceCode is the HTTP code for maintenance response.
	MaintenanceCode []int32 `json:"maintenance_code,omitempty"`
	// MaintenanceResponse is the body content to match for maintenance response.
	MaintenanceResponse string `json:"maintenance_response,omitempty"`
	// SslAttributes is the SSL attributes to use for HTTPS.
	SslAttributes HealthMonitorSSlattributes `json:"ssl_attributes,omitempty"`
	// ExactHttpRequest checks if the whole http request should match.
	ExactHttpRequest bool `json:"exact_http_request,omitempty"`
	// AuthType is the type of authentication to use.
	AuthType HealthMonitorAuth `json:"auth_type,omitempty"` // Handle enum conversion
	// HTTPRequestBody is the request body to send.
	HTTPRequestBody string `json:"http_request_body,omitempty"`
	// ResponseSize is the expected size of the response.
	ResponseSize int32 `json:"response_size,omitempty"`
	// HTTPHeaders is the list of headers to send.
	HTTPHeaders []string `json:"http_headers,omitempty"`
	// HTTPMethod is the HTTP method to use.
	HTTPMethod string `json:"http_method,omitempty"`
	// HTTPRequestHeaderPath is the path to use for headers.
	HTTPRequestHeaderPath string `json:"http_request_header_path,omitempty"`
}

// HealthMonitorSSlattributes defines the SSL attributes for HTTPS monitors.
type HealthMonitorSSlattributes struct {
	// PKI profile used to validate the SSL certificate presented by a server. It is a reference to an object of type PKIProfile. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PkiProfileRef string `json:"pki_profile_ref,omitempty"`

	// Fully qualified DNS hostname which will be used in the TLS SNI extension in server connections indicating SNI is enabled. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerName string `json:"server_name,omitempty"`

	// Service engines will present this SSL certificate to the server. It is a reference to an object of type SSLKeyAndCertificate. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SslKeyAndCertificateRef string `json:"ssl_key_and_certificate_ref,omitempty"`

	// SSL profile defines ciphers and SSL versions to be used for healthmonitor traffic to the back-end servers. It is a reference to an object of type SSLProfile. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// +kubebuilder:validation:Required
	SslProfileRef string `json:"ssl_profile_ref"`
}

// HealthMonitorStatus defines the observed state of HealthMonitor
type HealthMonitorStatus struct {
	// Status of the healthmonitor
	Status string `json:"status,omitempty"`
	// Error if any error was encountered
	Error string `json:"error"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=healthmonitors,scope=Namespaced
// +kubebuilder:subresource:status

// HealthMonitor is the Schema for the healthmonitors API
type HealthMonitor struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of HealthMonitor
	Spec HealthMonitorSpec `json:"spec,omitempty"`

	// Status defines the observed state of HealthMonitor
	// +optional
	Status HealthMonitorStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// HealthMonitorList contains a list of HealthMonitor
type HealthMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HealthMonitor `json:"items"`
}
