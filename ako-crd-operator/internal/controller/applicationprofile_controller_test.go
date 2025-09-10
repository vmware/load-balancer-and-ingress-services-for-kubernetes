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
)

func TestApplicationProfileController(t *testing.T) {
	tests := []struct {
		name         string
		ap           *akov1alpha1.ApplicationProfile
		prepare      func(mockAviClient *mock.MockAviClientInterface)
		prepareCache func(cache *mock.MockCacheOperation)
		want         *akov1alpha1.ApplicationProfile
		wantErr      bool
	}{
		{
			name: "success: add finalizer",
			ap: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				responseUUID := map[string]interface{}{
					"uuid": "123",
				}
				mockAviClient.EXPECT().AviSessionPost(constants.ApplicationProfileURL, gomock.Any(), gomock.Any(), gomock.Any()).Do(func(url string, request interface{}, response interface{}, params interface{}) {
					if resp, ok := response.(*map[string]interface{}); ok {
						*resp = responseUUID
					}
				}).Return(nil).AnyTimes()
			},
			want: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test",
					Finalizers:      []string{"applicationprofile.ako.vmware.com/finalizer"},
					ResourceVersion: "1001",
				},
				Status: akov1alpha1.ApplicationProfileStatus{
					UUID:              "123",
					BackendObjectName: "test-cluster--test",
					Tenant:            "admin",
					Conditions: []metav1.Condition{
						{
							Type:    "Ready",
							Status:  metav1.ConditionTrue,
							Reason:  "Created",
							Message: "ApplicationProfile created successfully on Avi Controller",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "success: add applicationprofile",
			ap: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test",
					Finalizers:      []string{"applicationprofile.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
					Namespace:       "default",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				responseUUID := map[string]interface{}{
					"uuid": "123",
				}
				mockAviClient.EXPECT().AviSessionPost(constants.ApplicationProfileURL, gomock.Any(), gomock.Any(), gomock.Any()).Do(func(url string, request interface{}, response interface{}, params interface{}) {
					if resp, ok := response.(*map[string]interface{}); ok {
						*resp = responseUUID
					}
				}).Return(nil).AnyTimes()
			},
			want: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test",
					Finalizers:      []string{"applicationprofile.ako.vmware.com/finalizer"},
					Namespace:       "default",
					ResourceVersion: "1001",
				},
				Status: akov1alpha1.ApplicationProfileStatus{
					UUID: "123",
					Conditions: []metav1.Condition{
						{
							Type:   "Ready",
							Status: metav1.ConditionTrue,
							// fake client isnt supporting time.UTC with nanoseconds precision
							LastTransitionTime: metav1.Time{Time: time.Now().Truncate(time.Second)},
							Reason:             "Created",
							Message:            "ApplicationProfile created successfully on Avi Controller",
						},
					},
					BackendObjectName: "test-cluster-default-test",
					Tenant:            "admin",
					LastUpdated:       &metav1.Time{Time: time.Now().Truncate(time.Second)},
				},
			},
			wantErr: false,
		},
		{
			name: "success: add applicationprofile with existing applicationprofile",
			ap: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test",
					Finalizers:      []string{"applicationprofile.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
					Namespace:       "default",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				responseBody := map[string]interface{}{
					"results": []interface{}{map[string]interface{}{"uuid": "123"}},
				}
				mockAviClient.EXPECT().AviSessionPost(constants.ApplicationProfileURL, gomock.Any(), gomock.Any(), gomock.Any()).Return(session.AviError{
					HttpStatusCode: http.StatusConflict,
					AviResult: session.AviResult{
						Message: &[]string{"already exists"}[0],
					},
				})
				mockAviClient.EXPECT().AviSessionGet(fmt.Sprintf("%s?name=%s", constants.ApplicationProfileURL, "test-cluster-default-test"), gomock.Any(), gomock.Any()).Do(func(url string, response interface{}, params interface{}) {
					if resp, ok := response.(*map[string]interface{}); ok {
						*resp = responseBody
					}
				}).Return(nil).AnyTimes()
				mockAviClient.EXPECT().AviSessionPut(constants.ApplicationProfileURL+"/123", gomock.Any(), gomock.Any(), gomock.Any()).Do(func(url string, request interface{}, response interface{}, params interface{}) {
					if resp, ok := response.(*map[string]interface{}); ok {
						*resp = responseBody
					}
				}).Return(nil).AnyTimes()

			},
			want: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test",
					Finalizers:      []string{"applicationprofile.ako.vmware.com/finalizer"},
					Namespace:       "default",
					ResourceVersion: "1001",
				},
				Status: akov1alpha1.ApplicationProfileStatus{
					UUID: "123",
					Conditions: []metav1.Condition{
						{
							Type:               "Ready",
							Status:             metav1.ConditionTrue,
							LastTransitionTime: metav1.Time{Time: time.Now().Truncate(time.Second)},
							Reason:             "Updated",
							Message:            "ApplicationProfile updated successfully on Avi Controller",
						},
					},
					BackendObjectName: "test-cluster-default-test",
					Tenant:            "admin",
					LastUpdated:       &metav1.Time{Time: time.Now().Truncate(time.Second)},
				},
			},
			wantErr: false,
		},
		{
			name: "success: update applicationprofile",
			ap: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test",
					Finalizers:      []string{"applicationprofile.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
					Namespace:       "default",
				},
				Status: akov1alpha1.ApplicationProfileStatus{
					UUID: "123",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				mockAviClient.EXPECT().AviSessionPut(constants.ApplicationProfileURL+"/123", gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			want: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test",
					Finalizers:      []string{"applicationprofile.ako.vmware.com/finalizer"},
					Namespace:       "default",
					ResourceVersion: "1001",
				},
				Status: akov1alpha1.ApplicationProfileStatus{
					UUID: "123",
					Conditions: []metav1.Condition{
						{
							Type:               "Ready",
							Status:             metav1.ConditionTrue,
							LastTransitionTime: metav1.Time{Time: time.Now().Truncate(time.Second)},
							Reason:             "Updated",
							Message:            "ApplicationProfile updated successfully on Avi Controller",
						},
					},
					BackendObjectName: "test-cluster-default-test",
					Tenant:            "admin",
					LastUpdated:       &metav1.Time{Time: time.Now().Truncate(time.Second)},
				},
			},
			wantErr: false,
		},
		{
			name: "success: update applicationprofile with no changes",
			ap: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					Name:            "test",
					Finalizers:      []string{"applicationprofile.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
					Namespace:       "default",
				},
				Status: akov1alpha1.ApplicationProfileStatus{
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
			want: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					Name:            "test",
					Finalizers:      []string{"applicationprofile.ako.vmware.com/finalizer"},
					Namespace:       "default",
					ResourceVersion: "1000",
				},
				Status: akov1alpha1.ApplicationProfileStatus{
					UUID:               "123",
					ObservedGeneration: 1,
					LastUpdated:        &metav1.Time{Time: time.Now().Truncate(time.Second)},
				},
			},
			wantErr: false,
		},
		{
			name: "success: delete applicationprofile",
			ap: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test",
					Finalizers:        []string{"applicationprofile.ako.vmware.com/finalizer"},
					DeletionTimestamp: &metav1.Time{Time: time.Now().Truncate(time.Second)},
				},
				Status: akov1alpha1.ApplicationProfileStatus{
					UUID: "123",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				mockAviClient.EXPECT().AviSessionDelete(constants.ApplicationProfileURL+"/123", gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "success: delete applicationprofile with 404 error (treated as success)",
			ap: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test",
					Finalizers:        []string{"applicationprofile.ako.vmware.com/finalizer"},
					DeletionTimestamp: &metav1.Time{Time: time.Now().Truncate(time.Second)},
				},
				Status: akov1alpha1.ApplicationProfileStatus{
					UUID: "123",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				mockAviClient.EXPECT().AviSessionDelete(constants.ApplicationProfileURL+"/123", gomock.Any(), gomock.Any(), gomock.Any()).Return(session.AviError{
					HttpStatusCode: 404,
				}).AnyTimes()
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "success: delete applicationprofile with empty UUID",
			ap: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test",
					Finalizers:        []string{"applicationprofile.ako.vmware.com/finalizer"},
					DeletionTimestamp: &metav1.Time{Time: time.Now().Truncate(time.Second)},
				},
				Status: akov1alpha1.ApplicationProfileStatus{
					UUID: "",
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "success: update applicationprofile with out-of-band changes (cache miss)",
			ap: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					Name:            "test",
					Finalizers:      []string{"applicationprofile.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
					Namespace:       "default",
				},
				Status: akov1alpha1.ApplicationProfileStatus{
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
				mockAviClient.EXPECT().AviSessionPut(constants.ApplicationProfileURL+"/123", gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			want: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					Name:            "test",
					Finalizers:      []string{"applicationprofile.ako.vmware.com/finalizer"},
					Namespace:       "default",
					ResourceVersion: "1001",
				},
				Status: akov1alpha1.ApplicationProfileStatus{
					UUID: "123",
					Conditions: []metav1.Condition{
						{
							Type:               "Ready",
							Status:             metav1.ConditionTrue,
							LastTransitionTime: metav1.Time{Time: time.Now().Truncate(time.Second)},
							Reason:             "Updated",
							Message:            "ApplicationProfile updated successfully on Avi Controller",
						},
					},
					BackendObjectName:  "test-cluster-default-test",
					Tenant:             "admin",
					LastUpdated:        &metav1.Time{Time: time.Now().Truncate(time.Second)},
					ObservedGeneration: 1,
				},
			},
			wantErr: false,
		},
		{
			name: "success: update applicationprofile with nil LastUpdated",
			ap: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					Name:            "test",
					Finalizers:      []string{"applicationprofile.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
					Namespace:       "default",
				},
				Status: akov1alpha1.ApplicationProfileStatus{
					UUID:               "123",
					ObservedGeneration: 1,
					LastUpdated:        nil, // Nil LastUpdated to test this path
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				mockAviClient.EXPECT().AviSessionPut(constants.ApplicationProfileURL+"/123", gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			want: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					Name:            "test",
					Finalizers:      []string{"applicationprofile.ako.vmware.com/finalizer"},
					Namespace:       "default",
					ResourceVersion: "1001",
				},
				Status: akov1alpha1.ApplicationProfileStatus{
					UUID: "123",
					Conditions: []metav1.Condition{
						{
							Type:               "Ready",
							Status:             metav1.ConditionTrue,
							LastTransitionTime: metav1.Time{Time: time.Now().Truncate(time.Second)},
							Reason:             "Updated",
							Message:            "ApplicationProfile updated successfully on Avi Controller",
						},
					},
					BackendObjectName:  "test-cluster-default-test",
					Tenant:             "admin",
					LastUpdated:        &metav1.Time{Time: time.Now().Truncate(time.Second)},
					ObservedGeneration: 1,
				},
			},
			wantErr: false,
		},
		// Negative Test Cases
		{
			name: "error: POST application profile creation fails with non-conflict error",
			ap: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test",
					Finalizers:      []string{"applicationprofile.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
					Namespace:       "default",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				mockAviClient.EXPECT().AviSessionPost(constants.ApplicationProfileURL, gomock.Any(), gomock.Any(), gomock.Any()).Return(session.AviError{
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
			ap: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test",
					Finalizers:      []string{"applicationprofile.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
					Namespace:       "default",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				mockAviClient.EXPECT().AviSessionPost(constants.ApplicationProfileURL, gomock.Any(), gomock.Any(), gomock.Any()).Return(session.AviError{
					HttpStatusCode: http.StatusConflict,
					AviResult: session.AviResult{
						Message: &[]string{"already exists"}[0],
					},
				})
				mockAviClient.EXPECT().AviSessionGet(fmt.Sprintf("%s?name=%s", constants.ApplicationProfileURL, "test-cluster-default-test"), gomock.Any(), gomock.Any()).Return(errors.New("GET failed")).AnyTimes()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error: POST conflict but extractUUID fails during recovery",
			ap: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test",
					Finalizers:      []string{"applicationprofile.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
					Namespace:       "default",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				responseBody := map[string]interface{}{
					"results": []interface{}{}, // Empty results to cause extractUUID to fail
				}
				mockAviClient.EXPECT().AviSessionPost(constants.ApplicationProfileURL, gomock.Any(), gomock.Any(), gomock.Any()).Return(session.AviError{
					HttpStatusCode: http.StatusConflict,
					AviResult: session.AviResult{
						Message: &[]string{"already exists"}[0],
					},
				})
				mockAviClient.EXPECT().AviSessionGet(fmt.Sprintf("%s?name=%s", constants.ApplicationProfileURL, "test-cluster-default-test"), gomock.Any(), gomock.Any()).Do(func(url string, response interface{}, params interface{}) {
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
			ap: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test",
					Finalizers:      []string{"applicationprofile.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
					Namespace:       "default",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				responseBody := map[string]interface{}{
					"results": []interface{}{map[string]interface{}{"uuid": "123"}},
				}
				mockAviClient.EXPECT().AviSessionPost(constants.ApplicationProfileURL, gomock.Any(), gomock.Any(), gomock.Any()).Return(session.AviError{
					HttpStatusCode: http.StatusConflict,
					AviResult: session.AviResult{
						Message: &[]string{"already exists"}[0],
					},
				})
				mockAviClient.EXPECT().AviSessionGet(fmt.Sprintf("%s?name=%s", constants.ApplicationProfileURL, "test-cluster-default-test"), gomock.Any(), gomock.Any()).Do(func(url string, response interface{}, params interface{}) {
					if resp, ok := response.(*map[string]interface{}); ok {
						*resp = responseBody
					}
				}).Return(nil).AnyTimes()
				mockAviClient.EXPECT().AviSessionPut(constants.ApplicationProfileURL+"/123", gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("PUT failed")).AnyTimes()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error: extractUUID fails after successful POST",
			ap: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test",
					Finalizers:      []string{"applicationprofile.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
					Namespace:       "default",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				responseUUID := map[string]interface{}{
					"invalid": "response", // Invalid response to cause extractUUID to fail
				}
				mockAviClient.EXPECT().AviSessionPost(constants.ApplicationProfileURL, gomock.Any(), gomock.Any(), gomock.Any()).Do(func(url string, request interface{}, response interface{}, params interface{}) {
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
			ap: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test",
					Finalizers:      []string{"applicationprofile.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
					Namespace:       "default",
				},
				Status: akov1alpha1.ApplicationProfileStatus{
					UUID: "123",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				mockAviClient.EXPECT().AviSessionPut(constants.ApplicationProfileURL+"/123", gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("PUT failed")).AnyTimes()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error: delete fails with non-404 error",
			ap: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test",
					Finalizers:        []string{"applicationprofile.ako.vmware.com/finalizer"},
					DeletionTimestamp: &metav1.Time{Time: time.Now().Truncate(time.Second)},
				},
				Status: akov1alpha1.ApplicationProfileStatus{
					UUID:   "123",
					Tenant: "admin",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				mockAviClient.EXPECT().AviSessionDelete(constants.ApplicationProfileURL+"/123", gomock.Any(), gomock.Any(), gomock.Any()).Return(session.AviError{
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
			ap: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test",
					Finalizers:        []string{"applicationprofile.ako.vmware.com/finalizer"},
					DeletionTimestamp: &metav1.Time{Time: time.Now().Truncate(time.Second)},
					ResourceVersion:   "1000",
				},
				Status: akov1alpha1.ApplicationProfileStatus{
					UUID:   "123",
					Tenant: "admin",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				mockAviClient.EXPECT().AviSessionDelete(constants.ApplicationProfileURL+"/123", gomock.Any(), gomock.Any(), gomock.Any()).Return(session.AviError{
					HttpStatusCode: 403,
					AviResult: session.AviResult{
						Message: &[]string{"Cannot delete, object is referred by: ['VirtualService custom-vs', 'Pool custom-pool']"}[0],
					},
				}).AnyTimes()
			},
			want: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test",
					Finalizers:        []string{"applicationprofile.ako.vmware.com/finalizer"},
					DeletionTimestamp: &metav1.Time{Time: time.Now().Truncate(time.Second)},
					ResourceVersion:   "1001",
				},
				Status: akov1alpha1.ApplicationProfileStatus{
					UUID: "123",
					Conditions: []metav1.Condition{
						{
							Type:               "Deleted",
							Status:             metav1.ConditionFalse,
							LastTransitionTime: metav1.Time{Time: time.Now().Truncate(time.Second)},
							Reason:             "DeletionSkipped",
							Message:            "Cannot delete, object is referred by: ['VirtualService custom-vs', 'Pool custom-pool']",
						},
					},
					Tenant: "admin",
				},
			},
			wantErr: false, // 403 doesn't cause requeue, it sets condition and waits
		},
		{
			name: "error: non-retryable error (status update without requeue)",
			ap: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test",
					Finalizers:      []string{"applicationprofile.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
					Namespace:       "default",
				},
			},
			prepare: func(mockAviClient *mock.MockAviClientInterface) {
				mockAviClient.EXPECT().AviSessionPost(constants.ApplicationProfileURL, gomock.Any(), gomock.Any(), gomock.Any()).Return(session.AviError{
					HttpStatusCode: 400, // Non-retryable error
					AviResult: session.AviResult{
						Message: &[]string{"bad request"}[0],
					},
				}).AnyTimes()
			},
			want: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test",
					Finalizers:      []string{"applicationprofile.ako.vmware.com/finalizer"},
					Namespace:       "default",
					ResourceVersion: "1001",
				},
				Status: akov1alpha1.ApplicationProfileStatus{
					Conditions: []metav1.Condition{
						{
							Type:               "Ready",
							Status:             metav1.ConditionFalse,
							LastTransitionTime: metav1.Time{Time: time.Now().Truncate(time.Second)},
							Reason:             "BadRequest",
							Message:            "Invalid ApplicationProfile specification: error from Controller: bad request",
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
			namespace := createNamespaceWithTenant(tt.ap.Namespace)

			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(tt.ap, namespace).WithStatusSubresource(tt.ap).Build()

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
			reconciler := &ApplicationProfileReconciler{
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
					Name:      tt.ap.Name,
					Namespace: tt.ap.Namespace,
				},
			}

			_, err := reconciler.Reconcile(ctx, req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Check if the application profile exists in the AVI controller
			ap := &akov1alpha1.ApplicationProfile{}
			err = fakeClient.Get(ctx, req.NamespacedName, ap)
			if err != nil {
				if tt.want == nil {
					return
				}
				if tt.wantErr {
					return
				}
				t.Errorf("Failed to get application profile: %v", err)
			}

			if tt.wantErr && tt.want == nil {
				return
			}

			// Compare everything except dynamic fields like LastUpdated and LastTransitionTime
			if tt.want != nil {
				assert.Equal(t, tt.want.ObjectMeta.Name, ap.ObjectMeta.Name)
				assert.Equal(t, tt.want.ObjectMeta.Finalizers, ap.ObjectMeta.Finalizers)
				assert.Equal(t, tt.want.ObjectMeta.ResourceVersion, ap.ObjectMeta.ResourceVersion)
				assert.Equal(t, tt.want.Status.UUID, ap.Status.UUID)
				assert.Equal(t, tt.want.Status.BackendObjectName, ap.Status.BackendObjectName)
				assert.Equal(t, tt.want.Status.Tenant, ap.Status.Tenant)
				if len(tt.want.Status.Conditions) > 0 && len(ap.Status.Conditions) > 0 {
					assert.Equal(t, tt.want.Status.Conditions[0].Type, ap.Status.Conditions[0].Type)
					assert.Equal(t, tt.want.Status.Conditions[0].Status, ap.Status.Conditions[0].Status)
					assert.Equal(t, tt.want.Status.Conditions[0].Reason, ap.Status.Conditions[0].Reason)
					assert.Equal(t, tt.want.Status.Conditions[0].Message, ap.Status.Conditions[0].Message)
					// LastTransitionTime and LastUpdated are dynamic, so we just verify they're set
					assert.NotZero(t, ap.Status.Conditions[0].LastTransitionTime)
					if tt.want.Status.LastUpdated != nil || ap.Status.LastUpdated != nil {
						assert.NotNil(t, ap.Status.LastUpdated)
					}
				}
			}
		})
	}
}

// TestApplicationProfileControllerKubernetesError tests error scenarios in Reconcile function
func TestApplicationProfileControllerKubernetesError(t *testing.T) {

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
						return k8serror.NewNotFound(akov1alpha1.GroupVersion.WithResource("applicationprofiles").GroupResource(), "test")
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
				ap := &akov1alpha1.ApplicationProfile{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
				}

				// Create namespace object with tenant annotation
				namespace := createNamespaceWithTenant("default")

				builder := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ap, namespace).WithStatusSubresource(ap).WithInterceptorFuncs(interceptor.Funcs{
					Patch: func(ctx context.Context, client client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
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
				ap := &akov1alpha1.ApplicationProfile{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "test",
						Namespace:         "default",
						DeletionTimestamp: &metav1.Time{Time: time.Now().Truncate(time.Second)},
						Finalizers:        []string{"applicationprofile.ako.vmware.com/finalizer"},
					},
				}

				// Create namespace object with tenant annotation
				namespace := createNamespaceWithTenant("default")

				builder := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ap, namespace).WithStatusSubresource(ap).WithInterceptorFuncs(interceptor.Funcs{
					Patch: func(ctx context.Context, client client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
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
				ap := &akov1alpha1.ApplicationProfile{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "test",
						Namespace:  "default",
						Finalizers: []string{"applicationprofile.ako.vmware.com/finalizer"},
					},
					Status: akov1alpha1.ApplicationProfileStatus{
						UUID: "123",
					},
				}

				// Create namespace object with tenant annotation
				namespace := createNamespaceWithTenant("default")

				builder := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ap, namespace).WithStatusSubresource(ap).WithInterceptorFuncs(interceptor.Funcs{
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
				mockAviClient.EXPECT().AviSessionPut(constants.ApplicationProfileURL+"/123", gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
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
			reconciler := &ApplicationProfileReconciler{
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
func TestApplicationProfileControllerTenantChange(t *testing.T) {
	tests := []struct {
		name             string
		initialAP        *akov1alpha1.ApplicationProfile
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
			initialAP: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test-ap-tenant-change",
					Namespace:       "test-namespace",
					Finalizers:      []string{"applicationprofile.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
				},
				Spec: akov1alpha1.ApplicationProfileSpec{
					Type: "APPLICATION_PROFILE_TYPE_HTTP",
				},
				Status: akov1alpha1.ApplicationProfileStatus{
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
					constants.ApplicationProfileURL,
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
			initialAP: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test-ap-no-change",
					Namespace:       "test-namespace",
					Finalizers:      []string{"applicationprofile.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
				},
				Spec: akov1alpha1.ApplicationProfileSpec{
					Type: "APPLICATION_PROFILE_TYPE_HTTP",
				},
				Status: akov1alpha1.ApplicationProfileStatus{
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
			initialAP: &akov1alpha1.ApplicationProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test-ap-new",
					Namespace:       "test-namespace",
					Finalizers:      []string{"applicationprofile.ako.vmware.com/finalizer"},
					ResourceVersion: "1000",
				},
				Spec: akov1alpha1.ApplicationProfileSpec{
					Type: "APPLICATION_PROFILE_TYPE_HTTP",
				},
				Status: akov1alpha1.ApplicationProfileStatus{
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
					constants.ApplicationProfileURL,
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
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(tt.initialAP, tt.updatedNamespace).WithStatusSubresource(tt.initialAP).Build()

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
			reconciler := &ApplicationProfileReconciler{
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
					Name:      tt.initialAP.Name,
					Namespace: tt.initialAP.Namespace,
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
			ap := &akov1alpha1.ApplicationProfile{}
			err = fakeClient.Get(ctx, req.NamespacedName, ap)
			assert.NoError(t, err)

			// Verify tenant is correctly set
			assert.Equal(t, tt.expectedTenant, ap.Status.Tenant, "Tenant should match expected value")

			// For tenant change scenarios, verify the UUID was cleared and recreated
			if tt.expectDeletion && tt.expectRecreation {
				assert.NotEmpty(t, ap.Status.UUID, "UUID should be set after recreation")
				assert.NotEqual(t, "existing-uuid-123", ap.Status.UUID, "UUID should be different after recreation")
			}

			// For no-change scenarios, verify UUID is preserved
			if !tt.expectDeletion && !tt.expectRecreation && tt.initialAP.Status.UUID != "" {
				assert.Equal(t, tt.initialAP.Status.UUID, ap.Status.UUID, "UUID should be preserved when no tenant change")
			}

			// For new resources, verify UUID is set
			if tt.expectRecreation && tt.initialAP.Status.UUID == "" {
				assert.NotEmpty(t, ap.Status.UUID, "UUID should be set for new resource")
			}
		})
	}
}
