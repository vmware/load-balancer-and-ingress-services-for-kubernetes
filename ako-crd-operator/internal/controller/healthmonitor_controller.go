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
	"github.com/vmware/alb-sdk/go/clients"
	"github.com/vmware/alb-sdk/go/session"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/api/v1alpha1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/constants"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// HealthMonitorReconciler reconciles a HealthMonitor object
type HealthMonitorReconciler struct {
	client.Client
	AviClient *clients.AviClient
	Scheme    *runtime.Scheme
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
	hm := &akov1alpha1.HealthMonitor{}
	err := r.Client.Get(ctx, req.NamespacedName, hm)
	if err != nil {
		if k8serror.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		utils.AviLog.Error(err, "Failed to get HealthMonitor: [%s/%s]", req.NamespacedName.Namespace, req.NamespacedName.Name)
		return ctrl.Result{}, err
	}
	if hm.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(hm, constants.HealthMonitorFinalizer) {
			controllerutil.AddFinalizer(hm, constants.HealthMonitorFinalizer)
			if err := r.Update(ctx, hm); err != nil {
				utils.AviLog.Error(err, "Failed to add finalizer to HealthMonitor: [%s/%s]", req.NamespacedName.Namespace, req.NamespacedName.Name)
				return ctrl.Result{}, err
			}
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
		return ctrl.Result{}, nil
	}
	if hm.Spec.Name == "" {
		hm.Spec.Name = hm.Name
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
		Named("healthmonitor").
		Complete(r)
}

// SetupWithManager sets up the controller with the Manager.
func (r *HealthMonitorReconciler) DeleteObject(ctx context.Context, hm *akov1alpha1.HealthMonitor) error {
	if hm.Status.Uuid != "" {
		if err := r.AviClient.HealthMonitor.Delete(hm.Status.Uuid); err != nil {
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
	// this is a POST Call
	if hm.Status.Uuid == "" {
		resp, err := r.createHealthMonitor(ctx, hm)
		if err != nil {
			utils.AviLog.Errorf("error creating healthmonitor: [%s/%s]: %s", hm.Namespace, hm.Name, err.Error())
			return err
		}
		uuid, err := extractUUID(resp)
		if err != nil {
			utils.AviLog.Errorf("error extracting UUID from healthmonitor: [%s/%s]: %s", hm.Namespace, hm.Name, err.Error())
		}
		hm.Status.Uuid = uuid
		if err := r.Status().Update(ctx, hm); err != nil {
			utils.AviLog.Errorf("unable to update healthmonitor status [%s/%s]: %s", hm.Namespace, hm.Name, err.Error())
			return err
		}
	} else {
		// this is a PUT Call
		resp := map[string]interface{}{}
		if err := r.AviClient.AviSession.Put(utils.GetUriEncoded("/api/healthmonitor/"+hm.Status.Uuid), hm.Spec, &resp); err != nil {
			utils.AviLog.Errorf("error updating healthmonitor [%s/%s]: %s", hm.Namespace, hm.Name, err.Error())
			return err
		}
		utils.AviLog.Infof("succesfully updated healthmonitor:[%s/%s]", hm.Namespace, hm.Name)
	}
	return nil
}

// createHealthMonitor will attempt to create a health monitor, if it already exists, it will return an object which contains the uuid
func (r *HealthMonitorReconciler) createHealthMonitor(ctx context.Context, hm *akov1alpha1.HealthMonitor) (map[string]interface{}, error) {
	resp := map[string]interface{}{}
	if err := r.AviClient.AviSession.Post(utils.GetUriEncoded("/api/healthmonitor"), hm.Spec, &resp); err != nil {
		utils.AviLog.Errorf("error posting healthmonitor: %s", err.Error())
		if aviError, ok := err.(session.AviError); ok {
			if aviError.HttpStatusCode == http.StatusConflict {
				utils.AviLog.Infof("healthmonitor [%s/%s] already exists. trying to get uuid", hm.Namespace, hm.Name)
				err := r.AviClient.AviSession.Get(utils.GetUriEncoded(fmt.Sprintf("/api/healthmonitor?name=%s", hm.Name)), resp)
				if err != nil {
					utils.AviLog.Errorf("error getting uuid for healthmonitor [%s/%s]: %s", hm.Namespace, hm.Name, err.Error())
					return nil, err
				}
				uuid, err := extractUUID(resp)
				if err != nil {
					utils.AviLog.Errorf("error extracting UUID from healthmonitor: [%s/%s]: %s", hm.Namespace, hm.Name, err.Error())
					return nil, err
				}
				utils.AviLog.Infof("updating healthmonitor: [%s/%s]", hm.Namespace, hm.Name)
				if err := r.AviClient.AviSession.Put(utils.GetUriEncoded("/api/healthmonitor/"+uuid), hm.Spec, &resp); err != nil {
					utils.AviLog.Errorf("error updating healthmonitor [%s/%s]: %s", hm.Namespace, hm.Name, err.Error())
					return nil, err
				}
				return resp, nil
			}
		}
		return nil, err
	}
	utils.AviLog.Infof("healthmonitor [%s/%s] succesfully created", hm.Namespace, hm.Name)
	return resp, nil
}

// extractUUID extracts the UUID from resp object
func extractUUID(resp map[string]interface{}) (string, error) {
	// Extract the results array
	results, ok := resp["results"].([]interface{})
	if !ok {
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
