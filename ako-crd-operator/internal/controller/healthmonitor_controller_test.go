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

package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/alb-sdk/go/session"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/api/v1alpha1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/constants"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/test/mock"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestHealthMonitorController(t *testing.T) {
	tests := []struct {
		name         string
		hm           *akov1alpha1.HealthMonitor
		prepare      func(mockAviClient *mock.MockAviClientInterface)
		prepareCache func(cache *mock.MockCacheOperation)
		want         *akov1alpha1.HealthMonitor
		wantErr      bool
	}{
		{
			name: "success: add finalizer",
			hm: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
			want: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
				},
			},
			wantErr: false,
		},
		{
			name: "success: add healthmonitor",
			hm: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
					Namespace:       "default",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				responseUUID := map[string]interface{}{
					"uuid": "123",
				}
				mockAviClient.EXPECT().AviSessionPost(constants.HealthMonitorURL, gomock.Any(), gomock.Any(), gomock.Any()).Do(func(url string, request interface{}, response interface{}, params ...interface{}) {
					if resp, ok := response.(*map[string]interface{}); ok {
						*resp = responseUUID
					}
				}).Return(nil).AnyTimes()
			},
			want: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					Namespace:       "default",
					ResourceVersion: "1001",
				},
				Status: akov1alpha1.HealthMonitorStatus{
					UUID: "123",
					Conditions: []metav1.Condition{
						{
							Type:   "Ready",
							Status: metav1.ConditionTrue,
							// fake client isnt supporting time.UTC with nanoseconds precision
							LastTransitionTime: metav1.Time{Time: time.Now().Truncate(time.Second)},
							Reason:             "Created",
							Message:            "HealthMonitor created successfully on Avi Controller",
						},
					},
					BackendObjectName: "test-cluster-default-test",
					LastUpdated:       &metav1.Time{Time: time.Now().Truncate(time.Second)},
					Tenant:            "admin",
				},
			},
			wantErr: false,
		},
		{
			name: "success: add healthmonitor with existing healthmonitor",
			hm: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
					Namespace:       "default",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				responseBody := map[string]interface{}{
					"results": []interface{}{map[string]interface{}{"uuid": "123"}},
				}
				mockAviClient.EXPECT().AviSessionPost(constants.HealthMonitorURL, gomock.Any(), gomock.Any(), gomock.Any()).Return(session.AviError{
					HttpStatusCode: http.StatusConflict,
					AviResult: session.AviResult{
						Message: &[]string{"already exists"}[0],
					},
				})
				mockAviClient.EXPECT().AviSessionGet(fmt.Sprintf("%s?name=%s", constants.HealthMonitorURL, "test-cluster-default-test"), gomock.Any(), gomock.Any()).Do(func(url string, response interface{}, params interface{}) {
					if resp, ok := response.(*map[string]interface{}); ok {
						*resp = responseBody
					}
				}).Return(nil).AnyTimes()
				mockAviClient.EXPECT().AviSessionPut(constants.HealthMonitorURL+"/123", gomock.Any(), gomock.Any(), gomock.Any()).Do(func(url string, request interface{}, response interface{}, params ...interface{}) {
					if resp, ok := response.(*map[string]interface{}); ok {
						*resp = responseBody
					}
				}).Return(nil).AnyTimes()

			},
			want: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					Namespace:       "default",
					ResourceVersion: "1001",
				},
				Status: akov1alpha1.HealthMonitorStatus{
					UUID: "123",
					Conditions: []metav1.Condition{
						{
							Type:               "Ready",
							Status:             metav1.ConditionTrue,
							LastTransitionTime: metav1.Time{Time: time.Now().Truncate(time.Second)},
							Reason:             "Updated",
							Message:            "HealthMonitor updated successfully on Avi Controller",
						},
					},
					BackendObjectName: "test-cluster-default-test",
					LastUpdated:       &metav1.Time{Time: time.Now().Truncate(time.Second)},
					Tenant:            "admin",
				},
			},
			wantErr: false,
		},
		{
			name: "success: update healthmonitor",
			hm: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
					Namespace:       "default",
				},
				Status: akov1alpha1.HealthMonitorStatus{
					UUID: "123",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				mockAviClient.EXPECT().AviSessionPut(constants.HealthMonitorURL+"/123", gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			want: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					Namespace:       "default",
					ResourceVersion: "1001",
				},
				Status: akov1alpha1.HealthMonitorStatus{
					UUID: "123",
					Conditions: []metav1.Condition{
						{
							Type:               "Ready",
							Status:             metav1.ConditionTrue,
							LastTransitionTime: metav1.Time{Time: time.Now().Truncate(time.Second)},
							Reason:             "Updated",
							Message:            "HealthMonitor updated successfully on Avi Controller",
						},
					},
					BackendObjectName: "test-cluster-default-test",
					LastUpdated:       &metav1.Time{Time: time.Now().Truncate(time.Second)},
					Tenant:            "admin",
				},
			},
			wantErr: false,
		},
		{
			name: "success: update healthmonitor with no changes",
			hm: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					Name:            "test",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
					Namespace:       "default",
				},
				Status: akov1alpha1.HealthMonitorStatus{
					UUID:               "123",
					ObservedGeneration: 1,
					LastUpdated:        &metav1.Time{Time: time.Now().Truncate(time.Second)},
				},
			},
			prepareCache: func(cache *mock.MockCacheOperation) {
				dataMap := map[string]interface{}{
					"uuid":           "123",
					"_last_modified": fmt.Sprintf("%d", time.Now().Add(-2*time.Minute).UnixMicro()),
				}
				cache.EXPECT().GetObjectByUUID(gomock.Any(), "123").Return(dataMap, true).AnyTimes()
			},
			want: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					Name:            "test",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					Namespace:       "default",
					ResourceVersion: "1000",
				},
				Status: akov1alpha1.HealthMonitorStatus{
					UUID:               "123",
					ObservedGeneration: 1,
					LastUpdated:        &metav1.Time{Time: time.Now().Truncate(time.Second)},
				},
			},
			wantErr: false,
		},
		{
			name: "success: delete healthmonitor",
			hm: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test",
					Finalizers:        []string{"healthmonitor.ako.vmware.com/finalizer"},
					DeletionTimestamp: &metav1.Time{Time: time.Now().Truncate(time.Second)},
				},
				Status: akov1alpha1.HealthMonitorStatus{
					UUID: "123",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				mockAviClient.EXPECT().AviSessionDelete(constants.HealthMonitorURL+"/123", gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "success: delete healthmonitor with 404 error (treated as success)",
			hm: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test",
					Finalizers:        []string{"healthmonitor.ako.vmware.com/finalizer"},
					DeletionTimestamp: &metav1.Time{Time: time.Now().Truncate(time.Second)},
				},
				Status: akov1alpha1.HealthMonitorStatus{
					UUID: "123",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				mockAviClient.EXPECT().AviSessionDelete(constants.HealthMonitorURL+"/123", gomock.Any(), gomock.Any(), gomock.Any()).Return(session.AviError{
					HttpStatusCode: 404,
				}).AnyTimes()
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "success: delete healthmonitor with empty UUID",
			hm: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test",
					Finalizers:        []string{"healthmonitor.ako.vmware.com/finalizer"},
					DeletionTimestamp: &metav1.Time{Time: time.Now().Truncate(time.Second)},
				},
				Status: akov1alpha1.HealthMonitorStatus{
					UUID: "",
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "success: update healthmonitor with out-of-band changes (cache miss)",
			hm: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					Name:            "test",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
					Namespace:       "default",
				},
				Status: akov1alpha1.HealthMonitorStatus{
					UUID:               "123",
					ObservedGeneration: 1,
					LastUpdated:        &metav1.Time{Time: time.Now().Truncate(time.Second)},
				},
			},
			prepareCache: func(cache *mock.MockCacheOperation) {
				// Cache miss - object not found in cache
				cache.EXPECT().GetObjectByUUID(gomock.Any(), "123").Return(nil, false).AnyTimes()
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				mockAviClient.EXPECT().AviSessionPut(constants.HealthMonitorURL+"/123", gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			want: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					Name:            "test",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					Namespace:       "default",
					ResourceVersion: "1001",
				},
				Status: akov1alpha1.HealthMonitorStatus{
					UUID: "123",
					Conditions: []metav1.Condition{
						{
							Type:               "Ready",
							Status:             metav1.ConditionTrue,
							LastTransitionTime: metav1.Time{Time: time.Now().Truncate(time.Second)},
							Reason:             "Updated",
							Message:            "HealthMonitor updated successfully on Avi Controller",
						},
					},
					BackendObjectName:  "test-cluster-default-test",
					LastUpdated:        &metav1.Time{Time: time.Now().Truncate(time.Second)},
					ObservedGeneration: 1,
					Tenant:             "admin",
				},
			},
			wantErr: false,
		},
		{
			name: "success: update healthmonitor with nil LastUpdated",
			hm: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					Name:            "test",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
					Namespace:       "default",
				},
				Status: akov1alpha1.HealthMonitorStatus{
					UUID:               "123",
					ObservedGeneration: 1,
					LastUpdated:        nil, // Nil LastUpdated to test this path
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				mockAviClient.EXPECT().AviSessionPut(constants.HealthMonitorURL+"/123", gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			want: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					Name:            "test",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					Namespace:       "default",
					ResourceVersion: "1001",
				},
				Status: akov1alpha1.HealthMonitorStatus{
					UUID: "123",
					Conditions: []metav1.Condition{
						{
							Type:               "Ready",
							Status:             metav1.ConditionTrue,
							LastTransitionTime: metav1.Time{Time: time.Now().Truncate(time.Second)},
							Reason:             "Updated",
							Message:            "HealthMonitor updated successfully on Avi Controller",
						},
					},
					BackendObjectName:  "test-cluster-default-test",
					LastUpdated:        &metav1.Time{Time: time.Now().Truncate(time.Second)},
					ObservedGeneration: 1,
					Tenant:             "admin",
				},
			},
			wantErr: false,
		},
		// Negative Test Cases
		{
			name: "error: POST health monitor creation fails with non-conflict error",
			hm: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
					Namespace:       "default",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				mockAviClient.EXPECT().AviSessionPost(constants.HealthMonitorURL, gomock.Any(), gomock.Any(), gomock.Any()).Return(session.AviError{
					HttpStatusCode: 500,
					AviResult: session.AviResult{
						Message: &[]string{"internal server error"}[0],
					},
				}).AnyTimes()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error: POST conflict but GET fails during recovery",
			hm: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
					Namespace:       "default",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				mockAviClient.EXPECT().AviSessionPost(constants.HealthMonitorURL, gomock.Any(), gomock.Any(), gomock.Any()).Return(session.AviError{
					HttpStatusCode: http.StatusConflict,
					AviResult: session.AviResult{
						Message: &[]string{"already exists"}[0],
					},
				})
				mockAviClient.EXPECT().AviSessionGet(fmt.Sprintf("%s?name=%s", constants.HealthMonitorURL, "test-cluster-default-test"), gomock.Any(), gomock.Any()).Return(errors.New("GET failed")).AnyTimes()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error: POST conflict but extractUUID fails during recovery",
			hm: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
					Namespace:       "default",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				responseBody := map[string]interface{}{
					"results": []interface{}{}, // Empty results to cause extractUUID to fail
				}
				mockAviClient.EXPECT().AviSessionPost(constants.HealthMonitorURL, gomock.Any(), gomock.Any(), gomock.Any()).Return(session.AviError{
					HttpStatusCode: http.StatusConflict,
					AviResult: session.AviResult{
						Message: &[]string{"already exists"}[0],
					},
				})
				mockAviClient.EXPECT().AviSessionGet(fmt.Sprintf("%s?name=%s", constants.HealthMonitorURL, "test-cluster-default-test"), gomock.Any(), gomock.Any()).Do(func(url string, response interface{}, params interface{}) {
					if resp, ok := response.(*map[string]interface{}); ok {
						*resp = responseBody
					}
				}).Return(nil).AnyTimes()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error: POST conflict but PUT fails during recovery",
			hm: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
					Namespace:       "default",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				responseBody := map[string]interface{}{
					"results": []interface{}{map[string]interface{}{"uuid": "123"}},
				}
				mockAviClient.EXPECT().AviSessionPost(constants.HealthMonitorURL, gomock.Any(), gomock.Any(), gomock.Any()).Return(session.AviError{
					HttpStatusCode: http.StatusConflict,
					AviResult: session.AviResult{
						Message: &[]string{"already exists"}[0],
					},
				})
				mockAviClient.EXPECT().AviSessionGet(fmt.Sprintf("%s?name=%s", constants.HealthMonitorURL, "test-cluster-default-test"), gomock.Any(), gomock.Any()).Do(func(url string, response interface{}, params interface{}) {
					if resp, ok := response.(*map[string]interface{}); ok {
						*resp = responseBody
					}
				}).Return(nil).AnyTimes()
				mockAviClient.EXPECT().AviSessionPut(constants.HealthMonitorURL+"/123", gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("PUT failed")).AnyTimes()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error: extractUUID fails after successful POST",
			hm: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
					Namespace:       "default",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				responseUUID := map[string]interface{}{
					"invalid": "response", // Invalid response to cause extractUUID to fail
				}
				mockAviClient.EXPECT().AviSessionPost(constants.HealthMonitorURL, gomock.Any(), gomock.Any(), gomock.Any()).Do(func(url string, request interface{}, response interface{}, params ...interface{}) {
					if resp, ok := response.(*map[string]interface{}); ok {
						*resp = responseUUID
					}
				}).Return(nil).AnyTimes()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error: PUT fails during update",
			hm: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
					Namespace:       "default",
				},
				Status: akov1alpha1.HealthMonitorStatus{
					UUID: "123",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				mockAviClient.EXPECT().AviSessionPut(constants.HealthMonitorURL+"/123", gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("PUT failed")).AnyTimes()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error: delete fails with non-404 error",
			hm: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test",
					Finalizers:        []string{"healthmonitor.ako.vmware.com/finalizer"},
					DeletionTimestamp: &metav1.Time{Time: time.Now().Truncate(time.Second)},
				},
				Status: akov1alpha1.HealthMonitorStatus{
					UUID:   "123",
					Tenant: "admin",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				mockAviClient.EXPECT().AviSessionDelete(constants.HealthMonitorURL+"/123", gomock.Any(), gomock.Any(), gomock.Any()).Return(session.AviError{
					HttpStatusCode: 500,
					AviResult: session.AviResult{
						Message: &[]string{"server error"}[0],
					},
				}).AnyTimes()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error: delete fails with 403 error (referenced by other objects)",
			hm: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test",
					Finalizers:        []string{"healthmonitor.ako.vmware.com/finalizer"},
					DeletionTimestamp: &metav1.Time{Time: time.Now().Truncate(time.Second)},
					ResourceVersion:   "1000",
				},
				Status: akov1alpha1.HealthMonitorStatus{
					UUID:   "123",
					Tenant: "admin",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				mockAviClient.EXPECT().AviSessionDelete(constants.HealthMonitorURL+"/123", gomock.Any(), gomock.Any(), gomock.Any()).Return(session.AviError{
					HttpStatusCode: 403,
					AviResult: session.AviResult{
						Message: &[]string{"Cannot delete, object is referred by: ['Pool custom-pool', 'VirtualService custom-vs']"}[0],
					},
				}).AnyTimes()
			},
			want: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test",
					Finalizers:        []string{"healthmonitor.ako.vmware.com/finalizer"},
					DeletionTimestamp: &metav1.Time{Time: time.Now().Truncate(time.Second)},
					ResourceVersion:   "1001",
				},
				Status: akov1alpha1.HealthMonitorStatus{
					UUID: "123",
					Conditions: []metav1.Condition{
						{
							Type:               "Deleted",
							Status:             metav1.ConditionFalse,
							LastTransitionTime: metav1.Time{Time: time.Now().Truncate(time.Second)},
							Reason:             "DeletionSkipped",
							Message:            "Cannot delete, object is referred by: ['Pool custom-pool', 'VirtualService custom-vs']",
						},
					},
					Tenant: "admin",
				},
			},
			wantErr: false, // 403 doesn't cause requeue, it sets condition and waits
		},
		{
			name: "error: non-retryable error (status update without requeue)",
			hm: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
					Namespace:       "default",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				mockAviClient.EXPECT().AviSessionPost(constants.HealthMonitorURL, gomock.Any(), gomock.Any(), gomock.Any()).Return(session.AviError{
					HttpStatusCode: 400, // Non-retryable error
					AviResult: session.AviResult{
						Message: &[]string{"bad request"}[0],
					},
				}).AnyTimes()
			},
			want: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					Namespace:       "default",
					ResourceVersion: "1001",
				},
				Status: akov1alpha1.HealthMonitorStatus{
					Conditions: []metav1.Condition{
						{
							Type:               "Ready",
							Status:             metav1.ConditionFalse,
							LastTransitionTime: metav1.Time{Time: time.Now().Truncate(time.Second)},
							Reason:             "BadRequest",
							Message:            "Invalid HealthMonitor specification: error from Controller: bad request",
						},
					},
					LastUpdated:        &metav1.Time{Time: time.Now().Truncate(time.Second)},
					ObservedGeneration: 0,
				},
			},
			wantErr: false, // Non-retryable error doesn't cause requeue
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake k8s client
			scheme := runtime.NewScheme()
			_ = akov1alpha1.AddToScheme(scheme)
			_ = corev1.AddToScheme(scheme)

			// Create namespace object with tenant annotation
			namespace := createNamespaceWithTenant(tt.hm.Namespace)

			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(tt.hm, namespace).WithStatusSubresource(tt.hm).Build()

			// Create mock AVI client
			mockAviClient := mock.NewMockAviClientInterface(gomock.NewController(t))

			if tt.prepare != nil {
				tt.prepare(mockAviClient)
			}

			mockCache := mock.NewMockCacheOperation(gomock.NewController(t))
			if tt.prepareCache != nil {
				tt.prepareCache(mockCache)
			}

			// Create reconciler
			reconciler := &HealthMonitorReconciler{
				Client:        fakeClient,
				AviClient:     mockAviClient,
				Scheme:        scheme,
				Logger:        utils.AviLog,
				EventRecorder: record.NewFakeRecorder(10),
				ClusterName:   "test-cluster",
				Cache:         mockCache,
			}

			// Test reconcile
			ctx := context.Background()
			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      tt.hm.Name,
					Namespace: tt.hm.Namespace,
				},
			}

			_, err := reconciler.Reconcile(ctx, req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Check if the health monitor exists in the AVI controller
			hm := &akov1alpha1.HealthMonitor{}
			err = fakeClient.Get(ctx, req.NamespacedName, hm)
			if err != nil {
				if tt.want == nil {
					return
				}
				if tt.wantErr {
					return
				}
				t.Errorf("Failed to get health monitor: %v", err)
			}

			if tt.wantErr && tt.want == nil {
				return
			}
			assert.Equal(t, tt.want, hm)
		})
	}
}

// TestHealthMonitorControllerKubernetesError tests error scenarios in Reconcile function
func TestHealthMonitorControllerKubernetesError(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (*fake.ClientBuilder, ctrl.Request)
		prepare func(mockAviClient *mock.MockAviClientInterface)
		wantErr bool
	}{
		{
			name: "error: client.Get fails with NotFound error",
			setup: func() (*fake.ClientBuilder, ctrl.Request) {
				// Create a client without the resource to simulate a different kind of error
				scheme := runtime.NewScheme()
				_ = akov1alpha1.AddToScheme(scheme)
				_ = corev1.AddToScheme(scheme)

				// Create namespace object with tenant annotation
				namespace := createNamespaceWithTenant("default")

				// Create a fake client that will return an error for Get operations
				builder := fake.NewClientBuilder().WithScheme(scheme).WithObjects(namespace).WithInterceptorFuncs(interceptor.Funcs{
					Get: func(ctx context.Context, client client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
						// Allow namespace Get requests to succeed, but fail for HealthMonitor
						if obj.GetObjectKind().GroupVersionKind().Kind == "Namespace" {
							return client.Get(ctx, key, obj, opts...)
						}
						return k8serror.NewNotFound(akov1alpha1.GroupVersion.WithResource("healthmonitors").GroupResource(), "test")
					},
				})

				req := ctrl.Request{
					NamespacedName: types.NamespacedName{
						Name:      "nonexistent",
						Namespace: "default",
					},
				}
				return builder, req
			},
			wantErr: false, // NotFound errors don't cause requeue
		},
		{
			name: "error: client.Get fails with Internal error",
			setup: func() (*fake.ClientBuilder, ctrl.Request) {
				// Create a client without the resource to simulate a different kind of error
				scheme := runtime.NewScheme()
				_ = akov1alpha1.AddToScheme(scheme)
				_ = corev1.AddToScheme(scheme)

				// Create namespace object with tenant annotation
				namespace := createNamespaceWithTenant("default")

				builder := fake.NewClientBuilder().WithScheme(scheme).WithObjects(namespace).WithInterceptorFuncs(interceptor.Funcs{
					Get: func(ctx context.Context, client client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
						// Allow namespace Get requests to succeed, but fail for HealthMonitor
						if obj.GetObjectKind().GroupVersionKind().Kind == "Namespace" {
							return client.Get(ctx, key, obj, opts...)
						}
						return k8serror.NewInternalError(errors.New("internal server error"))
					},
				})

				req := ctrl.Request{
					NamespacedName: types.NamespacedName{
						Name:      "nonexistent",
						Namespace: "default",
					},
				}
				return builder, req
			},
			wantErr: true,
		},
		{
			name: "error: client.Update finalizer fails with Internal error",
			setup: func() (*fake.ClientBuilder, ctrl.Request) {
				scheme := runtime.NewScheme()
				_ = akov1alpha1.AddToScheme(scheme)
				_ = corev1.AddToScheme(scheme)
				hm := &akov1alpha1.HealthMonitor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
				}

				// Create namespace object with tenant annotation
				namespace := createNamespaceWithTenant("default")

				builder := fake.NewClientBuilder().WithScheme(scheme).WithObjects(hm, namespace).WithStatusSubresource(hm).WithInterceptorFuncs(interceptor.Funcs{
					Update: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
						return k8serror.NewInternalError(errors.New("internal server error"))
					},
				})
				req := ctrl.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test",
						Namespace: "default",
					},
				}
				return builder, req
			},
			wantErr: true,
		},
		{
			name: "error: client.Update object deletion remove finalizer fails with Internal error",
			setup: func() (*fake.ClientBuilder, ctrl.Request) {
				scheme := runtime.NewScheme()
				_ = akov1alpha1.AddToScheme(scheme)
				_ = corev1.AddToScheme(scheme)
				hm := &akov1alpha1.HealthMonitor{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "test",
						Namespace:         "default",
						DeletionTimestamp: &metav1.Time{Time: time.Now().Truncate(time.Second)},
						Finalizers:        []string{"healthmonitor.ako.vmware.com/finalizer"},
					},
				}

				// Create namespace object with tenant annotation
				namespace := createNamespaceWithTenant("default")

				builder := fake.NewClientBuilder().WithScheme(scheme).WithObjects(hm, namespace).WithStatusSubresource(hm).WithInterceptorFuncs(interceptor.Funcs{
					Update: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
						return k8serror.NewInternalError(errors.New("internal server error"))
					},
				})
				req := ctrl.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test",
						Namespace: "default",
					},
				}
				return builder, req
			},
			wantErr: true,
		},
		{
			name: "error: client.Update object updation fails with Internal error",
			setup: func() (*fake.ClientBuilder, ctrl.Request) {
				scheme := runtime.NewScheme()
				_ = akov1alpha1.AddToScheme(scheme)
				_ = corev1.AddToScheme(scheme)
				hm := &akov1alpha1.HealthMonitor{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "test",
						Namespace:  "default",
						Finalizers: []string{"healthmonitor.ako.vmware.com/finalizer"},
					},
					Status: akov1alpha1.HealthMonitorStatus{
						UUID: "123",
					},
				}

				// Create namespace object with tenant annotation
				namespace := createNamespaceWithTenant("default")

				builder := fake.NewClientBuilder().WithScheme(scheme).WithObjects(hm, namespace).WithStatusSubresource(hm).WithInterceptorFuncs(interceptor.Funcs{
					SubResourceUpdate: func(ctx context.Context, client client.Client, subResourceName string, obj client.Object, opts ...client.SubResourceUpdateOption) error {
						return k8serror.NewInternalError(errors.New("internal server error"))
					},
				})
				req := ctrl.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test",
						Namespace: "default",
					},
				}
				return builder, req
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				mockAviClient.EXPECT().AviSessionPut(constants.HealthMonitorURL+"/123", gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder, req := tt.setup()
			fakeClient := builder.Build()

			mockAviClient := mock.NewMockAviClientInterface(gomock.NewController(t))
			mockCache := mock.NewMockCacheOperation(gomock.NewController(t))
			if tt.prepare != nil {
				tt.prepare(mockAviClient)
			}
			reconciler := &HealthMonitorReconciler{
				Client:        fakeClient,
				AviClient:     mockAviClient,
				Scheme:        runtime.NewScheme(),
				Logger:        utils.AviLog,
				EventRecorder: record.NewFakeRecorder(10),
				ClusterName:   "test-cluster",
				Cache:         mockCache,
			}

			ctx := context.Background()
			_, err := reconciler.Reconcile(ctx, req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// @AI-Generated
// [Generated by Cursor claude-4-sonnet]
func TestExtractUUID(t *testing.T) {
	tests := []struct {
		name    string
		resp    map[string]interface{}
		want    string
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful extraction from results array",
			resp: map[string]interface{}{
				"results": []interface{}{
					map[string]interface{}{
						"uuid": "test-uuid-123",
						"name": "test-healthmonitor",
					},
				},
			},
			want:    "test-uuid-123",
			wantErr: false,
		},
		{
			name: "successful extraction from POST response",
			resp: map[string]interface{}{
				"uuid": "post-uuid-456",
				"name": "test-healthmonitor",
			},
			want:    "post-uuid-456",
			wantErr: false,
		},
		{
			name: "results not found and no uuid in root",
			resp: map[string]interface{}{
				"data": "some-data",
			},
			want:    "",
			wantErr: true,
			errMsg:  "'results' not found or not an array",
		},
		{
			name: "results is not an array",
			resp: map[string]interface{}{
				"results": "not-an-array",
			},
			want:    "",
			wantErr: true,
			errMsg:  "'results' not found or not an array",
		},
		{
			name: "results array is empty",
			resp: map[string]interface{}{
				"results": []interface{}{},
			},
			want:    "",
			wantErr: true,
			errMsg:  "'results' array is empty",
		},
		{
			name: "first element in results is not a map",
			resp: map[string]interface{}{
				"results": []interface{}{
					"not-a-map",
				},
			},
			want:    "",
			wantErr: true,
			errMsg:  "first element in 'results' is not a map",
		},
		{
			name: "uuid not found in first result",
			resp: map[string]interface{}{
				"results": []interface{}{
					map[string]interface{}{
						"name": "test-healthmonitor",
					},
				},
			},
			want:    "",
			wantErr: true,
			errMsg:  "'uuid' not found or not a string",
		},
		{
			name: "uuid is not a string in first result",
			resp: map[string]interface{}{
				"results": []interface{}{
					map[string]interface{}{
						"uuid": 12345,
						"name": "test-healthmonitor",
					},
				},
			},
			want:    "",
			wantErr: true,
			errMsg:  "'uuid' not found or not a string",
		},
		{
			name: "uuid is not a string in root (POST response fallback)",
			resp: map[string]interface{}{
				"uuid": 12345, // Non-string uuid in root level
				"name": "test-healthmonitor",
			},
			want:    "",
			wantErr: true,
			errMsg:  "'results' not found or not an array",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractUUID(tt.resp)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

// @AI-Generated
// [Generated by Cursor claude-4-sonnet]
func TestHealthMonitorControllerSecretEvent(t *testing.T) {
	tests := []struct {
		name         string
		hm           *akov1alpha1.HealthMonitor
		secret       *corev1.Secret
		prepare      func(mockAviClient *mock.MockAviClientInterface)
		want         *akov1alpha1.HealthMonitor
		wantErr      bool
		secretExists bool
	}{
		{
			name: "success: healthmonitor with valid secret reference",
			hm: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test-hm",
					Namespace:       "default",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
				},
				Spec: akov1alpha1.HealthMonitorSpec{
					Type: "HEALTH_MONITOR_HTTP",
					Authentication: &akov1alpha1.HealthMonitorInfo{
						SecretRef: "test-secret",
					},
				},
			},
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test-secret",
					Namespace:       "default",
					ResourceVersion: "2000",
				},
				Type: "ako.vmware.com/basic-auth",
				Data: map[string][]byte{
					"username": []byte("testuser"),
					"password": []byte("testpass"),
				},
			},
			secretExists: true,
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				// Capture and validate the request contains proper auth credentials
				mockAviClient.EXPECT().AviSessionPost(constants.HealthMonitorURL, gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(url string, request interface{}, response interface{}, params ...interface{}) {
						// Validate secret resolution
						if req, ok := request.(*HealthMonitorRequest); ok {
							assert.NotNil(t, req.AuthCredentials)
							assert.Equal(t, "testuser", req.AuthCredentials.Username)
							assert.Equal(t, "testpass", req.AuthCredentials.Password)
							assert.Nil(t, req.Authentication) // Should be nil after resolution
						}
						// Set response
						if resp, ok := response.(*map[string]interface{}); ok {
							*resp = map[string]interface{}{"uuid": "test-uuid-123"}
						}
					}).Return(nil).AnyTimes()
			},
			want: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test-hm",
					Namespace:       "default",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1001",
				},
				Spec: akov1alpha1.HealthMonitorSpec{
					Type: "HEALTH_MONITOR_HTTP",
					Authentication: &akov1alpha1.HealthMonitorInfo{
						SecretRef: "test-secret",
					},
				},
				Status: akov1alpha1.HealthMonitorStatus{
					UUID:               "test-uuid-123",
					BackendObjectName:  "test-cluster-default-test-hm",
					DependencySum:      utils.Hash("2000"), // ResourceVersion checksum
					Tenant:             "admin",
					LastUpdated:        &metav1.Time{Time: time.Now().Truncate(time.Second)},
					ObservedGeneration: 0,
					Conditions: []metav1.Condition{
						{
							Type:               "Ready",
							Status:             metav1.ConditionTrue,
							LastTransitionTime: metav1.Time{Time: time.Now().Truncate(time.Second)},
							Reason:             "Created",
							Message:            "HealthMonitor created successfully on Avi Controller",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "success: healthmonitor without secret reference",
			hm: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test-hm-no-secret",
					Namespace:       "default",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
				},
				Spec: akov1alpha1.HealthMonitorSpec{
					Type: "HEALTH_MONITOR_HTTP",
				},
			},
			secretExists: false,
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				// Validate no auth credentials when no secret
				mockAviClient.EXPECT().AviSessionPost(constants.HealthMonitorURL, gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(url string, request interface{}, response interface{}, params ...interface{}) {
						// Validate no secret resolution
						if req, ok := request.(*HealthMonitorRequest); ok {
							assert.Nil(t, req.AuthCredentials)
							assert.Nil(t, req.Authentication)
						}
						// Set response
						if resp, ok := response.(*map[string]interface{}); ok {
							*resp = map[string]interface{}{"uuid": "test-uuid-456"}
						}
					}).Return(nil).AnyTimes()
			},
			want: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test-hm-no-secret",
					Namespace:       "default",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1001",
				},
				Spec: akov1alpha1.HealthMonitorSpec{
					Type: "HEALTH_MONITOR_HTTP",
				},
				Status: akov1alpha1.HealthMonitorStatus{
					UUID:               "test-uuid-456",
					BackendObjectName:  "test-cluster-default-test-hm-no-secret",
					Tenant:             "admin",
					DependencySum:      0, // No dependencies
					LastUpdated:        &metav1.Time{Time: time.Now().Truncate(time.Second)},
					ObservedGeneration: 0,
					Conditions: []metav1.Condition{
						{
							Type:               "Ready",
							Status:             metav1.ConditionTrue,
							LastTransitionTime: metav1.Time{Time: time.Now().Truncate(time.Second)},
							Reason:             "Created",
							Message:            "HealthMonitor created successfully on Avi Controller",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "success: healthmonitor update when secret changes (different dependency checksum)",
			hm: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test-hm-update",
					Namespace:       "default",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
					Generation:      1,
				},
				Spec: akov1alpha1.HealthMonitorSpec{
					Type: "HEALTH_MONITOR_HTTP",
					Authentication: &akov1alpha1.HealthMonitorInfo{
						SecretRef: "test-secret-updated",
					},
				},
				Status: akov1alpha1.HealthMonitorStatus{
					UUID:               "existing-uuid",
					ObservedGeneration: 1,
					DependencySum:      utils.Hash("1000"), // Old checksum
				},
			},
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test-secret-updated",
					Namespace:       "default",
					ResourceVersion: "3000", // Different resource version
				},
				Type: "ako.vmware.com/basic-auth",
				Data: map[string][]byte{
					"username": []byte("newuser"),
					"password": []byte("newpass"),
				},
			},
			secretExists: true,
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				// Validate updated credentials in PUT request
				mockAviClient.EXPECT().AviSessionPut(constants.HealthMonitorURL+"/existing-uuid", gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(url string, request interface{}, response interface{}, params ...interface{}) {
						// Validate secret resolution with new credentials
						if req, ok := request.(*HealthMonitorRequest); ok {
							assert.NotNil(t, req.AuthCredentials)
							assert.Equal(t, "newuser", req.AuthCredentials.Username)
							assert.Equal(t, "newpass", req.AuthCredentials.Password)
						}
					}).Return(nil).AnyTimes()
			},
			want: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test-hm-update",
					Namespace:       "default",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1001",
					Generation:      1,
				},
				Spec: akov1alpha1.HealthMonitorSpec{
					Type: "HEALTH_MONITOR_HTTP",
					Authentication: &akov1alpha1.HealthMonitorInfo{
						SecretRef: "test-secret-updated",
					},
				},
				Status: akov1alpha1.HealthMonitorStatus{
					UUID:               "existing-uuid",
					BackendObjectName:  "test-cluster-default-test-hm-update",
					DependencySum:      utils.Hash("3000"), // Updated checksum
					Tenant:             "admin",
					LastUpdated:        &metav1.Time{Time: time.Now().Truncate(time.Second)},
					ObservedGeneration: 1,
					Conditions: []metav1.Condition{
						{
							Type:               "Ready",
							Status:             metav1.ConditionTrue,
							LastTransitionTime: metav1.Time{Time: time.Now().Truncate(time.Second)},
							Reason:             "Updated",
							Message:            "HealthMonitor updated successfully on Avi Controller",
						},
					},
				},
			},
			wantErr: false,
		},
		// Error Cases
		{
			name: "error: secret not found",
			hm: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test-hm-missing-secret",
					Namespace:       "default",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
				},
				Spec: akov1alpha1.HealthMonitorSpec{
					Type: "HEALTH_MONITOR_HTTP",
					Authentication: &akov1alpha1.HealthMonitorInfo{
						SecretRef: "missing-secret",
					},
				},
			},
			secretExists: false,
			want: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test-hm-missing-secret",
					Namespace:       "default",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1001",
				},
				Spec: akov1alpha1.HealthMonitorSpec{
					Type: "HEALTH_MONITOR_HTTP",
					Authentication: &akov1alpha1.HealthMonitorInfo{
						SecretRef: "missing-secret",
					},
				},
				Status: akov1alpha1.HealthMonitorStatus{
					DependencySum:      0,
					LastUpdated:        &metav1.Time{Time: time.Now().Truncate(time.Second)},
					ObservedGeneration: 0,
					Conditions: []metav1.Condition{
						{
							Type:               "Ready",
							Status:             metav1.ConditionFalse,
							LastTransitionTime: metav1.Time{Time: time.Now().Truncate(time.Second)},
							Reason:             "UnresolvedRef",
							Message:            "Secret missing-secret not found",
						},
					},
				},
			},
			wantErr: false, // Non-retryable error
		},
		{
			name: "error: secret wrong type",
			hm: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test-hm-wrong-type",
					Namespace:       "default",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
				},
				Spec: akov1alpha1.HealthMonitorSpec{
					Type: "HEALTH_MONITOR_HTTP",
					Authentication: &akov1alpha1.HealthMonitorInfo{
						SecretRef: "wrong-type-secret",
					},
				},
			},
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wrong-type-secret",
					Namespace: "default",
				},
				Type: "kubernetes.io/tls", // Wrong type
				Data: map[string][]byte{
					"username": []byte("testuser"),
					"password": []byte("testpass"),
				},
			},
			secretExists: true,
			want: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test-hm-wrong-type",
					Namespace:       "default",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1001",
				},
				Spec: akov1alpha1.HealthMonitorSpec{
					Type: "HEALTH_MONITOR_HTTP",
					Authentication: &akov1alpha1.HealthMonitorInfo{
						SecretRef: "wrong-type-secret",
					},
				},
				Status: akov1alpha1.HealthMonitorStatus{
					DependencySum:      0,
					LastUpdated:        &metav1.Time{Time: time.Now().Truncate(time.Second)},
					ObservedGeneration: 0,
					Conditions: []metav1.Condition{
						{
							Type:               "Ready",
							Status:             metav1.ConditionFalse,
							LastTransitionTime: metav1.Time{Time: time.Now().Truncate(time.Second)},
							Reason:             "ConfigurationError",
							Message:            "Secret wrong-type-secret is not of type ako.vmware.com/basic-auth",
						},
					},
				},
			},
			wantErr: false, // Non-retryable error
		},
		{
			name: "error: secret missing username",
			hm: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test-hm-no-username",
					Namespace:       "default",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
				},
				Spec: akov1alpha1.HealthMonitorSpec{
					Type: "HEALTH_MONITOR_HTTP",
					Authentication: &akov1alpha1.HealthMonitorInfo{
						SecretRef: "no-username-secret",
					},
				},
			},
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "no-username-secret",
					Namespace: "default",
				},
				Type: "ako.vmware.com/basic-auth",
				Data: map[string][]byte{
					"password": []byte("testpass"),
					// Missing username
				},
			},
			secretExists: true,
			want: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test-hm-no-username",
					Namespace:       "default",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1001",
				},
				Spec: akov1alpha1.HealthMonitorSpec{
					Type: "HEALTH_MONITOR_HTTP",
					Authentication: &akov1alpha1.HealthMonitorInfo{
						SecretRef: "no-username-secret",
					},
				},
				Status: akov1alpha1.HealthMonitorStatus{
					DependencySum:      0,
					LastUpdated:        &metav1.Time{Time: time.Now().Truncate(time.Second)},
					ObservedGeneration: 0,
					Conditions: []metav1.Condition{
						{
							Type:               "Ready",
							Status:             metav1.ConditionFalse,
							LastTransitionTime: metav1.Time{Time: time.Now().Truncate(time.Second)},
							Reason:             "ConfigurationError",
							Message:            "Username not found in secret no-username-secret",
						},
					},
				},
			},
			wantErr: false, // Non-retryable error
		},
		{
			name: "error: secret missing password",
			hm: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test-hm-no-password",
					Namespace:       "default",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
				},
				Spec: akov1alpha1.HealthMonitorSpec{
					Type: "HEALTH_MONITOR_HTTP",
					Authentication: &akov1alpha1.HealthMonitorInfo{
						SecretRef: "no-password-secret",
					},
				},
			},
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "no-password-secret",
					Namespace: "default",
				},
				Type: "ako.vmware.com/basic-auth",
				Data: map[string][]byte{
					"username": []byte("testuser"),
					// Missing password
				},
			},
			secretExists: true,
			want: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test-hm-no-password",
					Namespace:       "default",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1001",
				},
				Spec: akov1alpha1.HealthMonitorSpec{
					Type: "HEALTH_MONITOR_HTTP",
					Authentication: &akov1alpha1.HealthMonitorInfo{
						SecretRef: "no-password-secret",
					},
				},
				Status: akov1alpha1.HealthMonitorStatus{
					DependencySum:      0,
					LastUpdated:        &metav1.Time{Time: time.Now().Truncate(time.Second)},
					ObservedGeneration: 0,
					Conditions: []metav1.Condition{
						{
							Type:               "Ready",
							Status:             metav1.ConditionFalse,
							LastTransitionTime: metav1.Time{Time: time.Now().Truncate(time.Second)},
							Reason:             "ConfigurationError",
							Message:            "Password not found in secret no-password-secret",
						},
					},
				},
			},
			wantErr: false, // Non-retryable error
		},
		{
			name: "error: secret with nil data",
			hm: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test-hm-nil-data",
					Namespace:       "default",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
				},
				Spec: akov1alpha1.HealthMonitorSpec{
					Type: "HEALTH_MONITOR_HTTP",
					Authentication: &akov1alpha1.HealthMonitorInfo{
						SecretRef: "nil-data-secret",
					},
				},
			},
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nil-data-secret",
					Namespace: "default",
				},
				Type: "ako.vmware.com/basic-auth",
				Data: nil, // Nil data
			},
			secretExists: true,
			want: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test-hm-nil-data",
					Namespace:       "default",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1001",
				},
				Spec: akov1alpha1.HealthMonitorSpec{
					Type: "HEALTH_MONITOR_HTTP",
					Authentication: &akov1alpha1.HealthMonitorInfo{
						SecretRef: "nil-data-secret",
					},
				},
				Status: akov1alpha1.HealthMonitorStatus{
					DependencySum:      0,
					LastUpdated:        &metav1.Time{Time: time.Now().Truncate(time.Second)},
					ObservedGeneration: 0,
					Conditions: []metav1.Condition{
						{
							Type:               "Ready",
							Status:             metav1.ConditionFalse,
							LastTransitionTime: metav1.Time{Time: time.Now().Truncate(time.Second)},
							Reason:             "ConfigurationError",
							Message:            "Secret data is nil in secret nil-data-secret",
						},
					},
				},
			},
			wantErr: false, // Non-retryable error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake k8s client
			scheme := runtime.NewScheme()
			_ = akov1alpha1.AddToScheme(scheme)
			_ = corev1.AddToScheme(scheme)

			objects := []client.Object{tt.hm}
			if tt.secretExists && tt.secret != nil {
				objects = append(objects, tt.secret)
			}

			// Create namespace object with tenant annotation
			namespace := createNamespaceWithTenant(tt.hm.Namespace)
			objects = append(objects, namespace)

			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).WithStatusSubresource(tt.hm).Build()

			// Create mock AVI client
			mockAviClient := mock.NewMockAviClientInterface(gomock.NewController(t))
			if tt.prepare != nil {
				tt.prepare(mockAviClient)
			}

			mockCache := mock.NewMockCacheOperation(gomock.NewController(t))

			// Create reconciler
			reconciler := &HealthMonitorReconciler{
				Client:        fakeClient,
				AviClient:     mockAviClient,
				Scheme:        scheme,
				Logger:        utils.AviLog,
				EventRecorder: record.NewFakeRecorder(10),
				ClusterName:   "test-cluster",
				Cache:         mockCache,
			}

			// Test reconcile
			ctx := context.Background()
			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      tt.hm.Name,
					Namespace: tt.hm.Namespace,
				},
			}

			_, err := reconciler.Reconcile(ctx, req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Check final state
			hm := &akov1alpha1.HealthMonitor{}
			err = fakeClient.Get(ctx, req.NamespacedName, hm)
			if err != nil {
				if tt.want == nil {
					return
				}
				if tt.wantErr {
					return
				}
				t.Errorf("Failed to get health monitor: %v", err)
			}

			if tt.wantErr && tt.want == nil {
				return
			}
			assert.Equal(t, tt.want, hm)
		})
	}
}

// @AI-Generated
// [Generated by Cursor claude-4-sonnet]
func TestHealthMonitorControllerSetupWithManager(t *testing.T) {
	// Test the positive scenario where SetupWithManager successfully sets up:
	// 1. Field indexer for secret references
	// 2. Controller with HealthMonitor watching
	// 3. Secret watching with proper handler and predicate

	scheme := runtime.NewScheme()
	_ = akov1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	// Create a fake client for testing
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	// Create reconciler with fake client
	reconciler := &HealthMonitorReconciler{
		Client:        fakeClient,
		AviClient:     mock.NewMockAviClientInterface(gomock.NewController(t)),
		Scheme:        scheme,
		Logger:        utils.AviLog,
		EventRecorder: record.NewFakeRecorder(10),
		ClusterName:   "test-cluster",
		Cache:         mock.NewMockCacheOperation(gomock.NewController(t)),
	}

	// Test that the SetupWithManager method exists and can be called
	// Note: We can't fully test SetupWithManager without a real manager,
	// but we can test the indexer function separately
	t.Log("SetupWithManager method exists and can be tested with integration tests")

	// Test the field indexer function logic (using refactored function)
	// Test indexer with HealthMonitor that has secret reference
	hmWithSecret := &akov1alpha1.HealthMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hm-with-secret",
			Namespace: "default",
		},
		Spec: akov1alpha1.HealthMonitorSpec{
			Type: "HEALTH_MONITOR_HTTP",
			Authentication: &akov1alpha1.HealthMonitorInfo{
				SecretRef: "test-secret",
			},
		},
	}

	result := healthMonitorSecretRefIndexer(hmWithSecret)
	assert.Equal(t, []string{"test-secret"}, result)

	// Test indexer with HealthMonitor without authentication
	hmWithoutAuth := &akov1alpha1.HealthMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hm-without-auth",
			Namespace: "default",
		},
		Spec: akov1alpha1.HealthMonitorSpec{
			Type: "HEALTH_MONITOR_HTTP",
		},
	}

	result = healthMonitorSecretRefIndexer(hmWithoutAuth)
	assert.Nil(t, result)

	// Test indexer with HealthMonitor with empty secret reference
	hmWithEmptySecret := &akov1alpha1.HealthMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hm-empty-secret",
			Namespace: "default",
		},
		Spec: akov1alpha1.HealthMonitorSpec{
			Type: "HEALTH_MONITOR_HTTP",
			Authentication: &akov1alpha1.HealthMonitorInfo{
				SecretRef: "",
			},
		},
	}

	result = healthMonitorSecretRefIndexer(hmWithEmptySecret)
	assert.Nil(t, result)

	// Test the secret handler function logic with fake client
	namespace := createNamespaceWithTenant("default")
	secretHandlerClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(hmWithSecret, hmWithoutAuth, namespace).Build()
	reconciler.Client = secretHandlerClient

	// Create a custom handler for testing since we can't use field selector with fake client
	handlerFunc := func(ctx context.Context, obj client.Object) []reconcile.Request {
		secret := obj.(*corev1.Secret)
		hmList := &akov1alpha1.HealthMonitorList{}

		// Since we can't use field selector with fake client, we'll list all and filter manually
		err := reconciler.List(ctx, hmList, &client.ListOptions{
			Namespace: secret.Namespace,
		})
		if err != nil {
			reconciler.Logger.Errorf("failed to list HealthMonitors for secret %s/%s: %v", secret.Namespace, secret.Name, err)
			return []reconcile.Request{}
		}

		requests := make([]reconcile.Request, 0)
		for _, item := range hmList.Items {
			if item.Spec.Authentication != nil && item.Spec.Authentication.SecretRef == secret.Name {
				requests = append(requests, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      item.Name,
						Namespace: item.Namespace,
					},
				})
			}
		}
		return requests
	}

	// Test handler with a secret that is referenced
	testSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
		Type: constants.HealthMonitorSecretType,
	}

	ctx := context.Background()
	requests := handlerFunc(ctx, testSecret)
	assert.Equal(t, 1, len(requests))
	assert.Equal(t, "hm-with-secret", requests[0].Name)
	assert.Equal(t, "default", requests[0].Namespace)

	// Test handler with a secret that is not referenced
	unusedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "unused-secret",
			Namespace: "default",
		},
		Type: constants.HealthMonitorSecretType,
	}

	requests = handlerFunc(ctx, unusedSecret)
	assert.Equal(t, 0, len(requests))

	// Test the secret predicate function (using refactored function)
	// Test predicate with correct secret type
	assert.True(t, healthMonitorSecretPredicate(testSecret))

	// Test predicate with wrong secret type
	wrongTypeSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wrong-type",
			Namespace: "default",
		},
		Type: corev1.SecretTypeTLS,
	}
	assert.False(t, healthMonitorSecretPredicate(wrongTypeSecret))

	// Test predicate with non-secret object
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-configmap",
			Namespace: "default",
		},
	}
	assert.False(t, healthMonitorSecretPredicate(configMap))

	// Test predicate with nil object
	assert.False(t, healthMonitorSecretPredicate(nil))
}

// @AI-Generated
// [Generated by Cursor claude-4-sonnet]
func TestHealthMonitorControllerTenantChange(t *testing.T) {
	tests := []struct {
		name             string
		initialHM        *akov1alpha1.HealthMonitor
		initialNamespace *corev1.Namespace
		updatedNamespace *corev1.Namespace
		prepare          func(mockAviClient *mock.MockAviClientInterface)
		prepareCache     func(cache *mock.MockCacheOperation)
		expectedTenant   string
		expectDeletion   bool
		expectRecreation bool
		wantErr          bool
	}{
		{
			name: "success: tenant change triggers deletion and recreation",
			initialHM: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test-hm-tenant-change",
					Namespace:       "test-namespace",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
				},
				Spec: akov1alpha1.HealthMonitorSpec{
					Type: "HEALTH_MONITOR_HTTP",
				},
				Status: akov1alpha1.HealthMonitorStatus{
					UUID:   "existing-uuid-123",
					Tenant: "tenant-a", // Initially set to tenant-a
				},
			},
			initialNamespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-namespace",
					Annotations: map[string]string{
						lib.TenantAnnotation: "tenant-a", // Initially tenant-a
					},
				},
			},
			updatedNamespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-namespace",
					Annotations: map[string]string{
						lib.TenantAnnotation: "tenant-b", // Changed to tenant-b
					},
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				// Expect deletion call for old tenant
				mockAviClient.EXPECT().AviSessionDelete(
					gomock.Any(), // URL
					gomock.Any(), // request
					gomock.Any(), // response
					gomock.Any(), // Options - tenant-a
				).Return(nil).Times(1)

				// Expect creation call for new tenant
				mockAviClient.EXPECT().AviSessionPost(
					constants.HealthMonitorURL,
					gomock.Any(), // Request body
					gomock.Any(), // Response
					gomock.Any(), // Options - tenant-b
				).DoAndReturn(func(url string, req interface{}, resp interface{}, opts ...interface{}) error {
					// Simulate successful creation response
					respMap := resp.(*map[string]interface{})
					(*respMap)["uuid"] = "new-uuid-456"
					return nil
				}).Times(1)
			},
			prepareCache: func(cache *mock.MockCacheOperation) {
				// Expect cache operations
				cache.EXPECT().GetObjectByUUID(gomock.Any(), gomock.Any()).Return(nil, false).AnyTimes()
			},
			expectedTenant:   "tenant-b",
			expectDeletion:   true,
			expectRecreation: true,
			wantErr:          false,
		},
		{
			name: "success: no tenant change, no deletion",
			initialHM: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test-hm-no-change",
					Namespace:       "test-namespace",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
				},
				Spec: akov1alpha1.HealthMonitorSpec{
					Type: "HEALTH_MONITOR_HTTP",
				},
				Status: akov1alpha1.HealthMonitorStatus{
					UUID:   "existing-uuid-123",
					Tenant: "tenant-a", // Same tenant
				},
			},
			initialNamespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-namespace",
					Annotations: map[string]string{
						lib.TenantAnnotation: "tenant-a", // Same tenant
					},
				},
			},
			updatedNamespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-namespace",
					Annotations: map[string]string{
						lib.TenantAnnotation: "tenant-a", // No change
					},
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				// Expect update call since the object already exists
				mockAviClient.EXPECT().AviSessionPut(
					gomock.Any(), // URL
					gomock.Any(), // Request body
					gomock.Any(), // Response
					gomock.Any(), // Options - tenant-a
				).Return(nil).Times(1)
			},
			prepareCache: func(cache *mock.MockCacheOperation) {
				cache.EXPECT().GetObjectByUUID(gomock.Any(), gomock.Any()).Return(nil, false).AnyTimes()
			},
			expectedTenant:   "tenant-a",
			expectDeletion:   false,
			expectRecreation: false,
			wantErr:          false,
		},
		{
			name: "success: new resource with tenant annotation",
			initialHM: &akov1alpha1.HealthMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test-hm-new",
					Namespace:       "test-namespace",
					Finalizers:      []string{"healthmonitor.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
				},
				Spec: akov1alpha1.HealthMonitorSpec{
					Type: "HEALTH_MONITOR_HTTP",
				},
				Status: akov1alpha1.HealthMonitorStatus{
					// No UUID or Tenant set (new resource)
				},
			},
			initialNamespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-namespace",
					Annotations: map[string]string{
						lib.TenantAnnotation: "tenant-b",
					},
				},
			},
			updatedNamespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-namespace",
					Annotations: map[string]string{
						lib.TenantAnnotation: "tenant-b", // Same tenant
					},
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				// Expect creation call for the tenant
				mockAviClient.EXPECT().AviSessionPost(
					constants.HealthMonitorURL,
					gomock.Any(), // Request body
					gomock.Any(), // Response
					gomock.Any(), // Options - tenant-b
				).DoAndReturn(func(url string, req interface{}, resp interface{}, opts ...interface{}) error {
					// Simulate successful creation response
					respMap := resp.(*map[string]interface{})
					(*respMap)["uuid"] = "new-uuid-789"
					return nil
				}).Times(1)
			},
			prepareCache: func(cache *mock.MockCacheOperation) {
				cache.EXPECT().GetObjectByUUID(gomock.Any(), gomock.Any()).Return(nil, false).AnyTimes()
			},
			expectedTenant:   "tenant-b",
			expectDeletion:   false,
			expectRecreation: true,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake k8s client
			scheme := runtime.NewScheme()
			_ = akov1alpha1.AddToScheme(scheme)
			_ = corev1.AddToScheme(scheme)

			// Start with the updated namespace (simulating the namespace change)
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(tt.initialHM, tt.updatedNamespace).WithStatusSubresource(tt.initialHM).Build()

			// Create mock AVI client
			mockAviClient := mock.NewMockAviClientInterface(gomock.NewController(t))
			if tt.prepare != nil {
				tt.prepare(mockAviClient)
			}

			mockCache := mock.NewMockCacheOperation(gomock.NewController(t))
			if tt.prepareCache != nil {
				tt.prepareCache(mockCache)
			}

			// Create reconciler
			reconciler := &HealthMonitorReconciler{
				Client:        fakeClient,
				AviClient:     mockAviClient,
				Scheme:        scheme,
				Logger:        utils.AviLog,
				EventRecorder: record.NewFakeRecorder(10),
				ClusterName:   "test-cluster",
				Cache:         mockCache,
			}

			// Test reconcile
			ctx := context.Background()
			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      tt.initialHM.Name,
					Namespace: tt.initialHM.Namespace,
				},
			}

			_, err := reconciler.Reconcile(ctx, req)
			if tt.wantErr {
				assert.Error(t, err)
				return
			} else {
				assert.NoError(t, err)
			}

			// Check final state
			hm := &akov1alpha1.HealthMonitor{}
			err = fakeClient.Get(ctx, req.NamespacedName, hm)
			assert.NoError(t, err)

			// Verify tenant is correctly set
			assert.Equal(t, tt.expectedTenant, hm.Status.Tenant, "Tenant should match expected value")

			// For tenant change scenarios, verify the UUID was cleared and recreated
			if tt.expectDeletion && tt.expectRecreation {
				assert.NotEmpty(t, hm.Status.UUID, "UUID should be set after recreation")
				assert.NotEqual(t, "existing-uuid-123", hm.Status.UUID, "UUID should be different after recreation")
			}

			// For no-change scenarios, verify UUID is preserved
			if !tt.expectDeletion && !tt.expectRecreation && tt.initialHM.Status.UUID != "" {
				assert.Equal(t, tt.initialHM.Status.UUID, hm.Status.UUID, "UUID should be preserved when no tenant change")
			}

			// For new resources, verify UUID is set
			if tt.expectRecreation && tt.initialHM.Status.UUID == "" {
				assert.NotEmpty(t, hm.Status.UUID, "UUID should be set for new resource")
			}
		})
	}
}
