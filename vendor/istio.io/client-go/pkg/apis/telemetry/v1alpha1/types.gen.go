// Copyright Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by kubetype-gen. DO NOT EDIT.

package v1alpha1

import (
	metav1alpha1 "istio.io/api/meta/v1alpha1"
	telemetryv1alpha1 "istio.io/api/telemetry/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// <!-- crd generation tags
// +cue-gen:Telemetry:groupName:telemetry.istio.io
// +cue-gen:Telemetry:versions:v1alpha1,v1
// +cue-gen:Telemetry:storageVersion
// +cue-gen:Telemetry:annotations:helm.sh/resource-policy=keep
// +cue-gen:Telemetry:labels:app=istio-pilot,chart=istio,istio=telemetry,heritage=Tiller,release=istio
// +cue-gen:Telemetry:subresource:status
// +cue-gen:Telemetry:scope:Namespaced
// +cue-gen:Telemetry:resource:categories=istio-io,telemetry-istio-io,shortNames=telemetry,plural=telemetries
// +cue-gen:Telemetry:preserveUnknownFields:false
// +cue-gen:Telemetry:printerColumn:name=Age,type=date,JSONPath=.metadata.creationTimestamp,description="CreationTimestamp
// is a timestamp representing the server time when this object was created. It
// is not guaranteed to be set in happens-before order across separate
// operations. Clients may not set this value. It is represented in RFC3339 form
// and is in UTC. Populated by the system. Read-only. Null for lists. More info:
// https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
// -->
//
// <!-- go code generation tags
// +kubetype-gen
// +kubetype-gen:groupVersion=telemetry.istio.io/v1alpha1
// +genclient
// +k8s:deepcopy-gen=true
// -->
// +kubebuilder:validation:XValidation:message="only one of targetRefs or selector can be set",rule="oneof(self.selector, self.targetRef, self.targetRefs)"
type Telemetry struct {
	v1.TypeMeta `json:",inline"`
	// +optional
	v1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec defines the implementation of this definition.
	// +optional
	Spec telemetryv1alpha1.Telemetry `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	Status metav1alpha1.IstioStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TelemetryList is a collection of Telemetries.
type TelemetryList struct {
	v1.TypeMeta `json:",inline"`
	// +optional
	v1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items       []*Telemetry `json:"items" protobuf:"bytes,2,rep,name=items"`
}
