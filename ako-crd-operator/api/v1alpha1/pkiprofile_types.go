/*


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
	v1 "sigs.k8s.io/gateway-api/apis/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.


// PKIProfileSpec defines the desired state of PKIProfile
type PKIProfileSpec struct {
	// CaCerts is a list of Certificate Authorities (Root and Intermediate) trusted that is used for certificate validation.
	CACertificateRefs []v1.ObjectReference `json:"caCertificateRefs,omitempty"`
}

// PKIProfileStatus defines the observed state of PKIProfile
type PKIProfileStatus struct {
	// Conditions represent the latest available observations of the PKIProfile's current state
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// UUID is unique identifier of the pki profile object
	UUID string `json:"uuid"`
	// ObservedGeneration is the observed generation by the operator
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// +optional
	// LastUpdated is the timestamp the object was last updated
	LastUpdated *metav1.Time `json:"lastUpdated"`
	// BackendObjectName is the name of the backend object
	BackendObjectName string `json:"backendObjectName,omitempty"`
	// Tenant is the tenant where the application profile is created
	Tenant string `json:"tenant,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Name",type="string",JSONPath=".spec.name",description="Name of the PKI Profile"
// +kubebuilder:printcolumn:name="CRL Check",type="boolean",JSONPath=".spec.crlCheck",description="CRL Check Enabled"
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="Current phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// PKIProfile is the Schema for the pkiprofiles API
type PKIProfile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PKIProfileSpec   `json:"spec,omitempty"`
	Status PKIProfileStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PKIProfileList contains a list of PKIProfile
type PKIProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PKIProfile `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PKIProfile{}, &PKIProfileList{})
}
