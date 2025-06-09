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
	"context"
	"flag"
	"fmt"

	"github.com/go-logr/zapr"
	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/api"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/api/models"
	crd "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha1/clientset/versioned"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	v1alpha2crd "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha2/clientset/versioned"
	v1beta1crd "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1beta1/clientset/versioned"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	advl4 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/service-apis/client/clientset/versioned"

	oshiftclient "github.com/openshift/client-go/route/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	svcapi "sigs.k8s.io/service-apis/pkg/client/clientset/versioned"

	"github.com/fsnotify/fsnotify"
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
	if lib.IsPrometheusEnabled() {
		lib.SetPrometheusRegistry()
	}
	akoApi := api.NewServer(lib.GetAkoApiServerPort(), []models.ApiModel{}, lib.IsPrometheusEnabled(), lib.GetPrometheusRegistry())
	akoApi.InitApi()
	lib.SetApiServerInstance(akoApi)
}

func InitializeAKC() {
	var err error
	kubeCluster := false
	utils.AviLog.Infof("AKO is running with version: %s", version)

	// set the logger for k8s as AviLogger.
	klog.SetLogger(zapr.NewLogger(utils.AviLog.Sugar.Desugar()))

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

	// Configure QPS and Burst with custom values to increase rate limit
	cfg.QPS = 100
	cfg.Burst = 100

	// Initialize akoControlConfig
	akoControlConfig := lib.AKOControlConfig()
	//Used to set vrf context, static routes
	isPrimaryAKO, err := strconv.ParseBool(os.Getenv("PRIMARY_AKO_FLAG"))
	if err != nil {
		isPrimaryAKO = true
	}
	akoControlConfig.SetEndpointSlicesEnabled(lib.GetEndpointSliceEnabled())
	akoControlConfig.SetAKOInstanceFlag(isPrimaryAKO)
	akoControlConfig.SetAKOBlockedNSList(lib.GetGlobalBlockedNSList())
	akoControlConfig.SetControllerVRFContext(lib.GetControllerVRFContext())
	akoControlConfig.SetAKOPrometheusFlag(lib.IsPrometheusEnabled())
	akoControlConfig.SetAKOFQDNReusePolicy(strings.ToLower(os.Getenv("FQDN_REUSE_POLICY")))

	var crdClient *crd.Clientset
	var advl4Client *advl4.Clientset
	var svcAPIClient *svcapi.Clientset

	isDefaultLBController, err := strconv.ParseBool(os.Getenv("DEFAULT_LB_CONTROLLER"))
	if err != nil {
		isDefaultLBController = true
	}
	akoControlConfig.SetDefaultLBController(isDefaultLBController)

	v1beta1crdClient, err := v1beta1crd.NewForConfig(cfg)
	if err != nil {
		utils.AviLog.Fatalf("Error building AKO CRD v1beta1 clientset: %s", err.Error())
	}

	if utils.IsWCP() {
		advl4Client, err = advl4.NewForConfig(cfg)
		if err != nil {
			utils.AviLog.Fatalf("Error building service-api v1alpha1pre1 clientset: %s", err.Error())
		}
		akoControlConfig.SetAdvL4Clientset(advl4Client)
		akoControlConfig.SetCRDClientsetAndEnableInfraSettingParam(v1beta1crdClient)
	} else {
		if lib.UseServicesAPI() {
			svcAPIClient, err = svcapi.NewForConfig(cfg)
			if err != nil {
				utils.AviLog.Fatalf("Error building service-api clientset: %s", err.Error())
			}
			akoControlConfig.SetServicesAPIClientset(svcAPIClient)
		}

		// This is kept as MCI and Service Import uses v1alpha1
		// In Next release, MCI and serviceImport should be taken out
		crdClient, err = crd.NewForConfig(cfg)
		if err != nil {
			utils.AviLog.Fatalf("Error building AKO CRD clientset: %s", err.Error())
		}
		akoControlConfig.SetCRDClientset(crdClient)

		akoControlConfig.Setv1beta1CRDClientset(v1beta1crdClient)
		v1alpha2crdClient, err := v1alpha2crd.NewForConfig(cfg)
		if err != nil {
			utils.AviLog.Fatalf("Error building AKO CRD v1alpha2 clientset: %s", err.Error())
		}
		akoControlConfig.Setv1alpha2CRDClientset(v1alpha2crdClient)

	}

	dynamicClient, err := lib.NewDynamicClientSet(cfg)
	if err != nil {
		utils.AviLog.Warnf("Error while creating dynamic client %v", err)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		utils.AviLog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	akoControlConfig.SetEventRecorder(lib.AKOEventComponent, kubeClient, false)

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

	// Check for kubernetes apiserver version compatibility with AKO version.
	if serverVersionInfo, err := kubeClient.Discovery().ServerVersion(); err != nil {
		utils.AviLog.Warnf("Error while fetching kubernetes apiserver version")
	} else {
		serverVersion := fmt.Sprintf("%s.%s", serverVersionInfo.Major, serverVersionInfo.Minor)
		utils.AviLog.Infof("Kubernetes cluster apiserver version %s", serverVersion)
		if lib.CompareVersions(serverVersion, ">", lib.GetK8sMaxSupportedVersion()) ||
			lib.CompareVersions(serverVersion, "<", lib.GetK8sMinSupportedVersion()) {
			akoControlConfig.PodEventf(corev1.EventTypeWarning, lib.AKOShutdown, "Unsupported kubernetes apiserver %s version detected", serverVersion)
			utils.AviLog.Fatalf("Unsupported kubernetes apiserver version detected. Please check the supportability guide.")
		}
	}

	oshiftClient, err := oshiftclient.NewForConfig(cfg)
	if err != nil && (lib.GetCNIPlugin() == lib.OPENSHIFT_CNI || lib.GetCNIPlugin() == lib.OVN_KUBERNETES_CNI) {
		utils.AviLog.Fatalf("Failed to initialize Openshift ClientSet")
	}

	registeredInformers, err := lib.InformersToRegister(kubeClient, oshiftClient)
	if err != nil {
		utils.AviLog.Fatalf("Failed to initialize informers: %v, shutting down AKO, going to reboot", err)
	}

	informersArg := make(map[string]interface{})
	informersArg[utils.INFORMERS_OPENSHIFT_CLIENT] = oshiftClient
	informersArg[utils.INFORMERS_AKO_CLIENT] = crdClient

	if lib.GetNamespaceToSync() != "" {
		informersArg[utils.INFORMERS_NAMESPACE] = lib.GetNamespaceToSync()
	}

	// Namespace bound Secret informers should be initialized for AKO in VDS,
	// For AKO in VCF, we will need to watch on Secrets across all namespaces.
	if !utils.IsVCFCluster() && utils.GetAdvancedL4() {
		informersArg[utils.INFORMERS_ADVANCED_L4] = true
	}

	utils.NewInformers(utils.KubeClientIntf{ClientSet: kubeClient}, registeredInformers, informersArg)
	lib.NewDynamicInformers(dynamicClient, false)
	if utils.IsWCP() {
		k8s.NewInfraSettingCRDInformer()
		k8s.NewAdvL4Informers(advl4Client)
	} else {
		k8s.NewCRDInformers()
		if lib.UseServicesAPI() {
			k8s.NewSvcApiInformers(svcAPIClient)
		}
	}
	istioUpdateCh := make(chan struct{})

	// Set Istio Informers.
	if lib.IsIstioEnabled() {
		lib.SetIstioInitialized(false)
		akoControlConfig.PodEventf(corev1.EventTypeNormal, "IstioEnabled", "Adding certificate watcher for Istio")
		utils.AviLog.Infof("Adding certificate watcher for Istio")
		istioCertWatcher, err := fsnotify.NewWatcher()
		if err != nil {
			utils.AviLog.Fatal(err)
		}
		defer istioCertWatcher.Close()

		go istioWatcherEvents(istioCertWatcher, kubeClient, &istioUpdateCh)

		_, err = os.Stat(lib.IstioCertOutputPath)
		if err == nil {
			err := istioCertWatcher.Add(lib.IstioCertOutputPath)
			if err != nil {
				utils.AviLog.Fatal(err)
			}
			akoControlConfig.PodEventf(corev1.EventTypeNormal, "IstioWatcher", "Added path to %s to Istio watcher", lib.IstioCertOutputPath)
			initIstioSecrets(kubeClient, &istioUpdateCh)
		} else if os.IsNotExist(err) {
			err := istioCertWatcher.Add("/etc")
			akoControlConfig.PodEventf(corev1.EventTypeNormal, "IstioWatcher", "Added path to /etc to Istio watcher")
			if err != nil {
				utils.AviLog.Fatal(err)
			}
		}

	}

	informers := k8s.K8sinformers{Cs: kubeClient, DynamicClient: dynamicClient, OshiftClient: oshiftClient}
	c := k8s.SharedAviController()
	stopCh := utils.SetupSignalHandler()
	ctrlCh := make(chan struct{})
	quickSyncCh := make(chan struct{})

	if utils.IsVCFCluster() {
		c.InitVCFHandlers(kubeClient, ctrlCh, stopCh)
	}

	err = k8s.PopulateControllerProperties(kubeClient)
	if err != nil {
		utils.AviLog.Warnf("Error while fetching secret for AKO bootstrap %s", err)
		lib.ShutdownApi()
	}

	aviRestClientPool := avicache.SharedAVIClients(lib.GetTenant())
	if aviRestClientPool == nil {
		utils.AviLog.Fatalf("Avi client not initialized")
	}

	if akoControlConfig.GetAKOAKOPrometheusFlag() {
		lib.RegisterPromMetrics()
	}

	if aviRestClientPool != nil && !avicache.IsAviClusterActive(aviRestClientPool.AviClient[0]) {
		akoControlConfig.PodEventf(corev1.EventTypeWarning, lib.AKOShutdown, "Avi Controller Cluster state is not Active")
		utils.AviLog.Fatalf("Avi Controller Cluster state is not Active, shutting down AKO")
	}

	akoControlConfig.SetLicenseType(aviRestClientPool.AviClient[0])

	if utils.GetAdvancedL4() {
		err, seGroupToUse := lib.FetchSEGroupWithMarkerSet(aviRestClientPool.AviClient[0])
		if err != nil {
			utils.AviLog.Warnf("Setting SEGroup with markerset failed: %s", err)
		}
		if seGroupToUse == "" {
			utils.AviLog.Infof("Continuing with Default-Group SEGroup")
			seGroupToUse = lib.DEFAULT_SE_GROUP
		}
		lib.SetSEGName(seGroupToUse)
	}

	err = c.HandleConfigMap(informers, ctrlCh, stopCh, quickSyncCh)
	if err != nil {
		utils.AviLog.Errorf("Handle configmap error during reboot, shutting down AKO. Error is: %v", err)
		return
	}
	if !utils.IsVCFCluster() {
		if _, err := lib.GetVipNetworkListEnv(); err != nil {
			utils.AviLog.Fatalf("Error in getting VIP network %s, shutting down AKO", err)
		}
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
	wgStatus := &sync.WaitGroup{}
	waitGroupMap["status"] = wgStatus
	wgLeaderElection := &sync.WaitGroup{}
	waitGroupMap["leaderElection"] = wgLeaderElection

	if lib.IsIstioEnabled() {
		utils.AviLog.Infof("Waiting on Istio resources creation")
		<-istioUpdateCh
	}

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
		wgLeaderElection.Wait()
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

func istioWatcherEvents(watcher *fsnotify.Watcher, kc *kubernetes.Clientset, istioUpdateCh *chan struct{}) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op == fsnotify.Write || event.Op == fsnotify.Create {
				utils.AviLog.Infof("Istio watcher event, modified file: %s", event.Name)

				if strings.HasSuffix(event.Name, "istio-output-certs") {
					watcher.Remove("/etc")
					utils.AviLog.Infof("removed path to /etc to istio watcher")
					watcher.Add(lib.IstioCertOutputPath)
					utils.AviLog.Infof("added path to %s to istio watcher", lib.IstioCertOutputPath)
					initIstioSecrets(kc, istioUpdateCh)
				} else {
					if strings.HasSuffix(event.Name, "cert-chain.pem") ||
						strings.HasSuffix(event.Name, "key.pem") ||
						strings.HasSuffix(event.Name, "root-cert.pem") {
						lib.CreateIstioSecretFromCert(event.Name, kc)
						if !lib.IsChanClosed(*istioUpdateCh) {
							if lib.GetIstioCertSet().Len() == 3 {
								close(*istioUpdateCh)
							}
						}
					}
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			utils.AviLog.Warnf("Istio watcher event, error: %s", err.Error())
		}
	}
}

func initIstioSecrets(kc *kubernetes.Clientset, istioUpdateCh *chan struct{}) {

	entries, err := os.ReadDir(lib.IstioCertOutputPath + "/")
	if err != nil {
		utils.AviLog.Warnf("Cannot read %s, error: %s", lib.IstioCertOutputPath, err.Error())
		return
	}
	if len(entries) == 0 {
		utils.AviLog.Infof("%s is empty", lib.IstioCertOutputPath)
		return
	}
	files := make([]fs.FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, error := entry.Info()
		if error != nil {
			utils.AviLog.Warnf("Cannot read %s, error: %s", lib.IstioCertOutputPath, err.Error())
			return
		}
		files = append(files, info)
	}

	if len(files) == 0 {
		utils.AviLog.Infof("%s is empty", lib.IstioCertOutputPath)
		return
	} else {
		for _, file := range files {
			if file.Name() == "cert-chain.pem" ||
				file.Name() == "key.pem" ||
				file.Name() == "root-cert.pem" {
				lib.CreateIstioSecretFromCert(lib.IstioCertOutputPath+"/"+file.Name(), kc)
			}
		}
	}
	utils.AviLog.Infof("%s initialized", lib.IstioSecret)
	close(*istioUpdateCh)
}
