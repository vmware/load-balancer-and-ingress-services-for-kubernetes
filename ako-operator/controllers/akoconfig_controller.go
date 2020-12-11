/*
Copyright 2020 VMware, Inc.
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

package controllers

import (
	"context"
	"sync"

	logr "github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-operator/api/v1alpha1"
)

const (
	CleanupFinalizer = "ako.vmware.com/cleanup"
)

var (
	rebootRequired = false
)

// AKOConfigReconciler reconciles a AKOConfig object
type AKOConfigReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

var objectList map[types.NamespacedName]runtime.Object

var objListOnce sync.Once

func getObjectList() map[types.NamespacedName]runtime.Object {
	objListOnce.Do(func() {
		objectList = make(map[types.NamespacedName]runtime.Object)
	})
	return objectList
}

func finalizerInList(finalizers []string, key string) bool {
	for _, f := range finalizers {
		if f == key {
			return true
		}
	}
	return false
}

func removeFinalizer(finalizers []string, key string) (result []string) {
	for _, f := range finalizers {
		if f == key {
			continue
		}
		result = append(result, f)
	}
	return result
}

// +kubebuilder:rbac:groups=ako.vmware.com,resources=akoconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ako.vmware.com,resources=akoconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apps",resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterroles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="policy",resources=podsecuritypolicies,verbs=get;list;watch;create;update;patch;delete

func (r *AKOConfigReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("ako-operator", req.NamespacedName)

	var ako akov1alpha1.AKOConfig
	err := r.Client.Get(ctx, req.NamespacedName, &ako)
	if err != nil && errors.IsNotFound(err) {
		// akoconfig object got deleted, before we come here, so just return
		return ctrl.Result{}, nil
	} else if err != nil {
		log.V(0).Info("unable to fetch AKOConfig object", "err", err)
		return ctrl.Result{}, err
	}

	if !ako.GetDeletionTimestamp().IsZero() {
		if finalizerInList(ako.GetFinalizers(), CleanupFinalizer) {
			if err := r.CleanupArtifacts(ctx, log); err != nil {
				return ctrl.Result{}, err
			}
			ako.Finalizers = removeFinalizer(ako.Finalizers, CleanupFinalizer)
			if err := r.Update(ctx, &ako); err != nil {
				return ctrl.Result{}, err
			}
		}
		// return from here, no more reconciliation as the AKOConfig is being deleted
		return ctrl.Result{}, nil
	}

	// reconcile all objects
	err = r.ReconcileAllArtifacts(ctx, ako, log)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *AKOConfigReconciler) ReconcileAllArtifacts(ctx context.Context, ako akov1alpha1.AKOConfig, log logr.Logger) error {
	secretNamespacedName := types.NamespacedName{Namespace: AviSystemNS, Name: AviSecretName}
	var aviSecret v1.Secret
	err := r.Get(ctx, secretNamespacedName, &aviSecret)
	if err != nil {
		log.Error(err, "secret named avi-secret is must for starting AKO controller")
		return err
	}
	// reconcile all the required artifacts for AKO
	err = createOrUpdateConfigMap(ctx, ako, log, r)
	if err != nil {
		return err
	}

	err = createOrUpdateServiceAccount(ctx, ako, log, r)
	if err != nil {
		return err
	}

	err = createOrUpdateClusterRole(ctx, ako, log, r)
	if err != nil {
		return err
	}

	err = createOrUpdateClusterroleBinding(ctx, ako, log, r)
	if err != nil {
		return err
	}

	err = createOrUpdatePodSecurityPolicy(ctx, ako, log, r)
	if err != nil {
		return err
	}

	err = createOrUpdateStatefulSet(ctx, ako, log, r, aviSecret)
	if err != nil {
		return err
	}

	return nil
}

func (r *AKOConfigReconciler) CleanupArtifacts(ctx context.Context, log logr.Logger) error {
	log.V(0).Info("cleaning up all the artifacts")
	objList := getObjectList()
	if len(objList) == 0 {
		// AKOConfig was deleted, but during the same time, the operator was restarted
		var cm corev1.ConfigMap
		if err := r.Get(ctx, getConfigMapName(), &cm); err != nil {
			log.V(0).Info("error getting configmap", "error", err)
		} else {
			objList[getConfigMapName()] = &cm
		}
		var sf appsv1.StatefulSet
		if err := r.Get(ctx, getSFNamespacedName(), &sf); err != nil {
			log.V(0).Info("error getting statefulset", "error", err)
		} else {
			objList[getSFNamespacedName()] = &sf
		}
		var cr rbacv1.ClusterRole
		if err := r.Get(ctx, getCRName(), &cr); err != nil {
			log.V(0).Info("error getting clusterrole", "error", err)
		} else {
			objList[getCRName()] = &cr
		}
		var crb rbacv1.ClusterRoleBinding
		if err := r.Get(ctx, getCRBName(), &crb); err != nil {
			log.V(0).Info("error getting clusterrolebinding", "error", err)
		} else {
			objList[getCRBName()] = &crb
		}
		var sa v1.ServiceAccount
		if err := r.Get(ctx, getSAName(), &sa); err != nil {
			log.V(0).Info("error getting serviceaccount", "error", err)
		} else {
			objList[getSAName()] = &sa
		}
		var psp policyv1beta1.PodSecurityPolicy
		if err := r.Get(ctx, getPSPName(), &psp); err != nil {
			log.V(0).Info("error getting podsecuritypolicy", "error", err)
		} else {
			objList[getPSPName()] = &psp
		}
	}
	for objName, obj := range objList {
		if err := r.deleteIfExists(ctx, objName, obj); err != nil {
			log.Error(err, "error while deleting object")
			return err
		}
	}
	return nil
}

func (r *AKOConfigReconciler) deleteIfExists(ctx context.Context, objNsName types.NamespacedName, object runtime.Object) error {
	err := r.Client.Get(ctx, objNsName, object)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if err == nil {
		if err := r.Client.Delete(ctx, object); err != nil {
			return err
		}
	}
	return nil
}

func (r *AKOConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&akov1alpha1.AKOConfig{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.StatefulSet{}).
		Complete(r)
}
