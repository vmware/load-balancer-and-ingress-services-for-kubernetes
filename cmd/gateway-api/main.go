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
	"os"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	gatewayclientset "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned"

	akogatewayk8s "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/k8s"
	akogatewaylib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	v1beta1crd "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1beta1/clientset/versioned"
)

var (
	masterURL  string
	kubeconfig string
	version    = "dev"
)

func init() {
	def_kube_config := os.Getenv("HOME") + "/.kube/config"
	flag.StringVar(&kubeconfig, "kubeconfig", def_kube_config, "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}

func main() {
	Initialize()
}

func Initialize() {

	os.Setenv(lib.ENABLE_EVH, "true")
	os.Setenv(lib.VIP_PER_NAMESPACE, "false")
	os.Setenv("SHARD_VS_SIZE", "LARGE")
	os.Setenv("PASSTHROUGH_SHARD_SIZE", "LARGE")
	os.Setenv(lib.DISABLE_STATIC_ROUTE_SYNC, "true")
	os.Setenv("PROMETHEUS_ENABLED", "false")
	os.Setenv("PRIMARY_AKO_FLAG", "false")

	utils.AviLog.SetLevel("DEBUG") // TODO: integrate the configmap to this pod and remove this hardcoding.
	utils.AviLog.Infof("AKO is running with version: %s", version)

	kubeCluster := false
	// Check if we are running inside kubernetes. Hence try authenticating with service token
	cfg, err := rest.InClusterConfig()
	if err != nil {
		utils.AviLog.Warnf("We are not running inside kubernetes cluster. %s", err.Error())
	} else {
		utils.AviLog.Infof("We are running inside kubernetes cluster. Won't use kubeconfig files.")
		kubeCluster = true
	}
	if !kubeCluster {
		cfg, err = clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
		utils.AviLog.Infof("master: %s", masterURL)
		if err != nil {
			utils.AviLog.Fatalf("Error building kubeconfig: %s", err.Error())
		}
	}

	akoControlConfig := akogatewaylib.AKOControlConfig()

	// Set the user with prefix
	_ = lib.AKOControlConfig()
	//TODO handle leader logic, must not be used with HA
	lib.AKOControlConfig().SetIsLeaderFlag(true)

	gwApiClient, err := gatewayclientset.NewForConfig(cfg)
	if err != nil {
		utils.AviLog.Fatalf("Error building gateway-api clientset: %s", err.Error())
	}
	akoControlConfig.SetGatewayAPIClientset(gwApiClient)

	dynamicClient, err := akogatewaylib.NewDynamicClientSet(cfg)
	if err != nil {
		utils.AviLog.Warnf("Error while creating dynamic client %v", err)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		utils.AviLog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	utils.AviLog.Infof("Successfully created kube client for ako-gateway-api")

	v1beta1crdClient, err := v1beta1crd.NewForConfig(cfg)
	if err != nil {
		utils.AviLog.Fatalf("Error building AKO CRD v1beta1 clientset: %s", err.Error())
	}
	// Enabling AviInfraSetting CR in GatewayAPI Controller
	akogatewaylib.AKOControlConfig().SetV1Beta1CRDClientSetAndEnableAviInfraSettingParam(v1beta1crdClient)

	akoControlConfig.SetEventRecorder(lib.AKOGatewayEventComponent, kubeClient, false)

	// POD_NAME is not set in case of a WCP cluster
	if os.Getenv("POD_NAME") == "" {
		pods, err := kubeClient.CoreV1().Pods(utils.GetAKONamespace()).List(context.TODO(), metav1.ListOptions{Limit: 1})
		if err != nil {
			utils.AviLog.Warnf("Error getting AKO pod details, %s.", err.Error())
		} else {
			for _, pod := range pods.Items {
				akoControlConfig.SaveAKOPodObjectMeta(&pod)
			}
		}
	} else {
		pod, err := kubeClient.CoreV1().Pods(utils.GetAKONamespace()).Get(context.TODO(), os.Getenv("POD_NAME"), metav1.GetOptions{})
		if err != nil {
			utils.AviLog.Warnf("Error getting AKO pod details, %s.", err.Error())
		}
		akoControlConfig.SaveAKOPodObjectMeta(pod)
	}
	registeredInformers, err := akogatewaylib.InformersToRegister(kubeClient)
	if err != nil {
		utils.AviLog.Fatalf("Failed to initialize informers: %v, shutting down AKO-Infra, going to reboot", err)
	}

	informersArg := make(map[string]interface{})

	utils.NewInformers(utils.KubeClientIntf{ClientSet: kubeClient}, registeredInformers, informersArg)
	akogatewaylib.NewDynamicInformers(dynamicClient, false)
	informers := k8s.K8sinformers{Cs: kubeClient, DynamicClient: dynamicClient}
	c := akogatewayk8s.SharedGatewayController()
	c.InitGatewayAPIInformers(gwApiClient)
	stopCh := utils.SetupSignalHandler()
	ctrlCh := make(chan struct{})
	quickSyncCh := make(chan struct{})

	akogatewayk8s.NewInfraSettingCRDInformer()

	if utils.IsVCFCluster() {
		// AKO will be primary by default in VCF deployments
		lib.AKOControlConfig().SetAKOInstanceFlag(true)
		k8s.SharedAviController().InitVCFHandlers(kubeClient, ctrlCh, stopCh)
	}

	lib.SetAKOUser(akogatewaylib.Prefix)
	lib.SetNamePrefix(akogatewaylib.Prefix)

	err = k8s.PopulateControllerProperties(kubeClient)
	if err != nil {
		utils.AviLog.Warnf("Error while fetching secret for AKO bootstrap %s", err)
		lib.ShutdownApi()
	}

	aviRestClientPool := avicache.SharedAVIClients(lib.GetTenant())
	if aviRestClientPool == nil {
		utils.AviLog.Fatalf("Avi client not initialized")
	}

	if aviRestClientPool != nil && !avicache.IsAviClusterActive(aviRestClientPool.AviClient[0]) {
		akoControlConfig.PodEventf(corev1.EventTypeWarning, lib.AKOShutdown, "Avi Controller Cluster state is not Active")
		utils.AviLog.Fatalf("Avi Controller Cluster state is not Active, shutting down AKO")
	}

	if utils.IsVCFCluster() {
		k8s.SharedAviController().InitVCFHandlers(kubeClient, ctrlCh, stopCh)
	}

	err = c.HandleConfigMap(informers, ctrlCh, stopCh, quickSyncCh)
	if err != nil {
		utils.AviLog.Errorf("Handle configmap error during reboot, shutting down AKO. Error is: %v", err)
		return
	}

	waitGroupMap := make(map[string]*sync.WaitGroup)
	wgIngestion := &sync.WaitGroup{}
	waitGroupMap["ingestion"] = wgIngestion
	wgFastRetry := &sync.WaitGroup{}
	waitGroupMap["fastretry"] = wgFastRetry
	wgSlowRetry := &sync.WaitGroup{}
	waitGroupMap["slowretry"] = wgSlowRetry
	wgGraph := &sync.WaitGroup{}
	waitGroupMap["graph"] = wgGraph
	wgStatus := &sync.WaitGroup{}
	waitGroupMap["status"] = wgStatus

	k8s.PopulateNodeCache(kubeClient)

	go c.InitController(informers, registeredInformers, ctrlCh, stopCh, quickSyncCh, waitGroupMap)

	<-stopCh
	close(ctrlCh)
	doneChan := make(chan struct{})
	go func() {
		defer close(doneChan)
		wgIngestion.Wait()
		wgGraph.Wait()
		wgFastRetry.Wait()
		wgStatus.Wait()
	}()

	// Timeout after 60 seconds.
	timeout := 60 * time.Second
	select {
	case <-doneChan:
		return
	case <-time.After(timeout):
		utils.AviLog.Warnf("Timed out while waiting for threads to return, going to stop AKO. Time waited 60 seconds")
		return
	}
}
