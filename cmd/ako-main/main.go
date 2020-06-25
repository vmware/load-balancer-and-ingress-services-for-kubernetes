/*
 * Copyright 2019-2020 VMware, Inc.
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
	"flag"
	"os"
	"sync"
	"time"

	"ako/pkg/k8s"
	"ako/pkg/lib"

	"github.com/avinetworks/container-lib/api"
	"github.com/avinetworks/container-lib/api/models"
	"github.com/avinetworks/container-lib/utils"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	masterURL  string
	kubeconfig string
)

func main() {
	go InitializeAKCApi()
	InitializeAKC()
}

func InitializeAKCApi() {
	akoApi := &api.ApiServer{
		Port:   "8080",
		Models: []models.ApiModel{},
	}

	akoApi.InitApi()
}

func InitializeAKC() {
	kubeCluster := false
	// Check if we are running inside kubernetes. Hence try authenticating with service token
	cfg, err := rest.InClusterConfig()
	if err != nil {
		utils.AviLog.Warnf("We are not running inside kubernetes cluster. %s", err.Error())
	} else {
		utils.AviLog.Info("We are running inside kubernetes cluster. Won't use kubeconfig files.")
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
	registeredInformers := []string{
		utils.ServiceInformer,
		utils.EndpointInformer,
		utils.IngressInformer,
		utils.SecretInformer,
		utils.NSInformer,
		utils.NodeInformer,
		utils.ConfigMapInformer,
	}
	if lib.GetNamespaceToSync() != "" {
		namespaceMap := make(map[string]interface{})
		namespaceMap[utils.INFORMERS_NAMESPACE] = lib.GetNamespaceToSync()
		utils.NewInformers(utils.KubeClientIntf{ClientSet: kubeClient}, registeredInformers, namespaceMap)
	} else {
		utils.NewInformers(utils.KubeClientIntf{ClientSet: kubeClient}, registeredInformers)
	}
	lib.NewDynamicInformers(dynamicClient)

	informers := k8s.K8sinformers{Cs: kubeClient, DynamicClient: dynamicClient}
	c := k8s.SharedAviController()
	stopCh := utils.SetupSignalHandler()
	ctrlCh := make(chan struct{})
	quickSyncCh := make(chan struct{})
	c.HandleConfigMap(informers, ctrlCh, stopCh, quickSyncCh)
	k8s.PopulateCache()
	k8s.PopulateNodeCache(kubeClient)
	waitGroupMap := make(map[string]*sync.WaitGroup)
	wgIngestion := &sync.WaitGroup{}
	waitGroupMap["ingestion"] = wgIngestion
	wgFastRetry := &sync.WaitGroup{}
	waitGroupMap["fastretry"] = wgFastRetry
	wgGraph := &sync.WaitGroup{}
	waitGroupMap["graph"] = wgGraph
	go c.InitController(informers, ctrlCh, stopCh, quickSyncCh, waitGroupMap)
	<-stopCh
	close(ctrlCh)
	timeoutChan := make(chan struct{})
	// Timeout after 10 seconds.
	timeout := 60 * time.Second
	go func() {
		defer close(timeoutChan)
		wgIngestion.Wait()
		wgGraph.Wait()
		wgFastRetry.Wait()
	}()
	select {
	case <-timeoutChan:
		utils.AviLog.Warnf("Timed out while waiting for threads to return, going to stop AKO. Time waited 60 seconds")
		return
	case <-time.After(timeout):
		return
	}

}

func init() {
	def_kube_config := os.Getenv("HOME") + "/.kube/config"
	flag.StringVar(&kubeconfig, "kubeconfig", def_kube_config, "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
