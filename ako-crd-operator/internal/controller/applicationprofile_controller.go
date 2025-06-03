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

	"github.com/vmware/alb-sdk/go/clients"
	"github.com/vmware/alb-sdk/go/session"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/api/v1alpha1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/constants"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ApplicationProfileReconciler reconciles a ApplicationProfile object
type ApplicationProfileReconciler struct {
	client.Client
	AviClient *clients.AviClient
	Scheme *runtime.Scheme
	Cache     cache.CacheOperation
}

type ApplicationProfileRequest struct {
	Name string `json:"name"`
	akov1alpha1.ApplicationProfileSpec

	namespace string
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
	ap := &akov1alpha1.ApplicationProfile{}
	err := r.Client.Get(ctx, req.NamespacedName, ap)
	if err != nil {
		if k8serror.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		utils.AviLog.Error(err, "Failed to get ApplicationProfile: [%s/%s]", req.NamespacedName.Namespace, req.NamespacedName.Name)
		return ctrl.Result{}, err
	}
	if ap.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(ap, constants.ApplicationProfileFinalizer) {
			controllerutil.AddFinalizer(ap, constants.ApplicationProfileFinalizer)
			if err := r.Update(ctx, ap); err != nil {
				utils.AviLog.Error(err, "Failed to add finalizer to ApplicationProfile: [%s/%s]", req.NamespacedName.Namespace, req.NamespacedName.Name)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	} else {
		// The object is being deleted
		if err := r.DeleteObject(ctx, ap); err != nil {
			return ctrl.Result{}, err
		}
		controllerutil.RemoveFinalizer(ap, constants.ApplicationProfileFinalizer)
		if err := r.Update(ctx, ap); err != nil {
			return ctrl.Result{}, err
		}
		utils.AviLog.Infof("succesfully deleted applicationprofile:[%s/%s]", ap.Namespace, ap.Name)
		return ctrl.Result{}, nil
	}
	if err := r.ReconcileIfRequired(ctx, ap); err != nil {
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

// SetupWithManager sets up the controller with the Manager.
func (r *ApplicationProfileReconciler) DeleteObject(ctx context.Context, ap *akov1alpha1.ApplicationProfile) error {
	if ap.Status.UUID != "" {
		if err := r.AviClient.ApplicationProfile.Delete(ap.Status.UUID); err != nil {
			utils.AviLog.Errorf("error deleting application profile: [%s/%s]: %s", ap.Namespace, ap.Name, err.Error())
			return err
		}
	} else {
		utils.AviLog.Warnf("error deleting application profile: [%s/%s]. uuid not present. possibly avi application profile object not created", ap.Namespace, ap.Name)
	}
	return nil
}

// TODO: Make this function generic
func (r *ApplicationProfileReconciler) ReconcileIfRequired(ctx context.Context, ap *akov1alpha1.ApplicationProfile) error {
	apReq := &ApplicationProfileRequest{
		ap.Name,
		ap.Spec,
		ap.Namespace,
	}
	// this is a POST Call
	if ap.Status.UUID == "" {
		resp, err := r.createApplicationProfile(ctx, apReq)
		if err != nil {
			utils.AviLog.Errorf("error creating application profile: [%s/%s]: %s", apReq.namespace, apReq.Name, err.Error())
			return err
		}
		uuid, err := extractUUID(resp)
		if err != nil {
			utils.AviLog.Errorf("error extracting UUID from application profile: [%s/%s]: %s", apReq.namespace, apReq.Name, err.Error())
		}
		ap.Status.UUID = uuid
	} else {
		// this is a PUT Call
		// check if no op by checking generation
		if ap.GetGeneration() == ap.Status.ObservedGeneration {
			// if no op from kubernetes side, check if op required from OOB changes by checking lastModified timestamp
			if ap.Status.LastUpdated != nil {
				dataMap, ok := r.Cache.GetObjectByUUID(ap.Status.UUID)
				if ok {
					if dataMap.GetLastModifiedTimeStamp().Before(ap.Status.LastUpdated.Time) {
						utils.AviLog.Debugf("no op for application profile [%s/%s]", apReq.namespace, apReq.Name)
						return nil
					}
				}
			}
			utils.AviLog.Debugf("overwriting application profile: [%s/%s]", apReq.namespace, apReq.Name)
		}
		resp := map[string]interface{}{}
		if err := r.AviClient.AviSession.Put(utils.GetUriEncoded(fmt.Sprintf("%s/%s", constants.ApplicationProfileURL, ap.Status.UUID)), apReq, &resp); err != nil {
			utils.AviLog.Errorf("error updating application profile [%s/%s]: %s", apReq.namespace, apReq.Name, err.Error())
			return err
		}
		utils.AviLog.Infof("succesfully updated application profile:[%s/%s]", apReq.namespace, apReq.Name)
	}
	ap.Status.LastUpdated = &metav1.Time{Time: time.Now().UTC()}
	ap.Status.ObservedGeneration = ap.Generation
	if err := r.Status().Update(ctx, ap); err != nil {
		utils.AviLog.Errorf("unable to update application profile status [%s/%s]: %s", apReq.namespace, apReq.Name, err.Error())
		return err
	}
	return nil
}

// createApplicationProfile will attempt to create a application profile, if it already exists, it will return an object which contains the uuid
func (r *ApplicationProfileReconciler) createApplicationProfile(ctx context.Context,apReq *ApplicationProfileRequest) (map[string]interface{}, error) {
	resp := map[string]interface{}{}
	if err := r.AviClient.AviSession.Post(utils.GetUriEncoded(constants.ApplicationProfileURL), apReq, &resp); err != nil {
		utils.AviLog.Errorf("error posting application profile: %s", err.Error())
		if aviError, ok := err.(session.AviError); ok {
			if aviError.HttpStatusCode == http.StatusConflict && strings.Contains(aviError.Error(), "already exists") {
				utils.AviLog.Infof("application profile [%s/%s] already exists. trying to get uuid", apReq.namespace, apReq.Name)
				err := r.AviClient.AviSession.Get(utils.GetUriEncoded(fmt.Sprintf("%s?name=%s", constants.ApplicationProfileURL, apReq.Name)), &resp)
				if err != nil {
					utils.AviLog.Errorf("error getting application profile [%s/%s]: %s", apReq.namespace, apReq.Name, err.Error())
					return nil, err
				}
				uuid, err := extractUUID(resp)
				if err != nil {
					utils.AviLog.Errorf("error extracting UUID from application profile: [%s/%s]: %s", apReq.namespace, apReq.Name, err.Error())
					return nil, err
				}
				utils.AviLog.Infof("updating application profile: [%s/%s]", apReq.namespace, apReq.Name)
				if err := r.AviClient.AviSession.Put(utils.GetUriEncoded(fmt.Sprintf("%s/%s", constants.ApplicationProfileURL, uuid)), apReq, &resp); err != nil {
					utils.AviLog.Errorf("error updating application profile [%s/%s]: %s", apReq.namespace, apReq.Name, err.Error())
					return nil, err
				}
				return resp, nil
			}
		}
		return nil, err
	}
	utils.AviLog.Infof("Application profile [%s/%s] successfully created", apReq.namespace, apReq.Name)
	return resp, nil
}
