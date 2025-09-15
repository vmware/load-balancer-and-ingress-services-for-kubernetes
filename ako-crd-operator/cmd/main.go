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

package main

import (
	"context"
	"flag"
	"os"

	"github.com/go-logr/zapr"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/constants"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/event"
	session2 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/session"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/healthz"

	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/api/v1alpha1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/controller"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = utils.AviLog.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(akov1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

// nolint:gocyclo
func main() {
	ctx := utils.LoggerWithContext(context.Background(), setupLog)
	var probeAddr string

	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")

	ctrl.SetLogger(zapr.NewLogger(utils.AviLog.Sugar.Desugar().Named("runtime")))

	cfg := ctrl.GetConfigOrDie()

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:                 scheme,
		HealthProbeBindAddress: probeAddr,
	})
	if err != nil {
		setupLog.Fatalf("unable to start manager. error: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		setupLog.Fatalf("Error building kubernetes clientset. error: %s", err.Error())
	}
	// event recorder
	// TODO: crd-specific event recorders
	eventRecorder := utils.NewEventRecorder("ako-crd-operator", kubeClient, false)
	eventManager := event.NewEventManager(eventRecorder, &v1.Pod{})

	// Initialize singleton session
	sessionManager := session2.InitializeSessionInstance(kubeClient, eventManager)
	if err := sessionManager.PopulateControllerProperties(ctx); err != nil {
		setupLog.Fatalf("Error populating controller properties. error: %s", err.Error())
	}

	sessionManager.CreateAviClients(ctx, 4)
	aviClients := sessionManager.GetAviClients()
	clusterName := os.Getenv("CLUSTER_NAME")

	cacheManager := cache.NewCache(
		session2.NewAviSessionClient(aviClients.AviClient[0]),
		clusterName)

	if err := cacheManager.PopulateCache(ctx, constants.HealthMonitorURL,
		constants.ApplicationProfileURL, constants.PKIProfileURL); err != nil {
		setupLog.Fatalf("unable to populate cacheManager. error: %s", err.Error())
	}
	utils.AviLog.SetLevel(GetEnvOrDefault("LOG_LEVEL", "INFO"))

	// Create Secret Controller first
	secretReconciler := controller.NewSecretReconciler(
		mgr.GetClient(),
		mgr.GetScheme(),
		clusterName,
	)
	secretReconciler.EventRecorder = mgr.GetEventRecorderFor(controller.SecretControllerName)

	// Setup Secret Controller with manager
	if err = secretReconciler.SetupWithManager(mgr); err != nil {
		setupLog.Fatalf("unable to create Secret Controller. error: %s", err.Error())
	}

	// Create controllers using factory methods
	_, err = controller.CreateNewHealthMonitorControllerAndSetupWithManager(
		mgr,
		session2.NewAviSessionClient(aviClients.AviClient[0]),
		cacheManager,
		clusterName,
		secretReconciler,
	)
	if err != nil {
		setupLog.Fatalf("unable to create HealthMonitor controller. error: %s", err.Error())
	}

	_, err = controller.CreateNewApplicationProfileControllerAndSetupWithManager(
		mgr,
		session2.NewAviSessionClient(aviClients.AviClient[1]),
		cacheManager,
		clusterName,
		secretReconciler,
	)
	if err != nil {
		setupLog.Fatalf("unable to create ApplicationProfile controller. error: %s", err.Error())
	}

	_, err = controller.CreateNewRouteBackendExtensionControllerAndSetupWithManager(
		mgr,
		session2.NewAviSessionClient(aviClients.AviClient[2]),
		clusterName,
		secretReconciler,
	)
	if err != nil {
		setupLog.Fatalf("unable to create RouteBackendExtension controller. error: %s", err.Error())
	}
	_, err = controller.CreateNewPKIProfileControllerAndSetupWithManager(
		mgr,
		session2.NewAviSessionClient(aviClients.AviClient[3]),
		cacheManager,
		clusterName,
		secretReconciler,
	)
	if err != nil {
		setupLog.Fatalf("unable to create PKIProfile controller. error: %s", err.Error())
	}

	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Fatalf("unable to set up health check. error: %s", err.Error())
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Fatalf("unable to set up ready check. error: %s", err.Error())
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Fatalf("problem running manager. error: %s", err.Error())
	}
}

// GetEnvOrDefault retrieves the value of the environment variable named by the key.
// If the variable is not present or its value is empty, it returns the
// defaultValue.
func GetEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
