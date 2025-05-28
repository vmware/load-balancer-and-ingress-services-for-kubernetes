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

package crd

import (
	"testing"
	"time"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	tests "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/gatewayapitests"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"
)

func setupHTTPRoute(t *testing.T, svcName1, svcName2, gatewayName, httpRouteName string) {
	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName1, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName2, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.3")

	parentRefs := tests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, []int32{8080, 6443})
	rule1 := tests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName1, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rule2 := tests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/bar"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName2, DEFAULT_NAMESPACE, "8081", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{rule1, rule2}

	hostnames := []gatewayv1.Hostname{"foo-8080.com", "foo-6443.com"}
	tests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)
}

func validateHTTPRouteWithT1LR(g *gomega.GomegaWithT, tenant, gatewayName, t1lr string) {
	modelName := lib.GetModelName(tenant, akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName))
	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 50*time.Second, 5*time.Second).Should(gomega.Equal(2))

	validateChildNode := func(childNode *avinodes.AviEvhVsNode) {
		g.Expect(childNode.Tenant).Should(gomega.Equal("nonadmin"))
		g.Expect(childNode.PoolGroupRefs).To(gomega.HaveLen(1))
		g.Expect(childNode.PoolGroupRefs[0].Tenant).Should(gomega.Equal("nonadmin"))
		g.Expect(childNode.PoolRefs).To(gomega.HaveLen(1))
		g.Expect(childNode.PoolRefs[0].Tenant).Should(gomega.Equal("nonadmin"))
		g.Expect(len(childNode.VHMatches)).To(gomega.Equal(2))
		g.Expect(childNode.PoolRefs[0].T1Lr).Should(gomega.Equal(t1lr))
	}

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	validateChildNode(nodes[0].EvhNodes[0]) // childNode1
	validateChildNode(nodes[0].EvhNodes[1]) // childNode2
}

func TestHTTPRouteWithInfraSetting(t *testing.T) {
	gatewayName := "gateway-01"
	gatewayClassName := "gateway-class-01"
	infraSettingName := "infrasetting-01"
	httpRouteName := "http-route-01"
	svcName1 := "avisvc-01"
	svcName2 := "avisvc-02"

	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)

	// Create AviInfraSetting and set the status to accepted
	integrationtest.SetupAviInfraSetting(t, infraSettingName, "", true)
	integrationtest.AnnotateAKONamespaceWithInfraSetting(t, DEFAULT_NAMESPACE, infraSettingName)
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

	modelName := lib.GetModelName("nonadmin", akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName))

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	setupHTTPRoute(t, svcName1, svcName2, gatewayName, httpRouteName)

	validateHTTPRouteWithT1LR(g, "nonadmin", gatewayName, "avi-domain-c9:1234")

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName2)
	tests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
	integrationtest.DeleteSecret(secrets[0], DEFAULT_NAMESPACE)
	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, DEFAULT_NAMESPACE)
	integrationtest.TeardownAviInfraSetting(t, infraSettingName)
}

func TestHTTPRouteCreateInfraSetting(t *testing.T) {
	gatewayName := "gateway-02"
	gatewayClassName := "gateway-class-02"
	infraSettingName := "infrasetting-02"
	httpRouteName := "http-route-02"
	svcName1 := "avisvc-03"
	svcName2 := "avisvc-04"

	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	integrationtest.AnnotateAKONamespaceWithInfraSetting(t, DEFAULT_NAMESPACE, infraSettingName)
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

	modelName := lib.GetModelName("nonadmin", akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName))

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	setupHTTPRoute(t, svcName1, svcName2, gatewayName, httpRouteName)

	validateHTTPRouteWithT1LR(g, "nonadmin", gatewayName, "test-t1lr")

	integrationtest.SetupAviInfraSetting(t, infraSettingName, "", true)

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return nodes[0].EvhNodes[0].PoolRefs[0].T1Lr == "avi-domain-c9:1234" &&
			nodes[0].EvhNodes[1].PoolRefs[0].T1Lr == "avi-domain-c9:1234"
	}, 25*time.Second).Should(gomega.Equal(true))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName2)
	tests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
	integrationtest.DeleteSecret(secrets[0], DEFAULT_NAMESPACE)
	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, DEFAULT_NAMESPACE)
	integrationtest.TeardownAviInfraSetting(t, infraSettingName)
}

func TestHTTPRouteUpdateInfraSettingStatusToAccepted(t *testing.T) {
	gatewayName := "gateway-03"
	gatewayClassName := "gateway-class-03"
	infraSettingName := "infrasetting-03"
	httpRouteName := "http-route-03"
	svcName1 := "avisvc-05"
	svcName2 := "avisvc-06"

	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)

	// Create AviInfraSetting but do not set the status to accepted
	integrationtest.SetupAviInfraSetting(t, infraSettingName, "")
	integrationtest.AnnotateAKONamespaceWithInfraSetting(t, DEFAULT_NAMESPACE, infraSettingName)
	integrationtest.AnnotateNamespaceWithTenant(t, DEFAULT_NAMESPACE, "nonadmin")

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

	modelName := lib.GetModelName("nonadmin", akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName))

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	setupHTTPRoute(t, svcName1, svcName2, gatewayName, httpRouteName)

	validateHTTPRouteWithT1LR(g, "nonadmin", gatewayName, "test-t1lr")

	integrationtest.SetAviInfraSettingStatus(t, infraSettingName, lib.StatusAccepted)

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return nodes[0].EvhNodes[0].PoolRefs[0].T1Lr == "avi-domain-c9:1234" &&
			nodes[0].EvhNodes[1].PoolRefs[0].T1Lr == "avi-domain-c9:1234"
	}, 25*time.Second).Should(gomega.Equal(true))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName2)
	tests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
	integrationtest.DeleteSecret(secrets[0], DEFAULT_NAMESPACE)
	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, DEFAULT_NAMESPACE)
	integrationtest.TeardownAviInfraSetting(t, infraSettingName)
}

func TestHTTPRouteUpdateInfraSettingStatusToRejected(t *testing.T) {
	gatewayName := "gateway-04"
	gatewayClassName := "gateway-class-04"
	infraSettingName := "infrasetting-04"
	httpRouteName := "http-route-04"
	svcName1 := "avisvc-07"
	svcName2 := "avisvc-08"

	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)

	integrationtest.SetupAviInfraSetting(t, infraSettingName, "", true)
	integrationtest.AnnotateAKONamespaceWithInfraSetting(t, DEFAULT_NAMESPACE, infraSettingName)
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

	modelName := lib.GetModelName("nonadmin", akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName))

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	setupHTTPRoute(t, svcName1, svcName2, gatewayName, httpRouteName)

	validateHTTPRouteWithT1LR(g, "nonadmin", gatewayName, "avi-domain-c9:1234")

	integrationtest.SetAviInfraSettingStatus(t, infraSettingName, lib.StatusRejected)

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return nodes[0].EvhNodes[0].PoolRefs[0].T1Lr == "test-t1lr" &&
			nodes[0].EvhNodes[1].PoolRefs[0].T1Lr == "test-t1lr"
	}, 25*time.Second).Should(gomega.Equal(true))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName2)
	tests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
	integrationtest.DeleteSecret(secrets[0], DEFAULT_NAMESPACE)
	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, DEFAULT_NAMESPACE)
	integrationtest.TeardownAviInfraSetting(t, infraSettingName)
}

func TestHTTPRouteDeleteInfraSetting(t *testing.T) {
	gatewayName := "gateway-05"
	gatewayClassName := "gateway-class-05"
	infraSettingName := "infrasetting-05"
	httpRouteName := "http-route-05"
	svcName1 := "avisvc-09"
	svcName2 := "avisvc-10"

	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)

	integrationtest.SetupAviInfraSetting(t, infraSettingName, "", true)
	integrationtest.AnnotateAKONamespaceWithInfraSetting(t, DEFAULT_NAMESPACE, infraSettingName)
	integrationtest.AnnotateNamespaceWithTenant(t, DEFAULT_NAMESPACE, "nonadmin")

	ports := []int32{8080}
	listeners := tests.GetListenersV1(ports, false, false)
	ports = []int32{6443}
	secrets := []string{"secret-05"}
	for _, secret := range secrets {
		integrationtest.AddSecret(secret, DEFAULT_NAMESPACE, "cert", "key")
	}
	tlsListeners := tests.GetListenersV1(ports, false, false, secrets...)
	listeners = append(listeners, tlsListeners...)

	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)
	g := gomega.NewGomegaWithT(t)

	modelName := lib.GetModelName("nonadmin", akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName))

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	setupHTTPRoute(t, svcName1, svcName2, gatewayName, httpRouteName)

	validateHTTPRouteWithT1LR(g, "nonadmin", gatewayName, "avi-domain-c9:1234")

	integrationtest.TeardownAviInfraSetting(t, infraSettingName)

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return nodes[0].EvhNodes[0].PoolRefs[0].T1Lr == "test-t1lr" &&
			nodes[0].EvhNodes[1].PoolRefs[0].T1Lr == "test-t1lr"
	}, 25*time.Second).Should(gomega.Equal(true))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName2)
	tests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
	integrationtest.DeleteSecret(secrets[0], DEFAULT_NAMESPACE)
	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, DEFAULT_NAMESPACE)
}
