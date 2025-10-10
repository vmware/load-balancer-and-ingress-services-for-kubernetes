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
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/session"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	// AviSecretName is the name of the secret containing AVI controller credentials
	AviSecretName = "avi-secret"

	// SecretControllerName is the name identifier for the secret controller
	SecretControllerName = "secret-controller"

	// DefaultRequeueDelay is the default delay for requeuing failed operations
	DefaultRequeueDelay = 30 * time.Second

	// DefaultNumClients is the default number of AVI clients to create
	DefaultNumClients = 3
)

// SecretReconciler reconciles AVI secret updates and manages session refresh
type SecretReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	Logger        *utils.AviLogger
	EventRecorder record.EventRecorder
	ClusterName   string

	// Reconciler registry
	reconcilers      []AviClientReconciler
	reconcilersMutex sync.RWMutex
}

// NewSecretReconciler creates a new SecretReconciler instance
func NewSecretReconciler(
	client client.Client,
	scheme *runtime.Scheme,
	clusterName string,
) *SecretReconciler {
	return &SecretReconciler{
		Client:      client,
		Scheme:      scheme,
		Logger:      utils.AviLog.WithName(SecretControllerName),
		ClusterName: clusterName,
		reconcilers: make([]AviClientReconciler, 0),
	}
}

// RegisterReconciler registers a reconciler to receive AVI client updates
func (r *SecretReconciler) RegisterReconciler(reconciler AviClientReconciler) error {
	r.reconcilersMutex.Lock()
	defer r.reconcilersMutex.Unlock()

	name := reconciler.GetReconcilerName()

	// Check if already registered
	for _, existing := range r.reconcilers {
		if existing.GetReconcilerName() == name {
			return fmt.Errorf("reconciler %s is already registered", name)
		}
	}

	r.reconcilers = append(r.reconcilers, reconciler)
	r.Logger.Infof("Registered reconciler: %s", name)
	return nil
}

// NotifyReconcilers notifies all registered reconcilers of AVI client updates
// Each reconciler gets a unique AVI client from the client pool
func (r *SecretReconciler) NotifyReconcilers(ctx context.Context, aviClients *utils.AviRestClientPool) error {
	r.reconcilersMutex.RLock()
	defer r.reconcilersMutex.RUnlock()

	if len(r.reconcilers) == 0 {
		r.Logger.Info("No reconcilers registered")
		return nil
	}

	// Ensure we have enough clients for all reconcilers
	if len(aviClients.AviClient) < len(r.reconcilers) {
		return fmt.Errorf("insufficient AVI clients: have %d, need %d", len(aviClients.AviClient), len(r.reconcilers))
	}

	var errors []error
	for i, reconciler := range r.reconcilers {
		name := reconciler.GetReconcilerName()
		r.Logger.Infof("Notifying reconciler %s of AVI client update", name)

		// Give each reconciler a unique AVI client
		aviClient := session.NewAviSessionClient(aviClients.AviClient[i])

		if err := reconciler.UpdateAviClient(aviClient); err != nil {
			r.Logger.Errorf("Failed to update AVI client for reconciler %s: %v", name, err)
			errors = append(errors, fmt.Errorf("reconciler %s: %w", name, err))
		} else {
			r.Logger.Infof("Successfully updated AVI client for reconciler %s", name)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to update %d reconcilers: %v", len(errors), errors)
	}

	return nil
}

// GetRegisteredReconcilers returns the list of registered reconciler names
func (r *SecretReconciler) GetRegisteredReconcilers() []string {
	r.reconcilersMutex.RLock()
	defer r.reconcilersMutex.RUnlock()

	names := make([]string, 0, len(r.reconcilers))
	for _, reconciler := range r.reconcilers {
		names = append(names, reconciler.GetReconcilerName())
	}
	return names
}

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile handles secret update events
func (r *SecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Logger.WithValues("secret", req.NamespacedName, "traceID", uuid.New().String())
	ctx = utils.LoggerWithContext(ctx, log)
	log.Debug("Reconciling AVI secret update")
	defer log.Debug("Reconciled AVI secret")

	// Only process avi-secret in the AKO namespace
	if req.Name != AviSecretName || req.Namespace != utils.GetAKONamespace() {
		log.Debug("Ignoring non-AVI secret")
		return ctrl.Result{}, nil
	}

	// Get the secret to verify it exists and has valid data
	secret := &corev1.Secret{}
	if err := r.Get(ctx, req.NamespacedName, secret); err != nil {
		if k8serror.IsNotFound(err) {
			log.Error("AVI secret not found, skipping reconciliation")
			return ctrl.Result{}, nil
		}
		log.Error("Failed to get AVI secret: " + err.Error())
		return ctrl.Result{RequeueAfter: DefaultRequeueDelay}, err
	}

	// Validate secret has required fields
	if err := r.validateSecretData(secret); err != nil {
		log.Error("Invalid AVI secret data: " + err.Error())
		return ctrl.Result{RequeueAfter: DefaultRequeueDelay}, err
	}

	// Update Avi session with new credentials
	aviClients, err := r.updateSessionInstance(ctx)
	if err != nil {
		log.Error("Failed to update the AVI session: " + err.Error())
		return ctrl.Result{RequeueAfter: DefaultRequeueDelay}, err
	}

	if aviClients == nil {
		log.Error("No AVI clients available after session update")
		return ctrl.Result{RequeueAfter: DefaultRequeueDelay}, fmt.Errorf("no AVI clients available")
	}

	// Notify all registered reconcilers with unique AVI clients
	if err := r.NotifyReconcilers(ctx, aviClients); err != nil {
		log.Error("Failed to notify reconcilers: " + err.Error())
		return ctrl.Result{RequeueAfter: DefaultRequeueDelay}, err
	}

	log.Info("Successfully processed AVI secret update")
	return ctrl.Result{}, nil
}

// validateSecretData validates that the secret contains required AVI controller credentials
func (r *SecretReconciler) validateSecretData(secret *corev1.Secret) error {
	requiredFields := []string{"username"}

	for _, field := range requiredFields {
		if _, exists := secret.Data[field]; !exists {
			return fmt.Errorf("required field %s missing from secret", field)
		}
	}

	// Must have either password or authtoken
	hasPassword := len(secret.Data["password"]) > 0
	hasAuthToken := len(secret.Data["authtoken"]) > 0

	if !hasPassword && !hasAuthToken {
		return fmt.Errorf("secret must contain either password or authtoken")
	}

	return nil
}

// updateSessionInstance updates the Avi session with new credentials
func (r *SecretReconciler) updateSessionInstance(ctx context.Context) (*utils.AviRestClientPool, error) {
	sessionManager := session.GetSessionInstance()

	// Update controller properties from the updated secret
	if err := sessionManager.UpdateAviClients(ctx, r.getRequiredClientCount()); err != nil {
		return nil, fmt.Errorf("failed to update the AVI session: %w", err)
	}

	aviClients := sessionManager.GetAviClients()
	if aviClients == nil {
		return nil, fmt.Errorf("failed to get AVI clients from the session object")
	}

	r.Logger.Infof("Successfully updated the AVI session with %d clients", len(aviClients.AviClient))
	return aviClients, nil
}

// getRequiredClientCount calculates the number of clients needed
func (r *SecretReconciler) getRequiredClientCount() int {
	numClients := DefaultNumClients
	r.reconcilersMutex.RLock()
	if len(r.reconcilers) > numClients {
		numClients = len(r.reconcilers)
	}
	r.reconcilersMutex.RUnlock()
	return numClients
}

// SetupWithManager sets up the controller with the Manager
func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Only watch avi-secret in the AKO namespace
	secretPredicate := predicate.NewPredicateFuncs(func(object client.Object) bool {
		return object.GetName() == AviSecretName && object.GetNamespace() == utils.GetAKONamespace()
	})

	// ResourceVersion predicate to detect actual changes to the secret
	resourceVersionPredicate := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Only trigger reconciliation if the resource version changed
			return e.ObjectOld.GetResourceVersion() != e.ObjectNew.GetResourceVersion()
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
		For(&corev1.Secret{}, builder.WithPredicates(secretPredicate, resourceVersionPredicate)).
		Complete(r)
}

// GetReconcilerName returns the reconciler name for interface implementation
func (r *SecretReconciler) GetReconcilerName() string {
	return SecretControllerName
}

// UpdateAviClient implements ReconcilerInterface (though not typically used for SecretReconciler)
func (r *SecretReconciler) UpdateAviClient(client session.AviClientInterface) error {
	// The secret reconciler manages its own session, so this is a no-op
	r.Logger.Debug("UpdateAviClient called on SecretReconciler (no-op)")
	return nil
}
