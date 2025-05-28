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
	"github.com/google/uuid"
	"github.com/vmware/alb-sdk/go/clients"
	"github.com/vmware/alb-sdk/go/session"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/api/v1alpha1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/constants"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/logger"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"strings"
	"time"
)

// HealthMonitorReconciler reconciles a HealthMonitor object
type HealthMonitorReconciler struct {
	client.Client
	AviClient *clients.AviClient
	Scheme    *runtime.Scheme
	Cache     cache.CacheOperation
	Logger    *logger.Logger
}

type HealthMonitorRequest struct {
	Name string `json:"name"`
	akov1alpha1.HealthMonitorSpec

	namespace string
}

// +kubebuilder:rbac:groups=ako.vmware.com,resources=healthmonitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ako.vmware.com,resources=healthmonitors/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ako.vmware.com,resources=healthmonitors/finalizers,verbs=update

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
	ctx = logger.WithContext(ctx, log)

	log.Infof("reconciling healthmonitor")
	hm := &akov1alpha1.HealthMonitor{}
	err := r.Client.Get(ctx, req.NamespacedName, hm)
	if err != nil {
		if k8serror.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get HealthMonitor")
		return ctrl.Result{}, err
	}
	if hm.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(hm, constants.HealthMonitorFinalizer) {
			controllerutil.AddFinalizer(hm, constants.HealthMonitorFinalizer)
			if err := r.Update(ctx, hm); err != nil {
				utils.AviLog.Error(err, "Failed to add finalizer to HealthMonitor")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	} else {
		// The object is being deleted
		if err := r.DeleteObject(ctx, hm); err != nil {
			return ctrl.Result{}, err
		}
		controllerutil.RemoveFinalizer(hm, constants.HealthMonitorFinalizer)
		if err := r.Update(ctx, hm); err != nil {
			return ctrl.Result{}, err
		}
		utils.AviLog.Infof("succesfully deleted healthmonitor:[%s/%s]", hm.Namespace, hm.Name)
		return ctrl.Result{}, nil
	}
	if err := r.ReconcileIfRequired(ctx, hm); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HealthMonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&akov1alpha1.HealthMonitor{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Named("healthmonitor").
		Complete(r)
}

// SetupWithManager sets up the controller with the Manager.
func (r *HealthMonitorReconciler) DeleteObject(ctx context.Context, hm *akov1alpha1.HealthMonitor) error {
	if hm.Status.UUID != "" {
		if err := r.AviClient.HealthMonitor.Delete(hm.Status.UUID); err != nil {
			utils.AviLog.Errorf("error deleting healthmonitor: [%s/%s]: %s", hm.Namespace, hm.Name, err.Error())
			return err
		}
	} else {
		utils.AviLog.Warnf("error deleting healthmonitor: [%s/%s]. uuid not present. possibly avi healthmonitor object not created", hm.Namespace, hm.Name)
	}
	return nil
}

// TODO: Make this function generic
func (r *HealthMonitorReconciler) ReconcileIfRequired(ctx context.Context, hm *akov1alpha1.HealthMonitor) error {
	log := logger.FromContext(ctx)
	hmReq := &HealthMonitorRequest{
		hm.Name,
		hm.Spec,
		hm.Namespace,
	}
	// this is a POST Call
	if hm.Status.UUID == "" {
		resp, err := r.createHealthMonitor(ctx, hmReq)
		if err != nil {
			log.Errorf("error creating healthmonitor: %s", err.Error())
			return err
		}
		uuid, err := extractUUID(resp)
		if err != nil {
			log.Errorf("error extracting UUID from healthmonitor: %s", err.Error())
			return err
		}
		hm.Status.UUID = uuid
	} else {
		// this is a PUT Call
		// check if no op by checking generation
		if hm.GetGeneration() == hm.Status.ObservedGeneration {
			// if no op from kubernetes side, check if op required from OOB changes by checking lastModified timestamp
			if hm.Status.LastUpdated != nil {
				dataMap, ok := r.Cache.GetObjectByUUID(hm.Status.UUID)
				if ok {
					if dataMap.GetLastModifiedTimeStamp().Before(hm.Status.LastUpdated.Time) {
						log.Debug("no op for healthmonitor")
						return nil
					}
				}
			}
			log.Debug("overwriting healthmonitor:")
		}
		resp := map[string]interface{}{}
		if err := r.AviClient.AviSession.Put(utils.GetUriEncoded(fmt.Sprintf("%s/%s", constants.HealthMonitorURL, hm.Status.UUID)), hmReq, &resp); err != nil {
			utils.AviLog.Errorf("error updating healthmonitor [%s/%s]: %s", hmReq.namespace, hmReq.Name, err.Error())
			return err
		}
		log.Info("succesfully updated healthmonitor")
	}

	hm.Status.LastUpdated = &metav1.Time{Time: time.Now().UTC()}
	hm.Status.ObservedGeneration = hm.Generation
	if err := r.Status().Update(ctx, hm); err != nil {
		utils.AviLog.Errorf("unable to update healthmonitor status [%s/%s]: %s", hmReq.namespace, hmReq.Name, err.Error())
		return err
	}

	return nil
}

// createHealthMonitor will attempt to create a health monitor, if it already exists, it will return an object which contains the uuid
func (r *HealthMonitorReconciler) createHealthMonitor(ctx context.Context, hmReq *HealthMonitorRequest) (map[string]interface{}, error) {
	resp := map[string]interface{}{}
	if err := r.AviClient.AviSession.Post(utils.GetUriEncoded(constants.HealthMonitorURL), hmReq, &resp); err != nil {
		utils.AviLog.Errorf("error posting healthmonitor: %s", err.Error())
		if aviError, ok := err.(session.AviError); ok {
			if aviError.HttpStatusCode == http.StatusConflict && strings.Contains(aviError.Error(), "already exists") {
				utils.AviLog.Infof("healthmonitor [%s/%s] already exists. trying to get uuid", hmReq.namespace, hmReq.Name)
				err := r.AviClient.AviSession.Get(utils.GetUriEncoded(fmt.Sprintf("%s?name=%s", constants.HealthMonitorURL, hmReq.Name)), &resp)
				if err != nil {
					utils.AviLog.Errorf("error getting healthmonitor [%s/%s]: %s", hmReq.namespace, hmReq.Name, err.Error())
					return nil, err
				}
				uuid, err := extractUUID(resp)
				if err != nil {
					utils.AviLog.Errorf("error extracting UUID from healthmonitor: [%s/%s]: %s", hmReq.namespace, hmReq.Name, err.Error())
					return nil, err
				}
				utils.AviLog.Infof("updating healthmonitor: [%s/%s]", hmReq.namespace, hmReq.Name)
				if err := r.AviClient.AviSession.Put(utils.GetUriEncoded(fmt.Sprintf("%s/%s", constants.HealthMonitorURL, uuid)), hmReq, &resp); err != nil {
					utils.AviLog.Errorf("error updating healthmonitor [%s/%s]: %s", hmReq.namespace, hmReq.Name, err.Error())
					return nil, err
				}
				return resp, nil
			}
		}
		return nil, err
	}
	utils.AviLog.Infof("healthmonitor [%s/%s] succesfully created", hmReq.namespace, hmReq.Name)
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
