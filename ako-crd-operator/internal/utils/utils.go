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
	"regexp"
	"strings"
	"time"

	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/constants"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/errors"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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
	if err == nil {
		return false
	}
	if akoCrdOperatorError, ok := err.(errors.AKOCRDOperatorError); ok {
		switch akoCrdOperatorError.HttpStatusCode {
		case 400:
			return false
		}
	}
	if aviError, ok := err.(session.AviError); ok {
		switch aviError.HttpStatusCode {
		case 400, 401, 403, 404, 409, 412, 422, 501:
			return false
		default:
			// For 5xx errors and other transient issues, retry
			return true
		}
	}

	if strings.Contains(err.Error(), constants.NoObject) {
		// non retryable
		return false
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
		LastTransitionTime: metav1.Time{Time: time.Now().UTC()},
		Reason:             "ConfigurationError",
		Message:            fmt.Sprintf("Non-retryable error: %s", err.Error()),
	}
	// check if ako-crd-operator error
	if akoCrdOperatorError, ok := err.(errors.AKOCRDOperatorError); ok {
		switch akoCrdOperatorError.HttpStatusCode {
		case 400:
			condition.Reason = akoCrdOperatorError.Reason
			condition.Message = akoCrdOperatorError.Message
		}
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
	resource.SetObservedGeneration(resource.GetGeneration())
	resource.SetLastUpdated(&metav1.Time{Time: time.Now().UTC()})

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

// createMarkers creates markers for the health monitor with cluster name and namespace
func CreateMarkers(clusterName, namespace string) []*models.RoleFilterMatchLabel {
	markers := []*models.RoleFilterMatchLabel{}

	// Add cluster name marker

	if clusterName != "" {
		clusterNameKey := "clustername"
		clusterMarker := &models.RoleFilterMatchLabel{
			Key:    &clusterNameKey,
			Values: []string{clusterName},
		}
		markers = append(markers, clusterMarker)
	}

	// Add namespace marker

	if namespace != "" {
		namespaceKey := "namespace"
		namespaceMarker := &models.RoleFilterMatchLabel{
			Key:    &namespaceKey,
			Values: []string{namespace},
		}
		markers = append(markers, namespaceMarker)
	}

	return markers
}

// ParseAviErrorMessage parses the error message from ALB SDK
// ALB SDK returns error message in string format: `map[... error: ...]`
func ParseAviErrorMessage(input string) string {
	re := regexp.MustCompile(`map\[.*error:(.*?)(?:\s+obj_name:.*?)?\]$`)
	match := re.FindStringSubmatch(input)
	if len(match) >= 2 {
		errorMsg := strings.TrimSpace(match[1])
		return errorMsg
	}
	return input
}

// NamespaceHandler is an interface that must be implemented by reconcilers using the generic namespace handler
type NamespaceHandler interface {
	List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error
	GetLogger() *utils.AviLogger
}

// ExtractItems is a function type that extracts client.Object items from a resource list
type ExtractItems[T client.ObjectList] func(T) []client.Object

// CreateGenericNamespaceHandler creates a generic namespace handler that can work with any resource type
// It takes the resource type name for logging, a function that creates a new empty resource list,
// and a function that extracts the items from the list
func CreateGenericNamespaceHandler[T client.ObjectList](
	handler NamespaceHandler,
	resourceTypeName string,
	newList func() T,
	extractItems ExtractItems[T],
) func(context.Context, client.Object) []reconcile.Request {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		ns := obj.(*corev1.Namespace)
		log := handler.GetLogger().WithValues("namespace", ns.Name)
		log.Info("Processing namespace update")

		list := newList()
		err := handler.List(ctx, list, &client.ListOptions{
			Namespace: ns.Name,
		})
		if err != nil {
			log.Errorf("failed to list %ss for namespace %s: %v", resourceTypeName, ns.Name, err)
			return []reconcile.Request{}
		}

		items := extractItems(list)
		requests := make([]reconcile.Request, 0, len(items))
		for _, item := range items {
			log.Debugf("enqueuing reconcile request for %s %s/%s", resourceTypeName, item.GetNamespace(), item.GetName())
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      item.GetName(),
					Namespace: item.GetNamespace(),
				},
			})
		}
		return requests
	}
}

// TenantAnnotationNamespacePredicate creates a generic namespace predicate that triggers reconciliation
// when tenant annotations are added, removed, or changed on namespaces
func TenantAnnotationNamespacePredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			if e.ObjectOld == nil || e.ObjectNew == nil {
				return false
			}
			oldNs, ok := e.ObjectOld.(*corev1.Namespace)
			if !ok {
				return false
			}
			newNs, ok := e.ObjectNew.(*corev1.Namespace)
			if !ok {
				return false
			}

			// Get tenant annotation values from both old and new namespaces
			oldTenant, oldHasTenant := oldNs.Annotations[lib.TenantAnnotation]
			newTenant, newHasTenant := newNs.Annotations[lib.TenantAnnotation]

			// Scenario 1: Tenant annotation was removed (old had it, new doesn't)
			if oldHasTenant && !newHasTenant {
				return true
			}

			// Scenario 2: Tenant annotation was added (old didn't have it, new does)
			if !oldHasTenant && newHasTenant {
				return true
			}

			// Scenario 3: Tenant annotation value was changed (both have it but different values)
			if oldHasTenant && newHasTenant && oldTenant != newTenant {
				return true
			}

			return false
		},
	}
}

// GetTenantInNamespace is a generic function that retrieves the tenant annotation from a namespace
// using the Kubernetes client. This follows the same pattern as the existing controller implementations.
func GetTenantInNamespace(ctx context.Context, client client.Client, namespace string) (string, error) {
	nsObj := corev1.Namespace{}
	if err := client.Get(ctx, types.NamespacedName{Name: namespace}, &nsObj); err != nil {
		return "", err
	}
	tenant, ok := nsObj.Annotations[lib.TenantAnnotation]
	if !ok || tenant == "" {
		return lib.GetTenant(), nil
	}
	return tenant, nil
}
