/*
 * Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
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

	akogatewaylib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-infra/ingestion"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	v1beta1crd "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1beta1/clientset/versioned"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	gatewayclientset "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned"
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
	clusterName := lib.GetClusterName()
	if !lib.GetVPCMode() {
		segExists := a.SetupSEGroup(aviCloud)
		clusterName, err = a.DeriveClusterNameToBeUsedInAKOUser(segExists)
		if err != nil {
			lib.AKOControlConfig().PodEventf(corev1.EventTypeWarning, "ClusterNameDerivationFailure", err.Error())
			utils.AviLog.Fatalf("Failed to derive Cluster name to be used in the AKO User, err: %s", err.Error())
		}
		c.AddAvailabilityZoneCREventHandler(stopCh)
	} else if lib.IsGatewayAPICapabilityEnabled() {
		gwApiClient, err := gatewayclientset.NewForConfig(cfg)
		if err != nil {
			utils.AviLog.Fatalf("Error building gateway-api clientset: %s", err.Error())
		}
		akogatewaylib.AKOControlConfig().SetGatewayAPIClientset(gwApiClient)
		err = akogatewaylib.CreateVCFGatewayClass()
		if err != nil {
			utils.AviLog.Fatalf("Error creating gateway class: %s", err.Error())
		}
	}
	lib.SetClusterName(clusterName)
	lib.SetAKOUser(lib.AKOPrefix)
	c.AddNamespaceEventHandler(stopCh)
	c.Sync()
	a.AnnotateSystemNamespace(lib.GetClusterID(), utils.CloudName, clusterName)
	c.AddNetworkInfoEventHandler(stopCh)

	// Add VKS capability event handler after system namespace is annotated
	c.AddVKSCapabilityEventHandler(stopCh)

	worker := c.InitFullSyncWorker()
	go worker.Run()

	<-stopCh
	worker.Shutdown()
	close(ctrlCh)
}

func init() {
	def_kube_config := os.Getenv("HOME") + "/.kube/config"
	flag.StringVar(&kubeconfig, "kubeconfig", def_kube_config, "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
