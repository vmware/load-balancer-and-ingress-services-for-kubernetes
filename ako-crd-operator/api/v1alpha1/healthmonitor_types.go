/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
	// Name is name of the spec and optional. If not present will be inferred from Metadata
	// +kubebuilder:validation:Optional
	Name string `json:"name"`
	// SendInterval is the frequency, in seconds, that pings are sent.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=3600
	// +kubebuilder:default:=10
	SendInterval int32 `json:"send_interval,omitempty"`

	// ReceiveTimeout is the timeout for receiving a ping response, in seconds.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=2400
	// +kubebuilder:default:=4
	ReceiveTimeout int32 `json:"receive_timeout,omitempty"`

	// SuccessfulChecks is the number of successful pings before marking up.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=50
	// +kubebuilder:default:=2
	SuccessfulChecks int32 `json:"successful_checks,omitempty"`

	// FailedChecks is the number of failed pings before marking down.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=50
	// +kubebuilder:default:=2
	FailedChecks int32 `json:"failed_checks,omitempty"`

	// Type is the type of health monitor.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="type is immutable"
	Type HealthMonitorType `json:"type,omitempty"`

	// MonitorPort is the port to use for the health check.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	MonitorPort int32 `json:"monitor_port,omitempty"`

	// Authentication defines the authentication information for HTTP/HTTPS monitors.
	// +optional
	Authentication *HealthMonitorInfo `json:"authentication,omitempty"`

	// IsFederated describes the object's replication scope. If the
	// field is set to false, then the object is visible within
	// the controller-cluster and its associated service-engines.
	// If the field is set to true, then the object is replicated
	// across the federation
	IsFederated bool `json:"is_federated,omitempty"`

	// TCP defines the TCP monitor configuration.
	// +optional
	TCP *TCPMonitor `json:"tcp_monitor,omitempty"`

	// HTTP defines the HTTP monitor configuration.
	// +optional
	HTTP *HTTPMonitor `json:"http_monitor,omitempty"`
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

// +kubebuilder:validation:XValidation:rule="self.tcp_half_open == false || (self.tcp_request == '' && self.tcp_response == '' && self.maintenance_response == '')",message="tcp_request, tcp_response, and maintenance_response cannot be set when tcp_half_open is true"

// TCPMonitor defines the TCP monitor configuration.
type TCPMonitor struct {
	// TcpRequest is the request to send for the TCP health check.
	// +optional
	// +kubebuilder:validation:MaxLength=1024
	TcpRequest string `json:"tcp_request,omitempty"`
	// TcpResponse is the expected response for the TCP health check.
	// +optional
	// +kubebuilder:validation:MaxLength=512
	TcpResponse string `json:"tcp_response,omitempty"`
	// MaintenanceResponse is the response to send when in maintenance mode.
	// +optional
	// +kubebuilder:validation:MaxLength=512
	MaintenanceResponse string `json:"maintenance_response,omitempty"`
	// TcpHalfOpen is a boolean to check if the tcp monitor is in half open mode
	// +optional
	TcpHalfOpen bool `json:"tcp_half_open,omitempty"`
}

// +kubebuilder:validation:Enum=HTTP_ANY;HTTP_1XX;HTTP_2XX;HTTP_3XX;HTTP_4XX;HTTP_5XX
type HTTPResponseCode string

const (
	HTTPAny HTTPResponseCode = "HTTP_ANY"
	HTTP1XX                  = "HTTP_1XX"
	HTTP2XX                  = "HTTP_2XX"
	HTTP3XX                  = "HTTP_3XX"
	HTTP4XX                  = "HTTP_4XX"
	HTTP5XX                  = "HTTP_5XX"
)

// HTTPMonitor defines the HTTP monitor configuration.
type HTTPMonitor struct {
	// HTTPRequest is the HTTP request to send.
	// +optional
	// +kubebuilder:validation:MaxLength=1024
	// +kubebuilder:default:="GET / HTTP/1.0"
	HTTPRequest string `json:"http_request,omitempty"`
	// HTTPResponseCode is the list of expected HTTP response codes.
	// +kubebuilder:validation:MinItems=1
	HTTPResponseCode []HTTPResponseCode `json:"http_response_code,omitempty"` // Use string array for enum values
	// HTTPResponse is a keyword to match in the response body.
	// +optional
	// +kubebuilder:validation:MaxLength=512
	HTTPResponse string `json:"http_response,omitempty"`
	// MaintenanceCode is the HTTP code for maintenance response.
	// +optional
	// +kubebuilder:validation:items:Minimum=101
	// +kubebuilder:validation:items:Maximum=599
	// +kubebuilder:validation:MaxItems=4
	// +kubebuilder:validation:items:Format=uint32
	MaintenanceCode []uint32 `json:"maintenance_code,omitempty"`
	// MaintenanceResponse is the body content to match for maintenance response.
	// +optional
	// +kubebuilder:validation: MaxLength=512
	MaintenanceResponse string `json:"maintenance_response,omitempty"`
	// ExactHttpRequest checks if the whole http request should match.
	// +optional
	ExactHttpRequest bool `json:"exact_http_request,omitempty"`
	// AuthType is the type of authentication to use.
	AuthType HealthMonitorAuth `json:"auth_type,omitempty"` // Handle enum conversion
	// HTTPRequestBody is the request body to send.
	// +optional
	HTTPRequestBody string `json:"http_request_body,omitempty"`
}

// HealthMonitorStatus defines the observed state of HealthMonitor
type HealthMonitorStatus struct {
	// Status of the healthmonitor
	Status string `json:"status,omitempty"`
	// Error if any error was encountered
	Error string `json:"error"`
	// UUID is unique identifier of the health monitor object
	Uuid string `json:"uuid"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=healthmonitors,scope=Namespaced
// +kubebuilder:subresource:status

// HealthMonitor is the Schema for the healthmonitors API
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=healthmonitors,shortName=hm,singular=healthmonitor,scope=Namespaced
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
// HealthMonitorList contains a list of HealthMonitor
type HealthMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HealthMonitor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HealthMonitor{}, &HealthMonitorList{})
}
