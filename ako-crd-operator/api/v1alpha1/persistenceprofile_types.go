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

// ServerHmDownRecovery defines the behavior when a persistent server has been marked down.
// +kubebuilder:validation:Enum=HM_DOWN_PICK_NEW_SERVER;HM_DOWN_ABORT_CONNECTION;HM_DOWN_CONTINUE_PERSISTENT_SERVER
type ServerHmDownRecovery string

const (
	// ServerHmDownPickNewServer indicates to pick a new server when the persistent server is down.
	ServerHmDownPickNewServer ServerHmDownRecovery = "HM_DOWN_PICK_NEW_SERVER"
	// ServerHmDownAbortConnection indicates to abort the connection when the persistent server is down.
	ServerHmDownAbortConnection ServerHmDownRecovery = "HM_DOWN_ABORT_CONNECTION"
	// ServerHmDownContinuePersistentServer indicates to continue using the persistent server even if it's down.
	ServerHmDownContinuePersistentServer ServerHmDownRecovery = "HM_DOWN_CONTINUE_PERSISTENT_SERVER"
)

// PersistenceType defines the method used to persist clients to the same server.
// +kubebuilder:validation:Enum=PERSISTENCE_TYPE_CLIENT_IP_ADDRESS;PERSISTENCE_TYPE_HTTP_COOKIE;PERSISTENCE_TYPE_TLS;PERSISTENCE_TYPE_CLIENT_IPV6_ADDRESS;PERSISTENCE_TYPE_CUSTOM_HTTP_HEADER;PERSISTENCE_TYPE_APP_COOKIE
type PersistenceType string

const (
	// PersistenceTypeClientIPAddress indicates persistence based on client IP address.
	PersistenceTypeClientIPAddress PersistenceType = "PERSISTENCE_TYPE_CLIENT_IP_ADDRESS"
	// PersistenceTypeHTTPCookie indicates persistence based on HTTP cookie.
	PersistenceTypeHTTPCookie PersistenceType = "PERSISTENCE_TYPE_HTTP_COOKIE"
	// PersistenceTypeTLS indicates persistence based on TLS session.
	PersistenceTypeTLS PersistenceType = "PERSISTENCE_TYPE_TLS"
	// PersistenceTypeClientIPV6Address indicates persistence based on client IPv6 address.
	PersistenceTypeClientIPV6Address PersistenceType = "PERSISTENCE_TYPE_CLIENT_IPV6_ADDRESS"
	// PersistenceTypeCustomHTTPHeader indicates persistence based on a custom HTTP header.
	PersistenceTypeCustomHTTPHeader PersistenceType = "PERSISTENCE_TYPE_CUSTOM_HTTP_HEADER"
	// PersistenceTypeAppCookie indicates persistence based on an application cookie.
	PersistenceTypeAppCookie PersistenceType = "PERSISTENCE_TYPE_APP_COOKIE"
)

// PersistenceProfileSpec defines the desired state of PersistenceProfile
// +kubebuilder:validation:XValidation:rule="((self.persistence_type == 'PERSISTENCE_TYPE_CLIENT_IP_ADDRESS' || self.persistence_type == 'PERSISTENCE_TYPE_CLIENT_IPV6_ADDRESS') && has(self.ip_persistence_profile) && !has(self.hdr_persistence_profile) && !has(self.app_cookie_persistence_profile) && !has(self.http_cookie_persistence_profile)) || (self.persistence_type == 'PERSISTENCE_TYPE_CUSTOM_HTTP_HEADER' && has(self.hdr_persistence_profile) && !has(self.ip_persistence_profile) && !has(self.app_cookie_persistence_profile) && !has(self.http_cookie_persistence_profile)) || (self.persistence_type == 'PERSISTENCE_TYPE_APP_COOKIE' && has(self.app_cookie_persistence_profile) && !has(self.ip_persistence_profile) && !has(self.hdr_persistence_profile) && !has(self.http_cookie_persistence_profile)) || (self.persistence_type == 'PERSISTENCE_TYPE_HTTP_COOKIE' && has(self.http_cookie_persistence_profile) && !has(self.ip_persistence_profile) && !has(self.hdr_persistence_profile) && !has(self.app_cookie_persistence_profile)) || (self.persistence_type == 'PERSISTENCE_TYPE_TLS' && !has(self.ip_persistence_profile) && !has(self.hdr_persistence_profile) && !has(self.app_cookie_persistence_profile) && !has(self.http_cookie_persistence_profile))", message="Invalid profile configuration for persistence_type. When persistence_type is CLIENT_IP_ADDRESS, CLIENT_IPV6_ADDRESS, CUSTOM_HTTP_HEADER, APP_COOKIE, or HTTP_COOKIE, its corresponding profile field (e.g. ipPersistenceProfile) must be set and other profile fields must be absent. For types like TLS, none of these specific profile fields should be set."
type PersistenceProfileSpec struct {

	// ServerHmDownRecovery specifies behavior when a persistent server has been marked down by a health monitor.
	// +kubebuilder:default:=HM_DOWN_PICK_NEW_SERVER
	ServerHmDownRecovery ServerHmDownRecovery `json:"server_hm_down_recovery,omitempty"`

	// PersistenceType is the method used to persist clients to the same server.
	// +kubebuilder:default:=PERSISTENCE_TYPE_CLIENT_IP_ADDRESS
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="type is immutable"
	PersistenceType PersistenceType `json:"persistence_type,omitempty"`

	// IPPersistenceProfile specifies the Client IP Persistence profile parameters.
	// +optional
	IPPersistenceProfile *IPPersistenceProfile `json:"ip_persistence_profile,omitempty"`

	// HdrPersistenceProfile specifies the custom HTTP Header Persistence profile parameters.
	// +optional
	HdrPersistenceProfile *HdrPersistenceProfile `json:"hdr_persistence_profile,omitempty"`

	// AppCookiePersistenceProfile specifies the Application Cookie Persistence profile parameters.
	// +optional
	AppCookiePersistenceProfile *AppCookiePersistenceProfile `json:"app_cookie_persistence_profile,omitempty"`

	// HTTPCookiePersistenceProfile specifies the HTTP Cookie Persistence profile parameters.
	// +optional
	HTTPCookiePersistenceProfile *HTTPCookiePersistenceProfile `json:"http_cookie_persistence_profile,omitempty"`

	// IsFederated describes the object's replication scope.
	// +kubebuilder:default:=false
	IsFederated bool `json:"is_federated,omitempty"`

	// Description is a user-friendly description of the persistence profile.
	// +optional
	Description string `json:"description,omitempty"`
}

// IPPersistenceProfile specifies the Client IP Persistence profile parameters.
type IPPersistenceProfile struct {
	// IPPersistentTimeout is the length of time after a client's connections have closed before expiring the client's persistence to a server.
	// +kubebuilder:default:=5
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=720
	IPPersistentTimeout int32 `json:"ip_persistent_timeout,omitempty"`

	// IPMask is the mask to be applied on client IP.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=128
	IPMask int32 `json:"ip_mask,omitempty"`
}

// HdrPersistenceProfile specifies the custom HTTP Header Persistence profile parameters.
type HdrPersistenceProfile struct {
	// PrstHdrName is the header name for custom header persistence.
	// +kubebuilder:validation:MaxLength=128
	PrstHdrName string `json:"prst_hdr_name,omitempty"`
}

// AppCookiePersistenceProfile specifies the Application Cookie Persistence profile parameters.
type AppCookiePersistenceProfile struct {
	// PrstHdrName is the header or cookie name for application cookie persistence.
	// +kubebuilder:validation:MaxLength=128
	PrstHdrName string `json:"prst_hdr_name,omitempty"`

	// Timeout is the length of time after a client's connections have closed before expiring the client's persistence to a server.
	// +kubebuilder:default:=20
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=720
	Timeout int32 `json:"timeout,omitempty"`

	// EncryptionKey is the key to use for cookie encryption.
	// +kubebuilder:validation:MaxLength=1024
	EncryptionKey string `json:"encryption_key,omitempty"`
}

// HTTPCookiePersistenceProfile specifies the HTTP Cookie Persistence profile parameters.
type HTTPCookiePersistenceProfile struct {
	// CookieName is the HTTP cookie name for cookie persistence.
	// +kubebuilder:validation:MaxLength=128
	CookieName string `json:"cookie_name,omitempty"`

	// Timeout is the maximum lifetime of any session cookie.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=14400
	Timeout int32 `json:"timeout,omitempty"`

	// AlwaysSendCookie indicates if a persistence cookie should always be sent.
	// +kubebuilder:default:=false
	AlwaysSendCookie bool `json:"always_send_cookie,omitempty"`

	// HTTPOnly sets the HttpOnly attribute in the cookie.
	// +kubebuilder:default:=false
	HTTPOnly bool `json:"http_only,omitempty"`

	// IsPersistentCookie indicates if the cookie is a persistent cookie.
	// +kubebuilder:default:=false
	IsPersistentCookie bool `json:"is_persistent_cookie,omitempty"`
}

// PersistenceProfileStatus defines the observed state of PersistenceProfile
type PersistenceProfileStatus struct {
	// Status of the PersistenceProfile
	Status string `json:"status,omitempty"`
	// Error if any error was encountered
	Error string `json:"error"`
	// UUID is unique identifier of the persistent profile object
	UUID string `json:"uuid"`
	// ObservedGeneration is the observed generation by the operator
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// LastUpdated is the timestamp the object was last updated
	LastUpdated *metav1.Time `json:"lastUpdated"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=persistenceprofiles,scope=Namespaced
// +kubebuilder:subresource:status

// PersistenceProfile is the Schema for the persistenceprofiles API
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=persistenceprofiles,shortName=pp,singular=persistenceprofile,scope=Namespaced
type PersistenceProfile struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of PersistenceProfile
	Spec PersistenceProfileSpec `json:"spec,omitempty"`

	// Status defines the observed state of PersistenceProfile
	// +optional
	Status PersistenceProfileStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// PersistenceProfileList contains a list of PersistenceProfile
type PersistenceProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PersistenceProfile `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PersistenceProfile{}, &PersistenceProfileList{})
}
