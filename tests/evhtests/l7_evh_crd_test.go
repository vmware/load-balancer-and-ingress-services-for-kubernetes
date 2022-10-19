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

package evhtests

import (
	"context"
	"flag"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/api"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha1"
	crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha1/clientset/versioned/fake"
	utils "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

var KubeClient *k8sfake.Clientset
var CRDClient *crdfake.Clientset
var ctrl *k8s.AviController
var akoApiServer *api.FakeApiServer
var keyChan chan string

var isVipPerNS = flag.String("isVipPerNS", "false", "is vip per namespace enabled")

func setVipPerNS(vipPerNS string) {
	if vipPerNS == "true" {
		os.Setenv("VIP_PER_NAMESPACE", "true")
	}
}

func GetModelName(hostname, namespace string) (string, string) {
	vsName := "cluster--Shared-L7-EVH-"
	if !lib.VIPPerNamespace() {
		vsName += strconv.Itoa(int(utils.Bkt(hostname, 8)))
		return "admin/" + vsName, vsName
	}
	vsName += "NS-" + namespace
	return "admin/" + vsName, vsName
}

func TestMain(m *testing.M) {
	flag.Parse()
	setVipPerNS(*isVipPerNS)

	os.Setenv("INGRESS_API", "extensionv1")
	os.Setenv("VIP_NETWORK_LIST", `[{"networkName":"net123"}]`)
	os.Setenv("CLUSTER_NAME", "cluster")
	os.Setenv("CLOUD_NAME", "CLOUD_VCENTER")
	os.Setenv("SEG_NAME", "Default-Group")
	os.Setenv("NODE_NETWORK_LIST", `[{"networkName":"net123","cidrs":["10.79.168.0/22"]}]`)
	os.Setenv("ENABLE_EVH", "true")
	os.Setenv("POD_NAMESPACE", utils.AKO_DEFAULT_NS)
	os.Setenv("SHARD_VS_SIZE", "LARGE")
	os.Setenv("AUTO_L4_FQDN", "default")

	akoControlConfig := lib.AKOControlConfig()
	KubeClient = k8sfake.NewSimpleClientset()
	CRDClient = crdfake.NewSimpleClientset()
	akoControlConfig.SetCRDClientset(CRDClient)
	akoControlConfig.SetAKOInstanceFlag(true)
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
	ctrl.HandleConfigMap(informers, ctrlCh, stopCh, quickSyncCh)
	integrationtest.KubeClient = KubeClient
	integrationtest.AddDefaultIngressClass()
	keyChan = make(chan string)
	ctrl.SetSEGroupCloudNameFromNSAnnotations()
	go ctrl.InitController(informers, registeredInformers, ctrlCh, stopCh, quickSyncCh, waitGroupMap)
	os.Exit(m.Run())
}

func SetupDomain() {
	mcache := cache.SharedAviObjCache()
	cloudObj := &cache.AviCloudPropertyCache{Name: "Default-Cloud", VType: "mock"}
	subdomains := []string{"avi.internal", ".com"}
	cloudObj.NSIpamDNS = subdomains
	mcache.CloudKeyCache.AviCacheAdd("Default-Cloud", cloudObj)
}

func SetUpTestForIngress(t *testing.T, modelNames ...string) {
	for _, model := range modelNames {
		objects.SharedAviGraphLister().Delete(model)
	}
	integrationtest.CreateSVC(t, "default", "avisvc", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEP(t, "default", "avisvc", false, false, "1.1.1")
}

func TearDownTestForIngress(t *testing.T, modelNames ...string) {
	// for _, model := range modelNames {
	// 	objects.SharedAviGraphLister().Delete(model)
	// }
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

func TearDownIngressForCacheSyncCheck(t *testing.T, modelName string) {
	if err := KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "foo-with-targets", metav1.DeleteOptions{}); err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), "my-secret", metav1.DeleteOptions{})
	TearDownTestForIngress(t, modelName)
}

func TestCreateUpdateDeleteHostRuleForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	hrname := "samplehr-foo"
	SetUpIngressForCacheSyncCheck(t, true, true, modelName)

	integrationtest.SetupHostRule(t, hrname, "foo.com", true)

	g.Eventually(func() string {
		hostrule, _ := CRDClient.AkoV1alpha1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/samplehr-foo", true)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(*nodes[0].EvhNodes[0].Enabled).To(gomega.Equal(true))
	//At rest layer, sslref are switched to Parent
	g.Expect(nodes[0].EvhNodes[0].SSLKeyCertAviRef).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].SSLKeyCertAviRef[0]).To(gomega.ContainSubstring("thisisaviref-sslkey"))
	g.Expect(nodes[0].EvhNodes[0].WafPolicyRef).To(gomega.ContainSubstring("thisisaviref-waf"))
	g.Expect(nodes[0].EvhNodes[0].AppProfileRef).To(gomega.ContainSubstring("thisisaviref-appprof"))
	g.Expect(nodes[0].EvhNodes[0].AnalyticsProfileRef).To(gomega.ContainSubstring("thisisaviref-analyticsprof"))
	g.Expect(nodes[0].EvhNodes[0].ErrorPageProfileRef).To(gomega.ContainSubstring("thisisaviref-errorprof"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicySetRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicySetRefs[0]).To(gomega.ContainSubstring("thisisaviref-httpps2"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicySetRefs[1]).To(gomega.ContainSubstring("thisisaviref-httpps1"))
	g.Expect(nodes[0].EvhNodes[0].VsDatascriptRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].VsDatascriptRefs[0]).To(gomega.ContainSubstring("thisisaviref-ds2"))
	g.Expect(nodes[0].EvhNodes[0].VsDatascriptRefs[1]).To(gomega.ContainSubstring("thisisaviref-ds1"))
	g.Expect(nodes[0].SSLProfileRef).To(gomega.ContainSubstring("thisisaviref-sslprof"))
	g.Expect(nodes[0].EvhNodes[0].VHDomainNames).To(gomega.ContainElement("bar.com"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].Hosts).To(gomega.ContainElement("bar.com"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Host).To(gomega.ContainElement("bar.com"))

	//Update with another fqdn
	hrUpdate := integrationtest.FakeHostRule{
		Name:              hrname,
		Namespace:         "default",
		Fqdn:              "foo.com",
		SslKeyCertificate: "thisisaviref-sslkey",
	}.HostRule()
	enableVirtualHost := true
	hrUpdate.Spec.VirtualHost.Gslb.Fqdn = "baz.com"
	hrUpdate.Spec.VirtualHost.EnableVirtualHost = &enableVirtualHost
	hrUpdate.ResourceVersion = "2"
	_, err := CRDClient.AkoV1alpha1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := CRDClient.AkoV1alpha1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))
	g.Eventually(func() []string {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes) > 0 {
			return nodes[0].EvhNodes[0].VHDomainNames
		}
		return []string{}
	}, 10*time.Second).Should(gomega.ContainElement("baz.com"))
	g.Eventually(func() int {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes[0].HttpPolicyRefs)
	}, 10*time.Second).Should(gomega.Equal(2))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Host).To(gomega.ContainElement("baz.com"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Host).NotTo(gomega.ContainElement("bar.com"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].Hosts).To(gomega.ContainElement("baz.com"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].Hosts).NotTo(gomega.ContainElement("bar.com"))

	//Delete/Disable
	hrUpdate = integrationtest.FakeHostRule{
		Name:              hrname,
		Namespace:         "default",
		Fqdn:              "foo.com",
		SslKeyCertificate: "thisisaviref-sslkey",
	}.HostRule()
	enableVirtualHost = false
	hrUpdate.Spec.VirtualHost.EnableVirtualHost = &enableVirtualHost
	hrUpdate.ResourceVersion = "3"
	_, err = CRDClient.AkoV1alpha1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			return *nodes[0].EvhNodes[0].Enabled
		}
		return true
	}, 25*time.Second).Should(gomega.Equal(false))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].EvhNodes[0].Enabled).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].WafPolicyRef).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].AppProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].AnalyticsProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].ErrorPageProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].VsDatascriptRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].SSLProfileRef).To(gomega.Equal(""))
	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestCreateDeleteSharedVSHostRuleForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	hrname := "samplehr-foo"
	SetUpIngressForCacheSyncCheck(t, true, true, modelName)

	fqdn := "cluster--Shared-L7-EVH-0.admin.com"
	if lib.VIPPerNamespace() {
		fqdn = "cluster--Shared-L7-EVH-NS-default.admin.com"
	}
	hostrule := integrationtest.FakeHostRule{
		Name:               hrname,
		Namespace:          "default",
		Fqdn:               fqdn,
		WafPolicy:          "thisisaviref-waf",
		ApplicationProfile: "thisisaviref-appprof",
		AnalyticsProfile:   "thisisaviref-analyticsprof",
		ErrorPageProfile:   "thisisaviref-errorprof",
		Datascripts:        []string{"thisisaviref-ds2", "thisisaviref-ds1"},
		HttpPolicySets:     []string{"thisisaviref-httpps2", "thisisaviref-httpps1"},
	}
	hrCreate := hostrule.HostRule()
	hrCreate.Spec.VirtualHost.TCPSettings = &v1alpha1.HostRuleTCPSettings{
		Listeners: []v1alpha1.HostRuleTCPListeners{
			{Port: 8081}, {Port: 8082}, {Port: 8083, EnableSSL: true},
		},
	}
	if _, err := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().HostRules("default").Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}

	g.Eventually(func() string {
		hostrule, _ := CRDClient.AkoV1alpha1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	vsKey := cache.NamespaceName{Namespace: "admin", Name: strings.Split(modelName, "/")[1]}
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/samplehr-foo", true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(*nodes[0].Enabled).To(gomega.Equal(true))
	g.Expect(nodes[0].WafPolicyRef).To(gomega.ContainSubstring("thisisaviref-waf"))
	g.Expect(nodes[0].AppProfileRef).To(gomega.ContainSubstring("thisisaviref-appprof"))
	g.Expect(nodes[0].AnalyticsProfileRef).To(gomega.ContainSubstring("thisisaviref-analyticsprof"))
	g.Expect(nodes[0].ErrorPageProfileRef).To(gomega.ContainSubstring("thisisaviref-errorprof"))
	g.Expect(nodes[0].HttpPolicySetRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].HttpPolicySetRefs[0]).To(gomega.ContainSubstring("thisisaviref-httpps2"))
	g.Expect(nodes[0].HttpPolicySetRefs[1]).To(gomega.ContainSubstring("thisisaviref-httpps1"))
	g.Expect(nodes[0].VsDatascriptRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].VsDatascriptRefs[0]).To(gomega.ContainSubstring("thisisaviref-ds2"))
	g.Expect(nodes[0].VsDatascriptRefs[1]).To(gomega.ContainSubstring("thisisaviref-ds1"))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(3))
	var ports []int
	for _, port := range nodes[0].PortProto {
		ports = append(ports, int(port.Port))
		if port.EnableSSL {
			g.Expect(int(port.Port)).To(gomega.Equal(8083))
		}
	}
	sort.Ints(ports)
	g.Expect(ports[0]).To(gomega.Equal(8081))
	g.Expect(ports[1]).To(gomega.Equal(8082))
	g.Expect(ports[2]).To(gomega.Equal(8083))

	integrationtest.TeardownHostRule(t, g, vsKey, hrname)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].Enabled).To(gomega.BeNil())
	g.Expect(nodes[0].SSLKeyCertAviRef).To(gomega.HaveLen(0))
	g.Expect(nodes[0].WafPolicyRef).To(gomega.Equal(""))
	g.Expect(nodes[0].AppProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].AnalyticsProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].ErrorPageProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].VsDatascriptRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].SSLProfileRef).To(gomega.Equal(""))
	ports = []int{}
	for _, port := range nodes[0].PortProto {
		ports = append(ports, int(port.Port))
		if port.EnableSSL {
			g.Expect(int(port.Port)).To(gomega.Equal(443))
		}
	}
	sort.Ints(ports)
	g.Expect(ports[0]).To(gomega.Equal(80))
	g.Expect(ports[1]).To(gomega.Equal(443))

	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestCreateHostRuleBeforeIngressForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	hrname := "samplehr-foo"
	integrationtest.SetupHostRule(t, hrname, "foo.com", true)

	g.Eventually(func() string {
		hostrule, _ := CRDClient.AkoV1alpha1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	SetUpIngressForCacheSyncCheck(t, true, true, modelName)

	g.Eventually(func() string {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes) == 1 && len(nodes[0].EvhNodes[0].SSLKeyCertAviRef) == 1 {
			return nodes[0].EvhNodes[0].SSLKeyCertAviRef[0]
		}
		return ""
	}, 10*time.Second).Should(gomega.ContainSubstring("thisisaviref-sslkey"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)

	g.Eventually(func() string {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes) == 1 && len(nodes[0].EvhNodes[0].SSLKeyCertAviRef) == 1 {
			return nodes[0].EvhNodes[0].SSLKeyCertAviRef[0]
		}
		return ""
	}, 10*time.Second).Should(gomega.Equal(""))
	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestGoodToBadHostRuleForEvh(t *testing.T) {
	// create insecure ingress, apply good secure hostrule, transition to bad
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	hrname := "samplehr-foo"
	SetUpIngressForCacheSyncCheck(t, false, false, modelName)
	integrationtest.SetupHostRule(t, hrname, "foo.com", true)

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/samplehr-foo", true)

	// update hostrule with bad ref
	hrUpdate := integrationtest.FakeHostRule{
		Name:               hrname,
		Namespace:          "default",
		Fqdn:               "voo.com",
		WafPolicy:          "thisisBADaviref",
		ApplicationProfile: "thisisaviref-appprof",
	}.HostRule()
	hrUpdate.ResourceVersion = "2"
	if _, err := CRDClient.AkoV1alpha1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}

	g.Eventually(func() string {
		hostrule, _ := CRDClient.AkoV1alpha1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Rejected"))

	// the last applied hostrule values would exist. At rest layer, all ssl avi ref will be assigned to parent
	// In Avi model, sslaviref will be associated with child
	g.Eventually(func() string {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes[0].SSLKeyCertAviRef) == 1 {
			return nodes[0].EvhNodes[0].SSLKeyCertAviRef[0]
		}
		return ""
	}, 50*time.Second).Should(gomega.ContainSubstring("thisisaviref"))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].EvhNodes[0].WafPolicyRef).To(gomega.ContainSubstring("thisisaviref-waf"))
	g.Expect(nodes[0].EvhNodes[0].AppProfileRef).To(gomega.ContainSubstring("thisisaviref-appprof"))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestInsecureHostAndHostruleForEvh(t *testing.T) {
	// create insecure ingress, insecure hostrule, hostrule should be applied in case of EVH
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	hrname := "samplehr-foo"
	SetUpIngressForCacheSyncCheck(t, false, false, modelName)
	integrationtest.SetupHostRule(t, hrname, "foo.com", false)

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 10*time.Second).Should(gomega.Equal(1))
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/samplehr-foo", true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].EvhNodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Host).To(gomega.ContainElement("bar.com"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].VHDomainNames).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].VHDomainNames).To(gomega.ContainElement("bar.com"))

	//Update host rule with another fqdn
	//Update with another fqdn
	hrUpdate := integrationtest.FakeHostRule{
		Name:      hrname,
		Namespace: "default",
		Fqdn:      "foo.com",
	}.HostRule()
	enableVirtualHost := true
	hrUpdate.Spec.VirtualHost.Gslb.Fqdn = "baz.com"
	hrUpdate.Spec.VirtualHost.EnableVirtualHost = &enableVirtualHost
	hrUpdate.ResourceVersion = "2"
	_, err := CRDClient.AkoV1alpha1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := CRDClient.AkoV1alpha1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 10*time.Second).Should(gomega.Equal(1))
	//sleep added as update is not getting reflected on evh nodes immediately.
	time.Sleep(5 * time.Second)
	g.Expect(nodes[0].EvhNodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Host).To(gomega.ContainElement("baz.com"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].VHDomainNames).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].VHDomainNames).To(gomega.ContainElement("baz.com"))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestHostruleAnalyticsPolicyUpdateForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	hrname := "ap-hr-foo"
	SetUpIngressForCacheSyncCheck(t, false, false, modelName)
	integrationtest.SetupHostRule(t, hrname, "foo.com", false)
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 10*time.Second).Should(gomega.Equal(1))
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/ap-hr-foo", true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	// Check the default value of AnalyticsPolicy
	g.Expect(nodes[0].EvhNodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].AnalyticsPolicy).To(gomega.BeNil())

	// Update host rule with AnalyticsPolicy
	hrUpdate := integrationtest.FakeHostRule{
		Name:      hrname,
		Namespace: "default",
		Fqdn:      "foo.com",
	}.HostRule()
	enabled := true
	analyticsPolicy := &v1alpha1.HostRuleAnalyticsPolicy{
		FullClientLogs: &v1alpha1.FullClientLogs{
			Enabled:  &enabled,
			Throttle: "LOW",
		},
		LogAllHeaders: &enabled,
	}
	hrUpdate.Spec.VirtualHost.AnalyticsPolicy = analyticsPolicy
	hrUpdate.ResourceVersion = "2"
	_, err := CRDClient.AkoV1alpha1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := CRDClient.AkoV1alpha1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	g.Eventually(func() int {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 10*time.Second).Should(gomega.Equal(1))

	// update is not getting reflected on evh nodes immediately. Hence adding a sleep of 5 seconds.
	time.Sleep(5 * time.Second)

	// Check whether the AnalyticsPolicy values are properly updated
	g.Expect(nodes[0].EvhNodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].AnalyticsPolicy).ShouldNot(gomega.BeNil())
	g.Expect(*nodes[0].EvhNodes[0].AnalyticsPolicy.AllHeaders).To(gomega.BeTrue())
	g.Expect(*nodes[0].EvhNodes[0].AnalyticsPolicy.FullClientLogs.Enabled).To(gomega.BeTrue())
	g.Expect(*nodes[0].EvhNodes[0].AnalyticsPolicy.FullClientLogs.Throttle).To(gomega.Equal(*lib.GetThrottle("LOW")))

	// Remove the analytics Policy and check whether it is removed from VS.
	hrUpdate.Spec.VirtualHost.AnalyticsPolicy = nil
	hrUpdate.ResourceVersion = "3"
	_, err = CRDClient.AkoV1alpha1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := CRDClient.AkoV1alpha1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	g.Eventually(func() int {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 10*time.Second).Should(gomega.Equal(1))

	// update is not getting reflected on evh nodes immediately. Hence adding a sleep of 5 seconds.
	time.Sleep(5 * time.Second)

	// Check whether the AnalyticsPolicy values are properly removed
	g.Expect(nodes[0].EvhNodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].AnalyticsPolicy).Should(gomega.BeNil())

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestHostruleFQDNAliasesForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	hrname := "fqdn-aliases-hr-foo"
	SetUpIngressForCacheSyncCheck(t, true, true, modelName)
	integrationtest.SetupHostRule(t, hrname, "foo.com", false)
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 10*time.Second).Should(gomega.Equal(1))
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/fqdn-aliases-hr-foo", true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	// Common function that takes care of all validations
	validateNode := func(node *avinodes.AviEvhVsNode, aliases []string) {
		g.Expect(node.VSVIPRefs).To(gomega.HaveLen(1))
		g.Expect(node.VSVIPRefs[0].FQDNs).Should(gomega.ContainElements(aliases))

		g.Expect(node.EvhNodes).To(gomega.HaveLen(1))
		g.Expect(node.EvhNodes[0].VHDomainNames).Should(gomega.ContainElements(aliases))
		g.Expect(node.EvhNodes[0].HttpPolicyRefs).To(gomega.HaveLen(2))
		for _, httpPolicyRef := range node.EvhNodes[0].HttpPolicyRefs {
			if httpPolicyRef.HppMap != nil {
				g.Expect(httpPolicyRef.HppMap).To(gomega.HaveLen(1))
				g.Expect(httpPolicyRef.HppMap[0].Host).Should(gomega.ContainElements(aliases))
			}
			if httpPolicyRef.RedirectPorts != nil {
				g.Expect(httpPolicyRef.RedirectPorts).To(gomega.HaveLen(1))
				g.Expect(httpPolicyRef.RedirectPorts[0].Hosts).Should(gomega.ContainElements(aliases))
			}
			g.Expect(httpPolicyRef.AviMarkers.Host).Should(gomega.ContainElements(aliases))
		}
	}

	// Check default values.
	validateNode(nodes[0], []string{"foo.com"})

	// Update host rule with a valid FQDN Aliases
	hrUpdate := integrationtest.FakeHostRule{
		Name:      hrname,
		Namespace: "default",
		Fqdn:      "foo.com",
	}.HostRule()
	aliases := []string{"alias1.com", "alias2.com"}
	hrUpdate.Spec.VirtualHost.FqdnType = v1alpha1.Exact
	hrUpdate.Spec.VirtualHost.Aliases = aliases
	hrUpdate.ResourceVersion = "2"
	_, err := CRDClient.AkoV1alpha1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := CRDClient.AkoV1alpha1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	g.Eventually(func() int {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 10*time.Second).Should(gomega.Equal(1))

	// update is not getting reflected on evh nodes immediately. Hence adding a sleep of 5 seconds.
	time.Sleep(5 * time.Second)

	// Check whether the Aliases are properly added to Parent and Child VSes.
	validateNode(nodes[0], aliases)

	// Append one more alias and check whether it is getting added to parent and child VS.
	aliases = append(aliases, "alias3.com")
	hrUpdate.Spec.VirtualHost.Aliases = aliases
	hrUpdate.ResourceVersion = "3"
	_, err = CRDClient.AkoV1alpha1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := CRDClient.AkoV1alpha1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	g.Eventually(func() int {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 10*time.Second).Should(gomega.Equal(1))

	// update is not getting reflected on evh nodes immediately. Hence adding a sleep of 5 seconds.
	time.Sleep(5 * time.Second)

	// Check whether the Aliases are properly added to Parent and Child VSes.
	validateNode(nodes[0], aliases)

	// Remove one alias from hostrule and check whether its reference is removed properly.
	aliases = aliases[1:]
	hrUpdate.Spec.VirtualHost.Aliases = aliases
	hrUpdate.ResourceVersion = "4"
	_, err = CRDClient.AkoV1alpha1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := CRDClient.AkoV1alpha1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	g.Eventually(func() int {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 10*time.Second).Should(gomega.Equal(1))

	// update is not getting reflected on evh nodes immediately. Hence adding a sleep of 5 seconds.
	time.Sleep(5 * time.Second)

	// Check whether the Alias reference is properly removed from Parent and Child VSes.
	validateNode(nodes[0], aliases)

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestHostruleFQDNAliasesForMultiPathIngressEvh(t *testing.T) {

	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	hrname := "fqdn-aliases-hr-multipath-foo"

	SetupDomain()
	SetUpTestForIngress(t, modelName)
	integrationtest.AddSecret("my-secret", "default", "tlsCert", "tlsKey")
	integrationtest.PollForCompletion(t, modelName, 5)
	ingressObject := integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"10.0.0.1"},
		HostNames:   []string{"v1"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			"my-secret": {"foo.com"},
		},
	}

	ingrFake := ingressObject.Ingress(true)
	if _, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)

	// Create the hostrule with a valid FQDN Aliases
	hrUpdate := integrationtest.FakeHostRule{
		Name:      hrname,
		Namespace: "default",
		Fqdn:      "foo.com",
	}.HostRule()
	aliases := []string{"alias1.foo.com", "alias2.foo.com"}
	hrUpdate.Spec.VirtualHost.FqdnType = v1alpha1.Exact
	hrUpdate.Spec.VirtualHost.Aliases = aliases
	hrUpdate.ResourceVersion = "2"
	_, err := CRDClient.AkoV1alpha1().HostRules("default").Create(context.TODO(), hrUpdate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := CRDClient.AkoV1alpha1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 30*time.Second).Should(gomega.Equal(1))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].VHDomainNames).Should(gomega.ContainElements(aliases))
	g.Expect(nodes[0].EvhNodes[0].AviMarkers).ShouldNot(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].AviMarkers.Host).To(gomega.HaveLen(len(aliases) + 1)) // aliases + host
	g.Expect(nodes[0].EvhNodes[0].AviMarkers.Host).Should(gomega.ContainElements(aliases))
	for _, httpPolicyRef := range nodes[0].EvhNodes[0].HttpPolicyRefs {
		if httpPolicyRef.HppMap != nil {
			g.Expect(httpPolicyRef.HppMap).To(gomega.HaveLen(2))
			g.Expect(httpPolicyRef.HppMap[0].Host).Should(gomega.ContainElements(aliases))
			g.Expect(httpPolicyRef.HppMap[1].Host).Should(gomega.ContainElements(aliases))
		}
		if httpPolicyRef.RedirectPorts != nil {
			g.Expect(httpPolicyRef.RedirectPorts).To(gomega.HaveLen(1))
			g.Expect(httpPolicyRef.RedirectPorts[0].Hosts).Should(gomega.ContainElements(aliases))
		}
		g.Expect(httpPolicyRef.AviMarkers.Host).To(gomega.HaveLen(len(aliases) + 1)) // aliases + host
		g.Expect(httpPolicyRef.AviMarkers.Host).Should(gomega.ContainElements(aliases))
	}

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheck(t, modelName)
}

// HttpRule tests

func TestHTTPRuleCreateDeleteForEvh(t *testing.T) {
	// ingress secure foo.com/foo /bar
	// create httprule /foo, nothing happens
	// create hostrule, httprule gets attached check on /foo /bar
	// delete hostrule, httprule gets detached
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	rrname := "samplerr-foo"

	SetupDomain()
	SetUpTestForIngress(t, modelName)
	integrationtest.AddSecret("my-secret", "default", "tlsCert", "tlsKey")
	integrationtest.PollForCompletion(t, modelName, 5)
	ingressObject := integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			"my-secret": {"foo.com"},
		},
	}

	ingrFake := ingressObject.Ingress(true)
	if _, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)

	poolFooKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--default-foo.com_foo-foo-with-targets-avisvc", lib.Pool)}
	poolBarKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--default-foo.com_bar-foo-with-targets-avisvc", lib.Pool)}
	httpRulePath := "/"
	integrationtest.SetupHTTPRule(t, rrname, "foo.com", httpRulePath)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolFooKey, "default/"+rrname+"/"+httpRulePath, true)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolBarKey, "default/"+rrname+"/"+httpRulePath, true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].LbAlgorithm).To(gomega.Equal("LB_ALGORITHM_CONSISTENT_HASH"))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].LbAlgorithmHash).To(gomega.Equal("LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS"))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].SslProfileRef).To(gomega.ContainSubstring("thisisaviref-sslprofile"))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].PkiProfile.CACert).To(gomega.Equal("httprule-destinationCA"))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].HealthMonitors).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].HealthMonitors[0]).To(gomega.ContainSubstring("thisisaviref-hm2"))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].HealthMonitors[1]).To(gomega.ContainSubstring("thisisaviref-hm1"))

	// delete httprule deletes refs as well
	integrationtest.TeardownHTTPRule(t, rrname)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolFooKey, "default/"+rrname+"/"+httpRulePath, false)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolBarKey, "default/"+rrname+"/"+httpRulePath, false)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].LbAlgorithm).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].SslProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].PkiProfile).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].HealthMonitors).To(gomega.HaveLen(0))

	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestHTTPRuleCreateDeleteWithPkiRefForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	rrname := "samplerr-foo"

	SetupDomain()
	SetUpTestForIngress(t, modelName)
	integrationtest.AddSecret("my-secret", "default", "tlsCert", "tlsKey")
	integrationtest.PollForCompletion(t, modelName, 5)
	ingressObject := integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			"my-secret": {"foo.com"},
		},
	}

	ingrFake := ingressObject.Ingress(true)
	if _, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)

	poolFooKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--default-foo.com_foo-foo-with-targets-avisvc", lib.Pool)}
	poolBarKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--default-foo.com_bar-foo-with-targets-avisvc", lib.Pool)}
	httpRulePath := "/"
	httprule := integrationtest.FakeHTTPRule{
		Name:      rrname,
		Namespace: "default",
		Fqdn:      "foo.com",
		PathProperties: []integrationtest.FakeHTTPRulePath{{
			Path:        httpRulePath,
			PkiProfile:  "thisisaviref-pkiprofile",
			LbAlgorithm: "LB_ALGORITHM_CONSISTENT_HASH",
		}},
	}

	rrCreate := httprule.HTTPRule()
	if _, err := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().HTTPRules("default").Create(context.TODO(), rrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HTTPRule: %v", err)
	}

	integrationtest.VerifyMetadataHTTPRule(t, g, poolFooKey, "default/"+rrname+"/"+httpRulePath, true)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolBarKey, "default/"+rrname+"/"+httpRulePath, true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].LbAlgorithm).To(gomega.Equal("LB_ALGORITHM_CONSISTENT_HASH"))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].PkiProfileRef).To(gomega.ContainSubstring("thisisaviref-pkiprofile"))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].PkiProfile).To(gomega.BeNil())

	// delete httprule deletes refs as well
	integrationtest.TeardownHTTPRule(t, rrname)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolFooKey, "default/"+rrname+"/"+httpRulePath, false)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolBarKey, "default/"+rrname+"/"+httpRulePath, false)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].LbAlgorithm).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].PkiProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].PkiProfile).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheck(t, modelName)
}
