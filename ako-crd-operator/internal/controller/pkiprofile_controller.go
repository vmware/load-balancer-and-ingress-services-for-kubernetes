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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// PKIProfileReconciler reconciles a PKIProfile object
type PKIProfileReconciler struct {
	client.Client
	AviClient     avisession.AviClientInterface
	Scheme        *runtime.Scheme
	Cache         cache.CacheOperation
	Logger        *utils.AviLogger
	EventRecorder record.EventRecorder
	ClusterName   string
}

// GetLogger returns the logger for the reconciler to implement NamespaceHandler interface
func (r *PKIProfileReconciler) GetLogger() *utils.AviLogger {
	return r.Logger
}

// We'll use models.PKIprofile directly instead of a custom struct

// +kubebuilder:rbac:groups=ako.vmware.com,resources=pkiprofiles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ako.vmware.com,resources=pkiprofiles/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ako.vmware.com,resources=pkiprofiles/finalizers,verbs=update;get;create;delete
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *PKIProfileReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Logger.WithValues("name", req.Name, "namespace", req.Namespace, "traceID", uuid.New().String())
	ctx = utils.LoggerWithContext(ctx, log)
	log.Debug("Reconciling PKIProfile")
	defer log.Debug("Reconciled PKIProfile")
	pki := &akov1alpha1.PKIProfile{}
	err := r.Client.Get(ctx, req.NamespacedName, pki)
	if err != nil {
		if k8serror.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error("Failed to get PKIProfile")
		return ctrl.Result{}, err
	}
	if pki.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(pki, constants.PKIProfileFinalizer) {
			controllerutil.AddFinalizer(pki, constants.PKIProfileFinalizer)
			if err := r.Update(ctx, pki); err != nil {
				log.Error("Failed to add finalizer to PKIProfile")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	} else {
		// The object is being deleted
		removeFinalizer := false
		if err, removeFinalizer = r.DeleteObject(ctx, pki); err != nil {
			return ctrl.Result{}, err
		}
		if removeFinalizer {
			controllerutil.RemoveFinalizer(pki, constants.PKIProfileFinalizer)
		} else {
			if err := r.Status().Update(ctx, pki); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: constants.RequeueInterval}, nil
		}
		if err := r.Update(ctx, pki); err != nil {
			return ctrl.Result{}, err
		}
		r.EventRecorder.Event(pki, corev1.EventTypeNormal, "Deleted", "PKIProfile deleted successfully from Avi Controller")
		log.Info("successfully deleted PKIProfile")
		return ctrl.Result{}, nil
	}
	// create or update - validate the object
	if err := r.ValidatedObject(ctx, pki); err != nil {
		// Check if the error is retryable
		if controllerutils.IsRetryableError(err) {
			// For 404(object not found) also, we are not retrying. So user has to update the object again to trigger
			// processing.
			// other way to retry for certain number of times for each object and then stop
			return ctrl.Result{RequeueAfter: constants.RequeueInterval}, err
		}
	}
	result, err := r.ReconcileIfRequired(ctx, pki)
	return result, err
}

func (r *PKIProfileReconciler) ValidatedObject(ctx context.Context, pki *akov1alpha1.PKIProfile) error {
	log := utils.LoggerFromContext(ctx)
	log.Info("Validating PKIProfile CRD")

	//TODO: Add more sophisticated certificate validation for pkiprofile if needed in future

	err := r.SetStatus(ctx, pki, "Ready", "ValidationSucceeded", "PKIProfile validation succeeded")
	if err != nil {
		log.Errorf("error in setting status: %s", err.Error())
		return err
	}
	log.Info("Accepted. Validated PKIProfile CRD")
	return nil
}

func (r *PKIProfileReconciler) SetStatus(ctx context.Context, pki *akov1alpha1.PKIProfile, conditionType string, reason string, message string) error {
	statusManager := &controllerutils.StatusManager{
		Client:        r.Client,
		EventRecorder: r.EventRecorder,
	}
	return statusManager.SetStatus(ctx, pki, conditionType, reason, message)
}

// SetupWithManager sets up the controller with the Manager.
func (r *PKIProfileReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&akov1alpha1.PKIProfile{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Named("pkiprofile").
		Watches(
			&corev1.Namespace{},
			handler.EnqueueRequestsFromMapFunc(controllerutils.CreateGenericNamespaceHandler(
				r,
				"pkiprofile",
				func() *akov1alpha1.PKIProfileList { return &akov1alpha1.PKIProfileList{} },
				func(list *akov1alpha1.PKIProfileList) []client.Object {
					objects := make([]client.Object, len(list.Items))
					for i, item := range list.Items {
						objects[i] = &item
					}
					return objects
				},
			)),
			builder.WithPredicates(controllerutils.TenantAnnotationNamespacePredicate()),
		).
		Complete(r)
}

// DeleteObject deletes the PKIProfile from Avi Controller and returns (error, bool)
// The boolean indicates whether the finalizer should be removed (true) or kept (false)
func (r *PKIProfileReconciler) DeleteObject(ctx context.Context, pki *akov1alpha1.PKIProfile) (error, bool) {
	log := utils.LoggerFromContext(ctx)
	if pki.Status.UUID != "" && pki.Status.Tenant != "" {
		if err := r.AviClient.AviSessionDelete(utils.GetUriEncoded(fmt.Sprintf("%s/%s", constants.PKIProfileURL, pki.Status.UUID)), nil, nil, session.SetOptTenant(pki.Status.Tenant)); err != nil {
			// Handle 404 as success case - object doesn't exist, which is the desired state for delete
			if aviError, ok := err.(session.AviError); ok {
				switch aviError.HttpStatusCode {
				case 404:
					log.Info("PKIProfile not found on Avi Controller (404), treating as successful deletion")
					return nil, true
				case 403:
					log.Errorf("PKIProfile is being referred by other objects, cannot be deleted. %s", aviError.Error())
					r.SetStatus(ctx, pki, "Deleted", "DeletionSkipped", controllerutils.ParseAviErrorMessage(*aviError.Message))
					return nil, false
				}
			}
			log.Errorf("error deleting pki profile: %s", err.Error())
			r.SetStatus(ctx, pki, "Deleted", "DeletionFailed", fmt.Sprintf("Failed to delete PKIProfile from Avi Controller: %v", err))
			return err, false
		}
	} else {
		log.Warn("error deleting pki profile. uuid not present. possibly avi pki profile object not created")
		r.SetStatus(ctx, pki, "Deleted", "DeletionSkipped", "UUID not present, PKIProfile may not have been created on Avi Controller")
	}
	return nil, true
}

// TODO: Make this function generic
func (r *PKIProfileReconciler) ReconcileIfRequired(ctx context.Context, pki *akov1alpha1.PKIProfile) (ctrl.Result, error) {
	log := utils.LoggerFromContext(ctx)
	namespaceTenant, err := controllerutils.GetTenantInNamespace(ctx, r.Client, pki.Namespace)
	if err != nil {
		log.Errorf("error getting tenant in namespace: %s", err.Error())
		// Check if the error is retryable
		if controllerutils.IsRetryableError(err) {
			log.Infof("Tenant lookup error is retryable, will retry: %v", err)
			return ctrl.Result{RequeueAfter: constants.RequeueInterval}, err
		}
		return ctrl.Result{}, err
	}

	// Check if tenant in status differs from tenant in namespace annotation
	// Only trigger tenant mismatch if status has a tenant set (not for new resources)
	if pki.Status.Tenant != "" && pki.Status.Tenant != namespaceTenant {
		log.Infof("Tenant update detected. Status tenant: %s, Namespace tenant: %s. Deleting PKIProfile from AVI.", pki.Status.Tenant, namespaceTenant)
		err, _ := r.DeleteObject(ctx, pki)
		if err != nil {
			log.Errorf("Failed to delete PKIProfile due to error: %s", err.Error())
			r.SetStatus(ctx, pki, "Deleted", "DeletionFailed", fmt.Sprintf("Failed to delete PKIProfile due to error: %v", err))
			// Check if the error is retryable
			if controllerutils.IsRetryableError(err) {
				log.Infof("Tenant update deletion error is retryable, will retry: %v", err)
				return ctrl.Result{RequeueAfter: constants.RequeueInterval}, err
			}
			return ctrl.Result{}, err
		}
		// Clear the status to force recreation with correct tenant
		pki.Status = akov1alpha1.PKIProfileStatus{}
		log.Info("PKIProfile deleted from AVI due to tenant update, status cleared for recreation")
	}

	// Ensure crl_check is always set to false
	crlCheckFalse := false
	pkiSpec := pki.Spec
	// Convert our custom SSLCertificate to AVI SDK SSLCertificate
	var aviCACerts []*models.SSLCertificate
	for _, cert := range pkiSpec.CACerts {
		if cert != nil {
			aviCert := &models.SSLCertificate{
				Certificate: cert.Certificate,
			}
			aviCACerts = append(aviCACerts, aviCert)
		}
	}

	// Create PKI profile using AVI SDK model directly
	name := fmt.Sprintf("%s-%s-%s", r.ClusterName, pki.Namespace, pki.Name)
	tenantRef := fmt.Sprintf("/api/tenant/?name=%s", namespaceTenant)
	createdBy := fmt.Sprintf("ako-crd-operator-%s", r.ClusterName)

	pkiReq := &models.PKIprofile{
		Name:      &name,
		CaCerts:   aviCACerts,
		CrlCheck:  &crlCheckFalse,
		TenantRef: &tenantRef,
		CreatedBy: &createdBy,
		Markers:   controllerutils.CreateMarkers(r.ClusterName, pki.Namespace),
	}
	// this is a POST Call
	if pki.Status.UUID == "" {
		resp, err := r.createPKIProfile(ctx, pkiReq, pki, namespaceTenant)
		if err != nil {
			log.Errorf("error creating pki profile: %s", err.Error())
			// Check if the error is retryable
			if controllerutils.IsRetryableError(err) {
				return ctrl.Result{RequeueAfter: constants.RequeueInterval}, err
			}
			r.SetStatus(ctx, pki, "Ready", "CreationFailed", fmt.Sprintf("Failed to create PKIProfile on Avi Controller: %v", err))
			return ctrl.Result{}, err
		}
		uuid, err := extractUUID(resp)
		if err != nil {
			log.Errorf("error extracting UUID from pki profile: %s", err.Error())
			// Check if the error is retryable
			if controllerutils.IsRetryableError(err) {
				return ctrl.Result{RequeueAfter: constants.RequeueInterval}, err
			}
			r.SetStatus(ctx, pki, "Ready", "UUIDExtractionFailed", fmt.Sprintf("Failed to extract UUID: %v", err))
			return ctrl.Result{}, err
		}
		pki.Status.UUID = uuid
		r.SetStatus(ctx, pki, "Ready", "Created", "PKIProfile created successfully on Avi Controller")
	} else {
		// this is a PUT Call
		// check if no op by checking generation
		if pki.GetGeneration() == pki.Status.ObservedGeneration {
			// if no op from kubernetes side, check if op required from OOB changes by checking lastModified timestamp
			if pki.Status.LastUpdated != nil {
				dataMap, ok := r.Cache.GetObjectByUUID(ctx, pki.Status.UUID)
				if ok {
					if dataMap.GetLastModifiedTimeStamp().Before(pki.Status.LastUpdated.Time) {
						log.Debug("no op for pki profile")
						return ctrl.Result{}, nil
					}
				}
			}
			log.Debugf("overwriting pki profile")
		}
		resp := map[string]interface{}{}
		if err := r.AviClient.AviSessionPut(utils.GetUriEncoded(fmt.Sprintf("%s/%s", constants.PKIProfileURL, pki.Status.UUID)), pkiReq, &resp, session.SetOptTenant(namespaceTenant)); err != nil {
			log.Errorf("error updating pki profile %s", err.Error())
			// Check if the error is retryable
			if controllerutils.IsRetryableError(err) {
				return ctrl.Result{RequeueAfter: constants.RequeueInterval}, err
			}
			r.SetStatus(ctx, pki, "Ready", "UpdateFailed", fmt.Sprintf("Failed to update PKIProfile on Avi Controller: %v", err))
			return ctrl.Result{}, err
		}
		// Note: We don't call SetStatus here because we're in the middle of reconciliation
		// and need to set BackendObjectName and Tenant fields later
		pki.Status.BackendObjectName = *pkiReq.Name
		pki.Status.Tenant = namespaceTenant
		r.SetStatus(ctx, pki, "Ready", "Updated", "PKIProfile updated successfully on Avi Controller")
	}
	return ctrl.Result{}, nil
}

// createPKIProfile will attempt to create a PKI profile, if it already exists, it will return an object which contains the uuid
func (r *PKIProfileReconciler) createPKIProfile(ctx context.Context, pkiReq *models.PKIprofile, pki *akov1alpha1.PKIProfile, tenant string) (map[string]interface{}, error) {
	log := utils.LoggerFromContext(ctx)
	if len(pkiReq.CaCerts) > 0 {
		for i, cert := range pkiReq.CaCerts {
			if cert != nil && cert.Certificate != nil {
				// Log the full certificate for debugging
				log.Debugf("Full CA Cert %d: %s", i, *cert.Certificate)
			} else {
				log.Warnf("CA Cert %d is nil or has nil Certificate", i)
			}
		}
	}

	// Marshal the request to JSON to see exactly what's being sent
	if reqBytes, err := json.Marshal(pkiReq); err == nil {
		log.Debugf("Full PKI request JSON: %s", string(reqBytes))
	} else {
		log.Warnf("Failed to marshal PKI request: %v", err)
	}

	resp := map[string]interface{}{}
	if err := r.AviClient.AviSessionPost(utils.GetUriEncoded(constants.PKIProfileURL), pkiReq, &resp, session.SetOptTenant(tenant)); err != nil {
		log.Errorf("error posting pki profile: %s", err.Error())
		if aviError, ok := err.(session.AviError); ok {
			if aviError.HttpStatusCode == http.StatusConflict && strings.Contains(aviError.Error(), "already exists") {
				log.Info("pki profile already exists. trying to get uuid")
				err := r.AviClient.AviSessionGet(utils.GetUriEncoded(fmt.Sprintf("%s?name=%s", constants.PKIProfileURL, *pkiReq.Name)), &resp)
				if err != nil {
					log.Errorf("error getting pki profile %s", err.Error())
					return nil, err
				}
				uuid, err := extractUUID(resp)
				if err != nil {
					log.Errorf("error extracting UUID from pki profile: %s", err.Error())
					return nil, err
				}
				log.Info("updating pki profile")
				if err := r.AviClient.AviSessionPut(utils.GetUriEncoded(fmt.Sprintf("%s/%s", constants.PKIProfileURL, uuid)), pkiReq, &resp, session.SetOptTenant(tenant)); err != nil {
					log.Errorf("error updating pki profile", err.Error())
					return nil, err
				}
				err = r.SetStatus(ctx, pki, "Ready", "Updated", "PKIProfile updated successfully on Avi Controller")
		        if err != nil {
				log.Errorf("error in setting status: %s", err.Error())
				return nil, 	err
				}
				
				return resp, nil
			}
		}
		return nil, err
	}
	err := r.SetStatus(ctx, pki, "Ready", "ValidationSucceeded", "PKIProfile created succeessfully")
	if err != nil {
		log.Errorf("error in setting status: %s", err.Error())
		return nil, err
	}
	return resp, nil
}
