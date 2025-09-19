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

	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/api/v1alpha1"
	crdlib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/lib"
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

// SetStatus sets the status, conditions, and events for a CRD object based on condition type, status, reason, and error message
func (sm *StatusManager) SetStatus(ctx context.Context, obj StatusObject, conditionType akov1alpha1.ObjectConditionType, conditionStatus metav1.ConditionStatus, reason akov1alpha1.ObjectConditionReason, message string) error {
	obj.SetController(crdlib.AKOCRDController)

	// Set additional status fields
	lastUpdated := metav1.Time{Time: time.Now().UTC()}
	obj.SetLastUpdated(&lastUpdated)
	obj.SetObservedGeneration(obj.GetGeneration())

	// Determine event type based on condition status
	var eventType string
	var eventReason string
	var eventMessage string

	switch conditionStatus {
	case metav1.ConditionTrue:
		eventType = corev1.EventTypeNormal
		eventReason = string(reason)
		if message == "" {
			eventMessage = fmt.Sprintf("Object %s successfully", reason)
		} else {
			eventMessage = message
		}
	case metav1.ConditionFalse:
		eventType = corev1.EventTypeWarning
		eventReason = string(reason)
		if message != "" {
			eventMessage = message
		} else {
			eventMessage = fmt.Sprintf("Object %s", reason)
		}
	default:
		// Default to warning for unknown condition status
		eventType = corev1.EventTypeWarning
		eventReason = string(reason)
		eventMessage = message
	}

	// Set condition
	conditions := SetCondition(obj.GetConditions(), metav1.Condition{
		Type:               string(conditionType),
		Status:             conditionStatus,
		LastTransitionTime: metav1.Time{Time: time.Now().UTC()},
		Reason:             string(reason),
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
