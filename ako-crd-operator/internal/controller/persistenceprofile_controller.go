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

	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/vmware/alb-sdk/go/clients"
	"github.com/vmware/alb-sdk/go/session"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/api/v1alpha1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/constants"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

// PersistenceProfileReconciler reconciles a PersistenceProfile object
type PersistenceProfileReconciler struct {
	client.Client
	AviClient *clients.AviClient
	Scheme    *runtime.Scheme
	Cache     cache.CacheOperation
}

type PersistenceProfileRequest struct {
	Name string `json:"name"`
	akov1alpha1.PersistenceProfileSpec

	namespace string
}

// +kubebuilder:rbac:groups=ako.vmware.com,resources=persistenceprofiles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ako.vmware.com,resources=persistenceprofiles/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ako.vmware.com,resources=persistenceprofiles/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PersistenceProfile object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *PersistenceProfileReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	pp := &akov1alpha1.PersistenceProfile{}
	err := r.Client.Get(ctx, req.NamespacedName, pp)
	if err != nil {
		if k8serror.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		utils.AviLog.Error(err, "Failed to get PersistenceProfile: [%s/%s]", req.NamespacedName.Namespace, req.NamespacedName.Name)
		return ctrl.Result{}, err
	}
	if pp.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(pp, constants.PersistenceProfileFinalizer) {
			controllerutil.AddFinalizer(pp, constants.PersistenceProfileFinalizer)
			if err := r.Update(ctx, pp); err != nil {
				utils.AviLog.Error(err, "Failed to add finalizer to PersistenceProfile: [%s/%s]", req.NamespacedName.Namespace, req.NamespacedName.Name)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	} else {
		// The object is being deleted
		if err := r.DeleteObject(ctx, pp); err != nil {
			return ctrl.Result{}, err
		}
		controllerutil.RemoveFinalizer(pp, constants.PersistenceProfileFinalizer)
		if err := r.Update(ctx, pp); err != nil {
			return ctrl.Result{}, err
		}
		utils.AviLog.Infof("succesfully deleted PersistenceProfile:[%s/%s]", pp.Namespace, pp.Name)
		return ctrl.Result{}, nil
	}
	if err := r.ReconcileIfRequired(ctx, pp); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PersistenceProfileReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&akov1alpha1.PersistenceProfile{}).
		Named("persistenceprofile").
		Complete(r)
}

// DeleteObject will attempt to delete the object from avi controller
func (r *PersistenceProfileReconciler) DeleteObject(ctx context.Context, pp *akov1alpha1.PersistenceProfile) error {
	if pp.Status.UUID != "" {
		if err := r.AviClient.ApplicationPersistenceProfile.Delete(pp.Status.UUID); err != nil {
			utils.AviLog.Errorf("error deleting PersistenceProfile: [%s/%s]: %s", pp.Namespace, pp.Name, err.Error())
			return err
		}
	} else {
		utils.AviLog.Warnf("error deleting PersistenceProfile: [%s/%s]. uuid not present. possibly avi PersistenceProfile object not created", pp.Namespace, pp.Name)
	}
	return nil
}

// TODO: Make this function generic
func (r *PersistenceProfileReconciler) ReconcileIfRequired(ctx context.Context, pp *akov1alpha1.PersistenceProfile) error {
	ppReq := &PersistenceProfileRequest{
		pp.Name,
		pp.Spec,
		pp.Namespace,
	}
	// this is a POST Call
	if pp.Status.UUID == "" {
		resp, err := r.createPersistenceProfile(ctx, ppReq)
		if err != nil {
			utils.AviLog.Errorf("error creating PersistenceProfile: [%s/%s]: %s", ppReq.namespace, ppReq.Name, err.Error())
			return err
		}
		uuid, err := extractUUID(resp)
		if err != nil {
			utils.AviLog.Errorf("error extracting UUID from PersistenceProfile: [%s/%s]: %s", ppReq.namespace, ppReq.Name, err.Error())
		}
		pp.Status.UUID = uuid
	} else {
		// this is a PUT Call
		// check if no op by checking generation
		if pp.GetGeneration() == pp.Status.ObservedGeneration {
			// if no op from kubernetes side, check if op required from OOB changes by checking lastModified timestamp
			if pp.Status.LastUpdated != nil {
				dataMap, ok := r.Cache.GetObjectByUUID(pp.Status.UUID)
				if ok {
					if dataMap.GetLastModifiedTimeStamp().Before(pp.Status.LastUpdated.Time) {
						utils.AviLog.Debugf("no op for PersistenceProfile [%s/%s]", ppReq.namespace, ppReq.Name)
						return nil
					}
				}
			}
			utils.AviLog.Debugf("overwriting PersistenceProfile: [%s/%s]", ppReq.namespace, ppReq.Name)
		}
		resp := map[string]interface{}{}
		if err := r.AviClient.AviSession.Put(utils.GetUriEncoded(fmt.Sprintf("%s/%s", constants.PersistenceProfileURL, pp.Status.UUID)), ppReq, &resp); err != nil {
			utils.AviLog.Errorf("error updating PersistenceProfile [%s/%s]: %s", ppReq.namespace, ppReq.Name, err.Error())
			return err
		}
		utils.AviLog.Infof("succesfully updated PersistenceProfile:[%s/%s]", ppReq.namespace, ppReq.Name)
	}

	pp.Status.LastUpdated = &metav1.Time{Time: time.Now().UTC()}
	pp.Status.ObservedGeneration = pp.Generation
	if err := r.Status().Update(ctx, pp); err != nil {
		utils.AviLog.Errorf("unable to update PersistenceProfile status [%s/%s]: %s", ppReq.namespace, ppReq.Name, err.Error())
		return err
	}

	return nil
}

// createPersistenceProfile will attempt to create a persistent prpfile, if it already exists, it will return an object which contains the uuid
func (r *PersistenceProfileReconciler) createPersistenceProfile(ctx context.Context, ppReq *PersistenceProfileRequest) (map[string]interface{}, error) {
	resp := map[string]interface{}{}
	if err := r.AviClient.AviSession.Post(utils.GetUriEncoded(constants.PersistenceProfileURL), ppReq, &resp); err != nil {
		utils.AviLog.Errorf("error posting PersistenceProfile: %s", err.Error())
		if aviError, ok := err.(session.AviError); ok {
			if aviError.HttpStatusCode == http.StatusConflict && strings.Contains(aviError.Error(), "already exists") {
				utils.AviLog.Infof("PersistenceProfile [%s/%s] already exists. trying to get uuid", ppReq.namespace, ppReq.Name)
				err := r.AviClient.AviSession.Get(utils.GetUriEncoded(fmt.Sprintf("%s?name=%s", constants.PersistenceProfileURL, ppReq.Name)), &resp)
				if err != nil {
					utils.AviLog.Errorf("error getting PersistenceProfile [%s/%s]: %s", ppReq.namespace, ppReq.Name, err.Error())
					return nil, err
				}
				uuid, err := extractUUID(resp)
				if err != nil {
					utils.AviLog.Errorf("error extracting UUID from PersistenceProfile: [%s/%s]: %s", ppReq.namespace, ppReq.Name, err.Error())
					return nil, err
				}
				utils.AviLog.Infof("updating PersistenceProfile: [%s/%s]", ppReq.namespace, ppReq.Name)
				if err := r.AviClient.AviSession.Put(utils.GetUriEncoded(fmt.Sprintf("%s/%s", constants.PersistenceProfileURL, uuid)), ppReq, &resp); err != nil {
					utils.AviLog.Errorf("error updating PersistenceProfile [%s/%s]: %s", ppReq.namespace, ppReq.Name, err.Error())
					return nil, err
				}
				return resp, nil
			}
		}
		return nil, err
	}
	utils.AviLog.Infof("PersistenceProfile [%s/%s] succesfully created", ppReq.namespace, ppReq.Name)
	return resp, nil
}
