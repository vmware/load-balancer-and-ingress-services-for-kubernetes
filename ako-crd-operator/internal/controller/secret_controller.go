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
	"reflect"

	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/google/uuid"
	avisession "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/session"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

// SecretReconciler reconciles avi-secret updates
type SecretReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	EventRecorder record.EventRecorder
	Logger        *utils.AviLogger
}

//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=secrets/status,verbs=get

// Reconcile handles avi-secret updates and refreshes AVI client pool
func (r *SecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Logger.WithValues("name", req.Name, "namespace", req.Namespace, "traceID", uuid.New().String())
	ctx = utils.LoggerWithContext(ctx, log)

	log.Debug("Reconciling Secret Resource")
	defer log.Debug("Reconciled Secret Resource")

	// Only process avi-secret in the AKO namespace
	if req.Name != lib.AviSecret || req.Namespace != utils.GetAKONamespace() {
		return ctrl.Result{}, nil
	}

	log.Infof("Reconciling avi-secret update: %s/%s", req.Namespace, req.Name)

	// Fetch the secret
	secret := &corev1.Secret{}
	if err := r.Get(ctx, req.NamespacedName, secret); err != nil {
		if k8serror.IsNotFound(err) {
			// Secret was deleted - this is a critical error
			log.Errorf("avi-secret was deleted: %s/%s", req.Namespace, req.Name)
			r.EventRecorder.Eventf(
				&corev1.Pod{}, // Using pod as reference since we don't have a specific resource
				corev1.EventTypeWarning,
				"AviSecretDeleted",
				"Critical: avi-secret %s/%s was deleted",
				req.Namespace, req.Name,
			)
			return ctrl.Result{}, err
		}
		log.Errorf("Failed to get avi-secret: %v", err)
		return ctrl.Result{}, err
	}

	// Update AVI client pool with new credentials from avi-secret
	// Using 3 clients to match the original setup in main.go
	sessionObj := avisession.GetGlobalSession()
	if err := sessionObj.UpdateAviClientsFromSecret(ctx, 3); err != nil {
		log.Errorf("Failed to update AVI client pool: %v", err)
		r.EventRecorder.Eventf(
			secret,
			corev1.EventTypeWarning,
			"AviClientPoolUpdateFailed",
			"Failed to update AVI client pool from avi-secret: %v",
			err,
		)
		return ctrl.Result{}, err
	}

	log.Infof("Successfully updated AVI client pool after avi-secret change")
	r.EventRecorder.Eventf(
		secret,
		corev1.EventTypeNormal,
		"AviClientPoolUpdated",
		"AVI client pool successfully updated with new credentials from avi-secret",
	)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Create predicate to only watch avi-secret in AKO namespace
	secretPredicate := predicate.NewPredicateFuncs(func(obj client.Object) bool {
		secret, ok := obj.(*corev1.Secret)
		if !ok {
			return false
		}
		return secret.Name == lib.AviSecret && secret.Namespace == utils.GetAKONamespace()
	})

	// Create update predicate to detect actual data changes
	updatePredicate := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldSecret, ok := e.ObjectOld.(*corev1.Secret)
			if !ok {
				return false
			}
			newSecret, ok := e.ObjectNew.(*corev1.Secret)
			if !ok {
				return false
			}

			// Only trigger reconciliation if the secret data actually changed
			return !reflect.DeepEqual(oldSecret.Data, newSecret.Data)
		},
		CreateFunc: func(e event.CreateEvent) bool {
			// Process create events for avi-secret
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// Process delete events for avi-secret (critical)
			return true
		},
		GenericFunc: func(e event.GenericEvent) bool {
			// Skip generic events
			return false
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}, builder.WithPredicates(secretPredicate, updatePredicate)).
		Complete(r)
}
