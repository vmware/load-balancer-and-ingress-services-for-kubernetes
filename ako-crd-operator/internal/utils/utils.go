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

package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/vmware/alb-sdk/go/session"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// StatusUpdater interface defines the methods needed to update status
type StatusUpdater interface {
	Status() client.SubResourceWriter
}

// ResourceWithStatus interface defines the common structure for resources with status
type ResourceWithStatus interface {
	client.Object
	GetGeneration() int64
	SetConditions([]metav1.Condition)
	GetConditions() []metav1.Condition
	SetObservedGeneration(int64)
	SetLastUpdated(*metav1.Time)
}

// IsRetryableError determines if an error from Avi Controller should be retried
func IsRetryableError(err error) bool {
	if aviError, ok := err.(session.AviError); ok {
		switch aviError.HttpStatusCode {
		case 400, 401, 403, 404, 409, 412, 422, 501: // Client errors and non-retryable conditions
			// 400: Bad Request - client error, don't retry
			// 401: Unauthorized - authentication issue, don't retry
			// 403: Forbidden - permission issue, don't retry
			// 404: Not Found - for critical dependencies, don't retry (Note: 404 during DELETE is actually success, but that's handled separately)
			// 409: Conflict - resource conflict, don't retry
			// 412: Precondition Failed - don't retry the same operation
			// 422: Unprocessable Entity - validation error, don't retry
			// 501: Not Implemented - feature not supported, don't retry
			return false
		default:
			// For 5xx errors and other transient issues, retry
			return true
		}
	}
	// For non-AviError types (network issues, timeouts, etc.), retry
	return true
}

// UpdateStatusWithNonRetryableError updates the resource status with failure condition
func UpdateStatusWithNonRetryableError(ctx context.Context, statusUpdater StatusUpdater, resource ResourceWithStatus, err error, resourceType string) {
	log := utils.LoggerFromContext(ctx)
	condition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionFalse,
		LastTransitionTime: metav1.Now(),
		Reason:             "ConfigurationError",
		Message:            fmt.Sprintf("Non-retryable error: %s", err.Error()),
	}

	// If it's an AviError, provide more specific information
	if aviError, ok := err.(session.AviError); ok {
		// Extract clean error message from Avi Controller response
		cleanErrorMsg := extractAviControllerErrorMessage(aviError)

		condition.Message = fmt.Sprintf("Avi Controller error (HTTP %d): %s", aviError.HttpStatusCode, cleanErrorMsg)
		switch aviError.HttpStatusCode {
		case 400:
			condition.Reason = "BadRequest"
			condition.Message = fmt.Sprintf("Invalid %s specification: %s", resourceType, cleanErrorMsg)
		case 401:
			condition.Reason = "Unauthorized"
			condition.Message = "Authentication failed with Avi Controller"
		case 403:
			condition.Reason = "Forbidden"
			condition.Message = fmt.Sprintf("Insufficient permissions to create/update %s on Avi Controller", resourceType)
		case 409:
			condition.Reason = "Conflict"
			condition.Message = fmt.Sprintf("Resource conflict on Avi Controller: %s", cleanErrorMsg)
		case 422:
			condition.Reason = "ValidationError"
			condition.Message = fmt.Sprintf("%s validation failed on Avi Controller: %s", resourceType, cleanErrorMsg)
		case 501:
			condition.Reason = "NotImplemented"
			condition.Message = fmt.Sprintf("%s feature not supported by Avi Controller version", resourceType)
		}
	}

	// Add or update the condition
	conditions := SetCondition(resource.GetConditions(), condition)
	resource.SetConditions(conditions)
	resource.SetLastUpdated(&metav1.Time{Time: time.Now().UTC()})
	resource.SetObservedGeneration(resource.GetGeneration())

	if err := statusUpdater.Status().Update(ctx, resource); err != nil {
		log.Errorf("Failed to update %s status with non-retryable error: %s", resourceType, err.Error())
	}
}

// extractAviControllerErrorMessage extracts the clean error message from Avi Controller response
func extractAviControllerErrorMessage(aviError session.AviError) string {
	errorStr := aviError.Error()
	// Look for other common error patterns in Avi responses
	if aviError.Message != nil && *aviError.Message != "" {
		return fmt.Sprintf("error from Controller: %s", *aviError.Message)
	}

	// Fallback to the original error if we can't parse it
	return errorStr
}

// SetCondition adds or updates a condition in the conditions slice
func SetCondition(conditions []metav1.Condition, newCondition metav1.Condition) []metav1.Condition {
	for i, condition := range conditions {
		if condition.Type == newCondition.Type {
			// Update existing condition
			conditions[i] = newCondition
			return conditions
		}
	}
	// Add new condition
	return append(conditions, newCondition)
}
