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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/google/uuid"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/api/v1alpha1"
	controllerutils "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/constants"
	avisession "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/session"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
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

// +kubebuilder:rbac:groups=ako.vmware.com,resources=routebackendextensions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ako.vmware.com,resources=routebackendextensions/status,verbs=get;update;patch

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

// SetupWithManager sets up the controller with the Manager.
func (r *RouteBackendExtensionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&akov1alpha1.RouteBackendExtension{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Named("routebackendextension").
		Complete(r)
}
func (r *RouteBackendExtensionReconciler) ValidatedObject(ctx context.Context, rbe *akov1alpha1.RouteBackendExtension) error {
	log := utils.LoggerFromContext(ctx)
	log.Info("Validating RouteBackendExtension CRD")
	resp := map[string]interface{}{}
	for _, hm := range rbe.Spec.HealthMonitor {
		// Check HM Present or not
		uri := fmt.Sprintf("%s?name=%s", constants.HealthMonitorURL, hm.Name)
		err := r.AviClient.AviSessionGet(utils.GetUriEncoded(uri), &resp)
		if err != nil {
			// This log message will change in multitenancy
			log.Errorf("error in getting healthmonitor: %s from tenant %s. Err: %s", hm.Name, lib.GetTenant(), err.Error())
			r.SetStatus(rbe, err.Error(), constants.REJECTED)
			return err
		} else if len(resp) == 0 || resp["count"] == nil || resp["count"].(float64) == float64(0) {
			log.Errorf("error in getting healthmonitor: %s from tenant %s. Object not found", hm.Name, lib.GetTenant())
			err = fmt.Errorf("error in getting healthmonitor: %s from tenant %s. Object not found", hm.Name, lib.GetTenant())
			r.SetStatus(rbe, err.Error(), constants.REJECTED)
			return err
		}
	}
	err := r.SetStatus(rbe, "", constants.ACCEPTED)
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
