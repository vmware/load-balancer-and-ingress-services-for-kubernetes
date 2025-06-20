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
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
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
				mockAviClient.EXPECT().AviSessionPost(constants.HealthMonitorURL, gomock.Any(), gomock.Any()).Do(func(url string, request interface{}, response interface{}) {
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
				mockAviClient.EXPECT().AviSessionPost(constants.HealthMonitorURL, gomock.Any(), gomock.Any()).Return(session.AviError{
					HttpStatusCode: http.StatusConflict,
					AviResult: session.AviResult{
						Message: &[]string{"already exists"}[0],
					},
				})
				mockAviClient.EXPECT().AviSessionGet(fmt.Sprintf("%s?name=%s", constants.HealthMonitorURL, "test-cluster-default-test"), gomock.Any()).Do(func(url string, response interface{}) {
					if resp, ok := response.(*map[string]interface{}); ok {
						*resp = responseBody
					}
				}).Return(nil).AnyTimes()
				mockAviClient.EXPECT().AviSessionPut(constants.HealthMonitorURL+"/123", gomock.Any(), gomock.Any()).Do(func(url string, request interface{}, response interface{}) {
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
				mockAviClient.EXPECT().AviSessionPut(constants.HealthMonitorURL+"/123", gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
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
				mockAviClient.EXPECT().AviSessionDelete(constants.HealthMonitorURL+"/123", gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
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
				mockAviClient.EXPECT().AviSessionDelete(constants.HealthMonitorURL+"/123", gomock.Any(), gomock.Any()).Return(session.AviError{
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
				mockAviClient.EXPECT().AviSessionPut(constants.HealthMonitorURL+"/123", gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
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
				mockAviClient.EXPECT().AviSessionPut(constants.HealthMonitorURL+"/123", gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
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
				mockAviClient.EXPECT().AviSessionPost(constants.HealthMonitorURL, gomock.Any(), gomock.Any()).Return(session.AviError{
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
				mockAviClient.EXPECT().AviSessionPost(constants.HealthMonitorURL, gomock.Any(), gomock.Any()).Return(session.AviError{
					HttpStatusCode: http.StatusConflict,
					AviResult: session.AviResult{
						Message: &[]string{"already exists"}[0],
					},
				})
				mockAviClient.EXPECT().AviSessionGet(fmt.Sprintf("%s?name=%s", constants.HealthMonitorURL, "test-cluster-default-test"), gomock.Any()).Return(errors.New("GET failed")).AnyTimes()
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
				mockAviClient.EXPECT().AviSessionPost(constants.HealthMonitorURL, gomock.Any(), gomock.Any()).Return(session.AviError{
					HttpStatusCode: http.StatusConflict,
					AviResult: session.AviResult{
						Message: &[]string{"already exists"}[0],
					},
				})
				mockAviClient.EXPECT().AviSessionGet(fmt.Sprintf("%s?name=%s", constants.HealthMonitorURL, "test-cluster-default-test"), gomock.Any()).Do(func(url string, response interface{}) {
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
				mockAviClient.EXPECT().AviSessionPost(constants.HealthMonitorURL, gomock.Any(), gomock.Any()).Return(session.AviError{
					HttpStatusCode: http.StatusConflict,
					AviResult: session.AviResult{
						Message: &[]string{"already exists"}[0],
					},
				})
				mockAviClient.EXPECT().AviSessionGet(fmt.Sprintf("%s?name=%s", constants.HealthMonitorURL, "test-cluster-default-test"), gomock.Any()).Do(func(url string, response interface{}) {
					if resp, ok := response.(*map[string]interface{}); ok {
						*resp = responseBody
					}
				}).Return(nil).AnyTimes()
				mockAviClient.EXPECT().AviSessionPut(constants.HealthMonitorURL+"/123", gomock.Any(), gomock.Any()).Return(errors.New("PUT failed")).AnyTimes()
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
				mockAviClient.EXPECT().AviSessionPost(constants.HealthMonitorURL, gomock.Any(), gomock.Any()).Do(func(url string, request interface{}, response interface{}) {
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
				mockAviClient.EXPECT().AviSessionPut(constants.HealthMonitorURL+"/123", gomock.Any(), gomock.Any()).Return(errors.New("PUT failed")).AnyTimes()
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
					UUID: "123",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				mockAviClient.EXPECT().AviSessionDelete(constants.HealthMonitorURL+"/123", gomock.Any(), gomock.Any()).Return(session.AviError{
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
					UUID: "123",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				mockAviClient.EXPECT().AviSessionDelete(constants.HealthMonitorURL+"/123", gomock.Any(), gomock.Any()).Return(session.AviError{
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
				mockAviClient.EXPECT().AviSessionPost(constants.HealthMonitorURL, gomock.Any(), gomock.Any()).Return(session.AviError{
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
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(tt.hm).WithStatusSubresource(tt.hm).Build()

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

				// Create a fake client that will return an error for Get operations
				builder := fake.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{
					Get: func(ctx context.Context, client client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
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

				builder := fake.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{
					Get: func(ctx context.Context, client client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
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
				hm := &akov1alpha1.HealthMonitor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
				}
				builder := fake.NewClientBuilder().WithScheme(scheme).WithObjects(hm).WithStatusSubresource(hm).WithInterceptorFuncs(interceptor.Funcs{
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
				hm := &akov1alpha1.HealthMonitor{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "test",
						Namespace:         "default",
						DeletionTimestamp: &metav1.Time{Time: time.Now().Truncate(time.Second)},
						Finalizers:        []string{"healthmonitor.ako.vmware.com/finalizer"},
					},
				}
				builder := fake.NewClientBuilder().WithScheme(scheme).WithObjects(hm).WithStatusSubresource(hm).WithInterceptorFuncs(interceptor.Funcs{
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
				builder := fake.NewClientBuilder().WithScheme(scheme).WithObjects(hm).WithStatusSubresource(hm).WithInterceptorFuncs(interceptor.Funcs{
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
				mockAviClient.EXPECT().AviSessionPut(constants.HealthMonitorURL+"/123", gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
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
				EventRecorder: &record.FakeRecorder{},
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
