/*
 * Copyright 2020-2021 VMware, Inc.
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

package dedicatedvstests

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha1/clientset/versioned/fake"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/api"
	utils "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

var KubeClient *k8sfake.Clientset
var CRDClient *crdfake.Clientset
var ctrl *k8s.AviController
var akoApiServer *api.FakeApiServer

func TestMain(m *testing.M) {
	os.Setenv("INGRESS_API", "extensionv1")
	os.Setenv("VIP_NETWORK_LIST", `[{"networkName":"net123"}]`)
	os.Setenv("CLUSTER_NAME", "cluster")
	os.Setenv("CLOUD_NAME", "CLOUD_VCENTER")
	os.Setenv("SEG_NAME", "Default-Group")
	os.Setenv("NODE_NETWORK_LIST", `[{"networkName":"net123","cidrs":["10.79.168.0/22"]}]`)
	os.Setenv("POD_NAMESPACE", utils.AKO_DEFAULT_NS)
	os.Setenv("SHARD_VS_SIZE", "DEDICATED")
	os.Setenv("AUTO_L4_FQDN", "default")

	akoControlConfig := lib.AKOControlConfig()
	akoControlConfig.SetAKOInstanceFlag(true)
	KubeClient = k8sfake.NewSimpleClientset()
	CRDClient = crdfake.NewSimpleClientset()
	akoControlConfig.SetCRDClientset(CRDClient)
	akoControlConfig.SetEventRecorder(lib.AKOEventComponent, KubeClient, true)
	data := map[string][]byte{
		"username": []byte("admin"),
		"password": []byte("admin"),
	}
	object := metav1.ObjectMeta{Name: "avi-secret", Namespace: utils.GetAKONamespace()}
	secret := &corev1.Secret{Data: data, ObjectMeta: object}
	KubeClient.CoreV1().Secrets(utils.GetAKONamespace()).Create(context.TODO(), secret, metav1.CreateOptions{})

	registeredInformers := []string{
		utils.ServiceInformer,
		utils.EndpointInformer,
		utils.IngressInformer,
		utils.IngressClassInformer,
		utils.SecretInformer,
		utils.NSInformer,
		utils.NodeInformer,
		utils.ConfigMapInformer,
	}
	utils.NewInformers(utils.KubeClientIntf{ClientSet: KubeClient}, registeredInformers)
	informers := k8s.K8sinformers{Cs: KubeClient}
	k8s.NewCRDInformers(CRDClient)

	mcache := cache.SharedAviObjCache()
	cloudObj := &cache.AviCloudPropertyCache{Name: "Default-Cloud", VType: "mock"}
	subdomains := []string{"avi.internal", ".com"}
	cloudObj.NSIpamDNS = subdomains
	mcache.CloudKeyCache.AviCacheAdd("Default-Cloud", cloudObj)

	akoApiServer = integrationtest.InitializeFakeAKOAPIServer()

	integrationtest.NewAviFakeClientInstance(KubeClient)
	defer integrationtest.AviFakeClientInstance.Close()

	ctrl = k8s.SharedAviController()
	stopCh := utils.SetupSignalHandler()
	ctrlCh := make(chan struct{})
	quickSyncCh := make(chan struct{})
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

	integrationtest.AddConfigMap(KubeClient)
	integrationtest.PollForSyncStart(ctrl, 10)

	ctrl.HandleConfigMap(informers, ctrlCh, stopCh, quickSyncCh)
	integrationtest.KubeClient = KubeClient
	integrationtest.AddDefaultIngressClass()
	ctrl.SetSEGroupCloudNameFromNSAnnotations()

	go ctrl.InitController(informers, registeredInformers, ctrlCh, stopCh, quickSyncCh, waitGroupMap)
	os.Exit(m.Run())
}

func SetUpTestForIngress(t *testing.T, modelNames ...string) {
	for _, model := range modelNames {
		objects.SharedAviGraphLister().Delete(model)
	}
	integrationtest.CreateSVC(t, "default", "avisvc", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEP(t, "default", "avisvc", false, false, "1.1.1")
}

func TearDownTestForIngress(t *testing.T, modelNames ...string) {
	for _, model := range modelNames {
		objects.SharedAviGraphLister().Delete(model)
	}
	integrationtest.DelSVC(t, "default", "avisvc")
	integrationtest.DelEP(t, "default", "avisvc")
}

func SetUpIngressForCacheSyncCheck(t *testing.T, tlsIngress, withSecret bool, modelNames ...string) {
	SetupDomain()
	SetUpTestForIngress(t, modelNames...)
	ingressObject := integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}
	if withSecret {
		integrationtest.AddSecret("my-secret", "default", "tlsCert", "tlsKey")
	}
	if tlsIngress {
		ingressObject.TlsSecretDNS = map[string][]string{
			"my-secret": {"foo.com"},
		}
	}
	ingrFake := ingressObject.Ingress()
	if _, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelNames[0], 5)
}

func SetupDomain() {
	mcache := cache.SharedAviObjCache()
	cloudObj := &cache.AviCloudPropertyCache{Name: "Default-Cloud", VType: "mock"}
	subdomains := []string{"avi.internal", ".com"}
	cloudObj.NSIpamDNS = subdomains
	mcache.CloudKeyCache.AviCacheAdd("Default-Cloud", cloudObj)
}

func TearDownIngressForCacheSyncCheck(t *testing.T, modelName string) {
	if err := KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "foo-with-targets", metav1.DeleteOptions{}); err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), "my-secret", metav1.DeleteOptions{})
	TearDownTestForIngress(t, modelName)
}
