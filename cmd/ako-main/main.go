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

	crd "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/client/v1alpha1/clientset/versioned"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/api"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/api/models"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	advl4 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/service-apis/client/clientset/versioned"

	oshiftclient "github.com/openshift/client-go/route/clientset/versioned"
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

	InitializeAKOApi()

	InitializeAKC()
}

func InitializeAKOApi() {
	akoApi := api.NewServer(lib.GetAkoApiServerPort(), []models.ApiModel{})
	akoApi.InitApi()
	lib.SetApiServerInstance(akoApi)
}

func InitializeAKC() {
	var err error
	kubeCluster := false
	utils.AviLog.Info("AKO is running with version: ", version)
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

	var crdClient *crd.Clientset
	var advl4Client *advl4.Clientset
	if lib.GetAdvancedL4() {
		advl4Client, err = advl4.NewForConfig(cfg)
		if err != nil {
			utils.AviLog.Fatalf("Error building service-api clientset: %s", err.Error())
		}
		lib.SetAdvL4Clientset(advl4Client)
	} else {
		crdClient, err = crd.NewForConfig(cfg)
		if err != nil {
			utils.AviLog.Fatalf("Error building AKO CRD clientset: %s", err.Error())
		}
		lib.SetCRDClientset(crdClient)
	}

	dynamicClient, err := lib.NewDynamicClientSet(cfg)
	if err != nil {
		utils.AviLog.Warnf("Error while creating dynamic client %v", err)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		utils.AviLog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	oshiftClient, err := oshiftclient.NewForConfig(cfg)
	if err != nil {
		utils.AviLog.Warnf("Error in creating openshift clientset")
	}

	registeredInformers := lib.InformersToRegister(oshiftClient, kubeClient)

	informersArg := make(map[string]interface{})
	informersArg[utils.INFORMERS_OPENSHIFT_CLIENT] = oshiftClient

	if lib.GetNamespaceToSync() != "" {
		informersArg[utils.INFORMERS_NAMESPACE] = lib.GetNamespaceToSync()
	}
	informersArg[utils.INFORMERS_ADVANCED_L4] = lib.GetAdvancedL4()
	utils.NewInformers(utils.KubeClientIntf{ClientSet: kubeClient}, registeredInformers, informersArg)
	lib.NewDynamicInformers(dynamicClient)
	if lib.GetAdvancedL4() {
		k8s.NewAdvL4Informers(advl4Client)
	} else {
		k8s.NewCRDInformers(crdClient)
	}

	informers := k8s.K8sinformers{Cs: kubeClient, DynamicClient: dynamicClient, OshiftClient: oshiftClient}
	c := k8s.SharedAviController()
	stopCh := utils.SetupSignalHandler()
	ctrlCh := make(chan struct{})
	quickSyncCh := make(chan struct{})
	err = c.HandleConfigMap(informers, ctrlCh, stopCh, quickSyncCh)
	if err != nil {
		utils.AviLog.Errorf("Handleconfigmap error during reboot, shutting down AKO")
		return
	}

	err = k8s.PopulateCache()
	if err != nil {
		c.DisableSync = true
		utils.AviLog.Errorf("failed to populate cache, disabling sync")
		lib.ShutdownApi()
	}
	c.InitializeNamespaceSync()
	k8s.PopulateNodeCache(kubeClient)
	waitGroupMap := make(map[string]*sync.WaitGroup)
	wgIngestion := &sync.WaitGroup{}
	waitGroupMap["ingestion"] = wgIngestion
	wgFastRetry := &sync.WaitGroup{}
	waitGroupMap["fastretry"] = wgFastRetry
	wgSlowRetry := &sync.WaitGroup{}
	waitGroupMap["slowretry"] = wgSlowRetry
	wgGraph := &sync.WaitGroup{}
	waitGroupMap["graph"] = wgGraph
	go c.InitController(informers, registeredInformers, ctrlCh, stopCh, quickSyncCh, waitGroupMap)
	<-stopCh
	close(ctrlCh)
	doneChan := make(chan struct{})
	go func() {
		defer close(doneChan)
		wgIngestion.Wait()
		wgGraph.Wait()
		wgFastRetry.Wait()
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

func init() {
	def_kube_config := os.Getenv("HOME") + "/.kube/config"
	flag.StringVar(&kubeconfig, "kubeconfig", def_kube_config, "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
