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
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/constants"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/event"
	session2 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/session"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

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
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(akov1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

// nolint:gocyclo
func main() {
	ctx := context.Background()
	var probeAddr string

	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")

	ctrl.SetLogger(zap.New())

	cfg := ctrl.GetConfigOrDie()

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:                 scheme,
		HealthProbeBindAddress: probeAddr,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		utils.AviLog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}
	// event recorder
	// TODO: crd-specific event recorders
	eventRecorder := utils.NewEventRecorder("ako-crd-operator", kubeClient, false)
	eventManager := event.NewEventManager(eventRecorder, &v1.Pod{})
	// setup controller properties
	sessionManager := session2.NewSession(kubeClient, eventManager)
	sessionManager.PopulateControllerProperties(ctx)
	sessionManager.CreateAviClients(ctx, 1)
	aviClients := sessionManager.GetAviClients()

	cache := cache.NewCache(sessionManager)
	if err := cache.PopulateCache(constants.HealthMonitorURL); err != nil {
		setupLog.Error(err, "unable to populate cache")
		os.Exit(1)
	}

	hmReconciler := &controller.HealthMonitorReconciler{
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
		AviClient: aviClients.AviClient[0],
		Cache:     cache,
	}
	if err = hmReconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "HealthMonitor")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
