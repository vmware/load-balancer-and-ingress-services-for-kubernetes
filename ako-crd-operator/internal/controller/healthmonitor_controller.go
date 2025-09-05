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
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/api/v1alpha1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/constants"
	akoerrors "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/errors"
	avisession "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/session"
	controllerutils "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/utils"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// HealthMonitorReconciler reconciles a HealthMonitor object
type HealthMonitorReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	Cache         cache.CacheOperation
	EventRecorder record.EventRecorder
	Logger        *utils.AviLogger
	ClusterName   string
	// AviClientForTesting is used only for testing to inject mock AVI client
	AviClientForTesting avisession.AviClientInterface
}

// GetLogger returns the logger for the reconciler to implement NamespaceHandler interface
func (r *HealthMonitorReconciler) GetLogger() *utils.AviLogger {
	return r.Logger
}

// getAviClient returns the AVI client for this reconciler
// Uses test client if available, otherwise gets from singleton session
func (r *HealthMonitorReconciler) getAviClient() avisession.AviClientInterface {
	if r.AviClientForTesting != nil {
		return r.AviClientForTesting
	}
	sessionObj := avisession.GetGlobalSession()
	return avisession.NewAviSessionClient(sessionObj.GetAviClients().AviClient[0])
}

type HealthMonitorRequest struct {
	Name string `json:"name"`
	akov1alpha1.HealthMonitorSpec
	Markers         []*models.RoleFilterMatchLabel `json:"markers,omitempty"`
	AuthCredentials *HealthMonitorAuthRequest      `json:"authentication,omitempty"`
}

type HealthMonitorAuthRequest struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// +kubebuilder:rbac:groups=ako.vmware.com,resources=healthmonitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ako.vmware.com,resources=healthmonitors/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ako.vmware.com,resources=healthmonitors/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the HealthMonitor object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.2/pkg/reconcile
func (r *HealthMonitorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Logger.WithValues("name", req.Name, "namespace", req.Namespace, "traceID", uuid.New().String())
	ctx = utils.LoggerWithContext(ctx, log)

	log.Debug("Reconciling HealthMonitor")
	defer log.Debug("Reconciled HealthMonitor")
	hm := &akov1alpha1.HealthMonitor{}
	err := r.Client.Get(ctx, req.NamespacedName, hm)
	if err != nil {
		if k8serror.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error("Failed to get HealthMonitor")
		return ctrl.Result{}, err
	}
	if hm.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(hm, constants.HealthMonitorFinalizer) {
			controllerutil.AddFinalizer(hm, constants.HealthMonitorFinalizer)
			if err := r.Update(ctx, hm); err != nil {
				log.Error("Failed to add finalizer to HealthMonitor")
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		removeFinalizer := false
		if err, removeFinalizer = r.DeleteObject(ctx, hm); err != nil {
			return ctrl.Result{}, err
		}
		if removeFinalizer {
			controllerutil.RemoveFinalizer(hm, constants.HealthMonitorFinalizer)
		} else {
			if err := r.Status().Update(ctx, hm); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: constants.RequeueInterval}, nil
		}
		if err := r.Update(ctx, hm); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	if err := r.ReconcileIfRequired(ctx, hm); err != nil {
		// Check if the error is retryable
		if !controllerutils.IsRetryableError(err) {
			// Update status with non-retryable error condition and don't return error (to avoid requeue)
			controllerutils.UpdateStatusWithNonRetryableError(ctx, r, hm, err, "HealthMonitor")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// healthMonitorSecretRefIndexer extracts secret references from HealthMonitor objects for efficient lookups
func healthMonitorSecretRefIndexer(obj client.Object) []string {
	hm := obj.(*akov1alpha1.HealthMonitor)
	if hm.Spec.Authentication == nil || hm.Spec.Authentication.SecretRef == "" {
		return nil
	}
	return []string{hm.Spec.Authentication.SecretRef}
}

// healthMonitorSecretHandler finds all HealthMonitors that reference a given secret and creates reconcile requests for them
func (r *HealthMonitorReconciler) healthMonitorSecretHandler(secretRefField string) func(context.Context, client.Object) []reconcile.Request {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		secret := obj.(*corev1.Secret)
		hmList := &akov1alpha1.HealthMonitorList{}

		// Use the index to find all HealthMonitors that reference this secret.
		err := r.List(ctx, hmList, &client.ListOptions{
			FieldSelector: fields.OneTermEqualSelector(secretRefField, secret.Name),
			Namespace:     secret.Namespace,
		})
		if err != nil {
			r.Logger.Errorf("failed to list HealthMonitors for secret %s/%s: %v", secret.Namespace, secret.Name, err)
			return []reconcile.Request{}
		}

		requests := make([]reconcile.Request, 0, len(hmList.Items))
		for _, item := range hmList.Items {
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      item.Name,
					Namespace: item.Namespace,
				},
			})
		}
		return requests
	}
}

// healthMonitorSecretPredicate ensures only secrets of the correct type are watched by the controller
func healthMonitorSecretPredicate(obj client.Object) bool {
	if obj == nil {
		return false
	}
	secret, ok := obj.(*corev1.Secret)
	if !ok {
		return false
	}
	return secret.Type == constants.HealthMonitorSecretType
}

// SetupWithManager sets up the controller with the Manager.
func (r *HealthMonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	secretRefField := ".spec.authentication.secret_ref"

	// Create an index on the secretRef field. This allows to quickly look up
	// HealthMonitors that reference a given secret.
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &akov1alpha1.HealthMonitor{}, secretRefField, healthMonitorSecretRefIndexer); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&akov1alpha1.HealthMonitor{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Named("healthmonitor").
		// Watch for changes to Secrets and enqueue reconcile requests for the HealthMonitors that reference them.
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.healthMonitorSecretHandler(secretRefField)),
			// Use a predicate to only watch for secrets of the correct type.
			builder.WithPredicates(predicate.NewPredicateFuncs(healthMonitorSecretPredicate)),
		).
		Watches(
			&corev1.Namespace{},
			handler.EnqueueRequestsFromMapFunc(controllerutils.CreateGenericNamespaceHandler(
				r,
				"healthmonitor",
				func() *akov1alpha1.HealthMonitorList { return &akov1alpha1.HealthMonitorList{} },
				func(list *akov1alpha1.HealthMonitorList) []client.Object {
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

// DeleteObject deletes the HealthMonitor from Avi Controller and returns (error, bool)
// The boolean indicates whether the finalizer should be removed (true) or kept (false)
func (r *HealthMonitorReconciler) DeleteObject(ctx context.Context, hm *akov1alpha1.HealthMonitor) (error, bool) {
	log := utils.LoggerFromContext(ctx)
	if hm.Status.UUID != "" && hm.Status.Tenant != "" {
		aviClient := r.getAviClient()
		if err := aviClient.AviSessionDelete(fmt.Sprintf("%s/%s", constants.HealthMonitorURL, hm.Status.UUID), nil, nil, session.SetOptTenant(hm.Status.Tenant)); err != nil {
			// Handle 404 as success case - object doesn't exist, which is the desired state for delete
			if aviError, ok := err.(session.AviError); ok {
				switch aviError.HttpStatusCode {
				case 404:
					log.Info("HealthMonitor not found on Avi Controller (404), treating as successful deletion")
					return nil, true
				case 403:
					log.Errorf("HealthMonitor is being referred by other objects, cannot be deleted. %s", aviError.Error())
					r.EventRecorder.Event(hm, corev1.EventTypeWarning, "DeletionSkipped", aviError.Error())
					hm.Status.Conditions = controllerutils.SetCondition(hm.Status.Conditions, metav1.Condition{
						Type:               "Deleted",
						Status:             metav1.ConditionFalse,
						LastTransitionTime: metav1.Time{Time: time.Now().UTC()},
						Reason:             "DeletionSkipped",
						Message:            controllerutils.ParseAviErrorMessage(*aviError.Message),
					})
					return nil, false
				}
			}
			log.Errorf("error deleting healthmonitor: %s", err.Error())
			r.EventRecorder.Event(hm, corev1.EventTypeWarning, "DeletionFailed", fmt.Sprintf("Failed to delete HealthMonitor from Avi Controller: %v", err))
			return err, false
		}
	} else {
		r.EventRecorder.Event(hm, corev1.EventTypeWarning, "DeletionSkipped", "UUID not present, HealthMonitor may not have been created on Avi Controller")
		log.Warn("error deleting healthmonitor. uuid not present. possibly avi healthmonitor object not created")
	}
	r.EventRecorder.Event(hm, corev1.EventTypeNormal, "Deleted", "HealthMonitor deleted successfully from Avi Controller")
	log.Info("successfully deleted healthmonitor")
	return nil, true
}

// TODO: Make this function generic
func (r *HealthMonitorReconciler) ReconcileIfRequired(ctx context.Context, hm *akov1alpha1.HealthMonitor) error {
	log := utils.LoggerFromContext(ctx)
	namespaceTenant, err := controllerutils.GetTenantInNamespace(ctx, r.Client, hm.Namespace)
	if err != nil {
		log.Errorf("error getting tenant in namespace: %s", err.Error())
		return err
	}
	// Check if tenant in status differs from tenant in namespace annotation
	// Only trigger tenant update if status has a tenant set (not for new resources)
	if hm.Status.Tenant != "" && hm.Status.Tenant != namespaceTenant {
		log.Infof("Tenant update detected. Status tenant: %s, Namespace tenant: %s. Deleting HealthMonitor from AVI.", hm.Status.Tenant, namespaceTenant)
		err, _ := r.DeleteObject(ctx, hm)
		if err != nil {
			log.Errorf("Failed to delete HealthMonitor due to error: %s", err.Error())
			r.EventRecorder.Event(hm, corev1.EventTypeWarning, "DeletionFailed", fmt.Sprintf("Failed to delete HealthMonitor due to error: %v", err))
			return err
		}
		// Clear the status to force recreation with correct tenant
		hm.Status = akov1alpha1.HealthMonitorStatus{}
		log.Info("HealthMonitor deleted from AVI due to tenant update, status cleared for recreation")
	}

	hmReq := &HealthMonitorRequest{
		Name:              fmt.Sprintf("%s-%s-%s", r.ClusterName, hm.Namespace, hm.Name),
		HealthMonitorSpec: hm.Spec,
		Markers:           controllerutils.CreateMarkers(r.ClusterName, hm.Namespace),
	}

	reconcile, dependencyChecksum, err := r.resolveRefsAndCheckDependencies(ctx, hm.Namespace, hm, hmReq)
	if err != nil {
		log.Errorf("error resolving refs: %s", err.Error())
		hm.Status.DependencySum = 0
		r.EventRecorder.Event(hm, corev1.EventTypeWarning, "RefResolutionFailed", fmt.Sprintf("Failed to resolve refs: %v", err))
		return err
	}

	// this is a POST Call
	if hm.Status.UUID == "" {
		resp, err := r.createHealthMonitor(ctx, hmReq, hm, namespaceTenant)
		if err != nil {
			r.EventRecorder.Event(hm, corev1.EventTypeWarning, "CreationFailed", fmt.Sprintf("Failed to create HealthMonitor on Avi Controller: %v", err))
			log.Errorf("error creating healthmonitor: %s", err.Error())
			return err
		}
		uuid, err := extractUUID(resp)
		if err != nil {
			r.EventRecorder.Event(hm, corev1.EventTypeWarning, "UUIDExtractionFailed", fmt.Sprintf("Failed to extract UUID: %v", err))
			log.Errorf("error extracting UUID from healthmonitor: %s", err.Error())
			return err
		}
		hm.Status.UUID = uuid
		r.EventRecorder.Event(hm, corev1.EventTypeNormal, "Created", "HealthMonitor created successfully on Avi Controller")
	} else {
		// this is a PUT Call
		// check if no op by checking generation and checksums
		if !reconcile {
			// if no op from kubernetes side, check if op required from OOB changes by checking lastModified timestamp
			if hm.Status.LastUpdated != nil {
				dataMap, ok := r.Cache.GetObjectByUUID(ctx, hm.Status.UUID)
				if ok {
					if dataMap.GetLastModifiedTimeStamp().Before(hm.Status.LastUpdated.Time) {
						log.Debug("no op for healthmonitor")
						return nil
					}
				}
			}
			log.Debug("overwriting healthmonitor")
		}
		resp := map[string]interface{}{}
		aviClient := r.getAviClient()
		if err := aviClient.AviSessionPut(utils.GetUriEncoded(fmt.Sprintf("%s/%s", constants.HealthMonitorURL, hm.Status.UUID)), hmReq, &resp); err != nil {
			log.Errorf("error updating healthmonitor %s", err.Error())
			r.EventRecorder.Event(hm, corev1.EventTypeWarning, "UpdateFailed", fmt.Sprintf("Failed to update HealthMonitor on Avi Controller: %v", err))
			return err
		}
		hm.Status.Conditions = controllerutils.SetCondition(hm.Status.Conditions, metav1.Condition{
			Type:               "Ready",
			Status:             metav1.ConditionTrue,
			LastTransitionTime: metav1.Time{Time: time.Now().UTC()},
			Reason:             "Updated",
			Message:            "HealthMonitor updated successfully on Avi Controller",
		})
		r.EventRecorder.Event(hm, corev1.EventTypeNormal, "Updated", "HealthMonitor updated successfully on Avi Controller")
		log.Info("successfully updated healthmonitor")
	}
	hm.Status.DependencySum = dependencyChecksum
	hm.Status.BackendObjectName = hmReq.Name
	hm.Status.Tenant = namespaceTenant
	lastUpdated := metav1.Time{Time: time.Now().UTC()}
	hm.Status.LastUpdated = &lastUpdated
	hm.Status.ObservedGeneration = hm.Generation
	if err := r.Status().Update(ctx, hm); err != nil {
		r.EventRecorder.Event(hm, corev1.EventTypeWarning, "StatusUpdateFailed", fmt.Sprintf("Failed to update HealthMonitor status: %v", err))
		log.Errorf("unable to update healthmonitor status: %s", err.Error())
		return err
	}
	return nil
}

// createHealthMonitor will attempt to create a health monitor, if it already exists, it will return an object which contains the uuid
func (r *HealthMonitorReconciler) createHealthMonitor(ctx context.Context, hmReq *HealthMonitorRequest, hm *akov1alpha1.HealthMonitor, tenant string) (map[string]interface{}, error) {
	log := utils.LoggerFromContext(ctx)
	resp := map[string]interface{}{}
	aviClient := r.getAviClient()
	if err := aviClient.AviSessionPost(utils.GetUriEncoded(constants.HealthMonitorURL), hmReq, &resp, session.SetOptTenant(tenant)); err != nil {
		log.Errorf("error posting healthmonitor: %s", err.Error())
		if aviError, ok := err.(session.AviError); ok {
			if aviError.HttpStatusCode == http.StatusConflict && strings.Contains(aviError.Error(), "already exists") {
				log.Info("healthmonitor already exists. trying to get uuid")
				err := aviClient.AviSessionGet(utils.GetUriEncoded(fmt.Sprintf("%s?name=%s", constants.HealthMonitorURL, hmReq.Name)), &resp)
				if err != nil {
					log.Errorf("error getting healthmonitor: %s", err.Error())
					return nil, err
				}
				uuid, err := extractUUID(resp)
				if err != nil {
					log.Errorf("error extracting UUID from healthmonitor: %s", err.Error())
					return nil, err
				}
				log.Info("updating healthmonitor")
				if err := aviClient.AviSessionPut(utils.GetUriEncoded(fmt.Sprintf("%s/%s", constants.HealthMonitorURL, uuid)), hmReq, &resp, session.SetOptTenant(tenant)); err != nil {
					log.Errorf("error updating healthmonitor: %s", err.Error())
					return nil, err
				}
				hm.Status.Conditions = controllerutils.SetCondition(hm.Status.Conditions, metav1.Condition{
					Type:               "Ready",
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.Time{Time: time.Now().UTC()},
					Reason:             "Updated",
					Message:            "HealthMonitor updated successfully on Avi Controller",
				})
				return resp, nil
			}
		}
		return nil, err
	}
	hm.Status.Conditions = controllerutils.SetCondition(hm.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Time{Time: time.Now().UTC()},
		Reason:             "Created",
		Message:            "HealthMonitor created successfully on Avi Controller",
	})
	log.Info("healthmonitor successfully created")
	return resp, nil
}

// extractUUID extracts the UUID from resp object
func extractUUID(resp map[string]interface{}) (string, error) {
	// Extract the results array
	results, ok := resp["results"].([]interface{})
	if !ok {
		// resp could be from POST call
		if uuid, ok := resp["uuid"].(string); ok {
			return uuid, nil
		}
		return "", errors.New("'results' not found or not an array")
	}

	// Check if the results array is empty
	if len(results) == 0 {
		return "", errors.New("'results' array is empty")
	}

	// Extract the first element from the results array (which is a map)
	firstResult, ok := results[0].(map[string]interface{})
	if !ok {
		return "", errors.New("first element in 'results' is not a map")
	}

	// Extract the UUID from the first result
	uuid, ok := firstResult["uuid"].(string)
	if !ok {
		return "", errors.New("'uuid' not found or not a string")
	}
	return uuid, nil
}

func (r *HealthMonitorReconciler) resolveRefsAndCheckDependencies(ctx context.Context, namespace string, hm *akov1alpha1.HealthMonitor, hmReq *HealthMonitorRequest) (bool, uint32, error) {
	reconcile := false
	secretResourceVersion, err := r.resolveSecret(ctx, namespace, hmReq)
	if err != nil {
		return reconcile, 0, err
	}
	dependencyChecksum := generateChecksum(secretResourceVersion)
	if hm.GetGeneration() != hm.Status.ObservedGeneration || hm.Status.DependencySum != dependencyChecksum {
		reconcile = true
	}
	return reconcile, dependencyChecksum, nil
}

func generateChecksum(checkSumFields ...string) uint32 {
	checksumString := []string{}
	for _, checksum := range checkSumFields {
		if checksum == "" {
			continue
		}
		checksumString = append(checksumString, checksum)
	}
	if len(checksumString) == 0 {
		return 0
	}
	return utils.Hash(strings.Join(checksumString, ":"))
}

func (r *HealthMonitorReconciler) resolveSecret(ctx context.Context, namespace string, hmReq *HealthMonitorRequest) (string, error) {
	log := utils.LoggerFromContext(ctx)
	if hmReq.Authentication == nil || hmReq.Authentication.SecretRef == "" {
		return "", nil
	}

	secret := &corev1.Secret{}
	if err := r.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: hmReq.Authentication.SecretRef}, secret); err != nil {
		log.Errorf("error getting secret: %s", err.Error())
		return "", akoerrors.AKOCRDOperatorError{
			HttpStatusCode: http.StatusBadRequest,
			Reason:         "UnresolvedRef",
			Message:        fmt.Sprintf("Secret %s not found", hmReq.Authentication.SecretRef),
		}
	}
	if secret.Type != constants.HealthMonitorSecretType {
		log.Errorf("secret is not of %s type", constants.HealthMonitorSecretType)
		return "", akoerrors.AKOCRDOperatorError{
			HttpStatusCode: http.StatusBadRequest,
			Reason:         "ConfigurationError",
			Message:        fmt.Sprintf("Secret %s is not of type %s", hmReq.Authentication.SecretRef, constants.HealthMonitorSecretType),
		}
	}

	if secret.Data == nil {
		log.Errorf("secret data is nil")
		return "", akoerrors.AKOCRDOperatorError{
			HttpStatusCode: http.StatusBadRequest,
			Reason:         "ConfigurationError",
			Message:        fmt.Sprintf("Secret data is nil in secret %s", hmReq.Authentication.SecretRef),
		}
	}
	username, ok := secret.Data["username"]
	if !ok {
		log.Errorf("username not found in secret")
		return "", akoerrors.AKOCRDOperatorError{
			HttpStatusCode: http.StatusBadRequest,
			Reason:         "ConfigurationError",
			Message:        fmt.Sprintf("Username not found in secret %s", hmReq.Authentication.SecretRef),
		}
	}
	password, ok := secret.Data["password"]
	if !ok {
		log.Errorf("password not found in secret")
		return "", akoerrors.AKOCRDOperatorError{
			HttpStatusCode: http.StatusBadRequest,
			Reason:         "ConfigurationError",
			Message:        fmt.Sprintf("Password not found in secret %s", hmReq.Authentication.SecretRef),
		}
	}
	hmReq.AuthCredentials = &HealthMonitorAuthRequest{
		Username: string(username),
		Password: string(password),
	}
	// explicitly set authentication with secretRef to nil to avoid sending it to the controller
	hmReq.Authentication = nil
	return secret.ResourceVersion, nil
}
