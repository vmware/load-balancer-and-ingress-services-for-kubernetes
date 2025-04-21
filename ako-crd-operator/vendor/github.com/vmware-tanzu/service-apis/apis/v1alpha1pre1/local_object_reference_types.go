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

package v1alpha1pre1

// ServicesDefaultLocalObjectReference identifies an API object within a
// known namespace that defaults group to core and resource to services
// if unspecified.
type ServicesDefaultLocalObjectReference struct {
	// Group is the group of the referent.  Omitting the value or specifying
	// the empty string indicates the core API group.  For example, use the
	// following to specify a service:
	//
	// fooRef:
	//   resource: services
	//   name: myservice
	//
	// Otherwise, if the core API group is not desired, specify the desired
	// group:
	//
	// fooRef:
	//   group: acme.io
	//   resource: foos
	//   name: myfoo
	//
	// +optional
	// +kubebuilder:default=core
	Group string `json:"group" protobuf:"bytes,1,opt,name=group"`
	// Resource is the API resource name of the referent. Omitting the value
	// or specifying the empty string indicates the services resource. For example,
	// use the following to specify a services resource:
	//
	// fooRef:
	//   name: myservice
	//
	// Otherwise, if the services resource is not desired, specify the desired
	// group:
	//
	// fooRef:
	//   group: acme.io
	//   resource: foos
	//   name: myfoo
	//
	// +optional
	// +kubebuilder:default=services
	Resource string `json:"resource" protobuf:"bytes,2,opt,name=resource"`
	// Name is the name of the referent.
	//
	// +kubebuilder:validation:Required
	Name string `json:"name" protobuf:"bytes,3,opt,name=name"`
}

// ConfigMapsDefaultLocalObjectReference identifies an API object within a
// known namespace that defaults group to core and resource to configmaps
// if unspecified.
type ConfigMapsDefaultLocalObjectReference struct {
	// Group is the group of the referent.  Omitting the value or specifying
	// the empty string indicates the core API group.  For example, use the
	// following to specify a configmaps:
	//
	// fooRef:
	//   resource: configmaps
	//   name: myconfigmap
	//
	// Otherwise, if the core API group is not desired, specify the desired
	// group:
	//
	// fooRef:
	//   group: acme.io
	//   resource: foos
	//   name: myfoo
	//
	// +optional
	// +kubebuilder:default=core
	Group string `json:"group" protobuf:"bytes,1,opt,name=group"`
	// Resource is the API resource name of the referent. Omitting the value
	// or specifying the empty string indicates the configmaps resource. For
	// example, use the following to specify a configmaps resource:
	//
	// fooRef:
	//   name: myconfigmap
	//
	// Otherwise, if the configmaps resource is not desired, specify the desired
	// group:
	//
	// fooRef:
	//   group: acme.io
	//   resource: foos
	//   name: myfoo
	//
	// +optional
	// +kubebuilder:default=configmaps
	Resource string `json:"resource" protobuf:"bytes,2,opt,name=resource"`
	// Name is the name of the referent.
	//
	// +kubebuilder:validation:Required
	Name string `json:"name" protobuf:"bytes,3,opt,name=name"`
}

// SecretsDefaultLocalObjectReference identifies an API object within a
// known namespace that defaults group to core and resource to secrets
// if unspecified.
type SecretsDefaultLocalObjectReference struct {
	// Group is the group of the referent.  Omitting the value or specifying
	// the empty string indicates the core API group.  For example, use the
	// following to specify a secrets resource:
	//
	// fooRef:
	//   resource: secrets
	//   name: mysecret
	//
	// Otherwise, if the core API group is not desired, specify the desired
	// group:
	//
	// fooRef:
	//   group: acme.io
	//   resource: foos
	//   name: myfoo
	//
	// +optional
	// +kubebuilder:default=core
	Group string `json:"group" protobuf:"bytes,1,opt,name=group"`
	// Resource is the API resource name of the referent. Omitting the value
	// or specifying the empty string indicates the secrets resource. For
	// example, use the following to specify a secrets resource:
	//
	// fooRef:
	//   name: mysecret
	//
	// Otherwise, if the secrets resource is not desired, specify the desired
	// group:
	//
	// fooRef:
	//   group: acme.io
	//   resource: foos
	//   name: myfoo
	//
	// +optional
	// +kubebuilder:default=secrets
	Resource string `json:"resource" protobuf:"bytes,2,opt,name=resource"`
	// Name is the name of the referent.
	//
	// +kubebuilder:validation:Required
	Name string `json:"name" protobuf:"bytes,3,opt,name=name"`
}
