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
	"flag"
	"fmt"
	"os"
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

	// Start VKS Cluster Watcher for event-driven processing
	if err := vksClusterWatcher.Start(stopCh); err != nil {
		utils.AviLog.Errorf("Failed to start VKS cluster watcher: %v", err.Error())
		// Don't fail startup - reconciliation loop will still work
	} else {
		utils.AviLog.Infof("Successfully started VKS cluster watcher")
	}

	// Initialize VKS admission webhook certificate management
	// This is only needed when VKS webhook is enabled
	vksWebhookEnabled := os.Getenv("VKS_WEBHOOK_ENABLED")
	if vksWebhookEnabled == "true" {
		utils.AviLog.Infof("VKS webhook enabled, initializing certificate management")

		// VKS webhook certificate management
		namespace := utils.GetAKONamespace()
		secretName := os.Getenv("VKS_WEBHOOK_SECRET_NAME")
		if secretName == "" {
			secretName = "ako-webhook-certs"
		}
		serviceName := os.Getenv("VKS_WEBHOOK_SERVICE_NAME")
		if serviceName == "" {
			serviceName = "ako-vks-webhook-service"
		}
		certDir := os.Getenv("VKS_WEBHOOK_CERT_DIR")
		if certDir == "" {
			certDir = "/etc/webhook/certs"
		}
		webhookConfigName := os.Getenv("VKS_WEBHOOK_CONFIG_NAME")
		if webhookConfigName == "" {
			webhookConfigName = "ako-vks-cluster-webhook"
		}

		// Initialize the certificate manager
		certManager := webhooks.NewVKSWebhookCertificateManager(
			kubeClient, namespace, secretName, serviceName, certDir, webhookConfigName)

		// Ensure certificates are ready before starting webhook
		if err := certManager.EnsureCertificates(context.TODO()); err != nil {
			utils.AviLog.Errorf("Failed to ensure VKS webhook certificates: %v", err.Error())
			// Don't fail startup - webhook will be disabled
		} else {
			utils.AviLog.Infof("VKS webhook certificates ready")
			// Start certificate rotation in background
			certManager.StartCertificateRotation(context.TODO(), 24*time.Hour)
		}

		utils.AviLog.Infof("VKS webhook certificate management initialized")
	}

	worker := c.InitFullSyncWorker()
	go worker.Run()

	<-stopCh
	worker.Shutdown()
	close(ctrlCh)
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
