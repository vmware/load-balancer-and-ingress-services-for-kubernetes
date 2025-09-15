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

	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/google/uuid"
	"github.com/vmware/alb-sdk/go/session"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/api/v1alpha1"
	controllerutils "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/utils"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/constants"
	avisession "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/session"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

const (
	routeBackendExtensionControllerName = "routebackendextension-controller"
)

// RouteBackendExtensionReconciler reconciles a RouteBackendExtension object
type RouteBackendExtensionReconciler struct {
	client.Client
	AviClient     avisession.AviClientInterface
	Scheme        *runtime.Scheme
	Cache         cache.CacheOperation
	Logger        *utils.AviLogger
	EventRecorder record.EventRecorder
	ClusterName   string
}

// GetLogger returns the logger for the reconciler to implement NamespaceHandler interface
func (r *RouteBackendExtensionReconciler) GetLogger() *utils.AviLogger {
	return r.Logger
}

// UpdateAviClient implements AviClientReconciler to update the AVI client when credentials change
func (r *RouteBackendExtensionReconciler) UpdateAviClient(client avisession.AviClientInterface) error {
	r.Logger.Info("Updating AVI client for RouteBackendExtension controller")
	r.AviClient = client
	return nil
}

// GetReconcilerName implements AviClientReconciler to return the reconciler name
func (r *RouteBackendExtensionReconciler) GetReconcilerName() string {
	return routeBackendExtensionControllerName
}

// +kubebuilder:rbac:groups=ako.vmware.com,resources=routebackendextensions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ako.vmware.com,resources=routebackendextensions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ako.vmware.com,resources=pkiprofiles,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RouteBackendExtension object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *RouteBackendExtensionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Logger.WithValues("name", req.Name, "namespace", req.Namespace, "traceID", uuid.New().String())
	ctx = utils.LoggerWithContext(ctx, log)
	log.Debug("Reconciling RouteBackendExtension CRD")
	defer log.Debug("Reconciled RouteBackendExtension CRD")
	rbe := &akov1alpha1.RouteBackendExtension{}
	err := r.Client.Get(ctx, req.NamespacedName, rbe)
	if err != nil {
		if k8serror.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error("Failed to get RouteBackendExtension CRD")
		return ctrl.Result{}, err
	}

	if !rbe.DeletionTimestamp.IsZero() {
		// The object is being deleted
		r.EventRecorder.Event(rbe, corev1.EventTypeNormal, "Deleted", "RouteBackendExtension CRD deleted successfully from Avi Controller")
		log.Info("Successfully deleted RouteBackendExtension CRD")
		return ctrl.Result{}, nil
	}
	// create or update - validate the object
	// When this CRD will have other crd object, this logic will change
	if err := r.ValidatedObject(ctx, rbe); err != nil {
		// Check if the error is retryable
		if controllerutils.IsRetryableError(err) {
			// For 404(object not found) also, we are not retrying. So user has to update the object again to trigger
			// processing.
			// other way to retry for certain number of times for each object and then stop
			return ctrl.Result{RequeueAfter: constants.RequeueInterval}, err
		}
	}
	return ctrl.Result{}, nil
}

const (
	// PKIProfileIndexKey is the index key for RouteBackendExtensions by PKIProfile reference
	PKIProfileIndexKey = "spec.pkiProfile.name"
)

// SetupWithManager sets up the controller with the Manager.
func (r *RouteBackendExtensionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Add indexer for RouteBackendExtension by PKIProfile reference
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &akov1alpha1.RouteBackendExtension{}, PKIProfileIndexKey, func(rawObj client.Object) []string {
		rbe := rawObj.(*akov1alpha1.RouteBackendExtension)
		if rbe.Spec.BackendTLS.PKIProfile != nil && rbe.Spec.BackendTLS.PKIProfile.Kind == akov1alpha1.ObjectKindCRD {
			return []string{rbe.Spec.BackendTLS.PKIProfile.Name}
		}
		return nil
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&akov1alpha1.RouteBackendExtension{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Named("routebackendextension").
		Watches(
			&corev1.Namespace{},
			handler.EnqueueRequestsFromMapFunc(controllerutils.CreateGenericNamespaceHandler(
				r,
				"routebackendextension",
				func() *akov1alpha1.RouteBackendExtensionList { return &akov1alpha1.RouteBackendExtensionList{} },
				func(list *akov1alpha1.RouteBackendExtensionList) []client.Object {
					objects := make([]client.Object, len(list.Items))
					for i, item := range list.Items {
						objects[i] = &item
					}
					return objects
				},
			)),
			builder.WithPredicates(controllerutils.TenantAnnotationNamespacePredicate()),
		).
		Watches(
			&akov1alpha1.PKIProfile{},
			handler.EnqueueRequestsFromMapFunc(r.findRouteBackendExtensionsForPKIProfile),
		).
		Complete(r)
}

// findRouteBackendExtensionsForPKIProfile finds all RouteBackendExtensions that reference the given PKIProfile
func (r *RouteBackendExtensionReconciler) findRouteBackendExtensionsForPKIProfile(ctx context.Context, obj client.Object) []ctrl.Request {
	pkiProfile, ok := obj.(*akov1alpha1.PKIProfile)
	if !ok {
		return nil
	}

	var rbeList akov1alpha1.RouteBackendExtensionList

	// Use indexer to find RouteBackendExtensions that reference this PKIProfile
	if err := r.Client.List(ctx, &rbeList,
		client.InNamespace(pkiProfile.Namespace),
		client.MatchingFields{PKIProfileIndexKey: pkiProfile.Name}); err != nil {
		r.Logger.Errorf("Failed to list RouteBackendExtensions by PKIProfile index in namespace %s: %v", pkiProfile.Namespace, err)
		return nil
	}

	// Pre-allocate requests slice with the expected capacity
	requests := make([]ctrl.Request, 0, len(rbeList.Items))

	// Convert found RouteBackendExtensions to reconcile requests
	for _, rbe := range rbeList.Items {
		requests = append(requests, ctrl.Request{
			NamespacedName: client.ObjectKey{
				Namespace: rbe.Namespace,
				Name:      rbe.Name,
			},
		})
		r.Logger.Debugf("PKIProfile %s/%s changed, queuing RouteBackendExtension %s/%s for reconciliation",
			pkiProfile.Namespace, pkiProfile.Name, rbe.Namespace, rbe.Name)
	}

	return requests
}

func (r *RouteBackendExtensionReconciler) ValidatedObject(ctx context.Context, rbe *akov1alpha1.RouteBackendExtension) error {
	log := utils.LoggerFromContext(ctx)
	log.Info("Validating RouteBackendExtension CRD")
	resp := map[string]interface{}{}
	tenant, err := controllerutils.GetTenantInNamespace(ctx, r.Client, rbe.Namespace)
	if err != nil {
		log.Errorf("error in getting tenant in namespace: %s", err.Error())
		return err
	}
	for _, hm := range rbe.Spec.HealthMonitor {
		// Check HM Present or not
		uri := fmt.Sprintf("%s?name=%s", constants.HealthMonitorURL, hm.Name)
		err := r.AviClient.AviSessionGet(utils.GetUriEncoded(uri), &resp, session.SetOptTenant(tenant))
		if err != nil {
			// This log message will change in multitenancy
			log.Errorf("error in getting healthmonitor: %s from tenant %s. Err: %s", hm.Name, tenant, err.Error())
			r.SetStatus(rbe, err.Error(), constants.REJECTED)
			return err
		} else if resp == nil {
			log.Errorf("error in getting healthmonitor: : %s from tenant %s. Count: 0.000000", hm.Name, tenant)
			err = fmt.Errorf("error in getting healthmonitor: %s from tenant %s. Object not found", hm.Name, tenant)
			r.SetStatus(rbe, err.Error(), constants.REJECTED)
			return err
		} else if len(resp) == 0 || resp["count"] == nil || resp["count"].(float64) == float64(0) {
			log.Errorf("error in getting healthmonitor: %s from tenant %s. Object not found", hm.Name, tenant)
			err = fmt.Errorf("error in getting healthmonitor: %s from tenant %s. Object not found", hm.Name, tenant)
			r.SetStatus(rbe, err.Error(), constants.REJECTED)
			return err
		}
	}

	// Check PKIProfile if present
	if rbe.Spec.BackendTLS != nil && rbe.Spec.BackendTLS.PKIProfile != nil && rbe.Spec.BackendTLS.PKIProfile.Kind == akov1alpha1.ObjectKindCRD {
		pkiProfile := &akov1alpha1.PKIProfile{}
		pkiProfileKey := client.ObjectKey{
			Namespace: rbe.Namespace,
			Name:      rbe.Spec.BackendTLS.PKIProfile.Name,
		}
		err := r.Client.Get(ctx, pkiProfileKey, pkiProfile)
		if err != nil {
			log.Errorf("error getting PKIProfile %s from namespace %s. Err: %s", rbe.Spec.BackendTLS.PKIProfile.Name, rbe.Namespace, err.Error())
			r.SetStatus(rbe, err.Error(), constants.REJECTED)
			return err
		}

		// Validate that the tenant of the PKI profile matches with the namespace tenant
		if pkiProfile.Status.Tenant != "" && pkiProfile.Status.Tenant != tenant {
			log.Errorf("PKIProfile %s tenant %s does not match namespace %s tenant %s",
				rbe.Spec.BackendTLS.PKIProfile.Name, pkiProfile.Status.Tenant, rbe.Namespace, tenant)
			err = fmt.Errorf("PKIProfile %s tenant %s does not match namespace %s tenant %s",
				rbe.Spec.BackendTLS.PKIProfile.Name, pkiProfile.Status.Tenant, rbe.Namespace, tenant)
			r.SetStatus(rbe, err.Error(), constants.REJECTED)
			return err
		}

		// Check if PKIProfile is ready by looking at its Ready condition
		isReady := false
		for _, condition := range pkiProfile.Status.Conditions {
			if condition.Type == "Ready" && condition.Status == "True" {
				isReady = true
				break
			}
		}

		if !isReady {
			log.Errorf("RBE is rejected beacause PKIProfile %s is not ready in namespace %s", rbe.Spec.BackendTLS.PKIProfile.Name, rbe.Namespace)
			err = fmt.Errorf("RBE is rejected beacause PKIProfile %s is not ready in namespace %s", rbe.Spec.BackendTLS.PKIProfile.Name, rbe.Namespace)
			r.SetStatus(rbe, err.Error(), constants.REJECTED)
			return err
		}
	}
	err = r.SetStatus(rbe, "", constants.ACCEPTED)
	if err != nil {
		log.Errorf("error in setting status: %s", err.Error())
		return err
	}
	log.Info("Accepted. Validated RouteBackendExtension CRD")
	return nil
}

func (r *RouteBackendExtensionReconciler) SetStatus(rbe *akov1alpha1.RouteBackendExtension, error1 string, status string) error {
	rbe.SetRouteBackendExtensionController(constants.AKOCRDController)
	rbe.Status.Error = error1
	rbe.Status.Status = status
	if r.Client == nil {
		log := utils.LoggerFromContext(context.Background())
		log.Errorf("r.Status() returned nil. Cannot update status for RouteBackendExtension: %s/%s", rbe.Namespace, rbe.Name)
		return fmt.Errorf("status client is nil")
	}
	err := r.Status().Update(context.Background(), rbe)
	return err
}

// CreateNewRouteBackendExtensionControllerAndSetupWithManager creates a new RouteBackendExtension controller,
// registers it with the Secret Controller, and sets it up with the manager
func CreateNewRouteBackendExtensionControllerAndSetupWithManager(
	mgr manager.Manager,
	aviClient avisession.AviClientInterface,
	clusterName string,
	secretReconciler *SecretReconciler,
) (*RouteBackendExtensionReconciler, error) {
	// Create the controller
	reconciler := &RouteBackendExtensionReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		AviClient:     aviClient,
		EventRecorder: mgr.GetEventRecorderFor(routeBackendExtensionControllerName),
		Logger:        utils.AviLog.WithName("routebackendextension"),
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
