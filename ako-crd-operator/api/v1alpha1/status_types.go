/*
Copyright 2025 VMware, Inc.
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

// ObjectStatus defines the type for object status conditions
// Currently used only for RouteBackendExtension CRD
type ObjectStatus string

const (
	// ObjectStatusAccepted indicates that a CRD resource has been successfully
	// validated and accepted for processing. This status is set when:
	// - All referenced objects exist and are accessible
	// - All validation checks have passed successfully
	// - Resource configuration is valid and ready for processing
	// - Dependencies are satisfied and in the expected state
	// Currently used only for RouteBackendExtension CRD
	ObjectStatusAccepted ObjectStatus = "Accepted"

	// ObjectStatusRejected indicates that a CRD resource has been rejected
	// due to validation failures. This status is set when:
	// - Referenced objects are not found or inaccessible
	// - Validation checks fail due to invalid configuration
	// - Dependencies are not satisfied or in an invalid state
	// - Any other validation error occurs during processing
	// When rejected, the resource will not be processed until the underlying issues are resolved.
	// Currently used only for RouteBackendExtension CRD
	ObjectStatusRejected ObjectStatus = "Rejected"
)
