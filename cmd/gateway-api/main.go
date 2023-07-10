/*
 * Copyright 2023-2024 VMware, Inc.
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
	"os"
	"sync"
	"time"

	gwk8s "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/k8s"
	gwlib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	gwApi "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned"
)

var (
	masterURL  string
	kubeconfig string
	version    = "dev"
)

func main() {
	Initialize()
}

func Initialize() {

	var err error
	kubeCluster := false
	utils.AviLog.Infof("AKO is running with version: %s", version)
	// Check if we are running inside kubernetes. Hence try authenticating with service token
	cfg, err := rest.InClusterConfig()
	if err != nil {
		utils.AviLog.Warnf("We are not running inside kubernetes cluster. %s", err.Error())
	} else {
		utils.AviLog.Info("We are running inside kubernetes cluster. Won't use kubeconfig files.")
		kubeCluster = true
	}
	if !kubeCluster {
		cfg, err = clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
		utils.AviLog.Infof("master: %s", masterURL)
		if err != nil {
			utils.AviLog.Fatalf("Error building kubeconfig: %s", err.Error())
		}
	}

	akoControlConfig := gwlib.AKOControlConfig()
	lib.SetAKOUser("ako-gw-")
	var gwApiClient *gwApi.Clientset

	gwApiClient, err = gwApi.NewForConfig(cfg)
	if err != nil {
		utils.AviLog.Fatalf("Error building gateway-api clientset: %s", err.Error())
	}
	akoControlConfig.SetGatewayAPIClientset(gwApiClient)

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		utils.AviLog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	utils.AviLog.Infof("Successfully created kube client for ako-gateway-api")

	akoControlConfig.SetEventRecorder(lib.AKOGatewayEventComponent, kubeClient, false)
	pod, err := kubeClient.CoreV1().Pods(utils.GetAKONamespace()).Get(context.TODO(), os.Getenv("POD_NAME"), metav1.GetOptions{})
	if err != nil {
		utils.AviLog.Warnf("Error getting AKO pod details, %s.", err.Error())
	}
	akoControlConfig.SaveAKOPodObjectMeta(pod)

	registeredInformers, err := gwlib.InformersToRegister(kubeClient)
	if err != nil {
		utils.AviLog.Fatalf("Failed to initialize informers: %v, shutting down AKO-Infra, going to reboot", err)
	}

	informersArg := make(map[string]interface{})

	utils.NewInformers(utils.KubeClientIntf{ClientSet: kubeClient}, registeredInformers, informersArg)

	informers := k8s.K8sinformers{Cs: kubeClient}
	c := gwk8s.SharedGatewayController()
	stopCh := utils.SetupSignalHandler()
	ctrlCh := make(chan struct{})
	quickSyncCh := make(chan struct{})

	err = k8s.PopulateControllerProperties(kubeClient)
	if err != nil {
		utils.AviLog.Warnf("Error while fetching secret for AKO bootstrap %s", err)
		lib.ShutdownApi()
	}

	aviRestClientPool := avicache.SharedAVIClients()
	if aviRestClientPool == nil {
		utils.AviLog.Fatalf("Avi client not initialized")
	}

	if aviRestClientPool != nil && !avicache.IsAviClusterActive(aviRestClientPool.AviClient[0]) {
		akoControlConfig.PodEventf(corev1.EventTypeWarning, lib.AKOShutdown, "Avi Controller Cluster state is not Active")
		utils.AviLog.Fatalf("Avi Controller Cluster state is not Active, shutting down AKO")
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
		//wgLeaderElection.Wait()
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
