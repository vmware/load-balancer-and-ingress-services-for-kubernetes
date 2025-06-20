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

package controller

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/api/v1alpha1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/constants"
	avisession "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/session"
	controllerutils "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ApplicationProfileReconciler reconciles a ApplicationProfile object
type ApplicationProfileReconciler struct {
	client.Client
	AviClient     avisession.AviClientInterface
	Scheme        *runtime.Scheme
	Cache         cache.CacheOperation
	Logger        *utils.AviLogger
	EventRecorder record.EventRecorder
	ClusterName   string
}

type ApplicationProfileRequest struct {
	Name string `json:"name"`
	akov1alpha1.ApplicationProfileSpec
	Markers []*models.RoleFilterMatchLabel `json:"markers,omitempty"`
}

// +kubebuilder:rbac:groups=ako.vmware.com,resources=applicationprofiles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ako.vmware.com,resources=applicationprofiles/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ako.vmware.com,resources=applicationprofiles/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ApplicationProfile object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *ApplicationProfileReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Logger.WithValues("name", req.Name, "namespace", req.Namespace, "traceID", uuid.New().String())
	ctx = utils.LoggerWithContext(ctx, log)
	log.Debug("Reconciling ApplicationProfile")
	defer log.Debug("Reconciled ApplicationProfile")
	ap := &akov1alpha1.ApplicationProfile{}
	err := r.Client.Get(ctx, req.NamespacedName, ap)
	if err != nil {
		if k8serror.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error("Failed to get ApplicationProfile")
		return ctrl.Result{}, err
	}
	if ap.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(ap, constants.ApplicationProfileFinalizer) {
			controllerutil.AddFinalizer(ap, constants.ApplicationProfileFinalizer)
			if err := r.Update(ctx, ap); err != nil {
				log.Error("Failed to add finalizer to ApplicationProfile")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	} else {
		// The object is being deleted
		removeFinalizer := false
		if err, removeFinalizer = r.DeleteObject(ctx, ap); err != nil {
			return ctrl.Result{}, err
		}
		if removeFinalizer {
			controllerutil.RemoveFinalizer(ap, constants.ApplicationProfileFinalizer)
		} else {
			if err := r.Status().Update(ctx, ap); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: constants.RequeueInterval}, nil
		}
		if err := r.Update(ctx, ap); err != nil {
			return ctrl.Result{}, err
		}
		r.EventRecorder.Event(ap, corev1.EventTypeNormal, "Deleted", "ApplicationProfile deleted successfully from Avi Controller")
		log.Info("succesfully deleted applicationprofile")
		return ctrl.Result{}, nil
	}
	if err := r.ReconcileIfRequired(ctx, ap); err != nil {
		// Check if the error is retryable
		if !controllerutils.IsRetryableError(err) {
			// Update status with non-retryable error condition and don't return error (to avoid requeue)
			controllerutils.UpdateStatusWithNonRetryableError(ctx, r, ap, err, "ApplicationProfile")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ApplicationProfileReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&akov1alpha1.ApplicationProfile{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Named("applicationprofile").
		Complete(r)
}

// DeleteObject deletes the ApplicationProfile from Avi Controller and returns (error, bool)
// The boolean indicates whether the finalizer should be removed (true) or kept (false)
func (r *ApplicationProfileReconciler) DeleteObject(ctx context.Context, ap *akov1alpha1.ApplicationProfile) (error, bool) {
	log := utils.LoggerFromContext(ctx)
	if ap.Status.UUID != "" {
		if err := r.AviClient.AviSessionDelete(utils.GetUriEncoded(fmt.Sprintf("%s/%s", constants.ApplicationProfileURL, ap.Status.UUID)), nil, nil); err != nil {
			// Handle 404 as success case - object doesn't exist, which is the desired state for delete
			if aviError, ok := err.(session.AviError); ok {
				switch aviError.HttpStatusCode {
				case 404:
					log.Info("ApplicationProfile not found on Avi Controller (404), treating as successful deletion")
					return nil, true
				case 403:
					log.Errorf("ApplicationProfile is being referred by other objects, cannot be deleted. %s", aviError.Error())
					r.EventRecorder.Event(ap, corev1.EventTypeWarning, "DeletionSkipped", aviError.Error())
					ap.Status.Conditions = controllerutils.SetCondition(ap.Status.Conditions, metav1.Condition{
						Type:               "Deleted",
						Status:             metav1.ConditionFalse,
						LastTransitionTime: metav1.Time{Time: time.Now().UTC()},
						Reason:             "DeletionSkipped",
						Message:            controllerutils.ParseAviErrorMessage(*aviError.Message),
					})
					return nil, false
				}
			}
			log.Errorf("error deleting application profile: %s", err.Error())
			r.EventRecorder.Event(ap, corev1.EventTypeWarning, "DeletionFailed", fmt.Sprintf("Failed to delete ApplicationProfile from Avi Controller: %v", err))
			return err, false
		}
	} else {
		r.EventRecorder.Event(ap, corev1.EventTypeWarning, "DeletionSkipped", "UUID not present, ApplicationProfile may not have been created on Avi Controller")
		log.Warn("error deleting application profile. uuid not present. possibly avi application profile object not created")
	}
	return nil, true
}

// TODO: Make this function generic
func (r *ApplicationProfileReconciler) ReconcileIfRequired(ctx context.Context, ap *akov1alpha1.ApplicationProfile) error {
	log := utils.LoggerFromContext(ctx)
	apReq := &ApplicationProfileRequest{
		Name:                   fmt.Sprintf("%s-%s-%s", r.ClusterName, ap.Namespace, ap.Name),
		ApplicationProfileSpec: ap.Spec,
		Markers:                controllerutils.CreateMarkers(r.ClusterName, ap.Namespace),
	}
	// this is a POST Call
	if ap.Status.UUID == "" {
		resp, err := r.createApplicationProfile(ctx, apReq, ap)
		if err != nil {
			r.EventRecorder.Event(ap, corev1.EventTypeWarning, "CreationFailed", fmt.Sprintf("Failed to create ApplicationProfile on Avi Controller: %v", err))
			log.Errorf("error creating application profile: %s", err.Error())
			return err
		}
		uuid, err := extractUUID(resp)
		if err != nil {
			r.EventRecorder.Event(ap, corev1.EventTypeWarning, "UUIDExtractionFailed", fmt.Sprintf("Failed to extract UUID: %v", err))
			log.Errorf("error extracting UUID from application profile: %s", err.Error())
			return err
		}
		ap.Status.UUID = uuid
		r.EventRecorder.Event(ap, corev1.EventTypeNormal, "Created", "ApplicationProfile created successfully on Avi Controller")
	} else {
		// this is a PUT Call
		// check if no op by checking generation
		if ap.GetGeneration() == ap.Status.ObservedGeneration {
			// if no op from kubernetes side, check if op required from OOB changes by checking lastModified timestamp
			if ap.Status.LastUpdated != nil {
				dataMap, ok := r.Cache.GetObjectByUUID(ctx, ap.Status.UUID)
				if ok {
					if dataMap.GetLastModifiedTimeStamp().Before(ap.Status.LastUpdated.Time) {
						log.Debug("no op for application profile")
						return nil
					}
				}
			}
			log.Debugf("overwriting applicationprofile")
		}
		resp := map[string]interface{}{}
		if err := r.AviClient.AviSessionPut(utils.GetUriEncoded(fmt.Sprintf("%s/%s", constants.ApplicationProfileURL, ap.Status.UUID)), apReq, &resp); err != nil {
			log.Errorf("error updating application profile %s", err.Error())
			r.EventRecorder.Event(ap, corev1.EventTypeWarning, "UpdateFailed", fmt.Sprintf("Failed to update ApplicationProfile on Avi Controller: %v", err))
			return err
		}
		ap.Status.Conditions = controllerutils.SetCondition(ap.Status.Conditions, metav1.Condition{
			Type:               "Ready",
			Status:             metav1.ConditionTrue,
			LastTransitionTime: metav1.Time{Time: time.Now().UTC()},
			Reason:             "Updated",
			Message:            "ApplicationProfile updated successfully on Avi Controller",
		})
		r.EventRecorder.Event(ap, corev1.EventTypeNormal, "Updated", "ApplicationProfile updated successfully on Avi Controller")
		log.Info("succesfully updated application profile")
	}
	ap.Status.BackendObjectName = apReq.Name
	lastUpdated := metav1.Time{Time: time.Now().UTC()}
	ap.Status.LastUpdated = &lastUpdated
	ap.Status.ObservedGeneration = ap.Generation
	if err := r.Status().Update(ctx, ap); err != nil {
		r.EventRecorder.Event(ap, corev1.EventTypeWarning, "StatusUpdateFailed", fmt.Sprintf("Failed to update ApplicationProfile status: %v", err))
		log.Errorf("unable to update application profile status %s", err.Error())
		return err
	}
	return nil
}

// createApplicationProfile will attempt to create a application profile, if it already exists, it will return an object which contains the uuid
func (r *ApplicationProfileReconciler) createApplicationProfile(ctx context.Context, apReq *ApplicationProfileRequest, ap *akov1alpha1.ApplicationProfile) (map[string]interface{}, error) {
	log := utils.LoggerFromContext(ctx)
	resp := map[string]interface{}{}
	if err := r.AviClient.AviSessionPost(utils.GetUriEncoded(constants.ApplicationProfileURL), apReq, &resp); err != nil {
		log.Errorf("error posting application profile: %s", err.Error())
		if aviError, ok := err.(session.AviError); ok {
			if aviError.HttpStatusCode == http.StatusConflict && strings.Contains(aviError.Error(), "already exists") {
				log.Info("application profile already exists. trying to get uuid")
				err := r.AviClient.AviSessionGet(utils.GetUriEncoded(fmt.Sprintf("%s?name=%s", constants.ApplicationProfileURL, apReq.Name)), &resp)
				if err != nil {
					log.Errorf("error getting application profile %s", err.Error())
					return nil, err
				}
				uuid, err := extractUUID(resp)
				if err != nil {
					log.Errorf("error extracting UUID from application profile: %s", err.Error())
					return nil, err
				}
				log.Info("updating application profile")
				if err := r.AviClient.AviSessionPut(utils.GetUriEncoded(fmt.Sprintf("%s/%s", constants.ApplicationProfileURL, uuid)), apReq, &resp); err != nil {
					log.Errorf("error updating application profile", err.Error())
					return nil, err
				}
				ap.Status.Conditions = controllerutils.SetCondition(ap.Status.Conditions, metav1.Condition{
					Type:               "Ready",
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.Time{Time: time.Now().UTC()},
					Reason:             "Updated",
					Message:            "ApplicationProfile updated successfully on Avi Controller",
				})
				return resp, nil
			}
		}
		return nil, err
	}
	ap.Status.Conditions = controllerutils.SetCondition(ap.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Time{Time: time.Now().UTC()},
		Reason:             "Created",
		Message:            "ApplicationProfile created successfully on Avi Controller",
	})
	log.Info("Application profile successfully created")
	return resp, nil
}
