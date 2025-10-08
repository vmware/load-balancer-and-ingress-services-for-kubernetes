/*
Copyright 2019-2025 VMware, Inc.
All Rights Reserved.

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

// ApplicationProfileType defines the type of application profile.
// +kubebuilder:validation:Enum=APPLICATION_PROFILE_TYPE_HTTP
type ApplicationProfileType string

const (
	// types of application profile

	// HTTP application proxy is enabled for this virtual service.
	APPLICATION_PROFILE_TYPE_HTTP ApplicationProfileType = "HTTP"
)

// XFFUpdate defines how X-Forwarded-For headers are handled
// +kubebuilder:validation:Enum=REPLACE_XFF_HEADERS;APPEND_TO_THE_XFF_HEADER;ADD_NEW_XFF_HEADER
type XFFUpdate string

const (
	// Drop all the incoming xff headers and add a new one with client's IP address
	REPLACE_XFF_HEADERS XFFUpdate = "Replace all incoming X-Forward-For headers with the Avi created header."
	// Appends all the incoming XFF headers and client's IP address together
	APPEND_TO_THE_XFF_HEADER XFFUpdate = "All incoming X-Forwarded-For headers will be appended to the Avi created header."
	// Add new XFF header with client's IP address
	ADD_NEW_XFF_HEADER XFFUpdate = "Simply add a new X-Forwarded-For header."
)

// TrueClientIPIndexDirection defines the direction to count IPs in the header.
// +kubebuilder:validation:Enum=LEFT;RIGHT
type TrueClientIPIndexDirection string

const (
	// From Left.
	LEFT TrueClientIPIndexDirection = "From Left."
	// From Right.
	RIGHT TrueClientIPIndexDirection = "From Right."
)

type TrueClientIPConfig struct {
	// +kubebuilder:validation:MaxItems=1
	// HTTP Headers to derive client IP from. If none specified and use_true_client_ip is set to true, it will use X-Forwarded-For header, if present.
	// +kubebuilder:validation:XValidation:rule="self.all(header, size(header) <= 128)",message="Each header name must be at most 128 characters long."
	Headers []string `json:"headers,omitempty"`

	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=1000
	// +optional
	// Position in the configured direction, in the specified header's value, to be used to set true client IP. If the value is greater than the number of IP addresses in the header, then the last IP address in the configured direction in the header will be used.
	IndexInHeader uint32 `json:"index_in_header,omitempty"`

	// +optional
	// Denotes the end from which to count the IPs in the specified header value.
	Direction TrueClientIPIndexDirection `json:"direction,omitempty"`
}

// +kubebuilder:validation:XValidation:rule="((!self.use_true_client_ip) && !has(self.true_client_ip))|| (self.use_true_client_ip && has(self.true_client_ip))",message="true_client_ip can only be configured if use_true_client_ip is true"
// +kubebuilder:validation:XValidation:rule="self.xff_enabled || (!has(self.xff_alternate_name) && !has(self.xff_update))",message="xff_alternate_name and xff_update can only be configured if xff_enabled is true"
// +kubebuilder:validation:XValidation:rule="!(has(self.xff_update) && self.xff_update == 'APPEND_TO_THE_XFF_HEADER' && has(self.xff_alternate_name))",message="if xff_update is APPEND_TO_THE_XFF_HEADER, xff_alternate_name must be an empty string."
type HTTPApplicationProfile struct {
	// Allows HTTP requests, not just TCP connections, to be load balanced across servers. Proxied TCP connections to servers may be reused by multiple clients to improve performance. This feature can not be enabled if the 'Preserve Client IP' feature is enabled.
	// +optional
	ConnectionMultiplexingEnabled bool `json:"connection_multiplexing_enabled,omitempty"`
	// The client's original IP address is inserted into an HTTP request header sent to the server. Servers may use this address for logging or other purposes, rather than Avi's source NAT address used in the Avi to server IP connection.
	// +optional
	XffEnabled bool `json:"xff_enabled,omitempty"`
	// Provide a custom name for the X-Forwarded-For header sent to the servers.
	// +optional
	XffAlternateName string `json:"xff_alternate_name,omitempty"`
	// Configure how incoming X-Forwarded-For headers from the client are handled.
	// +optional
	XffUpdate XFFUpdate `json:"xff_update,omitempty"`
	// Insert an X-Forwarded-Proto header in the request sent to the server. When the client connects via SSL, Avi terminates the SSL, and then forwards the requests to the servers via HTTP, so the servers can determine the original protocol via this header. In this example, the value will be 'https'.
	// +optional
	XForwardedProtoEnabled bool `json:"x_forwarded_proto_enabled,omitempty"`
	// The maximum time - in milliseconds - the server will wait between 2 consecutive read operations of a client request's body chunk. The value '0' specifies no timeout. This setting generally impacts the time allowed for a client to send a POST request with a body, it does not limit the time for the entire request body to be sent.
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100000000
	// (units) = MILLISECONDS,
	ClientBodyTimeout uint32 `json:"client_body_timeout,omitempty"`
	//"The max idle time allowed between HTTP requests over a Keep-alive connection."
	// +optional
	// +kubebuilder:validation:Minimum=10
	// +kubebuilder:validation:Maximum=100000000
	// (units) = MILLISECONDS,
	KeepaliveTimeout uint32 `json:"keepalive_timeout,omitempty"`
	// Use 'Keep-Alive' header timeout sent by application instead of sending the HTTP Keep-Alive Timeout.
	// +optional
	UseAppKeepaliveTimeout bool `json:"use_app_keepalive_timeout,omitempty"`
	// Maximum size for the client request body.  This limits the size of the client data that can be uploaded/posted as part of a single HTTP Request. The default value is 0 and means there is no size limit.
	// +optional
	// (units) = KB,
	ClientMaxBodySize uint32 `json:"client_max_body_size,omitempty"`
	// Send HTTP 'Keep-Alive' header to the client. By default, the timeout specified in the 'keepalive_timeout' field will be used unless the 'Use App Keepalive Timeout' flag is set, in which case the timeout sent by the application will be honored.
	// +optional
	KeepaliveHeader bool `json:"keepalive_header,omitempty"`
	// The max number of HTTP requests that can be sent over a Keep-Alive connection. '0' means unlimited.
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=1000000
	// (special_values) = "{\"0\": \"Unlimited requests on a connection\"}",
	MaxKeepaliveRequests int32 `json:"max_keepalive_requests,omitempty"`
	// If enabled, an HTTP request on an SSL port will result in connection close instead of a 400 response.
	// +optional
	ResetConnHttpOnSslPort bool `json:"reset_conn_http_on_ssl_port,omitempty"`
	// Size of HTTP buffer in kB.
	// +optional
	// (units) = KB,
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=256
	//(special_values) =  {'0': 'Auto compute the size of buffer'},
	HttpUpstreamBufferSize uint32 `json:"http_upstream_buffer_size,omitempty"`
	// Enable chunk body merge for chunked transfer encoding response.
	// +optional
	// +kubebuilder:default=true
	EnableChunkMerge bool `json:"enable_chunk_merge,omitempty"`
	// Detect client IP from user specified header.
	// +optional
	UseTrueClientIP bool `json:"use_true_client_ip,omitempty"`
	// Detect client IP from user specified header at the configured index in the specified direction.
	// +optional
	TrueClientIP *TrueClientIPConfig `json:"true_client_ip,omitempty"`
	// Maximum number of headers allowed in HTTP request and response.
	// +optional
	// +kubebuilder:default=256
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4096
	// (special_values) =  "{\"0\": \"unlimited headers in request and response\"}",
	MaxHeaderCount int32 `json:"max_header_count,omitempty"`
	// Close server-side connection when an error response is received.
	// +optional
	CloseServerSideConnectionOnError bool `json:"close_server_side_connection_on_error,omitempty"`
}

// ApplicationProfileSpec defines the desired state of ApplicationProfile
// Can be created in BASIC and ESSENTIAL license tiers
// +kubebuilder:resource:path=applicationprofiles,shortName=ap,singular=applicationprofile,scope=Namespaced
type ApplicationProfileSpec struct {
	// Type specifies which application layer proxy is enabled for the virtual service.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="applicatio_profile_type is immutable"
	Type ApplicationProfileType `json:"type,omitempty"`

	// HTTPProfile specifies the HTTP application proxy profile parameters.
	// +optional
	HTTPProfile *HTTPApplicationProfile `json:"http_profile,omitempty"`
}

// ApplicationProfileStatus defines the observed state of ApplicationProfile
type ApplicationProfileStatus struct {
	// UUID is unique identifier of the application profile object
	UUID string `json:"uuid"`
	// ObservedGeneration is the observed generation by the operator
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// +optional
	// LastUpdated is the timestamp the object was last updated
	LastUpdated *metav1.Time `json:"lastUpdated"`
	// Conditions is the list of conditions for the application profile
	// Supported condition types:
	// - "Programmed": Indicates whether the ApplicationProfile has been successfully
	//   processed and programmed on the Avi Controller
	//   Possible reasons for True: "Created", "Updated"
	//   Possible reasons for False: "CreationFailed", "UpdateFailed", "UUIDExtractionFailed", "DeletionFailed", "DeletionSkipped"
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// BackendObjectName is the name of the backend object
	BackendObjectName string `json:"backendObjectName,omitempty"`
	// Tenant is the tenant where the application profile is created
	Tenant string `json:"tenant,omitempty"`
	// Field is populated by AKO CRD operator as ako-crd-operator
	// +optional
	Controller string `json:"controller,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=applicationprofiles,scope=Namespaced
// +kubebuilder:subresource:status
// ApplicationProfile is the Schema for the applicationprofiles API
type ApplicationProfile struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of ApplicationProfile
	// +optional
	Spec ApplicationProfileSpec `json:"spec,omitempty"`

	// Status defines the observed state of ApplicationProfile
	// +optional
	Status ApplicationProfileStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// ApplicationProfileList contains a list of ApplicationProfile.
type ApplicationProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApplicationProfile `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ApplicationProfile{}, &ApplicationProfileList{})
}

// Methods to implement ResourceWithStatus interface

// SetConditions sets the conditions in the status
func (ap *ApplicationProfile) SetConditions(conditions []metav1.Condition) {
	ap.Status.Conditions = conditions
}

// GetConditions returns the conditions from the status
func (ap *ApplicationProfile) GetConditions() []metav1.Condition {
	return ap.Status.Conditions
}

// SetObservedGeneration sets the observed generation in the status
func (ap *ApplicationProfile) SetObservedGeneration(generation int64) {
	ap.Status.ObservedGeneration = generation
}

// SetLastUpdated sets the last updated timestamp in the status
func (ap *ApplicationProfile) SetLastUpdated(time *metav1.Time) {
	ap.Status.LastUpdated = time
}

// SetController sets the controller name in the status
func (ap *ApplicationProfile) SetController(controller string) {
	ap.Status.Controller = controller
}
