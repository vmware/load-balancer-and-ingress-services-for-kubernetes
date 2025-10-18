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

	"github.com/google/uuid"

	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/api/v1alpha1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/cache"
	crdlib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/lib"
	avisession "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/session"
	controllerutils "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/utils"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	pkiprofileControllerName = "pkiprofile-controller"
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
	StatusManager *controllerutils.StatusManager
}

// GetLogger returns the logger for the reconciler to implement NamespaceHandler interface
func (r *PKIProfileReconciler) GetLogger() *utils.AviLogger {
	return r.Logger
}

// UpdateAviClient implements AviClientReconciler to update the AVI client when credentials change
func (r *PKIProfileReconciler) UpdateAviClient(client avisession.AviClientInterface) error {
	r.Logger.Info("Updating AVI client for PKIProfile controller")
	r.AviClient = client
	return nil
}

// GetReconcilerName implements AviClientReconciler to return the reconciler name
func (r *PKIProfileReconciler) GetReconcilerName() string {
	return pkiprofileControllerName
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
	patch := client.MergeFrom(pki.DeepCopy())
	if pki.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(pki, crdlib.PKIProfileFinalizer) {
			controllerutil.AddFinalizer(pki, crdlib.PKIProfileFinalizer)
			if err := r.Patch(ctx, pki, patch); err != nil {
				log.Error("Failed to add finalizer to PKIProfile")
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		removeFinalizer := false
		if err, removeFinalizer = r.DeleteObject(ctx, pki); err != nil {
			return ctrl.Result{}, err
		}
		if removeFinalizer {
			controllerutil.RemoveFinalizer(pki, crdlib.PKIProfileFinalizer)
			if err := r.Patch(ctx, pki, patch); err != nil {
				return ctrl.Result{}, err
			}
		} else {
			if err := r.Status().Update(ctx, pki); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: crdlib.RequeueInterval}, nil
		}
		r.EventRecorder.Event(pki, corev1.EventTypeNormal, "Deleted", "PKIProfile deleted successfully from Avi Controller")
		log.Info("successfully deleted PKIProfile")
		return ctrl.Result{}, nil
	}

	if err := r.ReconcileIfRequired(ctx, pki); err != nil {
		// Check if the error is retryable
		if !controllerutils.IsRetryableError(err) {
			// Update status with non-retryable error condition and don't return error (to avoid requeue)
			controllerutils.UpdateStatusWithNonRetryableError(ctx, r, pki, err, "PKIProfile")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *PKIProfileReconciler) SetStatus(ctx context.Context, pki *akov1alpha1.PKIProfile, conditionType akov1alpha1.ObjectConditionType, conditionStatus metav1.ConditionStatus, reason akov1alpha1.ObjectConditionReason, message string) error {
	statusManager := &controllerutils.StatusManager{
		Client:        r.Client,
		EventRecorder: r.EventRecorder,
	}
	return statusManager.SetStatus(ctx, pki, conditionType, conditionStatus, reason, message)
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
		if err := r.AviClient.AviSessionDelete(utils.GetUriEncoded(fmt.Sprintf("%s/%s", crdlib.PKIProfileURL, pki.Status.UUID)), nil, nil, session.SetOptTenant(pki.Status.Tenant)); err != nil {
			// Handle 404 as success case - object doesn't exist, which is the desired state for delete
			if aviError, ok := err.(session.AviError); ok {
				switch aviError.HttpStatusCode {
				case 404:
					log.Info("PKIProfile not found on Avi Controller (404), treating as successful deletion")
					return nil, true
				case 403:
					log.Errorf("PKIProfile cannot be deleted. %s", aviError.Error())
					if statusErr := r.SetStatus(ctx, pki, akov1alpha1.ObjectConditionProgrammed, metav1.ConditionFalse, akov1alpha1.ObjectReasonDeletionSkipped, controllerutils.ParseAviErrorMessage(*aviError.Message)); statusErr != nil {
						return statusErr, false
					}
					return nil, false
				}
			}
			log.Errorf("error deleting pki profile: %s", err.Error())
			if statusErr := r.SetStatus(ctx, pki, akov1alpha1.ObjectConditionProgrammed, metav1.ConditionFalse, akov1alpha1.ObjectReasonDeletionFailed, fmt.Sprintf("Failed to delete PKIProfile from Avi Controller: %v", err)); statusErr != nil {
				return statusErr, false
			}
			return err, false
		}
	} else {
		log.Warn("error deleting pki profile. uuid not present. possibly avi pki profile object not created")
		if err := r.SetStatus(ctx, pki, akov1alpha1.ObjectConditionProgrammed, metav1.ConditionTrue, akov1alpha1.ObjectReasonDeletionSkipped, "UUID not present, PKIProfile may not have been created on Avi Controller"); err != nil {
			return err, false
		}
	}
	return nil, true
}

// TODO: Make this function generic
func (r *PKIProfileReconciler) ReconcileIfRequired(ctx context.Context, pki *akov1alpha1.PKIProfile) error {
	log := utils.LoggerFromContext(ctx)
	namespaceTenant, err := controllerutils.GetTenantInNamespace(ctx, r.Client, pki.Namespace)
	if err != nil {
		if statusErr := r.StatusManager.SetStatus(ctx, pki, akov1alpha1.ObjectConditionProgrammed, metav1.ConditionFalse, akov1alpha1.ObjectReasonCreationFailed, fmt.Sprintf("Error getting tenant in namespace %s: %s", pki.Namespace, err.Error())); statusErr != nil {
			return statusErr
		}
		log.Errorf("error getting tenant in namespace %s: %s", pki.Namespace, err.Error())
		return err
	}
	// Check if tenant in status differs from tenant in namespace annotation
	// Only trigger tenant mismatch if status has a tenant set (not for new resources)
	if pki.Status.Tenant != "" && pki.Status.Tenant != namespaceTenant {
		log.Infof("Tenant update detected. Status tenant: %s, Namespace tenant: %s. Deleting PKIProfile from AVI.", pki.Status.Tenant, namespaceTenant)
		err, _ := r.DeleteObject(ctx, pki)
		if err != nil {
			log.Errorf("Failed to delete PKIProfile due to error: %s", err.Error())
			if statusErr := r.SetStatus(ctx, pki, akov1alpha1.ObjectConditionProgrammed, metav1.ConditionFalse, akov1alpha1.ObjectReasonDeletionFailed, fmt.Sprintf("Failed to delete PKIProfile due to error: %v", err)); statusErr != nil {
				return statusErr
			}
			return err
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
	name := crdlib.GetObjectName(pki.Namespace, pki.Name)
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
		resp, err := r.createPKIProfile(ctx, pkiReq, namespaceTenant)
		if err != nil {
			log.Errorf("error creating pki profile: %s", err.Error())
			if statusErr := r.SetStatus(ctx, pki, akov1alpha1.ObjectConditionProgrammed, metav1.ConditionFalse, akov1alpha1.ObjectReasonCreationFailed, fmt.Sprintf("Failed to create PKIProfile on Avi Controller: %v", err)); statusErr != nil {
				return statusErr
			}
			return err
		}
		uuid, err := extractUUID(resp)
		if err != nil {
			log.Errorf("error extracting UUID from pki profile: %s", err.Error())
			if statusErr := r.SetStatus(ctx, pki, akov1alpha1.ObjectConditionProgrammed, metav1.ConditionFalse, akov1alpha1.ObjectReasonUUIDExtractionFailed, fmt.Sprintf("Failed to extract UUID: %v", err)); statusErr != nil {
				return statusErr
			}
			return err
		}
		pki.Status.UUID = uuid
		pki.Status.BackendObjectName = *pkiReq.Name
		pki.Status.Tenant = namespaceTenant
		if err := r.SetStatus(ctx, pki, akov1alpha1.ObjectConditionProgrammed, metav1.ConditionTrue, akov1alpha1.ObjectReasonCreated, "PKIProfile created successfully on Avi Controller"); err != nil {
			return err
		}
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
						return nil
					}
				}
			}
		}
		log.Debugf("overwriting pki profile")
		resp := map[string]interface{}{}
		if err := r.AviClient.AviSessionPut(utils.GetUriEncoded(fmt.Sprintf("%s/%s", crdlib.PKIProfileURL, pki.Status.UUID)), pkiReq, &resp, session.SetOptTenant(namespaceTenant)); err != nil {
			log.Errorf("error updating pki profile %s", err.Error())
			if statusErr := r.SetStatus(ctx, pki, akov1alpha1.ObjectConditionProgrammed, metav1.ConditionFalse, akov1alpha1.ObjectReasonUpdateFailed, fmt.Sprintf("Failed to update PKIProfile on Avi Controller: %v", err)); statusErr != nil {
				return statusErr
			}
			return err
		}
		pki.Status.BackendObjectName = *pkiReq.Name
		pki.Status.Tenant = namespaceTenant
		if err := r.SetStatus(ctx, pki, akov1alpha1.ObjectConditionProgrammed, metav1.ConditionTrue, akov1alpha1.ObjectReasonUpdated, "PKIProfile updated successfully on Avi Controller"); err != nil {
			return err
		}
	}
	return nil
}

// createPKIProfile will attempt to create a PKI profile, if it already exists, it will return an object which contains the uuid
func (r *PKIProfileReconciler) createPKIProfile(ctx context.Context, pkiReq *models.PKIprofile, tenant string) (map[string]interface{}, error) {
	log := utils.LoggerFromContext(ctx)
	resp := map[string]interface{}{}
	if err := r.AviClient.AviSessionPost(utils.GetUriEncoded(crdlib.PKIProfileURL), pkiReq, &resp, session.SetOptTenant(tenant)); err != nil {
		log.Errorf("error posting pki profile: %s", err.Error())
		if aviError, ok := err.(session.AviError); ok {
			if aviError.HttpStatusCode == http.StatusConflict && strings.Contains(aviError.Error(), "already exists") {
				log.Info("pki profile already exists. trying to get uuid")
				err := r.AviClient.AviSessionGet(utils.GetUriEncoded(fmt.Sprintf("%s?name=%s", crdlib.PKIProfileURL, *pkiReq.Name)), &resp, session.SetOptTenant(tenant))
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
				if err := r.AviClient.AviSessionPut(utils.GetUriEncoded(fmt.Sprintf("%s/%s", crdlib.PKIProfileURL, uuid)), pkiReq, &resp, session.SetOptTenant(tenant)); err != nil {
					log.Errorf("error updating pki profile", err.Error())
					return nil, err
				}
				return resp, nil
			}
		}
		return nil, err
	}
	return resp, nil
}

// CreateNewHealthMonitorControllerAndSetupWithManager creates a new HealthMonitor controller,
// registers it with the Secret Controller, and sets it up with the manager
func CreateNewPKIProfileControllerAndSetupWithManager(
	mgr manager.Manager,
	aviClient avisession.AviClientInterface,
	cache cache.CacheOperation,
	clusterName string,
	secretReconciler *SecretReconciler,
) (*PKIProfileReconciler, error) {
	// Create the controller
	reconciler := &PKIProfileReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		AviClient:     aviClient,
		Cache:         cache,
		EventRecorder: mgr.GetEventRecorderFor(pkiprofileControllerName),
		Logger:        utils.AviLog.WithName("pkiprofile"),
		ClusterName:   clusterName,
	}

	// Register with Secret Controller
	if err := secretReconciler.RegisterReconciler(reconciler); err != nil {
		return nil, err
	}

	// Setup with manager
	if err := reconciler.SetupWithManager(mgr); err != nil {
		return nil, err
	}

	return reconciler, nil
}
