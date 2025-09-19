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

	// This condition indicates whether a resource is programmed and has been
	// successfully processed by the controller.

	// Possible reasons for this condition to be True are:
	//
	// * "Created"
	// * "Updated"
	//
	// Possible reasons for this condition to be False are:
	//
	// * "CreationFailed"
	// * "UpdateFailed"
	// * "UUIDExtractionFailed"
	// * "DeletionFailed"
	// * "DeletionSkipped"
	//
	// Controllers may raise this condition with other reasons,
	// but should prefer to use the reasons listed above to improve
	// interoperability.
	ObjectConditionProgrammed ObjectConditionType = "Programmed"
)

const (
	// Success reasons
	ObjectReasonCreated ObjectConditionReason = "Created"
	ObjectReasonUpdated ObjectConditionReason = "Updated"

	// Failure reasons
	ObjectReasonCreationFailed       ObjectConditionReason = "CreationFailed"
	ObjectReasonUpdateFailed         ObjectConditionReason = "UpdateFailed"
	ObjectReasonDeletionFailed       ObjectConditionReason = "DeletionFailed"
	ObjectReasonDeletionSkipped      ObjectConditionReason = "DeletionSkipped"
	ObjectReasonUUIDExtractionFailed ObjectConditionReason = "UUIDExtractionFailed"
)
