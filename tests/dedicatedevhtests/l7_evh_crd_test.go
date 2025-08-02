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

package dedicatedevhtests

import (
	"context"
	"flag"
	"os"
	"sort"
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
	v1beta1crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1beta1/clientset/versioned/fake"

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
var v1beta1CRDClient *v1beta1crdfake.Clientset
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
	os.Setenv("POD_NAME", "ako-0")

	akoControlConfig := lib.AKOControlConfig()
	KubeClient = k8sfake.NewSimpleClientset()
	CRDClient = crdfake.NewSimpleClientset()
	v1beta1CRDClient = v1beta1crdfake.NewSimpleClientset()
	v1alpha2CRDClient = v1alpha2crdfake.NewSimpleClientset()
	akoControlConfig.Setv1beta1CRDClientset(v1beta1CRDClient)
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
		utils.IngressInformer,
		utils.IngressClassInformer,
		utils.SecretInformer,
		utils.NSInformer,
		utils.NodeInformer,
		utils.ConfigMapInformer,
	}

	registeredInformers = append(registeredInformers, utils.EndpointSlicesInformer)

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
	integrationtest.CreateEPS(t, "default", "avisvc", false, false, "1.1.1")
}

func TearDownTestForIngress(t *testing.T, modelNames ...string) {
	// for _, model := range modelNames {
	// 	objects.SharedAviGraphLister().Delete(model)
	// }
	integrationtest.DelSVC(t, "default", "avisvc")
	integrationtest.DelEPS(t, "default", "avisvc")
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

func SetUpIngressForCacheSyncCheckMultiPaths(t *testing.T, tlsIngress, withSecret bool, fqdns []string, paths []string, modelNames ...string) {
	SetupDomain()
	SetUpTestForIngress(t, modelNames...)
	ingressObject := integrationtest.FakeIngress{
		Name:        "app-root-test",
		Namespace:   "default",
		DnsNames:    fqdns,
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		Paths:       paths,
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

func TearDownIngressForCacheSyncCheckPath(t *testing.T, modelName string) {
	if err := KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "app-root-test", metav1.DeleteOptions{}); err != nil {
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
	}, 30*time.Second, 1*time.Second).Should(gomega.Equal("Accepted"))

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
		Name:                  hrname,
		Namespace:             "default",
		WafPolicy:             "thisisaviref-waf",
		ApplicationProfile:    "thisisaviref-appprof",
		AnalyticsProfile:      "thisisaviref-analyticsprof",
		ErrorPageProfile:      "thisisaviref-errorprof",
		Datascripts:           []string{"thisisaviref-ds2", "thisisaviref-ds1"},
		HttpPolicySets:        []string{"thisisaviref-httpps2", "thisisaviref-httpps1"},
		NetworkSecurityPolicy: "thisisaviref-networksecuritypolicyref",
	}
	hrObj := hostrule.HostRule()
	hrObj.Spec.VirtualHost.Fqdn = "foo.com"
	hrObj.Spec.VirtualHost.FqdnType = v1beta1.Contains
	hrObj.Spec.VirtualHost.TCPSettings = &v1beta1.HostRuleTCPSettings{
		Listeners: []v1beta1.HostRuleTCPListeners{
			{Port: 8081}, {Port: 8082, EnableSSL: true},
		},
	}

	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules("default").Create(context.TODO(), hrObj, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}

	g.Eventually(func() string {
		hostrule, _ := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
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
	if *isVipPerNS == "false" {
		// Not applicable in Vip per ns scenario
		g.Expect(*evhNode.NetworkSecurityPolicyRef).To(gomega.ContainSubstring("thisisaviref-networksecuritypolicyref"))
	}

	if *isVipPerNS != "true" {
		g.Expect(evhNode.PortProto).To(gomega.HaveLen(2))
		var portsWithHostRule []int
		for _, port := range nodes[0].PortProto {
			portsWithHostRule = append(portsWithHostRule, int(port.Port))
			if port.EnableSSL {
				g.Expect(int(port.Port)).Should(gomega.Equal(8082))
			}
		}
		sort.Ints(portsWithHostRule)
		g.Expect(portsWithHostRule[0]).To(gomega.Equal(8081))
		g.Expect(portsWithHostRule[1]).To(gomega.Equal(8082))
	}

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
	g.Expect(evhNode.SslKeyAndCertificateRefs).To(gomega.HaveLen(0))
	g.Expect(evhNode.WafPolicyRef).To(gomega.BeNil())
	g.Expect(evhNode.ApplicationProfileRef).To(gomega.BeNil())
	g.Expect(evhNode.AnalyticsProfileRef).To(gomega.BeNil())
	g.Expect(evhNode.ErrorPageProfileRef).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(evhNode.VsDatascriptRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(2))
	var portWithoutHostRule []int
	for _, port := range nodes[0].PortProto {
		portWithoutHostRule = append(portWithoutHostRule, int(port.Port))
		if port.EnableSSL {
			g.Expect(int(port.Port)).To(gomega.Equal(443))
		}
	}
	sort.Ints(portWithoutHostRule)
	g.Expect(portWithoutHostRule[0]).To(gomega.Equal(80))
	g.Expect(portWithoutHostRule[1]).To(gomega.Equal(443))

	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestApplyL7HostruleToDedicatedVS(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetDedicatedModel("foo.com", "default")
	hrname := "hr-cluster--foo.com-L7-dedicated"

	SetUpIngressForCacheSyncCheck(t, true, true, modelName)

	hostrule := integrationtest.FakeHostRule{
		Name:                  hrname,
		Namespace:             "default",
		WafPolicy:             "thisisaviref-waf",
		ApplicationProfile:    "thisisaviref-appprof",
		AnalyticsProfile:      "thisisaviref-analyticsprof",
		ErrorPageProfile:      "thisisaviref-errorprof",
		Datascripts:           []string{"thisisaviref-ds2", "thisisaviref-ds1"},
		HttpPolicySets:        []string{"thisisaviref-httpps2", "thisisaviref-httpps1"},
		NetworkSecurityPolicy: "thisisaviref-networksecuritypolicyref",
	}
	hrObj := hostrule.HostRule()
	hrObj.Spec.VirtualHost.Fqdn = "foo.com"
	hrObj.Spec.VirtualHost.FqdnType = v1beta1.Contains
	hrObj.Spec.VirtualHost.TCPSettings = &v1beta1.HostRuleTCPSettings{
		Listeners: []v1beta1.HostRuleTCPListeners{
			{Port: 8081}, {Port: 8082, EnableSSL: true},
		},
	}

	l7ruleName := "samplel7rule"
	integrationtest.SetupL7Rule(t, l7ruleName, g)

	hrObj.Spec.VirtualHost.L7Rule = l7ruleName

	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules("default").Create(context.TODO(), hrObj, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}

	g.Eventually(func() string {
		hostrule, _ := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
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
	if *isVipPerNS == "false" {
		// Not applicable in Vip per ns scenario
		g.Expect(*evhNode.NetworkSecurityPolicyRef).To(gomega.ContainSubstring("thisisaviref-networksecuritypolicyref"))
	}

	if *isVipPerNS != "true" {
		g.Expect(evhNode.PortProto).To(gomega.HaveLen(2))
		var portsWithHostRule []int
		for _, port := range nodes[0].PortProto {
			portsWithHostRule = append(portsWithHostRule, int(port.Port))
			if port.EnableSSL {
				g.Expect(int(port.Port)).Should(gomega.Equal(8082))
			}
		}
		sort.Ints(portsWithHostRule)
		g.Expect(portsWithHostRule[0]).To(gomega.Equal(8081))
		g.Expect(portsWithHostRule[1]).To(gomega.Equal(8082))
	}
	g.Expect(*evhNode.CloseClientConnOnConfigUpdate).To(gomega.Equal(true))
	g.Expect(*evhNode.AllowInvalidClientCert).To(gomega.Equal(true))
	g.Expect(*evhNode.IgnPoolNetReach).To(gomega.Equal(false))
	g.Expect(*evhNode.RemoveListeningPortOnVsDown).To(gomega.Equal(false))
	g.Expect(*evhNode.BotPolicyRef).To(gomega.ContainSubstring("thisisaviref-botpolicy"))
	g.Expect(*evhNode.SslSessCacheAvgSize).To(gomega.Equal(uint32(2024)))
	g.Expect(*evhNode.MinPoolsUp).To(gomega.Equal(uint32(0)))
	if *isVipPerNS == "false" {
		g.Expect(*evhNode.HostNameXlate).To(gomega.ContainSubstring("hostname.com"))
		g.Expect(*evhNode.SecurityPolicyRef).To(gomega.ContainSubstring("thisisaviref-secpolicy"))
	} else {
		g.Expect(evhNode.HostNameXlate).To(gomega.BeNil())
		g.Expect(evhNode.SecurityPolicyRef).To(gomega.BeNil())
	}

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
	g.Expect(evhNode.SslKeyAndCertificateRefs).To(gomega.HaveLen(0))
	g.Expect(evhNode.WafPolicyRef).To(gomega.BeNil())
	g.Expect(evhNode.ApplicationProfileRef).To(gomega.BeNil())
	g.Expect(evhNode.AnalyticsProfileRef).To(gomega.BeNil())
	g.Expect(evhNode.ErrorPageProfileRef).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(evhNode.VsDatascriptRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(2))
	var portWithoutHostRule []int
	for _, port := range nodes[0].PortProto {
		portWithoutHostRule = append(portWithoutHostRule, int(port.Port))
		if port.EnableSSL {
			g.Expect(int(port.Port)).To(gomega.Equal(443))
		}
	}
	sort.Ints(portWithoutHostRule)
	g.Expect(portWithoutHostRule[0]).To(gomega.Equal(80))
	g.Expect(portWithoutHostRule[1]).To(gomega.Equal(443))
	g.Expect(evhNode.CloseClientConnOnConfigUpdate).To(gomega.BeNil())
	g.Expect(evhNode.AllowInvalidClientCert).To(gomega.BeNil())
	g.Expect(evhNode.IgnPoolNetReach).To(gomega.BeNil())
	g.Expect(evhNode.RemoveListeningPortOnVsDown).To(gomega.BeNil())
	g.Expect(evhNode.BotPolicyRef).To(gomega.BeNil())
	g.Expect(evhNode.SslSessCacheAvgSize).To(gomega.BeNil())
	g.Expect(evhNode.MinPoolsUp).To(gomega.BeNil())
	g.Expect(evhNode.HostNameXlate).To(gomega.BeNil())
	g.Expect(evhNode.SecurityPolicyRef).To(gomega.BeNil())
	//Delete L7 Rule
	if err := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L7Rules("default").Delete(context.TODO(), l7ruleName, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("error in deleting l7Rule: %v", err)
	}
	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestHostruleSSLKeyCertToDedicatedVS(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetDedicatedModel("foo.com", "default")
	hrname := "hr-cluster--foo.com-L7-dedicated"

	SetUpIngressForCacheSyncCheck(t, true, true, modelName)

	integrationtest.SetupHostRule(t, hrname, "foo.com", true)

	g.Eventually(func() string {
		hostrule, _ := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
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
	g.Expect(evhNode.SslKeyAndCertificateRefs).To(gomega.HaveLen(1))
	g.Expect(evhNode.SslKeyAndCertificateRefs[0]).To(gomega.ContainSubstring("thisisaviref-sslkey"))

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
	g.Expect(evhNode.SslKeyAndCertificateRefs).To(gomega.HaveLen(0))

	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestHostRuleNoListenerDedicatedVS(t *testing.T) {
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
	hrObj.Spec.VirtualHost.FqdnType = v1beta1.Contains
	hrObj.Spec.VirtualHost.TCPSettings = &v1beta1.HostRuleTCPSettings{
		LoadBalancerIP: "80.80.80.80",
	}

	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules("default").Create(context.TODO(), hrObj, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}

	g.Eventually(func() string {
		hostrule, _ := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
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
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(2))
	var portsWithHostRule []int
	for _, port := range nodes[0].PortProto {
		portsWithHostRule = append(portsWithHostRule, int(port.Port))
		if port.EnableSSL {
			g.Expect(int(port.Port)).To(gomega.Equal(443))
		}
	}
	sort.Ints(portsWithHostRule)
	g.Expect(portsWithHostRule[0]).To(gomega.Equal(80))
	g.Expect(portsWithHostRule[1]).To(gomega.Equal(443))
	if *isVipPerNS != "true" {
		g.Expect(nodes[0].VSVIPRefs[0].IPAddress).To(gomega.Equal("80.80.80.80"))
	}

	integrationtest.TeardownHostRule(t, g, vsKey, hrname)
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/hr-cluster--foo.com-L7-dedicated", false)

	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestApplySSLHostRuleToInsecureDedicatedVS(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetDedicatedModel("foo.com", "default")
	hrname := "hr-cluster--foo.com-L7-dedicated"

	SetUpIngressForCacheSyncCheck(t, false, false, modelName)

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
	hostrule.SslKeyCertificate = "thisisaviref-sslkey"
	hostrule.SslProfile = "thisisaviref-sslprof"
	hrObj := hostrule.HostRule()
	hrObj.Spec.VirtualHost.Fqdn = "foo.com"
	hrObj.Spec.VirtualHost.FqdnType = v1beta1.Contains
	hrObj.Spec.VirtualHost.TCPSettings = &v1beta1.HostRuleTCPSettings{
		LoadBalancerIP: "80.80.80.80",
	}

	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules("default").Create(context.TODO(), hrObj, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}

	g.Eventually(func() string {
		hostrule, _ := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
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
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(2))
	var portsWithHostRule []int
	for _, port := range nodes[0].PortProto {
		portsWithHostRule = append(portsWithHostRule, int(port.Port))
		if port.EnableSSL {
			g.Expect(int(port.Port)).To(gomega.Equal(443))
		}
	}
	sort.Ints(portsWithHostRule)
	g.Expect(portsWithHostRule[0]).To(gomega.Equal(80))
	g.Expect(portsWithHostRule[1]).To(gomega.Equal(443))
	if *isVipPerNS != "true" {
		g.Expect(nodes[0].VSVIPRefs[0].IPAddress).To(gomega.Equal("80.80.80.80"))
	}

	integrationtest.TeardownHostRule(t, g, vsKey, hrname)
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/hr-cluster--foo.com-L7-dedicated", false)

	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestHostRuleUseRegexSecure(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetDedicatedModel("foo.com", "default")
	hrname := "hr-cluster--foo.com-L7-dedicated"
	fqdn := "foo.com"
	namespace := "default"

	SetUpIngressForCacheSyncCheck(t, true, true, modelName)

	hostrule := integrationtest.FakeHostRule{
		Name:      hrname,
		Namespace: namespace,
		Fqdn:      fqdn,
		UseRegex:  true,
	}
	hrCreate := hostrule.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

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

	g.Expect(evhNode.HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].Path).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].RedirectPath).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))
	g.Expect(evhNode.HttpPolicyRefs[1].HppMap).To(gomega.BeNil())

	integrationtest.TeardownHostRule(t, g, vsKey, hrname)
	time.Sleep(2 * time.Second)

	g.Expect(evhNode.HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].Path).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].RedirectPath).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))
	g.Expect(evhNode.HttpPolicyRefs[1].HppMap).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestHostRuleAppRootSecure(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetDedicatedModel("foo.com", "default")
	hrname := "hr-cluster--foo.com-L7-dedicated"
	fqdn := "foo.com"
	namespace := "default"
	appRootPath := "/foo"

	SetUpIngressForCacheSyncCheckMultiPaths(t, true, true, []string{fqdn}, []string{"/"}, modelName)

	hostrule := integrationtest.FakeHostRule{
		Name:                hrname,
		Namespace:           namespace,
		Fqdn:                fqdn,
		ApplicationRootPath: appRootPath,
	}
	hrCreate := hostrule.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

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

	g.Expect(evhNode.HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].Path[0]).Should(gomega.Equal(appRootPath))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(2))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].Path).To(gomega.Equal("/"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(80)))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].Protocol).To(gomega.Equal("HTTP"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].Path).To(gomega.Equal("/"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].Protocol).To(gomega.Equal("HTTPS"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].Path).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].RedirectPath).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))
	g.Expect(evhNode.HttpPolicyRefs[1].HppMap).To(gomega.BeNil())

	integrationtest.TeardownHostRule(t, g, vsKey, hrname)
	g.Eventually(func() bool {
		return evhNode.HttpPolicyRefs[0].RedirectPorts == nil
	}, 30*time.Second).Should(gomega.BeTrue())

	g.Expect(evhNode.HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].Path[0]).Should(gomega.Equal("/"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].Path).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].RedirectPath).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))
	g.Expect(evhNode.HttpPolicyRefs[1].HppMap).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheckPath(t, modelName)
}

func TestHostRuleRegexAppRootSecure(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetDedicatedModel("foo.com", "default")
	hrname := "hr-cluster--foo.com-L7-dedicated"
	fqdn := "foo.com"
	namespace := "default"
	appRootPath := "/foo"

	SetUpIngressForCacheSyncCheckMultiPaths(t, true, true, []string{fqdn, fqdn}, []string{"/something(/|$)(.*)", "/"}, modelName)

	hostrule := integrationtest.FakeHostRule{
		Name:                hrname,
		Namespace:           namespace,
		Fqdn:                fqdn,
		UseRegex:            true,
		ApplicationRootPath: appRootPath,
	}
	hrCreate := hostrule.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

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

	g.Expect(evhNode.HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(2))

	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].Path[0]).To(gomega.Equal("/something(/|$)(.*)"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[1].MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[1].MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[1].StringGroupRefs).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[1].Path[0]).To(gomega.Equal(appRootPath))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(2))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].Path).To(gomega.Equal("/"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(80)))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].Protocol).To(gomega.Equal("HTTP"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].Path).To(gomega.Equal("/"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].Protocol).To(gomega.Equal("HTTPS"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].Path).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].RedirectPath).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))
	g.Expect(evhNode.HttpPolicyRefs[1].HppMap).To(gomega.BeNil())

	integrationtest.TeardownHostRule(t, g, vsKey, hrname)
	g.Eventually(func() bool {
		return evhNode.HttpPolicyRefs[0].RedirectPorts == nil
	}, 30*time.Second).Should(gomega.BeTrue())

	g.Expect(evhNode.HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(2))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].Path[0]).To(gomega.Equal("/something(/|$)(.*)"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[1].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[1].Path[0]).To(gomega.Equal("/"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[1].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].Path).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].RedirectPath).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))
	g.Expect(evhNode.HttpPolicyRefs[1].HppMap).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheckPath(t, modelName)
}

func TestHostRuleUseRegexInsecure(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetDedicatedModel("foo.com", "default")
	hrname := "hr-cluster--foo.com-L7-dedicated"
	fqdn := "foo.com"
	namespace := "default"

	SetUpIngressForCacheSyncCheck(t, false, false, modelName)

	hostrule := integrationtest.FakeHostRule{
		Name:      hrname,
		Namespace: namespace,
		Fqdn:      fqdn,
		UseRegex:  true,
	}
	hrCreate := hostrule.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

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

	g.Expect(evhNode.HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(1))

	integrationtest.TeardownHostRule(t, g, vsKey, hrname)
	time.Sleep(2 * time.Second)

	g.Expect(evhNode.HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))

	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestHostRuleAppRootInsecure(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetDedicatedModel("foo.com", "default")
	hrname := "hr-cluster--foo.com-L7-dedicated"
	fqdn := "foo.com"
	namespace := "default"
	appRootPath := "/foo"

	SetUpIngressForCacheSyncCheckMultiPaths(t, false, false, []string{fqdn}, []string{"/"}, modelName)

	hostrule := integrationtest.FakeHostRule{
		Name:                hrname,
		Namespace:           namespace,
		Fqdn:                fqdn,
		ApplicationRootPath: appRootPath,
	}
	hrCreate := hostrule.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

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

	g.Expect(evhNode.HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].Path[0]).Should(gomega.Equal(appRootPath))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	if isVipPerNS != nil && *isVipPerNS == "true" {
		g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(2))
		g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].Path).To(gomega.Equal("/"))
		g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].RedirectPort).To(gomega.Equal(int32(443)))
		g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].Protocol).To(gomega.Equal("HTTPS"))
		g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
		g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
		g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	} else {
		g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(1))
	}
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].Path).To(gomega.Equal("/"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(80)))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].Protocol).To(gomega.Equal("HTTP"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPort).To(gomega.Equal("IS_IN"))

	integrationtest.TeardownHostRule(t, g, vsKey, hrname)
	g.Eventually(func() bool {
		return evhNode.HttpPolicyRefs[0].RedirectPorts == nil
	}, 30*time.Second).Should(gomega.BeTrue())

	g.Expect(evhNode.HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].Path[0]).Should(gomega.Equal("/"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))

	TearDownIngressForCacheSyncCheckPath(t, modelName)
}

func TestHostRuleRegexAppRootInsecure(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetDedicatedModel("foo.com", "default")
	hrname := "hr-cluster--foo.com-L7-dedicated"
	fqdn := "foo.com"
	namespace := "default"
	appRootPath := "/foo"

	SetUpIngressForCacheSyncCheckMultiPaths(t, false, false, []string{fqdn, fqdn}, []string{"/something(/|$)(.*)", "/"}, modelName)

	hostrule := integrationtest.FakeHostRule{
		Name:                hrname,
		Namespace:           namespace,
		Fqdn:                fqdn,
		UseRegex:            true,
		ApplicationRootPath: appRootPath,
	}
	hrCreate := hostrule.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

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

	g.Expect(evhNode.HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(2))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].Path[0]).To(gomega.Equal("/something(/|$)(.*)"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[1].MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[1].MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[1].StringGroupRefs).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[1].Path[0]).To(gomega.Equal(appRootPath))

	if isVipPerNS != nil && *isVipPerNS == "true" {
		g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(2))
		g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].Path).To(gomega.Equal("/"))
		g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].RedirectPort).To(gomega.Equal(int32(443)))
		g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].Protocol).To(gomega.Equal("HTTPS"))
		g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
		g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
		g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	} else {
		g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(1))
	}
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].Path).To(gomega.Equal("/"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(80)))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].Protocol).To(gomega.Equal("HTTP"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPort).To(gomega.Equal("IS_IN"))

	integrationtest.TeardownHostRule(t, g, vsKey, hrname)
	g.Eventually(func() bool {
		return evhNode.HttpPolicyRefs[0].RedirectPorts == nil
	}, 30*time.Second).Should(gomega.BeTrue())

	g.Expect(evhNode.HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(2))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].Path[0]).To(gomega.Equal("/something(/|$)(.*)"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[1].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[1].Path[0]).To(gomega.Equal("/"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[1].StringGroupRefs).To(gomega.HaveLen(0))

	TearDownIngressForCacheSyncCheckPath(t, modelName)
}

func TestHostRuleAppRootSecureListenerPorts(t *testing.T) {
	if lib.VIPPerNamespace() {
		t.Skip()
	}
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetDedicatedModel("foo.com", "default")
	hrname := "hr-cluster--foo.com-L7-dedicated"
	fqdn := "foo.com"
	namespace := "default"
	appRootPath := "/foo"

	SetUpIngressForCacheSyncCheckMultiPaths(t, true, true, []string{fqdn}, []string{"/"}, modelName)

	hostrule := integrationtest.FakeHostRule{
		Name:                hrname,
		Namespace:           namespace,
		Fqdn:                fqdn,
		ApplicationRootPath: appRootPath,
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
	hrCreate := hostrule.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

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

	g.Expect(evhNode.HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].Path[0]).Should(gomega.Equal(appRootPath))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(2))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].Path).To(gomega.Equal("/"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(8081)))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].Protocol).To(gomega.Equal("HTTP"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].Path).To(gomega.Equal("/"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].RedirectPort).To(gomega.Equal(int32(6443)))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].Protocol).To(gomega.Equal("HTTPS"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].Path).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].RedirectPath).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))
	g.Expect(evhNode.HttpPolicyRefs[1].HppMap).To(gomega.BeNil())

	integrationtest.TeardownHostRule(t, g, vsKey, hrname)
	g.Eventually(func() bool {
		return evhNode.HttpPolicyRefs[0].RedirectPorts == nil
	}, 30*time.Second).Should(gomega.BeTrue())

	g.Expect(evhNode.HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].Path[0]).Should(gomega.Equal("/"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].Path).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].RedirectPath).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))
	g.Expect(evhNode.HttpPolicyRefs[1].HppMap).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheckPath(t, modelName)
}

func TestHostRuleRegexAppRootSecureListenerPorts(t *testing.T) {
	if lib.VIPPerNamespace() {
		t.Skip()
	}
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetDedicatedModel("foo.com", "default")
	hrname := "hr-cluster--foo.com-L7-dedicated"
	fqdn := "foo.com"
	namespace := "default"
	appRootPath := "/foo"

	SetUpIngressForCacheSyncCheckMultiPaths(t, true, true, []string{fqdn, fqdn}, []string{"/something(/|$)(.*)", "/"}, modelName)

	hostrule := integrationtest.FakeHostRule{
		Name:                hrname,
		Namespace:           namespace,
		Fqdn:                fqdn,
		UseRegex:            true,
		ApplicationRootPath: appRootPath,
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
	hrCreate := hostrule.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

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

	g.Expect(evhNode.HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(2))

	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].Path[0]).To(gomega.Equal("/something(/|$)(.*)"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[1].MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[1].MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[1].StringGroupRefs).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[1].Path[0]).To(gomega.Equal(appRootPath))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(2))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].Path).To(gomega.Equal("/"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(8081)))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].Protocol).To(gomega.Equal("HTTP"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].Path).To(gomega.Equal("/"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].RedirectPort).To(gomega.Equal(int32(6443)))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].Protocol).To(gomega.Equal("HTTPS"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].Path).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].RedirectPath).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))
	g.Expect(evhNode.HttpPolicyRefs[1].HppMap).To(gomega.BeNil())

	integrationtest.TeardownHostRule(t, g, vsKey, hrname)
	g.Eventually(func() bool {
		return evhNode.HttpPolicyRefs[0].RedirectPorts == nil
	}, 30*time.Second).Should(gomega.BeTrue())

	g.Expect(evhNode.HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(2))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].Path[0]).To(gomega.Equal("/something(/|$)(.*)"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[1].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[1].Path[0]).To(gomega.Equal("/"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[1].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].Path).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].RedirectPath).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicyRefs[1].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))
	g.Expect(evhNode.HttpPolicyRefs[1].HppMap).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheckPath(t, modelName)
}

func TestHostRuleAppRootInsecureListenerPorts(t *testing.T) {
	if lib.VIPPerNamespace() {
		t.Skip()
	}
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetDedicatedModel("foo.com", "default")
	hrname := "hr-cluster--foo.com-L7-dedicated"
	fqdn := "foo.com"
	namespace := "default"
	appRootPath := "/foo"

	SetUpIngressForCacheSyncCheckMultiPaths(t, false, false, []string{fqdn}, []string{"/"}, modelName)

	hostrule := integrationtest.FakeHostRule{
		Name:                hrname,
		Namespace:           namespace,
		Fqdn:                fqdn,
		ApplicationRootPath: appRootPath,
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
	hrCreate := hostrule.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

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

	g.Expect(evhNode.HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].Path[0]).Should(gomega.Equal(appRootPath))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(2))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].Path).To(gomega.Equal("/"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(8081)))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].Protocol).To(gomega.Equal("HTTP"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].Path).To(gomega.Equal("/"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].RedirectPort).To(gomega.Equal(int32(6443)))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].Protocol).To(gomega.Equal("HTTPS"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPort).To(gomega.Equal("IS_IN"))

	integrationtest.TeardownHostRule(t, g, vsKey, hrname)
	g.Eventually(func() bool {
		return evhNode.HttpPolicyRefs[0].RedirectPorts == nil
	}, 30*time.Second).Should(gomega.BeTrue())

	g.Expect(evhNode.HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].Path[0]).Should(gomega.Equal("/"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))

	TearDownIngressForCacheSyncCheckPath(t, modelName)
}

func TestHostRuleRegexAppRootInsecureListenerPorts(t *testing.T) {
	if lib.VIPPerNamespace() {
		t.Skip()
	}
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetDedicatedModel("foo.com", "default")
	hrname := "hr-cluster--foo.com-L7-dedicated"
	fqdn := "foo.com"
	namespace := "default"
	appRootPath := "/foo"

	SetUpIngressForCacheSyncCheckMultiPaths(t, false, false, []string{fqdn, fqdn}, []string{"/something(/|$)(.*)", "/"}, modelName)

	hostrule := integrationtest.FakeHostRule{
		Name:                hrname,
		Namespace:           namespace,
		Fqdn:                fqdn,
		UseRegex:            true,
		ApplicationRootPath: appRootPath,
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
	hrCreate := hostrule.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

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

	g.Expect(evhNode.HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(2))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].Path[0]).To(gomega.Equal("/something(/|$)(.*)"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[1].MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[1].MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[1].StringGroupRefs).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[1].Path[0]).To(gomega.Equal(appRootPath))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(2))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].Path).To(gomega.Equal("/"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(8081)))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].Protocol).To(gomega.Equal("HTTP"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].Path).To(gomega.Equal("/"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].RedirectPort).To(gomega.Equal(int32(6443)))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].Protocol).To(gomega.Equal("HTTPS"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPort).To(gomega.Equal("IS_IN"))

	integrationtest.TeardownHostRule(t, g, vsKey, hrname)
	g.Eventually(func() bool {
		return evhNode.HttpPolicyRefs[0].RedirectPorts == nil
	}, 30*time.Second).Should(gomega.BeTrue())

	g.Expect(evhNode.HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(evhNode.HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(2))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].Path[0]).To(gomega.Equal("/something(/|$)(.*)"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[1].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[1].Path[0]).To(gomega.Equal("/"))
	g.Expect(evhNode.HttpPolicyRefs[0].HppMap[1].StringGroupRefs).To(gomega.HaveLen(0))

	TearDownIngressForCacheSyncCheckPath(t, modelName)
}

func TestHTTPRuleCreateDeleteWithEnableHTTP2ForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetDedicatedModel("foo.com", "default")
	rrName := "samplerr-foo"

	SetupDomain()
	secretName := "my-secret"
	ingressName := "foo-with-targets"
	svcName := "avisvc"

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

	var evhNode *avinodes.AviEvhVsNode
	if *isVipPerNS == "true" {
		evhNode = nodes[0].EvhNodes[0]
	} else {
		evhNode = nodes[0]
	}

	g.Expect(*evhNode.PoolRefs[0].EnableHttp2).Should(gomega.Equal(true))

	// delete httprule disables HTTP2
	integrationtest.TeardownHTTPRule(t, rrName)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolFooKey, "default/"+rrName+"/"+httpRulePath, false)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolBarKey, "default/"+rrName+"/"+httpRulePath, false)

	g.Expect(evhNode.PoolRefs[0].EnableHttp2).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestFQDNRestrictDedicatedSecureEVH(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	lib.AKOControlConfig().SetAKOFQDNReusePolicy("strict")
	modelName, _ := GetDedicatedModel("foo.com", "default")

	secretName := "my-secret"
	ingressName := "foo-with-targets"
	svcName := "avisvc"
	SetupDomain()
	SetUpTestForIngress(t, svcName, modelName)
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	integrationtest.PollForCompletion(t, modelName, 5)
	ingressObject := integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
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
	g.Expect(evhNode).To(gomega.Not(gomega.BeNil()))
	g.Expect(evhNode.PoolGroupRefs[0].AviMarkers.Host[0]).To(gomega.Equal("foo.com"))

	modelNameBar, _ := GetDedicatedModel("bar.com", "red")
	integrationtest.AddDefaultNamespace("red")
	integrationtest.CreateSVC(t, "red", "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, "red", "avisvc", false, false, "1.1.1")
	integrationtest.AddSecret(secretName, "red", "tlsCert", "tlsKey")
	integrationtest.PollForCompletion(t, modelNameBar, 5)

	ingressNameBar := "bar-with-targets"
	ingressObject2 := integrationtest.FakeIngress{
		Name:        ingressNameBar,
		Namespace:   "red",
		DnsNames:    []string{"bar.com"},
		Ips:         []string{"8.8.8.8"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
	}

	ingrFake2 := ingressObject2.Ingress(true)
	if _, err := KubeClient.NetworkingV1().Ingresses("red").Create(context.TODO(), ingrFake2, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelNameBar, 5)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelNameBar)
		return found
	}, 30*time.Second).Should(gomega.BeTrue())
	_, aviModel = objects.SharedAviGraphLister().Get(modelNameBar)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	if *isVipPerNS == "true" {
		evhNode = nodes[0].EvhNodes[0]
	} else {
		evhNode = nodes[0]
	}
	g.Expect(evhNode).To(gomega.Not(gomega.BeNil()))
	g.Expect(evhNode.PoolGroupRefs[0].AviMarkers.Host[0]).To(gomega.Equal("bar.com"))

	_, err := (integrationtest.FakeIngress{
		Name:      ingressNameBar,
		Namespace: "red",
		DnsNames:  []string{"foo.com"},
		Ips:       []string{"8.8.8.8"},
		Paths:     []string{"/foo", "/bar"},
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
		ServiceName: svcName,
	}).UpdateIngress()
	if err != nil {
		t.Fatalf("error in updating ingress %s", err)
	}
	integrationtest.PollForCompletion(t, modelNameBar, 5)
	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelNameBar)
		var isAviModelNil bool
		if *isVipPerNS == "true" {
			// TODO: Check if vipPerNS we need the VS deleted, currently it is not being deleted
			// the host is being deleted from NS
			isAviModelNil = len(aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()[0].EvhNodes) == 1
		} else {
			isAviModelNil = aviModel == nil
		}
		return found && isAviModelNil
	}, 30*time.Second).Should(gomega.BeTrue())

	if err := KubeClient.NetworkingV1().Ingresses("red").Delete(context.TODO(), "bar-with-targets", metav1.DeleteOptions{}); err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	if err := KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "foo-with-targets", metav1.DeleteOptions{}); err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	TearDownTestForIngress(t, modelName)
	TearDownTestForIngress(t, modelNameBar)
}
func TestApplyL7HostruleToDedicatedVSWithCommonProperties(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetDedicatedModel("foo.com", "default")
	hrname := "hr-cluster--foo.com-L7-dedicated"

	SetUpIngressForCacheSyncCheck(t, true, true, modelName)

	hostrule := integrationtest.FakeHostRule{
		Name:                  hrname,
		Namespace:             "default",
		WafPolicy:             "thisisaviref-waf",
		ErrorPageProfile:      "thisisaviref-errorprof",
		Datascripts:           []string{"thisisaviref-ds2", "thisisaviref-ds1"},
		HttpPolicySets:        []string{"thisisaviref-httpps2", "thisisaviref-httpps1"},
		NetworkSecurityPolicy: "thisisaviref-networksecuritypolicyref",
	}
	hrObj := hostrule.HostRule()
	hrObj.Spec.VirtualHost.Fqdn = "foo.com"
	hrObj.Spec.VirtualHost.FqdnType = v1beta1.Contains
	hrObj.Spec.VirtualHost.TCPSettings = &v1beta1.HostRuleTCPSettings{
		Listeners: []v1beta1.HostRuleTCPListeners{
			{Port: 8081}, {Port: 8082, EnableSSL: true},
		},
	}

	l7ruleName := "samplel7rule"
	integrationtest.SetupL7Rule(t, l7ruleName, g)

	hrObj.Spec.VirtualHost.L7Rule = l7ruleName

	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules("default").Create(context.TODO(), hrObj, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}

	g.Eventually(func() string {
		hostrule, _ := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
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
	g.Expect(*evhNode.ApplicationProfileRef).To(gomega.ContainSubstring("thisisaviref-appprofile-l7"))
	g.Expect(*evhNode.AnalyticsProfileRef).To(gomega.ContainSubstring("thisisaviref-analyticsprofile-l7"))
	g.Expect(evhNode.ErrorPageProfileRef).To(gomega.ContainSubstring("thisisaviref-errorprof"))
	g.Expect(evhNode.HttpPolicySetRefs).To(gomega.HaveLen(2))
	g.Expect(evhNode.HttpPolicySetRefs[0]).To(gomega.ContainSubstring("thisisaviref-httpps2"))
	g.Expect(evhNode.HttpPolicySetRefs[1]).To(gomega.ContainSubstring("thisisaviref-httpps1"))
	g.Expect(evhNode.VsDatascriptRefs).To(gomega.HaveLen(2))
	g.Expect(evhNode.VsDatascriptRefs[0]).To(gomega.ContainSubstring("thisisaviref-ds2"))
	g.Expect(evhNode.VsDatascriptRefs[1]).To(gomega.ContainSubstring("thisisaviref-ds1"))
	g.Expect(evhNode.AnalyticsPolicy).NotTo(gomega.BeNil())
	g.Expect(evhNode.AnalyticsPolicy.AllHeaders).NotTo(gomega.BeNil())
	g.Expect(*evhNode.AnalyticsPolicy.AllHeaders).To(gomega.Equal(false))
	g.Expect(evhNode.AnalyticsPolicy.FullClientLogs).NotTo(gomega.BeNil())
	g.Expect(evhNode.AnalyticsPolicy.FullClientLogs.Enabled).NotTo(gomega.BeNil())
	g.Expect(*evhNode.AnalyticsPolicy.FullClientLogs.Enabled).To(gomega.Equal(true))

	if *isVipPerNS == "false" {
		// Not applicable in Vip per ns scenario
		g.Expect(*evhNode.NetworkSecurityPolicyRef).To(gomega.ContainSubstring("thisisaviref-networksecuritypolicyref"))
	}

	if *isVipPerNS != "true" {
		g.Expect(evhNode.PortProto).To(gomega.HaveLen(2))
		var portsWithHostRule []int
		for _, port := range nodes[0].PortProto {
			portsWithHostRule = append(portsWithHostRule, int(port.Port))
			if port.EnableSSL {
				g.Expect(int(port.Port)).Should(gomega.Equal(8082))
			}
		}
		sort.Ints(portsWithHostRule)
		g.Expect(portsWithHostRule[0]).To(gomega.Equal(8081))
		g.Expect(portsWithHostRule[1]).To(gomega.Equal(8082))
	}
	g.Expect(*evhNode.CloseClientConnOnConfigUpdate).To(gomega.Equal(true))
	g.Expect(*evhNode.AllowInvalidClientCert).To(gomega.Equal(true))
	g.Expect(*evhNode.IgnPoolNetReach).To(gomega.Equal(false))
	g.Expect(*evhNode.RemoveListeningPortOnVsDown).To(gomega.Equal(false))
	g.Expect(*evhNode.BotPolicyRef).To(gomega.ContainSubstring("thisisaviref-botpolicy"))
	g.Expect(*evhNode.SslSessCacheAvgSize).To(gomega.Equal(uint32(2024)))
	g.Expect(*evhNode.MinPoolsUp).To(gomega.Equal(uint32(0)))
	if *isVipPerNS == "false" {
		g.Expect(*evhNode.HostNameXlate).To(gomega.ContainSubstring("hostname.com"))
		g.Expect(*evhNode.SecurityPolicyRef).To(gomega.ContainSubstring("thisisaviref-secpolicy"))
	} else {
		g.Expect(evhNode.HostNameXlate).To(gomega.BeNil())
		g.Expect(evhNode.SecurityPolicyRef).To(gomega.BeNil())
	}

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
	g.Expect(evhNode.SslKeyAndCertificateRefs).To(gomega.HaveLen(0))
	g.Expect(evhNode.WafPolicyRef).To(gomega.BeNil())
	g.Expect(evhNode.ApplicationProfileRef).To(gomega.BeNil())
	g.Expect(evhNode.AnalyticsProfileRef).To(gomega.BeNil())
	g.Expect(evhNode.ErrorPageProfileRef).To(gomega.Equal(""))
	g.Expect(evhNode.HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(evhNode.VsDatascriptRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(2))
	var portWithoutHostRule []int
	for _, port := range nodes[0].PortProto {
		portWithoutHostRule = append(portWithoutHostRule, int(port.Port))
		if port.EnableSSL {
			g.Expect(int(port.Port)).To(gomega.Equal(443))
		}
	}
	sort.Ints(portWithoutHostRule)
	g.Expect(portWithoutHostRule[0]).To(gomega.Equal(80))
	g.Expect(portWithoutHostRule[1]).To(gomega.Equal(443))
	g.Expect(evhNode.CloseClientConnOnConfigUpdate).To(gomega.BeNil())
	g.Expect(evhNode.AllowInvalidClientCert).To(gomega.BeNil())
	g.Expect(evhNode.IgnPoolNetReach).To(gomega.BeNil())
	g.Expect(evhNode.RemoveListeningPortOnVsDown).To(gomega.BeNil())
	g.Expect(evhNode.BotPolicyRef).To(gomega.BeNil())
	g.Expect(evhNode.SslSessCacheAvgSize).To(gomega.BeNil())
	g.Expect(evhNode.MinPoolsUp).To(gomega.BeNil())
	g.Expect(evhNode.HostNameXlate).To(gomega.BeNil())
	g.Expect(evhNode.SecurityPolicyRef).To(gomega.BeNil())
	g.Expect(evhNode.AnalyticsPolicy).To(gomega.BeNil())
	//Delete L7 Rule
	if err := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L7Rules("default").Delete(context.TODO(), l7ruleName, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("error in deleting l7Rule: %v", err)
	}
	TearDownIngressForCacheSyncCheck(t, modelName)
}
