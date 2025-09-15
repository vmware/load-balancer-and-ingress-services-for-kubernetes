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
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SSLCertificate defines a certificate structure matching AVI SDK
type SSLCertificate struct {
	// Certificate is the PEM-encoded certificate data
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Certificate *string `json:"certificate"`
}

// PKIProfileSpec defines the desired state of PKIProfile
type PKIProfileSpec struct {
	// CaCerts is a list of Certificate Authorities (Root and Intermediate) trusted that is used for certificate validation.
	// Matches AVI SDK PKIProfile.CaCerts structure
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	CACerts []*SSLCertificate `json:"ca_certs,omitempty"`
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
	// Field is populated by AKO CRD operator as ako-crd-operator
	// +optional
	Controller string `json:"controller,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:path=pkiprofiles,shortName=pp,singular=pkiprofile,scope=Namespaced
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

func (pki *PKIProfile) SetPKIProfileController(controller string) {
	pki.Status.Controller = controller
}

// StatusObject interface implementation for generic status management

func (pki *PKIProfile) SetController(controller string) {
	pki.Status.Controller = controller
}

func (pki *PKIProfile) GetConditions() []metav1.Condition {
	return pki.Status.Conditions
}

func (pki *PKIProfile) SetConditions(conditions []metav1.Condition) {
	pki.Status.Conditions = conditions
}

func (pki *PKIProfile) SetLastUpdated(lastUpdated *metav1.Time) {
	pki.Status.LastUpdated = lastUpdated
}

func (pki *PKIProfile) SetObservedGeneration(generation int64) {
	pki.Status.ObservedGeneration = generation
}
