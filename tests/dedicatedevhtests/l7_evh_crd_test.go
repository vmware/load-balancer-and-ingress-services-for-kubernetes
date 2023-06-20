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

package dedicatedevhtests

import (
	"context"
	"flag"
	"os"
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
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha2"
	crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha1/clientset/versioned/fake"
	v1alpha2crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha2/clientset/versioned/fake"
	utils "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

var KubeClient *k8sfake.Clientset
var CRDClient *crdfake.Clientset
var v1alpha2CRDClient *v1alpha2crdfake.Clientset
var ctrl *k8s.AviController
var akoApiServer *api.FakeApiServer
var keyChan chan string

var isVipPerNS = flag.String("isVipPerNS", "false", "is vip per namespace enabled")

func setVipPerNS(vipPerNS string) {
	if vipPerNS == "true" {
		os.Setenv("VIP_PER_NAMESPACE", "true")
	}
}

func GetDedicatedModel(host, namespace string) (string, string) {
	vsName := "cluster--Shared-L7-EVH-"
	if !lib.VIPPerNamespace() {
		vsName = lib.Encode(lib.GetNamePrefix()+host+lib.DedicatedSuffix, lib.EVHVS) + lib.EVHSuffix
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
	os.Setenv("SHARD_VS_SIZE", "DEDICATED")
	os.Setenv("AUTO_L4_FQDN", "default")

	akoControlConfig := lib.AKOControlConfig()
	KubeClient = k8sfake.NewSimpleClientset()
	CRDClient = crdfake.NewSimpleClientset()
	v1alpha2CRDClient = v1alpha2crdfake.NewSimpleClientset()
	akoControlConfig.SetCRDClientset(CRDClient)
	akoControlConfig.Setv1alpha2CRDClientset(v1alpha2CRDClient)
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
	integrationtest.CreateSVC(t, "default", "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
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

func TestCreateUpdateDeleteSSORuleForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetDedicatedModel("foo.com", "default")
	srname := "samplesr-foo"
	SetUpIngressForCacheSyncCheck(t, true, true, modelName)

	err := integrationtest.SetUpOAuthSecret()
	if err != nil {
		t.Fatalf("error in creating my-oauth-secret: %v", err)
	}
	// Sleeping for 5s for secret to be updated in informer
	time.Sleep(5 * time.Second)

	integrationtest.SetupSSORule(t, srname, "foo.com", "OAuth")

	g.Eventually(func() string {
		ssoRule, _ := v1alpha2CRDClient.AkoV1alpha2().SSORules("default").Get(context.TODO(), srname, metav1.GetOptions{})
		return ssoRule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com"+lib.DedicatedSuffix, lib.EVHVS) + "-EVH"}
	integrationtest.VerifyMetadataSSORule(t, g, sniVSKey, "default/samplesr-foo", true)
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes, ok := aviModel.(*avinodes.AviObjectGraph)
		if !ok {
			return 0
		}
		return len(nodes.GetAviEvhVS())
	}, 20*time.Second).Should(gomega.Equal(1))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	var evhNode *avinodes.AviEvhVsNode
	if *isVipPerNS == "true" {
		evhNode = nodes[0].EvhNodes[0]
	} else {
		evhNode = nodes[0]
	}
	g.Expect(*evhNode.SsoPolicyRef).To(gomega.ContainSubstring("thisisaviref-ssopolicyoauth"))
	g.Expect(*evhNode.OauthVsConfig.CookieName).To(gomega.Equal("MY_OAUTH_COOKIE"))
	g.Expect(*evhNode.OauthVsConfig.CookieTimeout).To(gomega.Equal(int32(120)))
	g.Expect(*evhNode.OauthVsConfig.LogoutURI).To(gomega.Equal("https://auth.com/oauth/logout"))
	g.Expect(*evhNode.OauthVsConfig.RedirectURI).To(gomega.Equal("https://auth.com/oauth/redirect"))
	g.Expect(*evhNode.OauthVsConfig.PostLogoutRedirectURI).To(gomega.Equal("https://auth.com/oauth/post-logout-redirect"))
	g.Expect(evhNode.OauthVsConfig.OauthSettings).To(gomega.HaveLen(1))
	g.Expect(*evhNode.OauthVsConfig.OauthSettings[0].AppSettings.ClientID).To(gomega.Equal("my-client-id"))
	g.Expect(*evhNode.OauthVsConfig.OauthSettings[0].AppSettings.ClientSecret).To(gomega.Equal("my-client-secret"))
	g.Expect(evhNode.OauthVsConfig.OauthSettings[0].AppSettings.Scopes).To(gomega.HaveLen(1))
	g.Expect(evhNode.OauthVsConfig.OauthSettings[0].AppSettings.Scopes[0]).To(gomega.Equal("scope-1"))
	g.Expect(*evhNode.OauthVsConfig.OauthSettings[0].AppSettings.OidcConfig.OidcEnable).To(gomega.Equal(true))
	g.Expect(*evhNode.OauthVsConfig.OauthSettings[0].AppSettings.OidcConfig.Profile).To(gomega.Equal(true))
	g.Expect(*evhNode.OauthVsConfig.OauthSettings[0].AppSettings.OidcConfig.Userinfo).To(gomega.Equal(true))
	g.Expect(*evhNode.OauthVsConfig.OauthSettings[0].AuthProfileRef).To(gomega.ContainSubstring("thisisaviref-authprofileoauth"))
	g.Expect(*evhNode.OauthVsConfig.OauthSettings[0].ResourceServer.AccessType).To(gomega.Equal(lib.ACCESS_TOKEN_TYPE_OPAQUE))
	g.Expect(*evhNode.OauthVsConfig.OauthSettings[0].ResourceServer.IntrospectionDataTimeout).To(gomega.Equal(int32(60)))
	g.Expect(*evhNode.OauthVsConfig.OauthSettings[0].ResourceServer.OpaqueTokenParams.ServerID).To(gomega.Equal("my-server-id"))
	g.Expect(*evhNode.OauthVsConfig.OauthSettings[0].ResourceServer.OpaqueTokenParams.ServerSecret).To(gomega.Equal("my-server-secret"))
	g.Expect(evhNode.OauthVsConfig.OauthSettings[0].ResourceServer.JwtParams).To(gomega.BeNil())
	g.Expect(evhNode.SamlSpConfig).To(gomega.BeNil())

	//Update with Oidc parameters as false
	srUpdate := integrationtest.FakeSSORule{
		Name:      srname,
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
		ssoRule, _ := v1alpha2CRDClient.AkoV1alpha2().SSORules("default").Get(context.TODO(), srname, metav1.GetOptions{})
		return ssoRule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	// update is not getting reflected on evh nodes immediately. Hence adding a sleep of 5 seconds.
	time.Sleep(5 * time.Second)

	g.Expect(*evhNode.OauthVsConfig.OauthSettings[0].AppSettings.OidcConfig.OidcEnable).To(gomega.Equal(false))
	g.Expect(*evhNode.OauthVsConfig.OauthSettings[0].AppSettings.OidcConfig.Userinfo).To(gomega.Equal(false))
	g.Expect(*evhNode.OauthVsConfig.OauthSettings[0].AppSettings.OidcConfig.Profile).To(gomega.Equal(false))

	// Delete/Disable
	integrationtest.TeardownSSORule(t, g, sniVSKey, srname)

	g.Expect(evhNode.SsoPolicyRef).To(gomega.BeNil())
	g.Expect(evhNode.OauthVsConfig).To(gomega.BeNil())
	g.Expect(evhNode.SamlSpConfig).To(gomega.BeNil())

	err = integrationtest.TearDownOAuthSecret()
	if err != nil {
		t.Fatalf("error in deleting my-oauth-secret: %v", err)
	}
	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestCreateUpdateDeleteSSORuleForEvhInsecure(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetDedicatedModel("foo.com", "default")
	srname := "samplesr-foo"
	// create insecure ingress, SSORule should be applied in case of EVH
	SetUpIngressForCacheSyncCheck(t, false, false, modelName)

	err := integrationtest.SetUpOAuthSecret()
	if err != nil {
		t.Fatalf("error in creating my-oauth-secret: %v", err)
	}
	// Sleeping for 5s for secret to be updated in informer
	time.Sleep(5 * time.Second)

	integrationtest.SetupSSORule(t, srname, "foo.com", "SAML")

	g.Eventually(func() string {
		ssoRule, _ := v1alpha2CRDClient.AkoV1alpha2().SSORules("default").Get(context.TODO(), srname, metav1.GetOptions{})
		return ssoRule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com"+lib.DedicatedSuffix, lib.EVHVS) + "-EVH"}
	integrationtest.VerifyMetadataSSORule(t, g, sniVSKey, "default/samplesr-foo", true)
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes, ok := aviModel.(*avinodes.AviObjectGraph)
		if !ok {
			return 0
		}
		return len(nodes.GetAviEvhVS())
	}, 20*time.Second).Should(gomega.Equal(1))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	var evhNode *avinodes.AviEvhVsNode
	if *isVipPerNS == "true" {
		evhNode = nodes[0].EvhNodes[0]
	} else {
		evhNode = nodes[0]
	}

	g.Expect(evhNode.OauthVsConfig).To(gomega.BeNil())
	g.Expect(*evhNode.SsoPolicyRef).To(gomega.ContainSubstring("thisisaviref-ssopolicysaml"))
	g.Expect(evhNode.SamlSpConfig.AcsIndex).To(gomega.BeNil())
	g.Expect(*evhNode.SamlSpConfig.AuthnReqAcsType).To(gomega.Equal("SAML_AUTHN_REQ_ACS_TYPE_NONE"))
	g.Expect(*evhNode.SamlSpConfig.CookieName).To(gomega.Equal("MY_SAML_COOKIE"))
	g.Expect(*evhNode.SamlSpConfig.CookieTimeout).To(gomega.Equal(int32(120)))
	g.Expect(*evhNode.SamlSpConfig.EntityID).To(gomega.Equal("my-entityid"))
	g.Expect(*evhNode.SamlSpConfig.SigningSslKeyAndCertificateRef).To(gomega.ContainSubstring("thisisaviref-sslkeyandcertrefsaml"))
	g.Expect(*evhNode.SamlSpConfig.SingleSignonURL).To(gomega.Equal("https://auth.com/sso/acs/"))
	g.Expect(*evhNode.SamlSpConfig.UseIdpSessionTimeout).To(gomega.Equal(false))

	//Update with oauth parameters instead of saml
	srUpdate := integrationtest.FakeSSORule{
		Name:      srname,
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
		ssoRule, _ := v1alpha2CRDClient.AkoV1alpha2().SSORules("default").Get(context.TODO(), srname, metav1.GetOptions{})
		return ssoRule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	// update is not getting reflected on evh nodes immediately. Hence adding a sleep of 5 seconds.
	time.Sleep(5 * time.Second)

	g.Expect(evhNode.SamlSpConfig).To(gomega.BeNil())
	g.Expect(*evhNode.SsoPolicyRef).To(gomega.ContainSubstring("thisisaviref-ssopolicyoauth"))
	g.Expect(*evhNode.OauthVsConfig.CookieName).To(gomega.Equal("MY_OAUTH_COOKIE"))
	g.Expect(*evhNode.OauthVsConfig.CookieTimeout).To(gomega.Equal(int32(120)))
	g.Expect(*evhNode.OauthVsConfig.LogoutURI).To(gomega.Equal("https://auth.com/oauth/logout"))
	g.Expect(*evhNode.OauthVsConfig.RedirectURI).To(gomega.Equal("https://auth.com/oauth/redirect"))
	g.Expect(*evhNode.OauthVsConfig.PostLogoutRedirectURI).To(gomega.Equal("https://auth.com/oauth/post-logout-redirect"))
	g.Expect(evhNode.OauthVsConfig.OauthSettings).To(gomega.HaveLen(1))
	g.Expect(*evhNode.OauthVsConfig.OauthSettings[0].AppSettings.ClientID).To(gomega.Equal("my-client-id"))
	g.Expect(*evhNode.OauthVsConfig.OauthSettings[0].AppSettings.ClientSecret).To(gomega.Equal("my-client-secret"))
	g.Expect(evhNode.OauthVsConfig.OauthSettings[0].AppSettings.Scopes).To(gomega.HaveLen(1))
	g.Expect(evhNode.OauthVsConfig.OauthSettings[0].AppSettings.Scopes[0]).To(gomega.Equal("scope-1"))
	g.Expect(*evhNode.OauthVsConfig.OauthSettings[0].AppSettings.OidcConfig.OidcEnable).To(gomega.Equal(true))
	g.Expect(*evhNode.OauthVsConfig.OauthSettings[0].AppSettings.OidcConfig.Profile).To(gomega.Equal(true))
	g.Expect(*evhNode.OauthVsConfig.OauthSettings[0].AppSettings.OidcConfig.Userinfo).To(gomega.Equal(true))
	g.Expect(*evhNode.OauthVsConfig.OauthSettings[0].AuthProfileRef).To(gomega.ContainSubstring("thisisaviref-authprofileoauth"))
	g.Expect(*evhNode.OauthVsConfig.OauthSettings[0].ResourceServer.AccessType).To(gomega.Equal(lib.ACCESS_TOKEN_TYPE_OPAQUE))
	g.Expect(*evhNode.OauthVsConfig.OauthSettings[0].ResourceServer.IntrospectionDataTimeout).To(gomega.Equal(int32(60)))
	g.Expect(*evhNode.OauthVsConfig.OauthSettings[0].ResourceServer.OpaqueTokenParams.ServerID).To(gomega.Equal("my-server-id"))
	g.Expect(*evhNode.OauthVsConfig.OauthSettings[0].ResourceServer.OpaqueTokenParams.ServerSecret).To(gomega.Equal("my-server-secret"))
	g.Expect(evhNode.OauthVsConfig.OauthSettings[0].ResourceServer.JwtParams).To(gomega.BeNil())

	// Delete/Disable
	integrationtest.TeardownSSORule(t, g, sniVSKey, srname)
	g.Expect(evhNode.SsoPolicyRef).To(gomega.BeNil())
	g.Expect(evhNode.OauthVsConfig).To(gomega.BeNil())
	g.Expect(evhNode.SamlSpConfig).To(gomega.BeNil())

	err = integrationtest.TearDownOAuthSecret()
	if err != nil {
		t.Fatalf("error in deleting my-oauth-secret: %v", err)
	}
	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestApplyHostruleToDedicatedVS(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetDedicatedModel("foo.com", "default")
	hrname := "hr-cluster--foo.com-L7-dedicated"

	SetUpIngressForCacheSyncCheck(t, true, true, modelName)

	hostrule := integrationtest.FakeHostRule{
		Name:               hrname,
		Namespace:          "default",
		WafPolicy:          "thisisaviref-waf",
		ApplicationProfile: "thisisaviref-appprof",
		AnalyticsProfile:   "thisisaviref-analyticsprof",
		ErrorPageProfile:   "thisisaviref-errorprof",
		Datascripts:        []string{"thisisaviref-ds2", "thisisaviref-ds1"},
		HttpPolicySets:     []string{"thisisaviref-httpps2", "thisisaviref-httpps1"},
	}
	hrObj := hostrule.HostRule()
	hrObj.Spec.VirtualHost.Fqdn = "foo.com"
	hrObj.Spec.VirtualHost.FqdnType = v1alpha1.Contains

	if _, err := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().HostRules("default").Create(context.TODO(), hrObj, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}

	g.Eventually(func() string {
		hostrule, _ := CRDClient.AkoV1alpha1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))

	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--fc94484f7312a22cfb5bcc517b05649e256c24a8-EVH"}
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/hr-cluster--foo.com-L7-dedicated", true)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.BeTrue())
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	var evhNode *avinodes.AviEvhVsNode
	if *isVipPerNS == "true" {
		evhNode = nodes[0].EvhNodes[0]
	} else {
		evhNode = nodes[0]
	}
	g.Expect(*evhNode.Enabled).To(gomega.Equal(true))
	g.Expect(*evhNode.WafPolicyRef).To(gomega.ContainSubstring("thisisaviref-waf"))
	g.Expect(*evhNode.ApplicationProfileRef).To(gomega.ContainSubstring("thisisaviref-appprof"))
	g.Expect(*evhNode.AnalyticsProfileRef).To(gomega.ContainSubstring("thisisaviref-analyticsprof"))
	g.Expect(evhNode.ErrorPageProfileRef).To(gomega.ContainSubstring("thisisaviref-errorprof"))
	g.Expect(evhNode.HttpPolicySetRefs).To(gomega.HaveLen(2))
	g.Expect(evhNode.HttpPolicySetRefs[0]).To(gomega.ContainSubstring("thisisaviref-httpps2"))
	g.Expect(evhNode.HttpPolicySetRefs[1]).To(gomega.ContainSubstring("thisisaviref-httpps1"))
	g.Expect(evhNode.VsDatascriptRefs).To(gomega.HaveLen(2))
	g.Expect(evhNode.VsDatascriptRefs[0]).To(gomega.ContainSubstring("thisisaviref-ds2"))
	g.Expect(evhNode.VsDatascriptRefs[1]).To(gomega.ContainSubstring("thisisaviref-ds1"))

	integrationtest.TeardownHostRule(t, g, vsKey, hrname)
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/hr-cluster--foo.com-L7-dedicated", false)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	if *isVipPerNS == "true" {
		evhNode = nodes[0].EvhNodes[0]
	} else {
		evhNode = nodes[0]
	}
	g.Expect(evhNode.Enabled).To(gomega.BeNil())
	g.Expect(evhNode.SSLKeyCertAviRef).To(gomega.HaveLen(0))
	g.Expect(evhNode.WafPolicyRef).To(gomega.BeNil())
	g.Expect(evhNode.ApplicationProfileRef).To(gomega.BeNil())
	g.Expect(evhNode.AnalyticsProfileRef).To(gomega.BeNil())
	g.Expect(evhNode.ErrorPageProfileRef).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(evhNode.VsDatascriptRefs).To(gomega.HaveLen(0))

	TearDownIngressForCacheSyncCheck(t, modelName)
}
