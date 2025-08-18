/*
 * Copyright 2025 VMware, Inc.
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

package multitenancy

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayfake "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned/fake"

	akogatewayapik8s "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/k8s"
	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	tests "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/gatewayapitests"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"
)

var ctrl *akogatewayapik8s.GatewayController

const (
	DEFAULT_NAMESPACE = "default"
)

func TestMain(m *testing.M) {
	tests.KubeClient = k8sfake.NewSimpleClientset()
	tests.GatewayClient = gatewayfake.NewSimpleClientset()
	integrationtest.KubeClient = tests.KubeClient
	testData := tests.GetL7RuleFakeData()
	tests.DynamicClient = dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), tests.GvrToKind, &testData)

	// Sets the environment variables
	os.Setenv("CLUSTER_NAME", "cluster")
	os.Setenv("CLOUD_NAME", "CLOUD_VCENTER")
	os.Setenv("SEG_NAME", "Default-Group")
	os.Setenv("POD_NAMESPACE", utils.AKO_DEFAULT_NS)
	os.Setenv("FULL_SYNC_INTERVAL", "300")
	os.Setenv("ENABLE_EVH", "true")
	os.Setenv("TENANT", "admin")
	os.Setenv("POD_NAME", "ako-0")

	// Set the user with prefix
	_ = lib.AKOControlConfig()
	lib.SetAKOUser(akogatewayapilib.Prefix)
	lib.SetNamePrefix(akogatewayapilib.Prefix)
	lib.AKOControlConfig().SetIsLeaderFlag(true)
	akoControlConfig := akogatewayapilib.AKOControlConfig()
	akoControlConfig.SetEventRecorder(lib.AKOGatewayEventComponent, tests.KubeClient, true)
	registeredInformers := []string{
		utils.ServiceInformer,
		utils.SecretInformer,
		utils.NSInformer,
	}
	registeredInformers = append(registeredInformers, utils.EndpointSlicesInformer)

	utils.AviLog.SetLevel("DEBUG")
	utils.NewInformers(utils.KubeClientIntf{ClientSet: tests.KubeClient}, registeredInformers, make(map[string]interface{}))
	data := map[string][]byte{
		"username": []byte("admin"),
		"password": []byte("admin"),
	}
	object := metav1.ObjectMeta{Name: "avi-secret", Namespace: utils.GetAKONamespace()}
	secret := &corev1.Secret{Data: data, ObjectMeta: object}
	tests.KubeClient.CoreV1().Secrets(utils.GetAKONamespace()).Create(context.TODO(), secret, metav1.CreateOptions{})

	akoApi := integrationtest.InitializeFakeAKOAPIServer()
	defer akoApi.ShutDown()

	tests.NewAviFakeClientInstance(tests.KubeClient, true)
	defer integrationtest.AviFakeClientInstance.Close()

	ctrl = akogatewayapik8s.SharedGatewayController()
	ctrl.DisableSync = false
	ctrl.InitGatewayAPIInformers(tests.GatewayClient)
	akoControlConfig.SetGatewayAPIClientset(tests.GatewayClient)

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

	integrationtest.AddConfigMap(tests.KubeClient)
	integrationtest.AddDefaultNamespace()
	go ctrl.InitController(k8s.K8sinformers{Cs: tests.KubeClient, DynamicClient: tests.DynamicClient}, registeredInformers, ctrlCh, stopCh, quickSyncCh, waitGroupMap)
	os.Exit(m.Run())
}

// Create a gw class, namespace and annotate namespace with non-admin tenant
// Create a simple gateway with 1 TLS and 1 non-tls listener
// check graph layer nodes tenant
func TestGatewayWithNonAdminTenant(t *testing.T) {
	gatewayName := "gateway-01"
	gatewayClassName := "gateway-class-01"
	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	integrationtest.AnnotateNamespaceWithTenant(t, DEFAULT_NAMESPACE, "nonadmin")

	ports := []int32{8080}
	listeners := tests.GetListenersV1(ports, false, false)
	ports = []int32{6443}
	secrets := []string{"secret-01"}
	for _, secret := range secrets {
		integrationtest.AddSecret(secret, DEFAULT_NAMESPACE, "cert", "key")
	}
	tlsListeners := tests.GetListenersV1(ports, false, false, secrets...)
	listeners = append(listeners, tlsListeners...)

	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)
	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	modelName := lib.GetModelName("nonadmin", akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName))

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Tenant).Should(gomega.Equal("nonadmin"))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(2))
	g.Expect(nodes[0].PortProto[1].EnableSSL).To(gomega.Equal(true))
	g.Expect(nodes[0].SSLKeyCertRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].SSLKeyCertRefs[0].Tenant).Should(gomega.Equal("nonadmin"))
	g.Expect(nodes[0].VSVIPRefs).To(gomega.HaveLen(1))
	g.Expect((nodes[0].VSVIPRefs[0]).Tenant).Should(gomega.Equal("nonadmin"))
	// default backend response
	g.Expect(nodes[0].HttpPolicyRefs[0].RequestRules[0].Match.Path.MatchStr[0]).To(gomega.Equal("/"))
	g.Expect(*nodes[0].HttpPolicyRefs[0].RequestRules[0].SwitchingAction.StatusCode).To(gomega.Equal("HTTP_LOCAL_RESPONSE_STATUS_CODE_404"))
	g.Expect(nodes[0].HttpPolicyRefs[0].Tenant).Should(gomega.Equal("nonadmin"))

	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
	integrationtest.DeleteSecret(secrets[0], DEFAULT_NAMESPACE)
	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, DEFAULT_NAMESPACE)
}

func TestGatewayAddTLSWithNonAdminTenant(t *testing.T) {
	gatewayName := "gateway-02"
	gatewayClassName := "gateway-class-02"
	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	integrationtest.AnnotateNamespaceWithTenant(t, DEFAULT_NAMESPACE, "nonadmin")

	ports := []int32{8080}
	listeners := tests.GetListenersV1(ports, false, false)
	ports = []int32{6443}
	secrets := []string{"secret-02"}

	tlsListeners := tests.GetListenersV1(ports, false, false, secrets...)
	listeners = append(listeners, tlsListeners...)

	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)
	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	modelName := lib.GetModelName("nonadmin", akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName))

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Tenant).Should(gomega.Equal("nonadmin"))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(1))
	g.Expect(nodes[0].SSLKeyCertRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].VSVIPRefs).To(gomega.HaveLen(1))
	g.Expect((nodes[0].VSVIPRefs[0]).Tenant).Should(gomega.Equal("nonadmin"))
	// default backend response
	g.Expect(nodes[0].HttpPolicyRefs[0].RequestRules[0].Match.Path.MatchStr[0]).To(gomega.Equal("/"))
	g.Expect(*nodes[0].HttpPolicyRefs[0].RequestRules[0].SwitchingAction.StatusCode).To(gomega.Equal("HTTP_LOCAL_RESPONSE_STATUS_CODE_404"))
	g.Expect(nodes[0].HttpPolicyRefs[0].Tenant).Should(gomega.Equal("nonadmin"))

	// Add secret for the tls listener
	for _, secret := range secrets {
		integrationtest.AddSecret(secret, DEFAULT_NAMESPACE, "cert", "key")
	}

	g.Eventually(func() bool {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].SSLKeyCertRefs) == 1 && nodes[0].SSLKeyCertRefs[0].Tenant == "nonadmin"
	}, 25*time.Second).Should(gomega.Equal(true))

	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
	integrationtest.DeleteSecret(secrets[0], DEFAULT_NAMESPACE)
	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, DEFAULT_NAMESPACE)
}

func TestGatewayAnnotateNamespaceWithNonAdminTenant(t *testing.T) {
	gatewayName := "gateway-03"
	gatewayClassName := "gateway-class-03"
	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)

	ports := []int32{8080}
	listeners := tests.GetListenersV1(ports, false, false)
	ports = []int32{6443}
	secrets := []string{"secret-03"}
	for _, secret := range secrets {
		integrationtest.AddSecret(secret, DEFAULT_NAMESPACE, "cert", "key")
	}
	tlsListeners := tests.GetListenersV1(ports, false, false, secrets...)
	listeners = append(listeners, tlsListeners...)

	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)
	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	modelName := lib.GetModelName(lib.GetTenant(), akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName))

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Tenant).Should(gomega.Equal("admin"))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(2))
	g.Expect(nodes[0].SSLKeyCertRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].SSLKeyCertRefs[0].Tenant).Should(gomega.Equal("admin"))
	g.Expect(nodes[0].VSVIPRefs).To(gomega.HaveLen(1))
	g.Expect((nodes[0].VSVIPRefs[0]).Tenant).Should(gomega.Equal("admin"))
	// default backend response
	g.Expect(nodes[0].HttpPolicyRefs[0].RequestRules[0].Match.Path.MatchStr[0]).To(gomega.Equal("/"))
	g.Expect(*nodes[0].HttpPolicyRefs[0].RequestRules[0].SwitchingAction.StatusCode).To(gomega.Equal("HTTP_LOCAL_RESPONSE_STATUS_CODE_404"))
	g.Expect(nodes[0].HttpPolicyRefs[0].Tenant).Should(gomega.Equal("admin"))

	integrationtest.AnnotateNamespaceWithTenant(t, DEFAULT_NAMESPACE, "nonadmin")

	g.Eventually(func() bool {
		_, model := objects.SharedAviGraphLister().Get(modelName)
		return model == nil
	}, 25*time.Second).Should(gomega.Equal(true))

	modelName = lib.GetModelName("nonadmin", akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName))
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Tenant).Should(gomega.Equal("nonadmin"))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(2))
	g.Expect(nodes[0].SSLKeyCertRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].SSLKeyCertRefs[0].Tenant).Should(gomega.Equal("nonadmin"))
	g.Expect(nodes[0].VSVIPRefs).To(gomega.HaveLen(1))
	g.Expect((nodes[0].VSVIPRefs[0]).Tenant).Should(gomega.Equal("nonadmin"))
	// default backend response
	g.Expect(nodes[0].HttpPolicyRefs[0].RequestRules[0].Match.Path.MatchStr[0]).To(gomega.Equal("/"))
	g.Expect(*nodes[0].HttpPolicyRefs[0].RequestRules[0].SwitchingAction.StatusCode).To(gomega.Equal("HTTP_LOCAL_RESPONSE_STATUS_CODE_404"))
	g.Expect(nodes[0].HttpPolicyRefs[0].Tenant).Should(gomega.Equal("nonadmin"))

	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
	integrationtest.DeleteSecret(secrets[0], DEFAULT_NAMESPACE)
	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, DEFAULT_NAMESPACE)
}

func TestGatewayDeannotateNamespaceWithNonAdminTenant(t *testing.T) {
	gatewayName := "gateway-04"
	gatewayClassName := "gateway-class-04"
	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	integrationtest.AnnotateNamespaceWithTenant(t, DEFAULT_NAMESPACE, "nonadmin")

	ports := []int32{8080}
	listeners := tests.GetListenersV1(ports, false, false)
	ports = []int32{6443}
	secrets := []string{"secret-04"}
	for _, secret := range secrets {
		integrationtest.AddSecret(secret, DEFAULT_NAMESPACE, "cert", "key")
	}
	tlsListeners := tests.GetListenersV1(ports, false, false, secrets...)
	listeners = append(listeners, tlsListeners...)

	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)
	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	modelName := lib.GetModelName("nonadmin", akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName))
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Tenant).Should(gomega.Equal("nonadmin"))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(2))
	g.Expect(nodes[0].SSLKeyCertRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].SSLKeyCertRefs[0].Tenant).Should(gomega.Equal("nonadmin"))
	g.Expect(nodes[0].VSVIPRefs).To(gomega.HaveLen(1))
	g.Expect((nodes[0].VSVIPRefs[0]).Tenant).Should(gomega.Equal("nonadmin"))
	// default backend response
	g.Expect(nodes[0].HttpPolicyRefs[0].RequestRules[0].Match.Path.MatchStr[0]).To(gomega.Equal("/"))
	g.Expect(*nodes[0].HttpPolicyRefs[0].RequestRules[0].SwitchingAction.StatusCode).To(gomega.Equal("HTTP_LOCAL_RESPONSE_STATUS_CODE_404"))
	g.Expect(nodes[0].HttpPolicyRefs[0].Tenant).Should(gomega.Equal("nonadmin"))

	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, DEFAULT_NAMESPACE)
	g.Eventually(func() bool {
		_, model := objects.SharedAviGraphLister().Get(modelName)
		return model == nil
	}, 25*time.Second).Should(gomega.Equal(true))

	modelName = lib.GetModelName(lib.GetTenant(), akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName))

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Tenant).Should(gomega.Equal("admin"))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(2))
	g.Expect(nodes[0].SSLKeyCertRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].SSLKeyCertRefs[0].Tenant).Should(gomega.Equal("admin"))
	g.Expect(nodes[0].VSVIPRefs).To(gomega.HaveLen(1))
	g.Expect((nodes[0].VSVIPRefs[0]).Tenant).Should(gomega.Equal("admin"))
	// default backend response
	g.Expect(nodes[0].HttpPolicyRefs[0].RequestRules[0].Match.Path.MatchStr[0]).To(gomega.Equal("/"))
	g.Expect(*nodes[0].HttpPolicyRefs[0].RequestRules[0].SwitchingAction.StatusCode).To(gomega.Equal("HTTP_LOCAL_RESPONSE_STATUS_CODE_404"))
	g.Expect(nodes[0].HttpPolicyRefs[0].Tenant).Should(gomega.Equal("admin"))

	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
	integrationtest.DeleteSecret(secrets[0], DEFAULT_NAMESPACE)
}

// Create a Gateway and VIP Type should be PUBLIC by default, add lb-vip-type annotation, and the VIP type field should get updated
func TestGatewayWithVipTypeAnnotation(t *testing.T) {
	gatewayName := "gateway-05"
	gatewayClassName := "gateway-class-05"
	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	integrationtest.AnnotateNamespaceWithTenant(t, DEFAULT_NAMESPACE, "nonadmin")

	ports := []int32{8080}
	listeners := tests.GetListenersV1(ports, false, false)
	ports = []int32{6443}
	secrets := []string{"secret-01"}
	for _, secret := range secrets {
		integrationtest.AddSecret(secret, DEFAULT_NAMESPACE, "cert", "key")
	}
	tlsListeners := tests.GetListenersV1(ports, false, false, secrets...)
	listeners = append(listeners, tlsListeners...)

	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)
	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	modelName := lib.GetModelName("nonadmin", akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName))

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Tenant).Should(gomega.Equal("nonadmin"))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(2))
	g.Expect(nodes[0].PortProto[1].EnableSSL).To(gomega.Equal(true))
	g.Expect(nodes[0].SSLKeyCertRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].SSLKeyCertRefs[0].Tenant).Should(gomega.Equal("nonadmin"))
	g.Expect(nodes[0].VSVIPRefs).To(gomega.HaveLen(1))
	g.Expect((nodes[0].VSVIPRefs[0]).Tenant).Should(gomega.Equal("nonadmin"))
	g.Expect(nodes[0].VSVIPRefs[0].LBVipType).Should(gomega.Equal("PUBLIC"))
	// default backend response
	g.Expect(nodes[0].HttpPolicyRefs[0].RequestRules[0].Match.Path.MatchStr[0]).To(gomega.Equal("/"))
	g.Expect(*nodes[0].HttpPolicyRefs[0].RequestRules[0].SwitchingAction.StatusCode).To(gomega.Equal("HTTP_LOCAL_RESPONSE_STATUS_CODE_404"))
	g.Expect(nodes[0].HttpPolicyRefs[0].Tenant).Should(gomega.Equal("nonadmin"))

	// Negative case of updating annotation to a non whitelisted value
	tests.UpdateGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners, "project")
	g.Consistently(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return nodes[0].VSVIPRefs[0].LBVipType != "PUBLIC"
	}, 25*time.Second).ShouldNot(gomega.Equal(true))

	tests.UpdateGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners, "private")

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return nodes[0].VSVIPRefs[0].LBVipType == "PRIVATE"
	}, 25*time.Second).Should(gomega.Equal(true))

	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
	integrationtest.DeleteSecret(secrets[0], DEFAULT_NAMESPACE)
	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, DEFAULT_NAMESPACE)
}
