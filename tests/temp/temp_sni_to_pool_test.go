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

package temp

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha1/clientset/versioned/fake"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/api"
	utils "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

// Use this file to execute tests that need special handling like - configmap create/delete/update operations etc.
// Pls delete this file/folder once this feature is deprecated in value of http caching on PG.

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
	os.Setenv("SHARD_VS_SIZE", "LARGE")
	os.Setenv("POD_NAME", "ako-0")

	akoControlConfig := lib.AKOControlConfig()
	KubeClient = k8sfake.NewSimpleClientset()
	CRDClient = crdfake.NewSimpleClientset()
	akoControlConfig.SetCRDClientset(CRDClient)
	akoControlConfig.SetAKOInstanceFlag(true)
	akoControlConfig.SetEventRecorder(lib.AKOEventComponent, KubeClient, true)

	registeredInformers := []string{
		utils.ServiceInformer,
		utils.EndpointSlicesInformer,
		utils.IngressInformer,
		utils.IngressClassInformer,
		utils.SecretInformer,
		utils.NSInformer,
		utils.NodeInformer,
		utils.ConfigMapInformer,
	}
	utils.NewInformers(utils.KubeClientIntf{ClientSet: KubeClient}, registeredInformers)
	informers := k8s.K8sinformers{Cs: KubeClient}
	k8s.NewCRDInformers()

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

	integrationtest.AddConfigMap(KubeClient)
	integrationtest.PollForSyncStart(ctrl, 10)

	ctrl.HandleConfigMap(informers, ctrlCh, stopCh, quickSyncCh)
	ctrl.SetSEGroupCloudNameFromNSAnnotations()

	integrationtest.KubeClient = KubeClient
	integrationtest.AddDefaultIngressClass()

	go ctrl.InitController(informers, registeredInformers, ctrlCh, stopCh, quickSyncCh, waitGroupMap)
	os.Exit(m.Run())
}
func TestSniPoolNoPGForSNI(t *testing.T) {
	/*
		-> Create Ingress with TLS key/secret and 2 paths
		-> Verify removing path works by updating Ingress with single path
		-> Verify adding path works by updating Ingress with 2 new paths
	*/
	UpdateConfigMap(lib.NO_PG_FOR_SNI, "true")
	g := gomega.NewGomegaWithT(t)
	integrationtest.AddSecret("my-secret", "default", "tlsCert", "tlsKey")
	modelName := "admin/cluster--Shared-L7-0"
	SetUpTestForIngress(t, modelName)
	ingrFake1 := (integrationtest.FakeIngress{
		Name:      "ingress-shp",
		Namespace: "default",
		DnsNames:  []string{"foo.com"},
		Ips:       []string{"8.8.8.8"},
		Paths:     []string{"/foo", "/bar"},
		HostNames: []string{"v1"},
		TlsSecretDNS: map[string][]string{
			"my-secret": {"foo.com"},
		},
		ServiceName: "avisvc",
	}).IngressMultiPath()
	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Eventually(len(nodes), 30*time.Second).Should(gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].PoolRefs)).Should(gomega.Equal(2))

	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}

	_, err = (integrationtest.FakeIngress{
		Name:      "ingress-shp",
		Namespace: "default",
		DnsNames:  []string{"foo.com"},
		Ips:       []string{"8.8.8.8"},
		Paths:     []string{"/foo"},
		HostNames: []string{"v1"},
		TlsSecretDNS: map[string][]string{
			"my-secret": {"foo.com"},
		},
		ServiceName: "avisvc",
	}).UpdateIngress()
	if err != nil {
		t.Fatalf("error in updating ingress %s", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)

	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Eventually(func() []*avinodes.AviPoolNode {
			return nodes[0].SniNodes[0].PoolRefs
		}, 10*time.Second).Should(gomega.HaveLen(1))
		g.Expect((nodes[0].SniNodes[0].PoolRefs[0].Name)).Should((gomega.Equal("cluster--default-foo.com_foo-ingress-shp--policy-to-pool")))
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}

	_, err = (integrationtest.FakeIngress{
		Name:      "ingress-shp",
		Namespace: "default",
		DnsNames:  []string{"foo.com"},
		Ips:       []string{"8.8.8.8"},
		Paths:     []string{"/foo", "/bar", "/baz"},
		HostNames: []string{"v1"},
		TlsSecretDNS: map[string][]string{
			"my-secret": {"foo.com"},
		},
		ServiceName: "avisvc",
	}).UpdateIngress()
	if err != nil {
		t.Fatalf("error in updating ingress %s", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)

	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Eventually(func() []*avinodes.AviPoolNode {
			return nodes[0].SniNodes[0].PoolRefs
		}, 10*time.Second).Should(gomega.HaveLen(3))

	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-shp", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), "my-secret", metav1.DeleteOptions{})
	VerifySNIIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, modelName)
	UpdateConfigMap(lib.NO_PG_FOR_SNI, "false")
}

func UpdateConfigMap(key, val string) {
	KubeClient.CoreV1().ConfigMaps(utils.GetAKONamespace()).Delete(context.TODO(), "avi-k8s-config", metav1.DeleteOptions{})
	aviCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: utils.GetAKONamespace(),
			Name:      "avi-k8s-config",
		},
		Data: make(map[string]string),
	}
	aviCM.Data[key] = val
	aviCM.ResourceVersion = "2"
	KubeClient.CoreV1().ConfigMaps(utils.GetAKONamespace()).Create(context.TODO(), aviCM, metav1.CreateOptions{})
	// Wait for the configmap changes to take effect
	time.Sleep(3 * time.Second)
}

func SetUpTestForIngress(t *testing.T, modelNames ...string) {
	for _, model := range modelNames {
		objects.SharedAviGraphLister().Delete(model)
	}
	integrationtest.CreateSVC(t, "default", "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, "default", "avisvc", false, false, "1.1.1")
}

func TearDownTestForIngress(t *testing.T, modelNames ...string) {
	for _, model := range modelNames {
		objects.SharedAviGraphLister().Delete(model)
	}
	integrationtest.DelSVC(t, "default", "avisvc")
	integrationtest.DelEPS(t, "default", "avisvc")
}

func VerifySNIIngressDeletion(t *testing.T, g *gomega.WithT, aviModel interface{}, sniCount int) {
	var nodes []*avinodes.AviVsNode
	g.Eventually(func() []*avinodes.AviVsNode {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].SniNodes
	}, 10*time.Second).Should(gomega.HaveLen(sniCount))

	g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(sniCount))
}
