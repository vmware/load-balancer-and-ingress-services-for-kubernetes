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
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha2"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1beta1"
	crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha1/clientset/versioned/fake"
	v1alpha2crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha2/clientset/versioned/fake"
	v1beta1crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1beta1/clientset/versioned/fake"

	utils "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

var (
	KubeClient           *k8sfake.Clientset
	CRDClient            *crdfake.Clientset
	v1alpha2CRDClient    *v1alpha2crdfake.Clientset
	v1beta1CRDClient     *v1beta1crdfake.Clientset
	ctrl                 *k8s.AviController
	akoApiServer         *api.FakeApiServer
	keyChan              chan string
	endpointSliceEnabled bool
	objNameMap           integrationtest.ObjectNameMap
)

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
	os.Setenv("POD_NAME", "ako-0")

	akoControlConfig := lib.AKOControlConfig()
	endpointSliceEnabled = lib.GetEndpointSliceEnabled()
	akoControlConfig.SetEndpointSlicesEnabled(endpointSliceEnabled)

	KubeClient = k8sfake.NewSimpleClientset()
	CRDClient = crdfake.NewSimpleClientset()
	v1alpha2CRDClient = v1alpha2crdfake.NewSimpleClientset()
	v1beta1CRDClient = v1beta1crdfake.NewSimpleClientset()
	akoControlConfig.SetCRDClientset(CRDClient)
	akoControlConfig.Setv1alpha2CRDClientset(v1alpha2CRDClient)
	akoControlConfig.Setv1beta1CRDClientset(v1beta1CRDClient)
	akoControlConfig.SetAKOInstanceFlag(true)
	akoControlConfig.SetEventRecorder(lib.AKOEventComponent, KubeClient, true)
	akoControlConfig.SetDefaultLBController(true)

	data := map[string][]byte{
		"username": []byte("admin"),
		"password": []byte("admin"),
	}
	object := metav1.ObjectMeta{Name: "avi-secret", Namespace: utils.GetAKONamespace()}
	secret := &corev1.Secret{Data: data, ObjectMeta: object}
	KubeClient.CoreV1().Secrets(utils.GetAKONamespace()).Create(context.TODO(), secret, metav1.CreateOptions{})

	registeredInformers := []string{
		utils.ServiceInformer,
		utils.IngressInformer,
		utils.IngressClassInformer,
		utils.SecretInformer,
		utils.NSInformer,
		utils.NodeInformer,
		utils.ConfigMapInformer,
	}
	if akoControlConfig.GetEndpointSlicesEnabled() {
		registeredInformers = append(registeredInformers, utils.EndpointSlicesInformer)
	} else {
		registeredInformers = append(registeredInformers, utils.EndpointInformer)
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
	integrationtest.AddDefaultNamespace()
	integrationtest.AddDefaultNamespace("red")
	go ctrl.InitController(informers, registeredInformers, ctrlCh, stopCh, quickSyncCh, waitGroupMap)
	objNameMap.InitMap()
	os.Exit(m.Run())
}

type IngressTestObject struct {
	ingressName string
	namespace   string
	dnsNames    []string
	ipAddrs     []string
	hostnames   []string
	paths       []string
	isTLS       bool
	withSecret  bool
	secretName  string
	serviceName string
	modelNames  []string
}

func (ing *IngressTestObject) FillParams() {
	if ing.namespace == "" {
		ing.namespace = "default"
	}
	if len(ing.dnsNames) == 0 {
		ing.dnsNames = append(ing.dnsNames, "foo.com")
	}
	if len(ing.ipAddrs) == 0 {
		ing.ipAddrs = append(ing.ipAddrs, "8.8.8.8")
	}
	if len(ing.hostnames) == 0 {
		ing.hostnames = append(ing.hostnames, "v1")
	}
	if len(ing.paths) == 0 {
		ing.paths = append(ing.paths, "/foo")
	}
}

func SetupDomain() {
	mcache := cache.SharedAviObjCache()
	cloudObj := &cache.AviCloudPropertyCache{Name: "Default-Cloud", VType: "mock"}
	subdomains := []string{"avi.internal", ".com"}
	cloudObj.NSIpamDNS = subdomains
	mcache.CloudKeyCache.AviCacheAdd("Default-Cloud", cloudObj)
}

func SetUpTestForIngress(t *testing.T, svcName string, modelNames ...string) {
	for _, model := range modelNames {
		objects.SharedAviGraphLister().Delete(model)
	}
	integrationtest.CreateSVC(t, "default", svcName, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, "default", svcName, false, false, "1.1.1")
}

func TearDownTestForIngress(t *testing.T, svcName string, modelNames ...string) {
	//for _, model := range modelNames {
	//	objects.SharedAviGraphLister().Delete(model)
	//}
	integrationtest.DelSVC(t, "default", svcName)
	integrationtest.DelEPorEPS(t, "default", svcName)
}

func SetUpIngressForCacheSyncCheck(t *testing.T, ingTestObj IngressTestObject) {
	SetupDomain()
	SetUpTestForIngress(t, ingTestObj.serviceName, ingTestObj.modelNames...)
	ingressObject := integrationtest.FakeIngress{
		Name:        ingTestObj.ingressName,
		Namespace:   ingTestObj.namespace,
		DnsNames:    ingTestObj.dnsNames,
		Ips:         ingTestObj.ipAddrs,
		HostNames:   ingTestObj.hostnames,
		Paths:       ingTestObj.paths,
		ServiceName: ingTestObj.serviceName,
	}
	if ingTestObj.withSecret {
		integrationtest.AddSecret(ingTestObj.secretName, ingTestObj.namespace, "tlsCert", "tlsKey")
	}
	if ingTestObj.isTLS {
		ingressObject.TlsSecretDNS = map[string][]string{
			ingTestObj.secretName: {"foo.com"},
		}
	}
	ingrFake := ingressObject.Ingress()
	if _, err := KubeClient.NetworkingV1().Ingresses(ingTestObj.namespace).Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, ingTestObj.modelNames[0], 5)
}

func TearDownIngressForCacheSyncCheck(t *testing.T, secretName, ingressName, svcName, modelName string) {
	if err := KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	if secretName != "" {
		KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), secretName, metav1.DeleteOptions{})
	}
	TearDownTestForIngress(t, svcName, modelName)
}

func TestCreateUpdateDeleteHostRuleForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	hrName := objNameMap.GenerateName("samplehr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	integrationtest.SetupHostRule(t, hrName, "foo.com", true)

	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/"+hrName, true)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(*nodes[0].EvhNodes[0].Enabled).To(gomega.Equal(true))
	//At rest layer, sslref are switched to Parent
	g.Expect(nodes[0].EvhNodes[0].SslKeyAndCertificateRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].SslKeyAndCertificateRefs[0]).To(gomega.ContainSubstring("thisisaviref-sslkey"))
	g.Expect(nodes[0].EvhNodes[0].ICAPProfileRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].ICAPProfileRefs[0]).To(gomega.ContainSubstring("thisisaviref-icapprof"))
	g.Expect(*nodes[0].EvhNodes[0].WafPolicyRef).To(gomega.ContainSubstring("thisisaviref-waf"))
	g.Expect(*nodes[0].EvhNodes[0].ApplicationProfileRef).To(gomega.ContainSubstring("thisisaviref-appprof"))
	g.Expect(*nodes[0].EvhNodes[0].AnalyticsProfileRef).To(gomega.ContainSubstring("thisisaviref-analyticsprof"))
	g.Expect(nodes[0].EvhNodes[0].ErrorPageProfileRef).To(gomega.ContainSubstring("thisisaviref-errorprof"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicySetRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicySetRefs[0]).To(gomega.ContainSubstring("thisisaviref-httpps2"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicySetRefs[1]).To(gomega.ContainSubstring("thisisaviref-httpps1"))
	g.Expect(nodes[0].EvhNodes[0].VsDatascriptRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].VsDatascriptRefs[0]).To(gomega.ContainSubstring("thisisaviref-ds2"))
	g.Expect(nodes[0].EvhNodes[0].VsDatascriptRefs[1]).To(gomega.ContainSubstring("thisisaviref-ds1"))
	g.Expect(*nodes[0].SslProfileRef).To(gomega.ContainSubstring("thisisaviref-sslprof"))
	g.Expect(nodes[0].EvhNodes[0].VHDomainNames).To(gomega.ContainElement("bar.com"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].Hosts).To(gomega.ContainElement("bar.com"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Host).To(gomega.ContainElement("bar.com"))
	// Hostrule with normal FQDN should not have reference to network security policy
	g.Expect(nodes[0].NetworkSecurityPolicyRef).Should(gomega.BeNil())

	//Update with another fqdn
	hrUpdate := integrationtest.FakeHostRule{
		Name:              hrName,
		Namespace:         "default",
		Fqdn:              "foo.com",
		SslKeyCertificate: "thisisaviref-sslkey",
	}.HostRule()
	enableVirtualHost := true
	hrUpdate.Spec.VirtualHost.Gslb.Fqdn = "baz.com"
	hrUpdate.Spec.VirtualHost.EnableVirtualHost = &enableVirtualHost
	hrUpdate.ResourceVersion = "2"
	_, err := v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
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
		Name:              hrName,
		Namespace:         "default",
		Fqdn:              "foo.com",
		SslKeyCertificate: "thisisaviref-sslkey",
	}.HostRule()
	enableVirtualHost = false
	hrUpdate.Spec.VirtualHost.EnableVirtualHost = &enableVirtualHost
	hrUpdate.ResourceVersion = "3"
	_, err = v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
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

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrName)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].EvhNodes[0].Enabled).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].ICAPProfileRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].WafPolicyRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].ApplicationProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].AnalyticsProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].ErrorPageProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].VsDatascriptRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].SslProfileRef).To(gomega.BeNil())
	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestCreateDeleteSharedVSHostRuleForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	hrName := objNameMap.GenerateName("samplehr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	fqdn := "cluster--Shared-L7-EVH-0.admin.com"
	if lib.VIPPerNamespace() {
		fqdn = "Shared-L7-EVH-NS"
	}
	hostrule := integrationtest.FakeHostRule{
		Name:                  hrName,
		Namespace:             "default",
		Fqdn:                  fqdn,
		WafPolicy:             "thisisaviref-waf",
		ApplicationProfile:    "thisisaviref-appprof",
		ICAPProfile:           []string{"thisisaviref-icapprof"},
		AnalyticsProfile:      "thisisaviref-analyticsprof",
		ErrorPageProfile:      "thisisaviref-errorprof",
		Datascripts:           []string{"thisisaviref-ds2", "thisisaviref-ds1"},
		HttpPolicySets:        []string{"thisisaviref-httpps2", "thisisaviref-httpps1"},
		NetworkSecurityPolicy: "thisisaviref-networksecuritypolicyref",
	}
	hrCreate := hostrule.HostRule()
	hrCreate.Spec.VirtualHost.TCPSettings = &v1beta1.HostRuleTCPSettings{
		Listeners: []v1beta1.HostRuleTCPListeners{
			{Port: 8081}, {Port: 8082}, {Port: 8083, EnableSSL: true},
		},
	}
	if lib.VIPPerNamespace() {
		hrCreate.Spec.VirtualHost.FqdnType = "Contains"
	}
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules("default").Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}

	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	vsKey := cache.NamespaceName{Namespace: "admin", Name: strings.Split(modelName, "/")[1]}
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/"+hrName, true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(*nodes[0].Enabled).To(gomega.Equal(true))
	g.Expect(nodes[0].ICAPProfileRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].ICAPProfileRefs[0]).To(gomega.ContainSubstring("thisisaviref-icapprof"))
	g.Expect(*nodes[0].WafPolicyRef).To(gomega.ContainSubstring("thisisaviref-waf"))
	g.Expect(*nodes[0].ApplicationProfileRef).To(gomega.ContainSubstring("thisisaviref-appprof"))
	g.Expect(*nodes[0].AnalyticsProfileRef).To(gomega.ContainSubstring("thisisaviref-analyticsprof"))
	g.Expect(nodes[0].ErrorPageProfileRef).To(gomega.ContainSubstring("thisisaviref-errorprof"))
	g.Expect(nodes[0].HttpPolicySetRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].HttpPolicySetRefs[0]).To(gomega.ContainSubstring("thisisaviref-httpps2"))
	g.Expect(nodes[0].HttpPolicySetRefs[1]).To(gomega.ContainSubstring("thisisaviref-httpps1"))
	g.Expect(nodes[0].VsDatascriptRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].VsDatascriptRefs[0]).To(gomega.ContainSubstring("thisisaviref-ds2"))
	g.Expect(nodes[0].VsDatascriptRefs[1]).To(gomega.ContainSubstring("thisisaviref-ds1"))
	g.Expect(*nodes[0].NetworkSecurityPolicyRef).To(gomega.ContainSubstring("thisisaviref-networksecuritypolicyref"))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(3))
	var ports []int
	for _, port := range nodes[0].PortProto {
		ports = append(ports, int(port.Port))
		if port.EnableSSL {
			g.Expect(int(port.Port)).Should(gomega.Equal(8083))
		}
	}
	sort.Ints(ports)
	g.Expect(ports[0]).To(gomega.Equal(8081))
	g.Expect(ports[1]).To(gomega.Equal(8082))
	g.Expect(ports[2]).To(gomega.Equal(8083))

	integrationtest.TeardownHostRule(t, g, vsKey, hrName)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].Enabled).To(gomega.BeNil())
	g.Expect(nodes[0].SslKeyAndCertificateRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].ICAPProfileRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].WafPolicyRef).To(gomega.BeNil())
	g.Expect(nodes[0].ApplicationProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].AnalyticsProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].ErrorPageProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].VsDatascriptRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].SslProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].NetworkSecurityPolicyRef).To(gomega.BeNil())
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

	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestCreateHostRuleBeforeIngressForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	hrName := objNameMap.GenerateName("samplehr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	integrationtest.SetupHostRule(t, hrName, "foo.com", true)

	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	g.Eventually(func() string {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes) == 1 && len(nodes[0].EvhNodes[0].SslKeyAndCertificateRefs) == 1 {
			return nodes[0].EvhNodes[0].SslKeyAndCertificateRefs[0]
		}
		return ""
	}, 10*time.Second).Should(gomega.ContainSubstring("thisisaviref-sslkey"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	integrationtest.TeardownHostRule(t, g, sniVSKey, hrName)

	g.Eventually(func() string {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes) == 1 && len(nodes[0].EvhNodes[0].SslKeyAndCertificateRefs) == 1 {
			return nodes[0].EvhNodes[0].SslKeyAndCertificateRefs[0]
		}
		return ""
	}, 10*time.Second).Should(gomega.Equal(""))
	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestGoodToBadHostRuleForEvh(t *testing.T) {
	// create insecure ingress, apply good secure hostrule, transition to bad
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	hrName := objNameMap.GenerateName("samplehr-foo")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       false,
		withSecret:  false,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)
	integrationtest.SetupHostRule(t, hrName, "foo.com", true)

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/"+hrName, true)

	// update hostrule with bad ref
	hrUpdate := integrationtest.FakeHostRule{
		Name:               hrName,
		Namespace:          "default",
		Fqdn:               "voo.com",
		WafPolicy:          "thisisBADaviref",
		ApplicationProfile: "thisisaviref-appprof",
	}.HostRule()
	hrUpdate.ResourceVersion = "2"
	if _, err := v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}

	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Rejected"))

	// the last applied hostrule values would exist. At rest layer, all ssl avi ref will be assigned to parent
	// In Avi model, sslaviref will be associated with child
	g.Eventually(func() string {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes[0].SslKeyAndCertificateRefs) == 1 {
			return nodes[0].EvhNodes[0].SslKeyAndCertificateRefs[0]
		}
		return ""
	}, 50*time.Second).Should(gomega.ContainSubstring("thisisaviref"))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(*nodes[0].EvhNodes[0].WafPolicyRef).To(gomega.ContainSubstring("thisisaviref-waf"))
	g.Expect(*nodes[0].EvhNodes[0].ApplicationProfileRef).To(gomega.ContainSubstring("thisisaviref-appprof"))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrName)
	TearDownIngressForCacheSyncCheck(t, "", ingressName, svcName, modelName)
}

func TestInsecureHostAndHostruleForEvh(t *testing.T) {
	// create insecure ingress, insecure hostrule, hostrule should be applied in case of EVH
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	hrName := objNameMap.GenerateName("samplehr-foo")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       false,
		withSecret:  false,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)
	integrationtest.SetupHostRule(t, hrName, "foo.com", false)

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 10*time.Second).Should(gomega.Equal(1))
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/"+hrName, true)
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
		Name:      hrName,
		Namespace: "default",
		Fqdn:      "foo.com",
	}.HostRule()
	enableVirtualHost := true
	hrUpdate.Spec.VirtualHost.Gslb.Fqdn = "baz.com"
	hrUpdate.Spec.VirtualHost.EnableVirtualHost = &enableVirtualHost
	hrUpdate.ResourceVersion = "2"
	_, err := v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
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

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrName)
	TearDownIngressForCacheSyncCheck(t, "", ingressName, svcName, modelName)
}

func TestHostruleAnalyticsPolicyUpdateForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	hrName := objNameMap.GenerateName("samplehr-foo")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       false,
		withSecret:  false,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)
	integrationtest.SetupHostRule(t, hrName, "foo.com", false)
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

	// Update host rule with AnalyticsPolicy - only with LogallHeaders
	hrUpdate := integrationtest.FakeHostRule{
		Name:      hrName,
		Namespace: "default",
		Fqdn:      "foo.com",
	}.HostRule()
	enabled := true
	analyticsPolicy := &v1beta1.HostRuleAnalyticsPolicy{
		LogAllHeaders: &enabled,
	}
	hrUpdate.Spec.VirtualHost.AnalyticsPolicy = analyticsPolicy
	hrUpdate.ResourceVersion = "2"
	_, err := v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	g.Eventually(func() bool {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes) == 1 && nodes[0].EvhNodes[0].AnalyticsPolicy != nil {
			return *nodes[0].EvhNodes[0].AnalyticsPolicy.AllHeaders
		}
		return false
	}, 10*time.Second).Should(gomega.Equal(true))

	// Update host rule with AnalyticsPolicy - with LogAllHeader, FullClientLogs.Enabled
	hrUpdate = integrationtest.FakeHostRule{
		Name:      hrName,
		Namespace: "default",
		Fqdn:      "foo.com",
	}.HostRule()

	analyticsPolicy = &v1beta1.HostRuleAnalyticsPolicy{
		FullClientLogs: &v1beta1.FullClientLogs{
			Enabled: &enabled,
		},
		LogAllHeaders: &enabled,
	}
	hrUpdate.Spec.VirtualHost.AnalyticsPolicy = analyticsPolicy
	hrUpdate.ResourceVersion = "3"
	_, err = v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	g.Eventually(func() bool {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes) == 1 && nodes[0].EvhNodes[0].AnalyticsPolicy != nil &&
			nodes[0].EvhNodes[0].AnalyticsPolicy.FullClientLogs != nil &&
			nodes[0].EvhNodes[0].AnalyticsPolicy.FullClientLogs.Enabled != nil {
			return *nodes[0].EvhNodes[0].AnalyticsPolicy.AllHeaders &&
				*nodes[0].EvhNodes[0].AnalyticsPolicy.FullClientLogs.Enabled
		}
		return false
	}, 10*time.Second).Should(gomega.Equal(true))

	// Update host rule with AnalyticsPolicy - All fields
	hrUpdate = integrationtest.FakeHostRule{
		Name:      hrName,
		Namespace: "default",
		Fqdn:      "foo.com",
	}.HostRule()

	analyticsPolicy = &v1beta1.HostRuleAnalyticsPolicy{
		FullClientLogs: &v1beta1.FullClientLogs{
			Enabled:  &enabled,
			Throttle: "LOW",
		},
		LogAllHeaders: &enabled,
	}
	hrUpdate.Spec.VirtualHost.AnalyticsPolicy = analyticsPolicy
	hrUpdate.ResourceVersion = "4"
	_, err = v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	g.Eventually(func() bool {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes) == 1 && nodes[0].EvhNodes[0].AnalyticsPolicy != nil &&
			nodes[0].EvhNodes[0].AnalyticsPolicy.FullClientLogs != nil &&
			nodes[0].EvhNodes[0].AnalyticsPolicy.FullClientLogs.Throttle != nil {
			return *nodes[0].EvhNodes[0].AnalyticsPolicy.FullClientLogs.Throttle == *lib.GetThrottle("LOW")
		}
		return false
	}, 10*time.Second).Should(gomega.Equal(true))

	g.Expect(*nodes[0].EvhNodes[0].AnalyticsPolicy.AllHeaders).To(gomega.BeTrue())
	g.Expect(*nodes[0].EvhNodes[0].AnalyticsPolicy.FullClientLogs.Enabled).To(gomega.BeTrue())

	// Remove the analytics Policy and check whether it is removed from VS.
	hrUpdate.Spec.VirtualHost.AnalyticsPolicy = nil
	hrUpdate.ResourceVersion = "5"
	_, err = v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	g.Eventually(func() bool {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes) == 1 {
			return nodes[0].EvhNodes[0].AnalyticsPolicy == nil
		}
		return false
	}, 10*time.Second).Should(gomega.Equal(true))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrName)
	TearDownIngressForCacheSyncCheck(t, "", ingressName, svcName, modelName)
}

func TestHostruleFQDNAliasesForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	hrName := objNameMap.GenerateName("samplehr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)
	integrationtest.SetupHostRule(t, hrName, "foo.com", false)
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
		Name:      hrName,
		Namespace: "default",
		Fqdn:      "foo.com",
	}.HostRule()
	aliases := []string{"alias1.com", "alias2.com"}
	hrUpdate.Spec.VirtualHost.FqdnType = v1beta1.Exact
	hrUpdate.Spec.VirtualHost.Aliases = aliases
	hrUpdate.ResourceVersion = "2"
	_, err := v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
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
	_, err = v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
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
	_, err = v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
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

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrName)
	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestHostruleFQDNAliasesForMultiPathIngressEvh(t *testing.T) {

	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	hrName := objNameMap.GenerateName("samplehr-foo")

	SetupDomain()
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, modelName)
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	integrationtest.PollForCompletion(t, modelName, 5)
	ingressObject := integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"10.0.0.1"},
		HostNames:   []string{"v1"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
	}

	ingrFake := ingressObject.Ingress(true)
	if _, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)

	// Create the hostrule with a valid FQDN Aliases
	hrUpdate := integrationtest.FakeHostRule{
		Name:      hrName,
		Namespace: "default",
		Fqdn:      "foo.com",
	}.HostRule()
	aliases := []string{"alias1.foo.com", "alias2.foo.com"}
	hrUpdate.Spec.VirtualHost.FqdnType = v1beta1.Exact
	hrUpdate.Spec.VirtualHost.Aliases = aliases
	hrUpdate.ResourceVersion = "2"
	_, err := v1beta1CRDClient.AkoV1beta1().HostRules("default").Create(context.TODO(), hrUpdate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 30*time.Second).Should(gomega.Equal(1))

	// Adding a sleep so that the HostRule is applied to the model
	time.Sleep(5 * time.Second)

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
	integrationtest.TeardownHostRule(t, g, sniVSKey, hrName)
	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestApplyHostruleToParentVS(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	hrName := objNameMap.GenerateName("samplehr-foo")

	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	hostrule := integrationtest.FakeHostRule{
		Name:               hrName,
		Namespace:          "default",
		WafPolicy:          "thisisaviref-waf",
		ApplicationProfile: "thisisaviref-appprof",
		AnalyticsProfile:   "thisisaviref-analyticsprof",
		ErrorPageProfile:   "thisisaviref-errorprof",
		Datascripts:        []string{"thisisaviref-ds2", "thisisaviref-ds1"},
		HttpPolicySets:     []string{"thisisaviref-httpps2", "thisisaviref-httpps1"},
	}
	hrObj := hostrule.HostRule()
	hrObj.Spec.VirtualHost.Fqdn = "Shared-L7-EVH-"
	hrObj.Spec.VirtualHost.FqdnType = v1beta1.Contains

	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules("default").Create(context.TODO(), hrObj, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}

	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))

	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-EVH-0"}
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/hr-cluster--Shared-L7-EVH-0", true)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.BeTrue())
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(*nodes[0].Enabled).To(gomega.Equal(true))
	g.Expect(*nodes[0].WafPolicyRef).To(gomega.ContainSubstring("thisisaviref-waf"))
	g.Expect(*nodes[0].ApplicationProfileRef).To(gomega.ContainSubstring("thisisaviref-appprof"))
	g.Expect(*nodes[0].AnalyticsProfileRef).To(gomega.ContainSubstring("thisisaviref-analyticsprof"))
	g.Expect(nodes[0].ErrorPageProfileRef).To(gomega.ContainSubstring("thisisaviref-errorprof"))
	g.Expect(nodes[0].HttpPolicySetRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].HttpPolicySetRefs[0]).To(gomega.ContainSubstring("thisisaviref-httpps2"))
	g.Expect(nodes[0].HttpPolicySetRefs[1]).To(gomega.ContainSubstring("thisisaviref-httpps1"))
	g.Expect(nodes[0].VsDatascriptRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].VsDatascriptRefs[0]).To(gomega.ContainSubstring("thisisaviref-ds2"))
	g.Expect(nodes[0].VsDatascriptRefs[1]).To(gomega.ContainSubstring("thisisaviref-ds1"))

	integrationtest.TeardownHostRule(t, g, vsKey, hrName)
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/hr-cluster--Shared-L7-EVH-0", false)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].Enabled).To(gomega.BeNil())
	g.Expect(nodes[0].SslKeyAndCertificateRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].WafPolicyRef).To(gomega.BeNil())
	g.Expect(nodes[0].ApplicationProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].AnalyticsProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].ErrorPageProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].VsDatascriptRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].SslProfileRef).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestHostRuleWithEmptyConfig(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	hrName := objNameMap.GenerateName("samplehr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	hostrule := integrationtest.FakeHostRule{
		Name:      hrName,
		Namespace: "default",
		Fqdn:      "foo.com",
	}
	hrObj := hostrule.HostRule()
	hrObj.ResourceVersion = "1"
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules("default").Create(context.TODO(), hrObj, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in creating HostRule: %v", err)
	}

	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/"+hrName, true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(*nodes[0].EvhNodes[0].Enabled).To(gomega.Equal(true))
	//At rest layer, sslref are switched to Parent
	g.Expect(nodes[0].EvhNodes[0].SslKeyAndCertificateRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].ICAPProfileRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].WafPolicyRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].ApplicationProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].AnalyticsProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].ErrorPageProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].VsDatascriptRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].SslProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].VHDomainNames).To(gomega.ContainElement("foo.com"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].Hosts).To(gomega.ContainElement("foo.com"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Host).To(gomega.ContainElement("foo.com"))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrName)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].EvhNodes[0].Enabled).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].ICAPProfileRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].WafPolicyRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].ApplicationProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].AnalyticsProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].ErrorPageProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].VsDatascriptRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].SslProfileRef).To(gomega.BeNil())
	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestSharedVSHostRuleNoListenerForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	hrName := objNameMap.GenerateName("samplehr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	fqdn := "cluster--Shared-L7-EVH-0.admin.com"
	if lib.VIPPerNamespace() {
		fqdn = "Shared-L7-EVH-NS"
	}
	hostrule := integrationtest.FakeHostRule{
		Name:               hrName,
		Namespace:          "default",
		Fqdn:               fqdn,
		WafPolicy:          "thisisaviref-waf",
		ApplicationProfile: "thisisaviref-appprof",
		ICAPProfile:        []string{"thisisaviref-icapprof"},
		AnalyticsProfile:   "thisisaviref-analyticsprof",
		ErrorPageProfile:   "thisisaviref-errorprof",
		Datascripts:        []string{"thisisaviref-ds2", "thisisaviref-ds1"},
		HttpPolicySets:     []string{"thisisaviref-httpps2", "thisisaviref-httpps1"},
	}
	hrCreate := hostrule.HostRule()
	hrCreate.Spec.VirtualHost.TCPSettings = &v1beta1.HostRuleTCPSettings{
		LoadBalancerIP: "80.80.80.80",
	}
	if lib.VIPPerNamespace() {
		hrCreate.Spec.VirtualHost.FqdnType = "Contains"
	}
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules("default").Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}

	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	vsKey := cache.NamespaceName{Namespace: "admin", Name: strings.Split(modelName, "/")[1]}
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/"+hrName, true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(*nodes[0].Enabled).To(gomega.Equal(true))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(2))
	var ports []int
	for _, port := range nodes[0].PortProto {
		ports = append(ports, int(port.Port))
		if port.EnableSSL {
			g.Expect(int(port.Port)).To(gomega.Equal(443))
		}
	}
	sort.Ints(ports)
	g.Expect(ports[0]).To(gomega.Equal(80))
	g.Expect(ports[1]).To(gomega.Equal(443))
	g.Expect(nodes[0].VSVIPRefs[0].IPAddress).To(gomega.Equal("80.80.80.80"))

	integrationtest.TeardownHostRule(t, g, vsKey, hrName)
	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

// HttpRule tests

func TestHTTPRuleCreateDeleteForEvh(t *testing.T) {
	// ingress secure foo.com/foo /bar
	// create httprule /foo, nothing happens
	// create hostrule, httprule gets attached check on /foo /bar
	// delete hostrule, httprule gets detached
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	rrName := objNameMap.GenerateName("samplerr-foo")

	SetupDomain()
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, modelName)
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	integrationtest.PollForCompletion(t, modelName, 5)
	ingressObject := integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
	}

	ingrFake := ingressObject.Ingress(true)
	if _, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)

	poolFooKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--default-foo.com_foo-"+ingressName+"-"+svcName, lib.Pool)}
	poolBarKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--default-foo.com_bar-"+ingressName+"-"+svcName, lib.Pool)}
	httpRulePath := "/"
	integrationtest.SetupHTTPRule(t, rrName, "foo.com", httpRulePath)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolFooKey, "default/"+rrName+"/"+httpRulePath, true)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolBarKey, "default/"+rrName+"/"+httpRulePath, true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(*nodes[0].EvhNodes[0].PoolRefs[0].LbAlgorithm).To(gomega.Equal("LB_ALGORITHM_CONSISTENT_HASH"))
	g.Expect(*nodes[0].EvhNodes[0].PoolRefs[0].LbAlgorithmHash).To(gomega.Equal("LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS"))
	g.Expect(*nodes[0].EvhNodes[0].PoolRefs[0].SslProfileRef).To(gomega.ContainSubstring("thisisaviref-sslprofile"))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].PkiProfile.CACert).To(gomega.Equal("httprule-destinationCA"))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].HealthMonitorRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm2"))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].HealthMonitorRefs[1]).To(gomega.ContainSubstring("thisisaviref-hm1"))

	// delete httprule deletes refs as well
	integrationtest.TeardownHTTPRule(t, rrName)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolFooKey, "default/"+rrName+"/"+httpRulePath, false)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolBarKey, "default/"+rrName+"/"+httpRulePath, false)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].LbAlgorithm).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].SslProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].PkiProfile).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].HealthMonitorRefs).To(gomega.HaveLen(0))

	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestHTTPRuleCreateDeleteWithPkiRefForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	rrName := objNameMap.GenerateName("samplerr-foo")

	SetupDomain()
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, modelName)
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	integrationtest.PollForCompletion(t, modelName, 5)
	ingressObject := integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
	}

	ingrFake := ingressObject.Ingress(true)
	if _, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)

	poolFooKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--default-foo.com_foo-"+ingressName+"-"+svcName, lib.Pool)}
	poolBarKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--default-foo.com_bar-"+ingressName+"-"+svcName, lib.Pool)}
	httpRulePath := "/"
	httprule := integrationtest.FakeHTTPRule{
		Name:      rrName,
		Namespace: "default",
		Fqdn:      "foo.com",
		PathProperties: []integrationtest.FakeHTTPRulePath{{
			Path:        httpRulePath,
			PkiProfile:  "thisisaviref-pkiprofile",
			LbAlgorithm: "LB_ALGORITHM_CONSISTENT_HASH",
		}},
	}

	rrCreate := httprule.HTTPRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HTTPRules("default").Create(context.TODO(), rrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HTTPRule: %v", err)
	}

	integrationtest.VerifyMetadataHTTPRule(t, g, poolFooKey, "default/"+rrName+"/"+httpRulePath, true)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolBarKey, "default/"+rrName+"/"+httpRulePath, true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(*nodes[0].EvhNodes[0].PoolRefs[0].LbAlgorithm).To(gomega.Equal("LB_ALGORITHM_CONSISTENT_HASH"))
	g.Expect(*nodes[0].EvhNodes[0].PoolRefs[0].PkiProfileRef).To(gomega.ContainSubstring("thisisaviref-pkiprofile"))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].PkiProfile).To(gomega.BeNil())

	// delete httprule deletes refs as well
	integrationtest.TeardownHTTPRule(t, rrName)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolFooKey, "default/"+rrName+"/"+httpRulePath, false)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolBarKey, "default/"+rrName+"/"+httpRulePath, false)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].LbAlgorithm).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].PkiProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].PkiProfile).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestHTTPRuleWithInvalidPath(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	rrName := objNameMap.GenerateName("samplerr-foo")

	SetupDomain()
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, modelName)
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	integrationtest.PollForCompletion(t, modelName, 5)
	ingressObject := integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
	}

	ingrFake := ingressObject.Ingress(true)
	if _, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)

	// create a httprule with a non-existing path
	integrationtest.SetupHTTPRule(t, rrName, "foo.com", "/invalidPath")

	time.Sleep(10 * time.Second)

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].EvhNodes[0].PoolRefs).To(gomega.HaveLen(2))

	// pool corresponding to the path "foo"
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].LbAlgorithm).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].LbAlgorithmHash).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].SslProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].PkiProfile).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].HealthMonitorRefs).To(gomega.HaveLen(0))

	// pool corresponding to the path "bar"
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[1].LbAlgorithm).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[1].LbAlgorithmHash).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[1].SslProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[1].PkiProfile).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[1].HealthMonitorRefs).To(gomega.HaveLen(0))

	// delete httprule must not change any configs
	integrationtest.TeardownHTTPRule(t, rrName)

	time.Sleep(10 * time.Second)

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].EvhNodes[0].PoolRefs).To(gomega.HaveLen(2))

	// pool corresponding to the path "foo"
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].LbAlgorithm).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].SslProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].PkiProfile).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].HealthMonitorRefs).To(gomega.HaveLen(0))

	// pool corresponding to the path "bar"
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[1].LbAlgorithm).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[1].SslProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[1].PkiProfile).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[1].HealthMonitorRefs).To(gomega.HaveLen(0))

	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestHTTPRuleCreateDeleteWithEnableHTTP2ForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	rrName := objNameMap.GenerateName("samplerr-foo")

	SetupDomain()
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, modelName)
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	integrationtest.PollForCompletion(t, modelName, 5)
	ingressObject := integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
	}

	ingrFake := ingressObject.Ingress(true)
	if _, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)

	poolFooKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--default-foo.com_foo-"+ingressName+"-"+svcName, lib.Pool)}
	poolBarKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--default-foo.com_bar-"+ingressName+"-"+svcName, lib.Pool)}
	httpRulePath := "/"
	httprule := integrationtest.FakeHTTPRule{
		Name:      rrName,
		Namespace: "default",
		Fqdn:      "foo.com",
		PathProperties: []integrationtest.FakeHTTPRulePath{{
			Path:        httpRulePath,
			EnableHTTP2: true,
		}},
	}

	rrCreate := httprule.HTTPRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HTTPRules("default").Create(context.TODO(), rrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HTTPRule: %v", err)
	}

	integrationtest.VerifyMetadataHTTPRule(t, g, poolFooKey, "default/"+rrName+"/"+httpRulePath, true)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolBarKey, "default/"+rrName+"/"+httpRulePath, true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(*nodes[0].EvhNodes[0].PoolRefs[0].EnableHttp2).To(gomega.Equal(true))

	// delete httprule disables HTTP2
	integrationtest.TeardownHTTPRule(t, rrName)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolFooKey, "default/"+rrName+"/"+httpRulePath, false)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolBarKey, "default/"+rrName+"/"+httpRulePath, false)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].EnableHttp2).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestCreateUpdateDeleteSSORuleForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	srName := objNameMap.GenerateName("samplesr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	err := integrationtest.SetUpOAuthSecret()
	if err != nil {
		t.Fatalf("error in creating my-oauth-secret: %v", err)
	}
	// Sleeping for 5s for secret to be updated in informer
	time.Sleep(5 * time.Second)

	integrationtest.SetupSSORule(t, srName, "foo.com", "OAuth")

	g.Eventually(func() string {
		ssoRule, _ := v1alpha2CRDClient.AkoV1alpha2().SSORules("default").Get(context.TODO(), srName, metav1.GetOptions{})
		return ssoRule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	integrationtest.VerifyMetadataSSORule(t, g, sniVSKey, "default/samplesr-foo", true)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	g.Expect(*nodes[0].EvhNodes[0].SsoPolicyRef).To(gomega.ContainSubstring("thisisaviref-ssopolicyoauth"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.CookieName).To(gomega.Equal("MY_OAUTH_COOKIE"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.CookieTimeout).To(gomega.Equal(int32(120)))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.LogoutURI).To(gomega.Equal("https://auth.com/oauth/logout"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.RedirectURI).To(gomega.Equal("https://auth.com/oauth/redirect"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.PostLogoutRedirectURI).To(gomega.Equal("https://auth.com/oauth/post-logout-redirect"))
	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings).To(gomega.HaveLen(1))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.ClientID).To(gomega.Equal("my-client-id"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.ClientSecret).To(gomega.Equal("my-client-secret"))
	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.Scopes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.Scopes[0]).To(gomega.Equal("scope-1"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.OidcConfig.OidcEnable).To(gomega.Equal(true))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.OidcConfig.Profile).To(gomega.Equal(true))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.OidcConfig.Userinfo).To(gomega.Equal(true))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AuthProfileRef).To(gomega.ContainSubstring("thisisaviref-authprofileoauth"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].ResourceServer.AccessType).To(gomega.Equal(lib.ACCESS_TOKEN_TYPE_OPAQUE))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].ResourceServer.IntrospectionDataTimeout).To(gomega.Equal(int32(60)))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].ResourceServer.OpaqueTokenParams.ServerID).To(gomega.Equal("my-server-id"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].ResourceServer.OpaqueTokenParams.ServerSecret).To(gomega.Equal("my-server-secret"))
	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].ResourceServer.JwtParams).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].SamlSpConfig).To(gomega.BeNil())

	//Update with Oidc parameters as false
	srUpdate := integrationtest.FakeSSORule{
		Name:      srName,
		Namespace: "default",
		Fqdn:      "foo.com",
		SSOType:   "OAuth",
	}.SSORule()
	srUpdate.ResourceVersion = "2"
	oidcEnable, profile, userinfo := false, false, false
	srUpdate.Spec.OauthVsConfig.OauthSettings[0].AppSettings.OidcConfig = &v1alpha2.OIDCConfig{
		OidcEnable: &oidcEnable,
		Profile:    &profile,
		Userinfo:   &userinfo,
	}
	_, err = v1alpha2CRDClient.AkoV1alpha2().SSORules("default").Update(context.TODO(), srUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating SSORule: %v", err)
	}
	g.Eventually(func() string {
		ssoRule, _ := v1alpha2CRDClient.AkoV1alpha2().SSORules("default").Get(context.TODO(), srName, metav1.GetOptions{})
		return ssoRule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	// update is not getting reflected on evh nodes immediately. Hence adding a sleep of 5 seconds.
	time.Sleep(5 * time.Second)

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.OidcConfig.OidcEnable).To(gomega.Equal(false))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.OidcConfig.Userinfo).To(gomega.Equal(false))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.OidcConfig.Profile).To(gomega.Equal(false))

	// Delete/Disable
	integrationtest.TeardownSSORule(t, g, sniVSKey, srName)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	g.Expect(nodes[0].EvhNodes[0].SsoPolicyRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].SamlSpConfig).To(gomega.BeNil())

	err = integrationtest.TearDownOAuthSecret()
	if err != nil {
		t.Fatalf("error in deleting my-oauth-secret: %v", err)
	}
	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestCreateUpdateDeleteSSORuleForEvhInsecure(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	srName := objNameMap.GenerateName("samplesr-foo")

	// create insecure ingress, SSORule should be applied in case of EVH
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       false,
		withSecret:  false,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	err := integrationtest.SetUpOAuthSecret()
	if err != nil {
		t.Fatalf("error in creating my-oauth-secret: %v", err)
	}
	// Sleeping for 5s for secret to be updated in informer
	time.Sleep(5 * time.Second)

	integrationtest.SetupSSORule(t, srName, "foo.com", "SAML")

	g.Eventually(func() string {
		ssoRule, _ := v1alpha2CRDClient.AkoV1alpha2().SSORules("default").Get(context.TODO(), srName, metav1.GetOptions{})
		return ssoRule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	integrationtest.VerifyMetadataSSORule(t, g, sniVSKey, "default/samplesr-foo", true)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig).To(gomega.BeNil())
	g.Expect(*nodes[0].EvhNodes[0].SsoPolicyRef).To(gomega.ContainSubstring("thisisaviref-ssopolicysaml"))
	g.Expect(nodes[0].EvhNodes[0].SamlSpConfig.AcsIndex).To(gomega.BeNil())
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.AuthnReqAcsType).To(gomega.Equal("SAML_AUTHN_REQ_ACS_TYPE_NONE"))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.CookieName).To(gomega.Equal("MY_SAML_COOKIE"))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.CookieTimeout).To(gomega.Equal(int32(120)))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.EntityID).To(gomega.Equal("my-entityid"))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.SigningSslKeyAndCertificateRef).To(gomega.ContainSubstring("thisisaviref-sslkeyandcertrefsaml"))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.SingleSignonURL).To(gomega.Equal("https://auth.com/sso/acs/"))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.UseIdpSessionTimeout).To(gomega.Equal(false))

	//Update with oauth parameters instead of saml
	srUpdate := integrationtest.FakeSSORule{
		Name:      srName,
		Namespace: "default",
		Fqdn:      "foo.com",
		SSOType:   "OAuth",
	}.SSORule()
	srUpdate.ResourceVersion = "2"
	_, err = v1alpha2CRDClient.AkoV1alpha2().SSORules("default").Update(context.TODO(), srUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating SSORule: %v", err)
	}
	g.Eventually(func() string {
		ssoRule, _ := v1alpha2CRDClient.AkoV1alpha2().SSORules("default").Get(context.TODO(), srName, metav1.GetOptions{})
		return ssoRule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	// update is not getting reflected on evh nodes immediately. Hence adding a sleep of 5 seconds.
	time.Sleep(5 * time.Second)

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].EvhNodes[0].SamlSpConfig).To(gomega.BeNil())
	g.Expect(*nodes[0].EvhNodes[0].SsoPolicyRef).To(gomega.ContainSubstring("thisisaviref-ssopolicyoauth"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.CookieName).To(gomega.Equal("MY_OAUTH_COOKIE"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.CookieTimeout).To(gomega.Equal(int32(120)))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.LogoutURI).To(gomega.Equal("https://auth.com/oauth/logout"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.RedirectURI).To(gomega.Equal("https://auth.com/oauth/redirect"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.PostLogoutRedirectURI).To(gomega.Equal("https://auth.com/oauth/post-logout-redirect"))
	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings).To(gomega.HaveLen(1))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.ClientID).To(gomega.Equal("my-client-id"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.ClientSecret).To(gomega.Equal("my-client-secret"))
	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.Scopes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.Scopes[0]).To(gomega.Equal("scope-1"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.OidcConfig.OidcEnable).To(gomega.Equal(true))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.OidcConfig.Profile).To(gomega.Equal(true))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.OidcConfig.Userinfo).To(gomega.Equal(true))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AuthProfileRef).To(gomega.ContainSubstring("thisisaviref-authprofileoauth"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].ResourceServer.AccessType).To(gomega.Equal(lib.ACCESS_TOKEN_TYPE_OPAQUE))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].ResourceServer.IntrospectionDataTimeout).To(gomega.Equal(int32(60)))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].ResourceServer.OpaqueTokenParams.ServerID).To(gomega.Equal("my-server-id"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].ResourceServer.OpaqueTokenParams.ServerSecret).To(gomega.Equal("my-server-secret"))
	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].ResourceServer.JwtParams).To(gomega.BeNil())

	// Delete/Disable
	integrationtest.TeardownSSORule(t, g, sniVSKey, srName)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	g.Expect(nodes[0].EvhNodes[0].SsoPolicyRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].SamlSpConfig).To(gomega.BeNil())

	err = integrationtest.TearDownOAuthSecret()
	if err != nil {
		t.Fatalf("error in deleting my-oauth-secret: %v", err)
	}
	TearDownIngressForCacheSyncCheck(t, "", ingressName, svcName, modelName)
}

func TestCreateUpdateDeleteSSORuleForEvhJwt(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	srName := objNameMap.GenerateName("samplesr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	err := integrationtest.SetUpOAuthSecret()
	if err != nil {
		t.Fatalf("error in creating my-oauth-secret: %v", err)
	} else {
		// Sleeping for 5s for secret to be updated in informer
		time.Sleep(5 * time.Second)
	}
	integrationtest.SetupSSORule(t, srName, "foo.com", "OAuth")

	g.Eventually(func() string {
		ssoRule, _ := v1alpha2CRDClient.AkoV1alpha2().SSORules("default").Get(context.TODO(), srName, metav1.GetOptions{})
		return ssoRule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	integrationtest.VerifyMetadataSSORule(t, g, sniVSKey, "default/samplesr-foo", true)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	g.Expect(*nodes[0].EvhNodes[0].SsoPolicyRef).To(gomega.ContainSubstring("thisisaviref-ssopolicyoauth"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.CookieName).To(gomega.Equal("MY_OAUTH_COOKIE"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.CookieTimeout).To(gomega.Equal(int32(120)))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.LogoutURI).To(gomega.Equal("https://auth.com/oauth/logout"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.RedirectURI).To(gomega.Equal("https://auth.com/oauth/redirect"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.PostLogoutRedirectURI).To(gomega.Equal("https://auth.com/oauth/post-logout-redirect"))
	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings).To(gomega.HaveLen(1))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.ClientID).To(gomega.Equal("my-client-id"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.ClientSecret).To(gomega.Equal("my-client-secret"))
	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.Scopes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.Scopes[0]).To(gomega.Equal("scope-1"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.OidcConfig.OidcEnable).To(gomega.Equal(true))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.OidcConfig.Profile).To(gomega.Equal(true))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.OidcConfig.Userinfo).To(gomega.Equal(true))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AuthProfileRef).To(gomega.ContainSubstring("thisisaviref-authprofileoauth"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].ResourceServer.AccessType).To(gomega.Equal(lib.ACCESS_TOKEN_TYPE_OPAQUE))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].ResourceServer.IntrospectionDataTimeout).To(gomega.Equal(int32(60)))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].ResourceServer.OpaqueTokenParams.ServerID).To(gomega.Equal("my-server-id"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].ResourceServer.OpaqueTokenParams.ServerSecret).To(gomega.Equal("my-server-secret"))
	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].ResourceServer.JwtParams).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].SamlSpConfig).To(gomega.BeNil())

	//Update with Opaque token parameters instead of jwt
	srUpdate := integrationtest.FakeSSORule{
		Name:      srName,
		Namespace: "default",
		Fqdn:      "foo.com",
		SSOType:   "OAuth",
	}.SSORule()
	srUpdate.ResourceVersion = "2"
	accessType := lib.ACCESS_TOKEN_TYPE_JWT
	audience := "my-audience"
	srUpdate.Spec.OauthVsConfig.OauthSettings[0].ResourceServer.AccessType = &accessType
	srUpdate.Spec.OauthVsConfig.OauthSettings[0].ResourceServer.JwtParams = &v1alpha2.JWTValidationParams{Audience: &audience}
	srUpdate.Spec.OauthVsConfig.OauthSettings[0].ResourceServer.OpaqueTokenParams = nil

	_, err = v1alpha2CRDClient.AkoV1alpha2().SSORules("default").Update(context.TODO(), srUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating SSORule: %v", err)
	}
	g.Eventually(func() string {
		ssoRule, _ := v1alpha2CRDClient.AkoV1alpha2().SSORules("default").Get(context.TODO(), srName, metav1.GetOptions{})
		return ssoRule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	// update is not getting reflected on evh nodes immediately. Hence adding a sleep of 5 seconds.
	time.Sleep(5 * time.Second)

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(*nodes[0].EvhNodes[0].SsoPolicyRef).To(gomega.ContainSubstring("thisisaviref-ssopolicyoauth"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.CookieName).To(gomega.Equal("MY_OAUTH_COOKIE"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.CookieTimeout).To(gomega.Equal(int32(120)))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.LogoutURI).To(gomega.Equal("https://auth.com/oauth/logout"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.RedirectURI).To(gomega.Equal("https://auth.com/oauth/redirect"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.PostLogoutRedirectURI).To(gomega.Equal("https://auth.com/oauth/post-logout-redirect"))
	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings).To(gomega.HaveLen(1))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.ClientID).To(gomega.Equal("my-client-id"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.ClientSecret).To(gomega.Equal("my-client-secret"))
	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.Scopes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.Scopes[0]).To(gomega.Equal("scope-1"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.OidcConfig.OidcEnable).To(gomega.Equal(true))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.OidcConfig.Profile).To(gomega.Equal(true))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.OidcConfig.Userinfo).To(gomega.Equal(true))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AuthProfileRef).To(gomega.ContainSubstring("thisisaviref-authprofileoauth"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].ResourceServer.AccessType).To(gomega.Equal(lib.ACCESS_TOKEN_TYPE_JWT))
	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].ResourceServer.IntrospectionDataTimeout).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].ResourceServer.OpaqueTokenParams).To(gomega.BeNil())
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].ResourceServer.JwtParams.Audience).To(gomega.Equal("my-audience"))
	g.Expect(nodes[0].EvhNodes[0].SamlSpConfig).To(gomega.BeNil())

	// Delete/Disable
	integrationtest.TeardownSSORule(t, g, sniVSKey, srName)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	g.Expect(nodes[0].EvhNodes[0].SsoPolicyRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].SamlSpConfig).To(gomega.BeNil())

	err = integrationtest.TearDownOAuthSecret()
	if err != nil {
		t.Fatalf("error in deleting my-oauth-secret: %v", err)
	}
	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestCreateUpdateDeleteSSORuleForEvhSamlACS(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	srName := objNameMap.GenerateName("samplesr-foo")

	// create insecure ingress, SSORule should be applied in case of EVH
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       false,
		withSecret:  false,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	integrationtest.SetupSSORule(t, srName, "foo.com", "SAML")

	g.Eventually(func() string {
		ssoRule, _ := v1alpha2CRDClient.AkoV1alpha2().SSORules("default").Get(context.TODO(), srName, metav1.GetOptions{})
		return ssoRule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	integrationtest.VerifyMetadataSSORule(t, g, sniVSKey, "default/samplesr-foo", true)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig).To(gomega.BeNil())
	g.Expect(*nodes[0].EvhNodes[0].SsoPolicyRef).To(gomega.ContainSubstring("thisisaviref-ssopolicysaml"))
	g.Expect(nodes[0].EvhNodes[0].SamlSpConfig.AcsIndex).To(gomega.BeNil())
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.AuthnReqAcsType).To(gomega.Equal("SAML_AUTHN_REQ_ACS_TYPE_NONE"))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.CookieName).To(gomega.Equal("MY_SAML_COOKIE"))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.CookieTimeout).To(gomega.Equal(int32(120)))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.EntityID).To(gomega.Equal("my-entityid"))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.SigningSslKeyAndCertificateRef).To(gomega.ContainSubstring("thisisaviref-sslkeyandcertrefsaml"))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.SingleSignonURL).To(gomega.Equal("https://auth.com/sso/acs/"))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.UseIdpSessionTimeout).To(gomega.Equal(false))

	//Update saml parameters with acs type url
	srUpdate := integrationtest.FakeSSORule{
		Name:      srName,
		Namespace: "default",
		Fqdn:      "foo.com",
		SSOType:   "SAML",
	}.SSORule()
	acsType := "SAML_AUTHN_REQ_ACS_TYPE_URL"
	acsIndex := int32(64)
	srUpdate.Spec.SamlSpConfig.AuthnReqAcsType = &acsType
	// setting AcsIndex but it will still be nil as act type is not index
	srUpdate.Spec.SamlSpConfig.AcsIndex = &acsIndex
	srUpdate.ResourceVersion = "2"
	_, err := v1alpha2CRDClient.AkoV1alpha2().SSORules("default").Update(context.TODO(), srUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating SSORule: %v", err)
	}
	g.Eventually(func() string {
		ssoRule, _ := v1alpha2CRDClient.AkoV1alpha2().SSORules("default").Get(context.TODO(), srName, metav1.GetOptions{})
		return ssoRule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	// update is not getting reflected on evh nodes immediately. Hence adding a sleep of 5 seconds.
	time.Sleep(5 * time.Second)

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig).To(gomega.BeNil())
	g.Expect(*nodes[0].EvhNodes[0].SsoPolicyRef).To(gomega.ContainSubstring("thisisaviref-ssopolicysaml"))
	g.Expect(nodes[0].EvhNodes[0].SamlSpConfig.AcsIndex).To(gomega.BeNil())
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.AuthnReqAcsType).To(gomega.Equal("SAML_AUTHN_REQ_ACS_TYPE_URL"))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.CookieName).To(gomega.Equal("MY_SAML_COOKIE"))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.CookieTimeout).To(gomega.Equal(int32(120)))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.EntityID).To(gomega.Equal("my-entityid"))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.SigningSslKeyAndCertificateRef).To(gomega.ContainSubstring("thisisaviref-sslkeyandcertrefsaml"))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.SingleSignonURL).To(gomega.Equal("https://auth.com/sso/acs/"))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.UseIdpSessionTimeout).To(gomega.Equal(false))

	//Update saml parameters with acs type index
	srUpdate = integrationtest.FakeSSORule{
		Name:      srName,
		Namespace: "default",
		Fqdn:      "foo.com",
		SSOType:   "SAML",
	}.SSORule()
	acsType = lib.SAML_AUTHN_REQ_ACS_TYPE_INDEX
	acsIndex = int32(64)
	srUpdate.Spec.SamlSpConfig.AuthnReqAcsType = &acsType
	// setting AcsIndex but it will still be nil as act type is not index
	srUpdate.Spec.SamlSpConfig.AcsIndex = &acsIndex
	srUpdate.ResourceVersion = "3"
	_, err = v1alpha2CRDClient.AkoV1alpha2().SSORules("default").Update(context.TODO(), srUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating SSORule: %v", err)
	}
	g.Eventually(func() string {
		ssoRule, _ := v1alpha2CRDClient.AkoV1alpha2().SSORules("default").Get(context.TODO(), srName, metav1.GetOptions{})
		return ssoRule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	// update is not getting reflected on evh nodes immediately. Hence adding a sleep of 5 seconds.
	time.Sleep(5 * time.Second)

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig).To(gomega.BeNil())
	g.Expect(*nodes[0].EvhNodes[0].SsoPolicyRef).To(gomega.ContainSubstring("thisisaviref-ssopolicysaml"))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.AcsIndex).To(gomega.Equal(int32(64)))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.AuthnReqAcsType).To(gomega.Equal(lib.SAML_AUTHN_REQ_ACS_TYPE_INDEX))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.CookieName).To(gomega.Equal("MY_SAML_COOKIE"))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.CookieTimeout).To(gomega.Equal(int32(120)))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.EntityID).To(gomega.Equal("my-entityid"))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.SigningSslKeyAndCertificateRef).To(gomega.ContainSubstring("thisisaviref-sslkeyandcertrefsaml"))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.SingleSignonURL).To(gomega.Equal("https://auth.com/sso/acs/"))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.UseIdpSessionTimeout).To(gomega.Equal(false))

	// Delete/Disable
	integrationtest.TeardownSSORule(t, g, sniVSKey, srName)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	g.Expect(nodes[0].EvhNodes[0].SsoPolicyRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].SamlSpConfig).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestCreateSSORuleBeforeIngressForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	srName := objNameMap.GenerateName("samplesr-foo")
	err := integrationtest.SetUpOAuthSecret()
	if err != nil {
		t.Fatalf("error in creating my-oauth-secret: %v", err)
	}
	// Sleeping for 5s for secret to be updated in informer
	time.Sleep(5 * time.Second)

	// creating SSORule before ingress
	integrationtest.SetupSSORule(t, srName, "foo.com", "OAuth")

	g.Eventually(func() string {
		ssoRule, _ := v1alpha2CRDClient.AkoV1alpha2().SSORules("default").Get(context.TODO(), srName, metav1.GetOptions{})
		return ssoRule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	modelName, _ := GetModelName("foo.com", "default")
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 10*time.Second).Should(gomega.Equal(1))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	g.Expect(*nodes[0].EvhNodes[0].SsoPolicyRef).To(gomega.ContainSubstring("thisisaviref-ssopolicyoauth"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.CookieName).To(gomega.Equal("MY_OAUTH_COOKIE"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.CookieTimeout).To(gomega.Equal(int32(120)))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.LogoutURI).To(gomega.Equal("https://auth.com/oauth/logout"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.RedirectURI).To(gomega.Equal("https://auth.com/oauth/redirect"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.PostLogoutRedirectURI).To(gomega.Equal("https://auth.com/oauth/post-logout-redirect"))
	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings).To(gomega.HaveLen(1))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.ClientID).To(gomega.Equal("my-client-id"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.ClientSecret).To(gomega.Equal("my-client-secret"))
	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.Scopes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.Scopes[0]).To(gomega.Equal("scope-1"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.OidcConfig.OidcEnable).To(gomega.Equal(true))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.OidcConfig.Profile).To(gomega.Equal(true))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AppSettings.OidcConfig.Userinfo).To(gomega.Equal(true))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].AuthProfileRef).To(gomega.ContainSubstring("thisisaviref-authprofileoauth"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].ResourceServer.AccessType).To(gomega.Equal(lib.ACCESS_TOKEN_TYPE_OPAQUE))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].ResourceServer.IntrospectionDataTimeout).To(gomega.Equal(int32(60)))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].ResourceServer.OpaqueTokenParams.ServerID).To(gomega.Equal("my-server-id"))
	g.Expect(*nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].ResourceServer.OpaqueTokenParams.ServerSecret).To(gomega.Equal("my-server-secret"))
	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig.OauthSettings[0].ResourceServer.JwtParams).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].SamlSpConfig).To(gomega.BeNil())

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	integrationtest.TeardownSSORule(t, g, sniVSKey, srName)
	err = integrationtest.TearDownOAuthSecret()
	if err != nil {
		t.Fatalf("error in deleting my-oauth-secret: %v", err)
	}
	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestGoodToBadSSORuleForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	srName := objNameMap.GenerateName("samplesr-foo")

	// create insecure ingress, SSORule should be applied in case of EVH
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       false,
		withSecret:  false,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	integrationtest.SetupSSORule(t, srName, "foo.com", "SAML")

	g.Eventually(func() string {
		ssoRule, _ := v1alpha2CRDClient.AkoV1alpha2().SSORules("default").Get(context.TODO(), srName, metav1.GetOptions{})
		return ssoRule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	integrationtest.VerifyMetadataSSORule(t, g, sniVSKey, "default/samplesr-foo", true)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig).To(gomega.BeNil())
	g.Expect(*nodes[0].EvhNodes[0].SsoPolicyRef).To(gomega.ContainSubstring("thisisaviref-ssopolicysaml"))
	g.Expect(nodes[0].EvhNodes[0].SamlSpConfig.AcsIndex).To(gomega.BeNil())
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.AuthnReqAcsType).To(gomega.Equal("SAML_AUTHN_REQ_ACS_TYPE_NONE"))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.CookieName).To(gomega.Equal("MY_SAML_COOKIE"))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.CookieTimeout).To(gomega.Equal(int32(120)))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.EntityID).To(gomega.Equal("my-entityid"))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.SigningSslKeyAndCertificateRef).To(gomega.ContainSubstring("thisisaviref-sslkeyandcertrefsaml"))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.SingleSignonURL).To(gomega.Equal("https://auth.com/sso/acs/"))
	g.Expect(*nodes[0].EvhNodes[0].SamlSpConfig.UseIdpSessionTimeout).To(gomega.Equal(false))

	//Update with bad sso policy ref
	srUpdate := integrationtest.FakeSSORule{
		Name:      srName,
		Namespace: "default",
		Fqdn:      "foo.com",
		SSOType:   "SAML",
	}.SSORule()
	badSsoPloicyRef := "thisisBADssopolicyref"
	srUpdate.Spec.SsoPolicyRef = &badSsoPloicyRef
	srUpdate.ResourceVersion = "2"
	_, err := v1alpha2CRDClient.AkoV1alpha2().SSORules("default").Update(context.TODO(), srUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating SSORule: %v", err)
	}
	g.Eventually(func() string {
		ssoRule, _ := v1alpha2CRDClient.AkoV1alpha2().SSORules("default").Get(context.TODO(), srName, metav1.GetOptions{})
		return ssoRule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Rejected"))

	// update is not getting reflected on evh nodes immediately. Hence adding a sleep of 5 seconds.
	time.Sleep(5 * time.Second)

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	// the last applied SSORule values would exist.
	g.Expect(*nodes[0].EvhNodes[0].SsoPolicyRef).To(gomega.ContainSubstring("thisisaviref-ssopolicysaml"))

	// Delete/Disable
	integrationtest.TeardownSSORule(t, g, sniVSKey, srName)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	g.Expect(nodes[0].EvhNodes[0].SsoPolicyRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].OauthVsConfig).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].SamlSpConfig).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheck(t, "", ingressName, svcName, modelName)
}

func TestCreateUpdateDeleteL7RuleInHostRule(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName, _ := GetModelName("foo.com", "default")
	hrName := objNameMap.GenerateName("samplehr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)
	integrationtest.SetupHostRule(t, hrName, "foo.com", true)
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/"+hrName, true)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	l7ruleName := objNameMap.GenerateName("samplel7rule")
	integrationtest.SetupL7Rule(t, l7ruleName, g)

	//Update hostrule with L7rule
	hrUpdate := integrationtest.FakeHostRule{
		Name:      hrName,
		Namespace: "default",
		Fqdn:      "foo.com",
	}.HostRule()
	hrUpdate.Spec.VirtualHost.L7Rule = l7ruleName
	hrUpdate.ResourceVersion = "2"
	_, err := v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, err := lib.AKOControlConfig().CRDInformers().HostRuleInformer.Lister().HostRules("default").Get(hrName)
		if err != nil {
			return ""
		}
		return hostrule.Status.Status
	}, 30*time.Second, 1*time.Second).Should(gomega.Equal("Accepted"))
	g.Eventually(func() *bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes) > 0 {
			return nodes[0].EvhNodes[0].AllowInvalidClientCert
		}
		return nil
	}, 25*time.Second, 1*time.Second).ShouldNot(gomega.BeNil())

	g.Eventually(func() interface{} {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return nodes[0].EvhNodes[0].AviVsNodeGeneratedFields
	}, 25*time.Second, 1*time.Second).ShouldNot(gomega.BeNil())

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(*nodes[0].EvhNodes[0].CloseClientConnOnConfigUpdate).To(gomega.Equal(true))
	g.Expect(*nodes[0].EvhNodes[0].AllowInvalidClientCert).To(gomega.Equal(true))
	g.Expect(*nodes[0].EvhNodes[0].IgnPoolNetReach).To(gomega.Equal(false))
	g.Expect(*nodes[0].EvhNodes[0].RemoveListeningPortOnVsDown).To(gomega.Equal(false))
	g.Expect(*nodes[0].EvhNodes[0].BotPolicyRef).To(gomega.ContainSubstring("thisisaviref-botpolicy"))
	g.Expect(*nodes[0].EvhNodes[0].SslSessCacheAvgSize).To(gomega.Equal(uint32(2024)))
	g.Expect(*nodes[0].EvhNodes[0].MinPoolsUp).To(gomega.Equal(uint32(0)))
	g.Expect(nodes[0].EvhNodes[0].HostNameXlate).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].SecurityPolicyRef).To(gomega.BeNil())
	//remove L7rule from hostrule
	hrUpdate = integrationtest.FakeHostRule{
		Name:      hrName,
		Namespace: "default",
		Fqdn:      "foo.com",
	}.HostRule()
	hrUpdate.ResourceVersion = "3"
	_, err = v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() *bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			return nodes[0].EvhNodes[0].AllowInvalidClientCert
		}
		v := true
		return &v
	}, 25*time.Second).Should(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].CloseClientConnOnConfigUpdate).Should(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].CloseClientConnOnConfigUpdate).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].AllowInvalidClientCert).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].IgnPoolNetReach).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].RemoveListeningPortOnVsDown).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].BotPolicyRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].SslSessCacheAvgSize).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].MinPoolsUp).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].HostNameXlate).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].SecurityPolicyRef).To(gomega.BeNil())

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrName)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	g.Expect(nodes[0].EvhNodes[0].Enabled).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].ICAPProfileRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].WafPolicyRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].ApplicationProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].AnalyticsProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].ErrorPageProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].VsDatascriptRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].SslProfileRef).To(gomega.BeNil())

	if err := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L7Rules("default").Delete(context.TODO(), l7ruleName, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("error in deleting l7Rule: %v", err)
	}
	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}
func TestDeleteL7RulePresentInHostRule(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName, _ := GetModelName("foo.com", "default")
	hrName := objNameMap.GenerateName("samplehr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)
	integrationtest.SetupHostRule(t, hrName, "foo.com", true)
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/"+hrName, true)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	l7ruleName := objNameMap.GenerateName("samplel7rule")
	integrationtest.SetupL7Rule(t, l7ruleName, g)
	//Update hostrule with L7rule
	hrUpdate := integrationtest.FakeHostRule{
		Name:      hrName,
		Namespace: "default",
		Fqdn:      "foo.com",
	}.HostRule()
	hrUpdate.Spec.VirtualHost.L7Rule = l7ruleName
	hrUpdate.ResourceVersion = "2"
	_, err := v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))
	g.Eventually(func() *bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes) > 0 {
			return nodes[0].EvhNodes[0].AllowInvalidClientCert
		}
		return nil
	}, 25*time.Second).ShouldNot(gomega.BeNil())
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(*nodes[0].EvhNodes[0].CloseClientConnOnConfigUpdate).To(gomega.Equal(true))
	g.Expect(*nodes[0].EvhNodes[0].AllowInvalidClientCert).To(gomega.Equal(true))
	g.Expect(*nodes[0].EvhNodes[0].IgnPoolNetReach).To(gomega.Equal(false))
	g.Expect(*nodes[0].EvhNodes[0].RemoveListeningPortOnVsDown).To(gomega.Equal(false))
	g.Expect(*nodes[0].EvhNodes[0].BotPolicyRef).To(gomega.ContainSubstring("thisisaviref-botpolicy"))
	g.Expect(*nodes[0].EvhNodes[0].SslSessCacheAvgSize).To(gomega.Equal(uint32(2024)))
	g.Expect(*nodes[0].EvhNodes[0].MinPoolsUp).To(gomega.Equal(uint32(0)))
	g.Expect(nodes[0].EvhNodes[0].HostNameXlate).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].SecurityPolicyRef).To(gomega.BeNil())
	//Delete L7 Rule
	if err := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L7Rules("default").Delete(context.TODO(), l7ruleName, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("error in deleting l7Rule: %v", err)
	}
	g.Eventually(func() *bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			return nodes[0].EvhNodes[0].AllowInvalidClientCert
		}
		v := true
		return &v
	}, 25*time.Second).Should(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].CloseClientConnOnConfigUpdate).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].AllowInvalidClientCert).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].IgnPoolNetReach).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].RemoveListeningPortOnVsDown).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].BotPolicyRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].SslSessCacheAvgSize).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].MinPoolsUp).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].HostNameXlate).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].SecurityPolicyRef).To(gomega.BeNil())

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrName)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].EvhNodes[0].Enabled).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].ICAPProfileRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].WafPolicyRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].ApplicationProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].AnalyticsProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].ErrorPageProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].VsDatascriptRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].SslProfileRef).To(gomega.BeNil())
	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)

}
func TestChangeL7RuleInHostRule(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName, _ := GetModelName("foo.com", "default")
	hrName := objNameMap.GenerateName("samplehr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)
	integrationtest.SetupHostRule(t, hrName, "foo.com", true)
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/"+hrName, true)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	l7ruleName := objNameMap.GenerateName("samplel7rule")
	integrationtest.SetupL7Rule(t, l7ruleName, g)
	//Update hostrule with L7rule
	hrUpdate := integrationtest.FakeHostRule{
		Name:      hrName,
		Namespace: "default",
		Fqdn:      "foo.com",
	}.HostRule()
	hrUpdate.Spec.VirtualHost.L7Rule = l7ruleName
	hrUpdate.ResourceVersion = "2"
	_, err := v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))
	g.Eventually(func() *bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes) > 0 {
			return nodes[0].EvhNodes[0].AllowInvalidClientCert
		}
		return nil
	}, 25*time.Second).ShouldNot(gomega.BeNil())
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(*nodes[0].EvhNodes[0].CloseClientConnOnConfigUpdate).To(gomega.Equal(true))
	g.Expect(*nodes[0].EvhNodes[0].AllowInvalidClientCert).To(gomega.Equal(true))
	g.Expect(*nodes[0].EvhNodes[0].IgnPoolNetReach).To(gomega.Equal(false))
	g.Expect(*nodes[0].EvhNodes[0].RemoveListeningPortOnVsDown).To(gomega.Equal(false))
	g.Expect(*nodes[0].EvhNodes[0].BotPolicyRef).To(gomega.ContainSubstring("thisisaviref-botpolicy"))
	g.Expect(*nodes[0].EvhNodes[0].SslSessCacheAvgSize).To(gomega.Equal(uint32(2024)))
	g.Expect(*nodes[0].EvhNodes[0].MinPoolsUp).To(gomega.Equal(uint32(0)))
	g.Expect(nodes[0].EvhNodes[0].HostNameXlate).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].SecurityPolicyRef).To(gomega.BeNil())
	l7ruleName2 := "samplel7rule2"
	integrationtest.SetupL7Rule(t, l7ruleName2, g)
	l7rule2 := integrationtest.FakeL7Rule{Name: l7ruleName2,
		Namespace:                     "default",
		AllowInvalidClientCert:        false,
		BotPolicyRef:                  "thisisaviref-botpolicy",
		CloseClientConnOnConfigUpdate: true,
		HostNameXlate:                 "hostname.com",
		IgnPoolNetReach:               false,
		MinPoolsUp:                    0,
		SecurityPolicyRef:             "thisisaviref-secpolicy",
		RemoveListeningPortOnVsDown:   true,
		SslSessCacheAvgSize:           2024,
	}.L7Rule()
	l7rule2.ResourceVersion = "2"
	_, err = v1alpha2CRDClient.AkoV1alpha2().L7Rules("default").Update(context.TODO(), l7rule2, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating L7Rule: %v", err)
	}
	g.Eventually(func() string {
		l7Rule, _ := v1alpha2CRDClient.AkoV1alpha2().L7Rules("default").Get(context.TODO(), l7ruleName2, metav1.GetOptions{})
		return l7Rule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	//Update hostrule with L7rule2
	hrUpdate = integrationtest.FakeHostRule{
		Name:      hrName,
		Namespace: "default",
		Fqdn:      "foo.com",
	}.HostRule()
	hrUpdate.Spec.VirtualHost.L7Rule = l7ruleName2
	hrUpdate.ResourceVersion = "3"
	_, err = v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}

	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))
	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes) > 0 {
			return *nodes[0].EvhNodes[0].AllowInvalidClientCert
		}
		return true
	}, 25*time.Second).Should(gomega.Equal(false))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(*nodes[0].EvhNodes[0].RemoveListeningPortOnVsDown).To(gomega.Equal(true))
	g.Expect(*nodes[0].EvhNodes[0].CloseClientConnOnConfigUpdate).To(gomega.Equal(true))
	g.Expect(*nodes[0].EvhNodes[0].AllowInvalidClientCert).To(gomega.Equal(false))
	g.Expect(*nodes[0].EvhNodes[0].IgnPoolNetReach).To(gomega.Equal(false))
	g.Expect(*nodes[0].EvhNodes[0].BotPolicyRef).To(gomega.ContainSubstring("thisisaviref-botpolicy"))
	g.Expect(*nodes[0].EvhNodes[0].SslSessCacheAvgSize).To(gomega.Equal(uint32(2024)))
	g.Expect(*nodes[0].EvhNodes[0].MinPoolsUp).To(gomega.Equal(uint32(0)))
	g.Expect(nodes[0].EvhNodes[0].HostNameXlate).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].SecurityPolicyRef).To(gomega.BeNil())
	integrationtest.TeardownHostRule(t, g, sniVSKey, hrName)
	if err := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L7Rules("default").Delete(context.TODO(), l7ruleName, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("error in deleting l7Rule: %v", err)
	}
	if err := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L7Rules("default").Delete(context.TODO(), l7ruleName2, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("error in deleting l7Rule: %v", err)
	}
	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestValidToInvalidL7rule(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName, _ := GetModelName("foo.com", "default")
	hrName := objNameMap.GenerateName("samplehr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)
	integrationtest.SetupHostRule(t, hrName, "foo.com", true)
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/"+hrName, true)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	l7ruleName := objNameMap.GenerateName("samplel7rule")
	integrationtest.SetupL7Rule(t, l7ruleName, g)
	//Update hostrule with L7rule
	hrUpdate := integrationtest.FakeHostRule{
		Name:      hrName,
		Namespace: "default",
		Fqdn:      "foo.com",
	}.HostRule()
	hrUpdate.Spec.VirtualHost.L7Rule = l7ruleName
	hrUpdate.ResourceVersion = "2"
	_, err := v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))
	g.Eventually(func() *bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes) > 0 {
			return nodes[0].EvhNodes[0].AllowInvalidClientCert
		}
		return nil
	}, 25*time.Second).ShouldNot(gomega.BeNil())
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(*nodes[0].EvhNodes[0].CloseClientConnOnConfigUpdate).To(gomega.Equal(true))
	l7rule := integrationtest.FakeL7Rule{Name: l7ruleName,
		Namespace:                     "default",
		AllowInvalidClientCert:        false,
		BotPolicyRef:                  "invalidbotpolicy",
		CloseClientConnOnConfigUpdate: true,
		HostNameXlate:                 "hostname.com",
		IgnPoolNetReach:               false,
		MinPoolsUp:                    0,
		SecurityPolicyRef:             "thisisaviref-secpolicy",
		RemoveListeningPortOnVsDown:   true,
		SslSessCacheAvgSize:           2024,
	}.L7Rule()
	val := true
	l7rule.Spec.RemoveListeningPortOnVsDown = &val
	invalidBotPolicyRef := "invalidBotPolicy"
	l7rule.Spec.BotPolicyRef = &invalidBotPolicyRef
	l7rule.ResourceVersion = "2"

	_, err = v1alpha2CRDClient.AkoV1alpha2().L7Rules("default").Update(context.TODO(), l7rule, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating L7Rule: %v", err)
	}

	g.Eventually(func() string {
		l7Rule, _ := v1alpha2CRDClient.AkoV1alpha2().L7Rules("default").Get(context.TODO(), l7ruleName, metav1.GetOptions{})
		return l7Rule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Rejected"))
	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes) > 0 {
			return *nodes[0].EvhNodes[0].AllowInvalidClientCert
		}
		return false
	}, 25*time.Second).Should(gomega.Equal(true))
	botPolicy := "thisisaviref-bop"
	l7rule.Spec.BotPolicyRef = &botPolicy
	l7rule.ResourceVersion = "3"
	_, err = v1alpha2CRDClient.AkoV1alpha2().L7Rules("default").Update(context.TODO(), l7rule, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating L7Rule: %v", err)
	}

	g.Eventually(func() string {
		l7Rule, _ := v1alpha2CRDClient.AkoV1alpha2().L7Rules("default").Get(context.TODO(), l7ruleName, metav1.GetOptions{})
		return l7Rule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes) > 0 {
			return *nodes[0].EvhNodes[0].AllowInvalidClientCert
		}
		return true
	}, 25*time.Second).Should(gomega.Equal(false))

	g.Expect(*nodes[0].EvhNodes[0].CloseClientConnOnConfigUpdate).To(gomega.Equal(true))
	g.Expect(*nodes[0].EvhNodes[0].AllowInvalidClientCert).To(gomega.Equal(false))
	g.Expect(*nodes[0].EvhNodes[0].IgnPoolNetReach).To(gomega.Equal(false))
	g.Expect(*nodes[0].EvhNodes[0].RemoveListeningPortOnVsDown).To(gomega.Equal(true))
	g.Expect(*nodes[0].EvhNodes[0].BotPolicyRef).To(gomega.ContainSubstring("thisisaviref-bop"))
	g.Expect(*nodes[0].EvhNodes[0].SslSessCacheAvgSize).To(gomega.Equal(uint32(2024)))
	g.Expect(*nodes[0].EvhNodes[0].MinPoolsUp).To(gomega.Equal(uint32(0)))
	g.Expect(nodes[0].EvhNodes[0].HostNameXlate).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].SecurityPolicyRef).To(gomega.BeNil())
	integrationtest.TeardownHostRule(t, g, sniVSKey, hrName)
	if err := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L7Rules("default").Delete(context.TODO(), l7ruleName, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("error in deleting l7Rule: %v", err)
	}
	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestDeleteHostRuleWithActiveL7Rule(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName, _ := GetModelName("foo.com", "default")
	hrName := objNameMap.GenerateName("samplehr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)
	integrationtest.SetupHostRule(t, hrName, "foo.com", true)
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/"+hrName, true)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	l7ruleName := objNameMap.GenerateName("samplel7rule")
	integrationtest.SetupL7Rule(t, l7ruleName, g)
	//Update hostrule with L7rule
	hrUpdate := integrationtest.FakeHostRule{
		Name:      hrName,
		Namespace: "default",
		Fqdn:      "foo.com",
	}.HostRule()
	hrUpdate.Spec.VirtualHost.L7Rule = l7ruleName
	hrUpdate.ResourceVersion = "2"
	_, err := v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))
	g.Eventually(func() *bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes) > 0 {
			return nodes[0].EvhNodes[0].AllowInvalidClientCert
		}
		return nil
	}, 25*time.Second).ShouldNot(gomega.BeNil())
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(*nodes[0].EvhNodes[0].CloseClientConnOnConfigUpdate).To(gomega.Equal(true))
	g.Expect(*nodes[0].EvhNodes[0].AllowInvalidClientCert).To(gomega.Equal(true))
	g.Expect(*nodes[0].EvhNodes[0].IgnPoolNetReach).To(gomega.Equal(false))
	g.Expect(*nodes[0].EvhNodes[0].RemoveListeningPortOnVsDown).To(gomega.Equal(false))
	g.Expect(*nodes[0].EvhNodes[0].BotPolicyRef).To(gomega.ContainSubstring("thisisaviref-botpolicy"))
	g.Expect(*nodes[0].EvhNodes[0].SslSessCacheAvgSize).To(gomega.Equal(uint32(2024)))
	g.Expect(*nodes[0].EvhNodes[0].MinPoolsUp).To(gomega.Equal(uint32(0)))
	g.Expect(nodes[0].EvhNodes[0].HostNameXlate).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].SecurityPolicyRef).To(gomega.BeNil())
	//Delete Hostrule
	integrationtest.TeardownHostRule(t, g, sniVSKey, hrName)
	g.Eventually(func() *bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			return nodes[0].EvhNodes[0].AllowInvalidClientCert
		}
		v := true
		return &v
	}, 25*time.Second).Should(gomega.BeNil())
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].EvhNodes[0].Enabled).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].ICAPProfileRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].WafPolicyRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].ApplicationProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].AnalyticsProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].ErrorPageProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].VsDatascriptRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].CloseClientConnOnConfigUpdate).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].AllowInvalidClientCert).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].IgnPoolNetReach).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].RemoveListeningPortOnVsDown).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].BotPolicyRef).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].SslSessCacheAvgSize).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].MinPoolsUp).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].HostNameXlate).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].SecurityPolicyRef).To(gomega.BeNil())
	g.Expect(nodes[0].SslProfileRef).To(gomega.BeNil())
	//Delete L7 Rule
	if err := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L7Rules("default").Delete(context.TODO(), l7ruleName, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("error in deleting l7Rule: %v", err)
	}
	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)

}

func TestHostRuleUseRegexSecure(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	fqdn := "foo.com"
	namespace := "default"

	hrName := objNameMap.GenerateName("samplehr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	hostrule := integrationtest.FakeHostRule{
		Name:      hrName,
		Namespace: namespace,
		Fqdn:      fqdn,
		UseRegex:  true,
	}
	hrCreate := hostrule.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
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

	g.Expect(nodes[0].EvhNodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].Path).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].RedirectPath).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].HppMap).To(gomega.BeNil())

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrName)
	time.Sleep(2 * time.Second)

	g.Expect(nodes[0].EvhNodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].Path).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].RedirectPath).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].HppMap).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestHostRuleAppRootSecure(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	fqdn := "foo.com"
	namespace := "default"
	appRootPath := "/foo"
	hrName := objNameMap.GenerateName("samplehr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
		dnsNames:    []string{fqdn},
		paths:       []string{"/"},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	hostrule := integrationtest.FakeHostRule{
		Name:                hrName,
		Namespace:           namespace,
		Fqdn:                fqdn,
		ApplicationRootPath: "/foo",
	}
	hrCreate := hostrule.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
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

	g.Expect(nodes[0].EvhNodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0]).Should(gomega.Equal(appRootPath))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].Path).To(gomega.Equal("/"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].Protocol).To(gomega.Equal("HTTP"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(80)))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].Path).To(gomega.Equal("/"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].Protocol).To(gomega.Equal("HTTPS"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].Path).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].RedirectPath).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].HppMap).To(gomega.BeNil())

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrName)
	g.Eventually(func() bool {
		return nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts == nil
	}, 30*time.Second).Should(gomega.BeTrue())

	g.Expect(nodes[0].EvhNodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0]).To(gomega.Equal("/"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].Path).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].RedirectPath).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].HppMap).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestHostRuleRegexAppRootSecure(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	fqdn := "foo.com"
	namespace := "default"
	appRootPath := "/foo"

	hrName := objNameMap.GenerateName("samplehr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
		dnsNames:    []string{fqdn, fqdn},
		paths:       []string{"/something(/|$)(.*)", "/"},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	hostrule := integrationtest.FakeHostRule{
		Name:                hrName,
		Namespace:           namespace,
		Fqdn:                fqdn,
		UseRegex:            true,
		ApplicationRootPath: "/foo",
	}
	hrCreate := hostrule.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
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

	g.Expect(nodes[0].EvhNodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0]).To(gomega.Equal("/something(/|$)(.*)"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].StringGroupRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].Path[0]).To(gomega.Equal(appRootPath))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].Path).To(gomega.Equal("/"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].Protocol).To(gomega.Equal("HTTP"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(80)))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].Path).To(gomega.Equal("/"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].Protocol).To(gomega.Equal("HTTPS"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].Path).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].RedirectPath).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].HppMap).To(gomega.BeNil())

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrName)
	g.Eventually(func() bool {
		return nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts == nil
	}, 30*time.Second).Should(gomega.BeTrue())

	g.Expect(nodes[0].EvhNodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0]).To(gomega.Equal("/something(/|$)(.*)"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].Path[0]).To(gomega.Equal("/"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].Path).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].RedirectPath).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].HppMap).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestHostRuleUseRegexInsecure(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	fqdn := "foo.com"
	namespace := "default"

	hrName := objNameMap.GenerateName("samplehr-foo")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       false,
		withSecret:  false,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	hostrule := integrationtest.FakeHostRule{
		Name:      hrName,
		Namespace: namespace,
		Fqdn:      fqdn,
		UseRegex:  true,
	}
	hrCreate := hostrule.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
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

	g.Expect(nodes[0].EvhNodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(1))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrName)
	time.Sleep(2 * time.Second)

	g.Expect(nodes[0].EvhNodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))

	TearDownIngressForCacheSyncCheck(t, "", ingressName, svcName, modelName)
}

func TestHostRuleAppRootInsecure(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	fqdn := "foo.com"
	namespace := "default"
	appRootPath := "/foo"

	hrName := objNameMap.GenerateName("samplehr-foo")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       false,
		withSecret:  false,
		serviceName: svcName,
		modelNames:  []string{modelName},
		dnsNames:    []string{fqdn},
		paths:       []string{"/"},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	hostrule := integrationtest.FakeHostRule{
		Name:                hrName,
		Namespace:           namespace,
		Fqdn:                fqdn,
		ApplicationRootPath: "/foo",
	}
	hrCreate := hostrule.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
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

	g.Expect(nodes[0].EvhNodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0]).Should(gomega.Equal(appRootPath))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].Path).To(gomega.Equal("/"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].Protocol).To(gomega.Equal("HTTP"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(80)))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].Path).To(gomega.Equal("/"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].Protocol).To(gomega.Equal("HTTPS"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPort).To(gomega.Equal("IS_IN"))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrName)
	g.Eventually(func() bool {
		return nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts == nil
	}, 30*time.Second).Should(gomega.BeTrue())

	g.Expect(nodes[0].EvhNodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0]).To(gomega.Equal("/"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))

	TearDownIngressForCacheSyncCheck(t, "", ingressName, svcName, modelName)
}

func TestHostRuleRegexAppRootInsecure(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	fqdn := "foo.com"
	namespace := "default"
	appRootPath := "/foo"

	hrName := objNameMap.GenerateName("samplehr-foo")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       false,
		withSecret:  false,
		serviceName: svcName,
		modelNames:  []string{modelName},
		dnsNames:    []string{fqdn, fqdn},
		paths:       []string{"/something(/|$)(.*)", "/"},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	hostrule := integrationtest.FakeHostRule{
		Name:                hrName,
		Namespace:           namespace,
		Fqdn:                fqdn,
		UseRegex:            true,
		ApplicationRootPath: "/foo",
	}
	hrCreate := hostrule.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
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

	g.Expect(nodes[0].EvhNodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0]).To(gomega.Equal("/something(/|$)(.*)"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].StringGroupRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].Path[0]).To(gomega.Equal(appRootPath))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].Path).To(gomega.Equal("/"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].Protocol).To(gomega.Equal("HTTP"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(80)))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].Path).To(gomega.Equal("/"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].Protocol).To(gomega.Equal("HTTPS"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPort).To(gomega.Equal("IS_IN"))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrName)
	g.Eventually(func() bool {
		return nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts == nil
	}, 30*time.Second).Should(gomega.BeTrue())

	g.Expect(nodes[0].EvhNodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0]).To(gomega.Equal("/something(/|$)(.*)"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].Path[0]).To(gomega.Equal("/"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].StringGroupRefs).To(gomega.HaveLen(0))

	TearDownIngressForCacheSyncCheck(t, "", ingressName, svcName, modelName)
}

func TestHostRuleAppRootSecureListenerPorts(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	fqdn := "foo.com"
	namespace := "default"
	appRootPath := "/foo"
	hrName := objNameMap.GenerateName("samplehr-foo")
	sharedHrName := objNameMap.GenerateName("samplehr-shared")
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
		dnsNames:    []string{fqdn},
		paths:       []string{"/"},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	sharedFQDN := "Shared-L7-EVH"
	hostruleShared := integrationtest.FakeHostRule{
		Name:      sharedHrName,
		Namespace: namespace,
		Fqdn:      sharedFQDN,
		FqdnType:  "Contains",
		ListenerPorts: []integrationtest.ListenerPorts{
			{
				Port:      8081,
				EnableSSL: false,
			},
			{
				Port:      6443,
				EnableSSL: true,
			},
		},
	}

	hrCreate := hostruleShared.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), sharedHrName, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	hostrule := integrationtest.FakeHostRule{
		Name:                hrName,
		Namespace:           namespace,
		Fqdn:                fqdn,
		ApplicationRootPath: "/foo",
	}
	hrCreate = hostrule.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
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

	g.Expect(nodes[0].EvhNodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0]).Should(gomega.Equal(appRootPath))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].Path).To(gomega.Equal("/"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].Protocol).To(gomega.Equal("HTTP"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(8081)))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].Path).To(gomega.Equal("/"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].Protocol).To(gomega.Equal("HTTPS"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].RedirectPort).To(gomega.Equal(int32(6443)))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].Path).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].RedirectPath).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].HppMap).To(gomega.BeNil())

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrName)
	integrationtest.TeardownHostRule(t, g, sniVSKey, sharedHrName)
	g.Eventually(func() bool {
		return nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts == nil
	}, 30*time.Second).Should(gomega.BeTrue())

	g.Expect(nodes[0].EvhNodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0]).To(gomega.Equal("/"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].Path).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].RedirectPath).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].HppMap).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestHostRuleRegexAppRootSecureListenerPorts(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	fqdn := "foo.com"
	namespace := "default"
	appRootPath := "/foo"

	hrName := objNameMap.GenerateName("samplehr-foo")
	sharedHrName := objNameMap.GenerateName("samplehr-shared")
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
		dnsNames:    []string{fqdn, fqdn},
		paths:       []string{"/something(/|$)(.*)", "/"},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	hostrule := integrationtest.FakeHostRule{
		Name:                hrName,
		Namespace:           namespace,
		Fqdn:                fqdn,
		ApplicationRootPath: "/foo",
		UseRegex:            true,
	}
	hrCreate := hostrule.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	sharedFQDN := "Shared-L7-EVH"
	hostruleShared := integrationtest.FakeHostRule{
		Name:      sharedHrName,
		Namespace: namespace,
		Fqdn:      sharedFQDN,
		FqdnType:  "Contains",
		ListenerPorts: []integrationtest.ListenerPorts{
			{
				Port:      8081,
				EnableSSL: false,
			},
			{
				Port:      6443,
				EnableSSL: true,
			},
		},
	}

	hrCreate = hostruleShared.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), sharedHrName, metav1.GetOptions{})
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

	g.Expect(nodes[0].EvhNodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0]).To(gomega.Equal("/something(/|$)(.*)"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].StringGroupRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].Path[0]).To(gomega.Equal(appRootPath))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].Path).To(gomega.Equal("/"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].Protocol).To(gomega.Equal("HTTP"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(8081)))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].Path).To(gomega.Equal("/"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].Protocol).To(gomega.Equal("HTTPS"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].RedirectPort).To(gomega.Equal(int32(6443)))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].Path).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].RedirectPath).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].HppMap).To(gomega.BeNil())

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrName)
	integrationtest.TeardownHostRule(t, g, sniVSKey, sharedHrName)
	g.Eventually(func() bool {
		return nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts == nil
	}, 30*time.Second).Should(gomega.BeTrue())

	g.Expect(nodes[0].EvhNodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0]).To(gomega.Equal("/something(/|$)(.*)"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].Path[0]).To(gomega.Equal("/"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].Path).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].RedirectPath).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[1].HppMap).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestHostRuleAppRootInsecureListenerPorts(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	fqdn := "foo.com"
	namespace := "default"
	appRootPath := "/foo"

	hrName := objNameMap.GenerateName("samplehr-foo")
	sharedHrName := objNameMap.GenerateName("samplehr-shared")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       false,
		withSecret:  false,
		serviceName: svcName,
		modelNames:  []string{modelName},
		dnsNames:    []string{fqdn},
		paths:       []string{"/"},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	sharedFQDN := "Shared-L7-EVH"
	hostruleShared := integrationtest.FakeHostRule{
		Name:      sharedHrName,
		Namespace: namespace,
		Fqdn:      sharedFQDN,
		FqdnType:  "Contains",
		ListenerPorts: []integrationtest.ListenerPorts{
			{
				Port:      8081,
				EnableSSL: false,
			},
			{
				Port:      6443,
				EnableSSL: true,
			},
		},
	}

	hrCreate := hostruleShared.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), sharedHrName, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	hostrule := integrationtest.FakeHostRule{
		Name:                hrName,
		Namespace:           namespace,
		Fqdn:                fqdn,
		ApplicationRootPath: "/foo",
	}
	hrCreate = hostrule.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
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

	g.Expect(nodes[0].EvhNodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0]).Should(gomega.Equal(appRootPath))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].Path).To(gomega.Equal("/"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].Protocol).To(gomega.Equal("HTTP"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(8081)))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].Path).To(gomega.Equal("/"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].Protocol).To(gomega.Equal("HTTPS"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].RedirectPort).To(gomega.Equal(int32(6443)))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPort).To(gomega.Equal("IS_IN"))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrName)
	integrationtest.TeardownHostRule(t, g, sniVSKey, sharedHrName)
	g.Eventually(func() bool {
		return nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts == nil
	}, 30*time.Second).Should(gomega.BeTrue())

	g.Expect(nodes[0].EvhNodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0]).To(gomega.Equal("/"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))

	TearDownIngressForCacheSyncCheck(t, "", ingressName, svcName, modelName)
}

func TestHostRuleRegexAppRootInsecureListenerPorts(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	fqdn := "foo.com"
	namespace := "default"
	appRootPath := "/foo"

	hrName := objNameMap.GenerateName("samplehr-foo")
	sharedHrName := objNameMap.GenerateName("samplehr-shared")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       false,
		withSecret:  false,
		serviceName: svcName,
		modelNames:  []string{modelName},
		dnsNames:    []string{fqdn, fqdn},
		paths:       []string{"/something(/|$)(.*)", "/"},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	hostrule := integrationtest.FakeHostRule{
		Name:                hrName,
		Namespace:           namespace,
		Fqdn:                fqdn,
		ApplicationRootPath: "/foo",
		UseRegex:            true,
	}
	hrCreate := hostrule.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	sharedFQDN := "Shared-L7-EVH"
	hostruleShared := integrationtest.FakeHostRule{
		Name:      sharedHrName,
		Namespace: namespace,
		Fqdn:      sharedFQDN,
		FqdnType:  "Contains",
		ListenerPorts: []integrationtest.ListenerPorts{
			{
				Port:      8081,
				EnableSSL: false,
			},
			{
				Port:      6443,
				EnableSSL: true,
			},
		},
	}

	hrCreate = hostruleShared.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), sharedHrName, metav1.GetOptions{})
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

	g.Expect(nodes[0].EvhNodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0]).To(gomega.Equal("/something(/|$)(.*)"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].StringGroupRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].Path[0]).To(gomega.Equal(appRootPath))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].Path).To(gomega.Equal("/"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].Protocol).To(gomega.Equal("HTTP"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(8081)))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].Path).To(gomega.Equal("/"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].Protocol).To(gomega.Equal("HTTPS"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].RedirectPort).To(gomega.Equal(int32(6443)))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPort).To(gomega.Equal("IS_IN"))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrName)
	integrationtest.TeardownHostRule(t, g, sniVSKey, sharedHrName)
	g.Eventually(func() bool {
		return nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts == nil
	}, 30*time.Second).Should(gomega.BeTrue())

	g.Expect(nodes[0].EvhNodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0]).To(gomega.Equal("/something(/|$)(.*)"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].Path[0]).To(gomega.Equal("/"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicyRefs[0].HppMap[1].StringGroupRefs).To(gomega.HaveLen(0))

	TearDownIngressForCacheSyncCheck(t, "", ingressName, svcName, modelName)
}

func TestApplyHostruleToParentVSWithEmptyDomains(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	integrationtest.SetEmptyDomainList()
	defer integrationtest.ResetMiddleware()
	modelName, _ := GetModelName("zoo.com", "default")
	hrName := objNameMap.GenerateName("samplehr-zoo")

	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("zoo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
		dnsNames:    []string{"zoo.com"},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	hostrule := integrationtest.FakeHostRule{
		Name:               hrName,
		Namespace:          "default",
		WafPolicy:          "thisisaviref-waf",
		ApplicationProfile: "thisisaviref-appprof",
		AnalyticsProfile:   "thisisaviref-analyticsprof",
		ErrorPageProfile:   "thisisaviref-errorprof",
		Datascripts:        []string{"thisisaviref-ds2", "thisisaviref-ds1"},
		HttpPolicySets:     []string{"thisisaviref-httpps2", "thisisaviref-httpps1"},
	}
	hrObj := hostrule.HostRule()
	hrObj.Spec.VirtualHost.Fqdn = "Shared-L7-EVH-"
	hrObj.Spec.VirtualHost.FqdnType = v1beta1.Contains

	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules("default").Create(context.TODO(), hrObj, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}

	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrName, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))

	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-EVH-0"}
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/hr-cluster--Shared-L7-EVH-0", true)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.BeTrue())
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(*nodes[0].Enabled).To(gomega.Equal(true))
	g.Expect(*nodes[0].WafPolicyRef).To(gomega.ContainSubstring("thisisaviref-waf"))
	g.Expect(*nodes[0].ApplicationProfileRef).To(gomega.ContainSubstring("thisisaviref-appprof"))
	g.Expect(*nodes[0].AnalyticsProfileRef).To(gomega.ContainSubstring("thisisaviref-analyticsprof"))
	g.Expect(nodes[0].ErrorPageProfileRef).To(gomega.ContainSubstring("thisisaviref-errorprof"))
	g.Expect(nodes[0].HttpPolicySetRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].HttpPolicySetRefs[0]).To(gomega.ContainSubstring("thisisaviref-httpps2"))
	g.Expect(nodes[0].HttpPolicySetRefs[1]).To(gomega.ContainSubstring("thisisaviref-httpps1"))
	g.Expect(nodes[0].VsDatascriptRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].VsDatascriptRefs[0]).To(gomega.ContainSubstring("thisisaviref-ds2"))
	g.Expect(nodes[0].VsDatascriptRefs[1]).To(gomega.ContainSubstring("thisisaviref-ds1"))

	integrationtest.TeardownHostRule(t, g, vsKey, hrName)
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/hr-cluster--Shared-L7-EVH-0", false)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].Enabled).To(gomega.BeNil())
	g.Expect(nodes[0].SslKeyAndCertificateRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].WafPolicyRef).To(gomega.BeNil())
	g.Expect(nodes[0].ApplicationProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].AnalyticsProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].ErrorPageProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].VsDatascriptRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].SslProfileRef).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)

	mcache := cache.SharedAviObjCache()
	cloudObj := &cache.AviCloudPropertyCache{Name: "Default-Cloud", VType: "mock"}
	subdomains := []string{"avi.internal", ".com"}
	cloudObj.NSIpamDNS = subdomains
	mcache.CloudKeyCache.AviCacheAdd("Default-Cloud", cloudObj)

}
