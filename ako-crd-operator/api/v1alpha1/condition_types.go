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

// ObjectConditionType is a generic type of condition that can be associated with any resource.
type ObjectConditionType string

// ObjectConditionReason defines the set of reasons that explain why a
// particular condition type has been raised.
type ObjectConditionReason string

const (
	// Generic Condition Types - can be used by any resource type

	// This condition indicates whether a resource is ready and has been
	// successfully processed by the controller. It is a positive-polarity summary
	// condition, and so should always be present on the resource with
	// ObservedGeneration set.
	
	// Possible reasons for this condition to be True are:
	//
	// * "ValidationSucceeded"
	// * "Created"
	// * "Updated"
	//
	// Possible reasons for this condition to be False are:
	//
	// * "ValidationFailed"
	// * "CreationFailed"
	// * "UpdateFailed"
	// * "UUIDExtractionFailed"
	//
	// Controllers may raise this condition with other reasons,
	// but should prefer to use the reasons listed above to improve
	// interoperability.
	ObjectConditionReady ObjectConditionType = "Ready"

	// This condition indicates whether a resource has been successfully
	// deleted from the underlying Avi Controller. This condition is typically
	// present during the deletion lifecycle of the resource.
	//
	// Possible reasons for this condition to be True are:
	//
	// * "Deleted"
	//
	// Possible reasons for this condition to be False are:
	//
	// * "DeletionFailed"
	// * "DeletionSkipped"
	//
	// Controllers may raise this condition with other reasons,
	// but should prefer to use the reasons listed above to improve
	// interoperability.
	ObjectConditionDeleted ObjectConditionType = "Deleted"

	// This condition indicates whether a resource has been accepted
	// for processing. It reflects the validation state of the resource.
	//
	// Possible reasons for this condition to be True are:
	//
	// * "Accepted"
	// * "ValidationSucceeded"
	//
	// Possible reasons for this condition to be False are:
	//
	// * "Rejected"
	// * "ValidationFailed"
	// * "NotFound"
	// * "NotReady"
	// * "TenantMismatch"
	//
	// Controllers may raise this condition with other reasons,
	// but should prefer to use the reasons listed above to improve
	// interoperability.
	ObjectConditionAccepted ObjectConditionType = "Accepted"

	// This condition indicates whether a resource is Rejected
	// for processing. It reflects the validation state of the resource.
	// This condition must be set only for RBE CRD
	ObjectConditionRejected ObjectConditionType = "Rejected"
)

const (
	// Generic Condition Reasons - can be used by any resource type

	// Success reasons
	ObjectReasonValidationSucceeded ObjectConditionReason = "ValidationSucceeded"
	ObjectReasonCreated             ObjectConditionReason = "Created"
	ObjectReasonUpdated             ObjectConditionReason = "Updated"
	ObjectReasonDeleted             ObjectConditionReason = "Deleted"
	ObjectReasonAccepted            ObjectConditionReason = "Accepted"

	// Failure reasons
	ObjectReasonValidationFailed     ObjectConditionReason = "ValidationFailed"
	ObjectReasonCreationFailed       ObjectConditionReason = "CreationFailed"
	ObjectReasonUpdateFailed         ObjectConditionReason = "UpdateFailed"
	ObjectReasonDeletionFailed       ObjectConditionReason = "DeletionFailed"
	ObjectReasonDeletionSkipped      ObjectConditionReason = "DeletionSkipped"
	ObjectReasonUUIDExtractionFailed ObjectConditionReason = "UUIDExtractionFailed"
	ObjectReasonRejected             ObjectConditionReason = "Rejected"
	ObjectReasonNotFound             ObjectConditionReason = "NotFound"
	ObjectReasonNotReady             ObjectConditionReason = "NotReady"
	ObjectReasonTenantMismatch       ObjectConditionReason = "TenantMismatch"
)
