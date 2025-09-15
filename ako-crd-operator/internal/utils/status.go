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

package utils

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/constants"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

// StatusObject defines the interface that CRD objects must implement to use the generic SetStatus function
type StatusObject interface {
	client.Object
	GetGeneration() int64
	SetController(controller string)
	GetConditions() []metav1.Condition
	SetConditions(conditions []metav1.Condition)
	SetLastUpdated(lastUpdated *metav1.Time)
	SetObservedGeneration(generation int64)
}

// StatusManager provides generic status management functionality for CRD objects
type StatusManager struct {
	Client        client.Client
	EventRecorder record.EventRecorder
}

// SetStatus sets the status, conditions, and events for a CRD object based on condition type, reason, and error message
func (sm *StatusManager) SetStatus(ctx context.Context, obj StatusObject, conditionType string, reason string, message string) error {
	obj.SetController(constants.AKOCRDController)

	// Set additional status fields
	lastUpdated := metav1.Time{Time: time.Now().UTC()}
	obj.SetLastUpdated(&lastUpdated)
	obj.SetObservedGeneration(obj.GetGeneration())

	// Determine condition status based on reason
	var conditionStatus metav1.ConditionStatus
	var eventType string
	var eventReason string
	var eventMessage string

	// Map reasons to condition values
	switch reason {
	case "ValidationSucceeded", "Created", "Updated", "Deleted":
		conditionStatus = metav1.ConditionTrue
		eventType = corev1.EventTypeNormal
		eventReason = reason
		if message == "" {
			eventMessage = fmt.Sprintf("Object %s successfully", reason)
		} else {
			eventMessage = message
		}
	case "ValidationFailed", "CreationFailed", "UpdateFailed", "UUIDExtractionFailed", "DeletionFailed", "DeletionSkipped":
		conditionStatus = metav1.ConditionFalse
		eventType = corev1.EventTypeWarning
		eventReason = reason
		if message != "" {
			eventMessage = message
		} else {
			eventMessage = fmt.Sprintf("Object %s", reason)
		}
	default:
		// Default to failed state for unknown reasons
		conditionStatus = metav1.ConditionFalse
		eventType = corev1.EventTypeWarning
		eventReason = reason
		eventMessage = message
	}

	// Set condition
	conditions := SetCondition(obj.GetConditions(), metav1.Condition{
		Type:               conditionType,
		Status:             conditionStatus,
		LastTransitionTime: metav1.Time{Time: time.Now().UTC()},
		Reason:             reason,
		Message:            eventMessage,
	})
	obj.SetConditions(conditions)

	// Record event
	sm.EventRecorder.Event(obj, eventType, eventReason, eventMessage)

	if sm.Client == nil {
		log := utils.LoggerFromContext(ctx)
		log.Errorf("Client is nil. Cannot update status for object: %s/%s", obj.GetNamespace(), obj.GetName())
		return fmt.Errorf("status client is nil")
	}
	err := sm.Client.Status().Update(ctx, obj)
	if err != nil {
		sm.EventRecorder.Event(obj, corev1.EventTypeWarning, "StatusUpdateFailed", fmt.Sprintf("Failed to update object status: %v", err))
	}
	return err
}
