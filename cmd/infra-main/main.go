/*
 * Copyright 2021 VMware, Inc.
 * All Rights Reserved.
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*   http://www.apache.org/licenses/LICENSE-2.0
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*/

package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-infra/ingestion"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-infra/webhooks"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	v1beta1crd "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1beta1/clientset/versioned"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	masterURL  string
	kubeconfig string
	version    = "dev"
)

func main() {
	InitializeAKOInfra()
}

func InitializeAKOInfra() {
	if !utils.IsVCFCluster() {
		utils.AviLog.Fatalf("Not running in vcf cluster, shutting down")
	}

	var err error
	kubeCluster := false
	utils.AviLog.Infof("AKO-Infra is running with version: %s", version)
	// Check if we are running inside kubernetes. Hence try authenticating with service token
	cfg, err := rest.InClusterConfig()
	if err != nil {
		utils.AviLog.Warnf("We are not running inside kubernetes cluster. %s", err.Error())
	} else {
		utils.AviLog.Infof("We are running inside kubernetes cluster. Won't use kubeconfig files.")
		kubeCluster = true
	}
	if kubeCluster == false {
		cfg, err = clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
		utils.AviLog.Infof("master: %s", masterURL)
		if err != nil {
			utils.AviLog.Fatalf("Error building kubeconfig: %s", err.Error())
		}
	}

	dynamicClient, err := lib.NewDynamicClientSet(cfg)
	if err != nil {
		utils.AviLog.Warnf("Error while creating dynamic client %v", err)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		utils.AviLog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	v1beta1crdClient, err := v1beta1crd.NewForConfig(cfg)
	if err != nil {
		utils.AviLog.Fatalf("Error building AKO CRD v1beta1 clientset: %s", err.Error())
	}

	utils.AviLog.Infof("Successfully created kube client for ako-infra")

	lib.AKOControlConfig().SetEventRecorder(lib.AKOEventComponent, kubeClient, false)
	pods, err := kubeClient.CoreV1().Pods(utils.GetAKONamespace()).List(context.TODO(), metav1.ListOptions{Limit: 1})
	if err != nil {
		utils.AviLog.Warnf("Error getting AKO pod details, %s.", err.Error())
	}
	for _, pod := range pods.Items {
		lib.AKOControlConfig().SaveAKOPodObjectMeta(&pod)
	}

	registeredInformers, err := lib.InformersToRegister(kubeClient, nil)
	if err != nil {
		utils.AviLog.Fatalf("Failed to initialize informers: %v, shutting down AKO-Infra, going to reboot", err)
	}

	informersArg := make(map[string]interface{})

	utils.NewInformers(utils.KubeClientIntf{ClientSet: kubeClient}, registeredInformers, informersArg)
	lib.NewDynamicInformers(dynamicClient, true)

	lib.AKOControlConfig().SetCRDClientsetAndEnableInfraSettingParam(v1beta1crdClient)
	k8s.NewInfraSettingCRDInformer()

	c := ingestion.SharedVCFK8sController()

	stopCh := utils.SetupSignalHandler()
	ctrlCh := make(chan struct{})

	transportZone := c.HandleVCF(stopCh, ctrlCh)
	lib.VCFInitialized = true
	lib.SetAKOUser(lib.AKOPrefix)

	// Checking/Setting up Avi pre-reqs
	a := ingestion.NewAviControllerInfra(kubeClient)

	a.InitInfraController()
	// Check for kubernetes apiserver version compatibility with AKO version.
	if serverVersionInfo, err := kubeClient.Discovery().ServerVersion(); err != nil {
		utils.AviLog.Warnf("Error while fetching kubernetes apiserver version")
	} else {
		serverVersion := fmt.Sprintf("%s.%s", serverVersionInfo.Major, serverVersionInfo.Minor)
		utils.AviLog.Infof("Kubernetes cluster apiserver version %s", serverVersion)
		if lib.CompareVersions(serverVersion, ">", lib.GetK8sMaxSupportedVersion()) ||
			lib.CompareVersions(serverVersion, "<", lib.GetK8sMinSupportedVersion()) {
			utils.AviLog.Fatalf("Unsupported kubernetes apiserver version detected. Please check the supportability guide.")
		}
	}

	c.InitNetworkingHandler()
	lib.RunAviInfraSettingInformer(stopCh)
	c.AddSecretEventHandler(stopCh)
	aviCloud, err := a.DeriveCloudMappedToTZ(transportZone)
	if err != nil {
		lib.AKOControlConfig().PodEventf(corev1.EventTypeWarning, "CloudMatchingTZNotFound", err.Error())
		utils.AviLog.Fatalf("Failed to derive cloud, err: %s", err.Error())
	}
	if !lib.GetVPCMode() {
		a.SetupSEGroup(aviCloud)
		c.AddAvailabilityZoneCREventHandler(stopCh)
	}
	c.AddNamespaceEventHandler(stopCh)
	c.Sync()
	a.AnnotateSystemNamespace(lib.GetClusterID(), utils.CloudName)
	c.AddNetworkInfoEventHandler(stopCh)

	// Initialize VKS Dependency Manager with AddonInstall integration
	// This manages both cluster dependencies and the global AddonInstall resource
	vksDependencyManager := ingestion.NewVKSDependencyManager(kubeClient, dynamicClient)
	if err := vksDependencyManager.EnsureGlobalAddonInstall(context.TODO()); err != nil {
		utils.AviLog.Warnf("Failed to ensure VKS global AddonInstall resource: %v", err.Error())
		// Don't fail startup - this is not critical for non-VKS environments
	} else {
		utils.AviLog.Infof("Successfully ensured VKS global AddonInstall resource")
	}

	// Initialize VKS Cluster Watcher for event-driven dependency management
	// This watches cluster events and triggers dependency creation/deletion
	vksClusterWatcher := ingestion.NewVKSClusterWatcher(kubeClient, dynamicClient)

	// Initialize Avi Controller connection for both dependency manager and cluster watcher
	// This should be done after cloud and SEG setup is complete
	aviHost := lib.GetControllerIP()
	aviVersion := lib.GetControllerVersion()
	aviCloudName := *aviCloud.Name // Extract name from Cloud object
	aviTenant := lib.GetTenant()
	aviVrf := lib.GetVrf()

	if aviHost != "" {
		vksDependencyManager.InitializeAviControllerConnection(aviHost, aviVersion, aviCloudName, aviTenant, aviVrf)

		// Initialize VKS Management Service integration with vCenter URL
		// This enables VMCI proxy support for guest cluster connectivity
		vCenterURL := getVCenterURLFromConfig(kubeClient)
		if vCenterURL != "" {
			vksDependencyManager.InitializeVKSManagementService(vCenterURL)
			utils.AviLog.Infof("Initialized VKS Management Service integration with vCenter: %s", vCenterURL)
		} else {
			utils.AviLog.Warnf("vCenter URL not available - VKS Management Service integration disabled")
		}

		utils.AviLog.Infof("Initialized VKS components with Avi Controller: %s", aviHost)
	} else {
		utils.AviLog.Warnf("Avi Controller IP not available - VKS dependency generation may be delayed")
	}

	// Initialize VKS Webhook Server FIRST for admission control
	// This provides proactive cluster labeling and must start before the cluster watcher
	// to prevent race conditions where clusters are created before webhook is ready
	vksWebhookEnabled := os.Getenv("VKS_WEBHOOK_ENABLED") == "true"
	if vksWebhookEnabled {
		vksWebhookPort := os.Getenv("VKS_WEBHOOK_PORT")
		if vksWebhookPort == "" {
			vksWebhookPort = "9443" // Default webhook port
		}
		vksWebhookCertDir := "/etc/webhook/certs" // Standard webhook cert directory

		// Initialize certificate manager for webhook TLS
		certManager := webhooks.NewVKSWebhookCertificateManager(
			kubeClient,
			utils.GetAKONamespace(),
			"vks-webhook-certs",
			"ako-webhook-service",
			vksWebhookCertDir,
			"vks-cluster-webhook-config",
		)

		// Ensure certificates are ready
		if err := certManager.EnsureCertificatesOnStartup(context.TODO()); err != nil {
			utils.AviLog.Errorf("Failed to ensure webhook certificates: %v", err.Error())
		} else {
			utils.AviLog.Infof("VKS webhook certificates ready")

			// Initialize webhook server
			vksWebhook := webhooks.NewVKSClusterWebhook(kubeClient)

			// Start webhook server
			go func() {
				if err := startVKSWebhookServer(vksWebhook, vksWebhookPort, vksWebhookCertDir, stopCh); err != nil {
					utils.AviLog.Errorf("VKS webhook server failed: %v", err.Error())
				}
			}()

			utils.AviLog.Infof("Started VKS webhook server on port %s", vksWebhookPort)
		}
	} else {
		utils.AviLog.Infof("VKS webhook disabled - using cluster watcher for reactive processing")
	}

	// Start VKS Cluster Watcher AFTER webhook for event-driven processing
	// This handles existing clusters and provides backup for any missed webhook events
	if err := vksClusterWatcher.Start(stopCh); err != nil {
		utils.AviLog.Errorf("Failed to start VKS cluster watcher: %v", err.Error())
		// Don't fail startup - webhook will handle new clusters
	} else {
		utils.AviLog.Infof("Successfully started VKS cluster watcher")
	}

	// Start VKS dependency manager reconciler
	// This handles periodic reconciliation and resource watching automatically
	if err := vksDependencyManager.StartReconciler(context.TODO()); err != nil {
		utils.AviLog.Errorf("Failed to start VKS dependency manager reconciler: %v", err.Error())
	} else {
		utils.AviLog.Infof("Successfully started VKS dependency manager reconciler")
	}

	utils.AviLog.Infof("AKO-Infra initialization complete")
	<-stopCh
}

// startVKSWebhookServer starts the VKS webhook HTTP server with TLS
func startVKSWebhookServer(webhook *webhooks.VKSClusterWebhook, port, certDir string, stopCh <-chan struct{}) error {
	// Set up TLS configuration
	certPath := filepath.Join(certDir, "tls.crt")
	keyPath := filepath.Join(certDir, "tls.key")

	// Load TLS certificate
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return fmt.Errorf("failed to load TLS certificate: %w", err)
	}

	// Create TLS configuration
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	// Create HTTP server
	mux := http.NewServeMux()
	mux.Handle("/mutate-cluster-x-k8s-io-v1beta1-cluster", webhook)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:      ":" + port,
		Handler:   mux,
		TLSConfig: tlsConfig,
	}

	// Start server in goroutine
	go func() {
		utils.AviLog.Infof("Starting VKS webhook server on port %s", port)
		if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			utils.AviLog.Errorf("VKS webhook server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-stopCh
	utils.AviLog.Infof("Shutting down VKS webhook server")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		utils.AviLog.Errorf("VKS webhook server shutdown error: %v", err)
		return err
	}

	utils.AviLog.Infof("VKS webhook server stopped")
	return nil
}

// getVCenterURLFromConfig retrieves vCenter URL from wcp-cluster-config ConfigMap
func getVCenterURLFromConfig(kubeClient kubernetes.Interface) string {
	// Try to get vCenter URL from wcp-cluster-config ConfigMap in vmware-system-capw namespace
	configMap, err := kubeClient.CoreV1().ConfigMaps("vmware-system-capw").Get(context.TODO(), "wcp-cluster-config", metav1.GetOptions{})
	if err != nil {
		utils.AviLog.Debugf("Failed to get wcp-cluster-config ConfigMap: %v", err)
		return ""
	}

	// Look for vCenter URL in the ConfigMap data
	if vcenterURL, exists := configMap.Data["vcenter-server"]; exists && vcenterURL != "" {
		utils.AviLog.Infof("Found vCenter URL in wcp-cluster-config: %s", vcenterURL)
		return vcenterURL
	}

	// Also check for alternative key names that might contain vCenter URL
	if vcenterURL, exists := configMap.Data["vcenter-url"]; exists && vcenterURL != "" {
		utils.AviLog.Infof("Found vCenter URL in wcp-cluster-config: %s", vcenterURL)
		return vcenterURL
	}

	if vcenterURL, exists := configMap.Data["vcenter_url"]; exists && vcenterURL != "" {
		utils.AviLog.Infof("Found vCenter URL in wcp-cluster-config: %s", vcenterURL)
		return vcenterURL
	}

	utils.AviLog.Debugf("vCenter URL not found in wcp-cluster-config ConfigMap")
	return ""
}

func init() {
	def_kube_config := os.Getenv("HOME") + "/.kube/config"
	flag.StringVar(&kubeconfig, "kubeconfig", def_kube_config, "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
