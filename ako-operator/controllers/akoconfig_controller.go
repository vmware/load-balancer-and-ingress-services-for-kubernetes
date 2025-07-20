/*
Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
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

//nolint:unparam
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
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

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
	Config *rest.Config
	Log    logr.Logger
	Scheme *runtime.Scheme
}

var objectList map[types.NamespacedName]client.Object

var objListOnce sync.Once

func getObjectList() map[types.NamespacedName]client.Object {
	objListOnce.Do(func() {
		objectList = make(map[types.NamespacedName]client.Object)
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

// +kubebuilder:rbac:groups="",resources=*,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=configmaps;configmaps/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets;secrets/status;secrets/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts;serviceaccounts/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services;services/status,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create;update;patch
// +kubebuilder:rbac:groups=ako.vmware.com,resources=akoconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ako.vmware.com,resources=akoconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ako.vmware.com,resources=akoconfigs/finalizers,verbs=update
// +kubebuilder:rbac:groups=ako.vmware.com,resources=aviinfrasettings;aviinfrasettings/status;aviinfrasettings/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ako.vmware.com,resources=httprules;httprules/status;httprules/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ako.vmware.com,resources=hostrules;hostrules/status;hostrules/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crd.projectcalico.org,resources=blockaffinities;blockaffinities/status,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apiextensions.k8s.io",resources=customresourcedefinitions;customresourcedefinitions/status;customresourcedefinitions/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apps",resources=statefulsets;statefulsets/status;statefulsets/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=extensions,resources=ingresses; ingresses/status,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=network.openshift.io,resources=hostsubnets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingressclasses;ingressclasses/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses;ingresses/status,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.x-k8s.io,resources=gatewayclasses;gatewayclasses/status;gatewayclasses/finalizers;gateways;gateways/status;gateways/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=policy;extensions,resources=podsecuritypolicies;podsecuritypolicies/finalizers,verbs=use;get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterroles;clusterroles/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterrolebindings;clusterrolebindings/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=route.openshift.io,resources=routes;routes/status,verbs=get;list;watch;create;update;patch;delete

func (r *AKOConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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

			patch := client.MergeFrom(ako.DeepCopy())
			ako.Finalizers = removeFinalizer(ako.Finalizers, CleanupFinalizer)
			if err := r.Patch(context.TODO(), &ako, patch); err != nil {
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

	checkDeprecatedFields(ako, log)

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

	err = createCRDs(r.Config, log)
	if err != nil {
		return err
	}

	err = createOrUpdateGatewayClass(ctx, ako, log, r)
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

		var gwClass gatewayv1beta1.GatewayClass
		if err := r.Get(ctx, getGWClassName(), &gwClass); err != nil {
			log.V(0).Info("error getting gatewayclass", "error", err)
		} else {
			objList[getGWClassName()] = &gwClass
		}
	}
	for objName, obj := range objList {
		if err := r.deleteIfExists(ctx, objName, obj); err != nil {
			log.Error(err, "error while deleting object")
			return err
		}
	}
	err := deleteCRDs(r.Config, log)
	if err != nil {
		log.Error(err, "error while deleting crds")
		return err
	}
	return nil
}

func (r *AKOConfigReconciler) deleteIfExists(ctx context.Context, objNsName types.NamespacedName, object client.Object) error {
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
