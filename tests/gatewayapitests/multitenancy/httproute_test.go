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
	"testing"
	"time"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	akogatewayapitests "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/gatewayapitests"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"
)

func TestHTTPRouteWithNonAdminTenant(t *testing.T) {
	gatewayClassName := "gateway-class-01"
	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	integrationtest.AnnotateNamespaceWithTenant(t, DEFAULT_NAMESPACE, "nonadmin")

	gatewayName := "gateway-01"
	httpRouteName := "http-route-01"
	svcName1 := "avisvc-01"
	svcName2 := "avisvc-02"
	ports := []int32{8080, 8081}
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	modelName := lib.GetModelName("nonadmin", akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName))

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName1, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName2, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.3")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule1 := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName1, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rule2 := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/bar"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName2, DEFAULT_NAMESPACE, "8081", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{rule1, rule2}

	hostnames := []gatewayv1.Hostname{"foo-8080.com", "foo-8081.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 50*time.Second, 5*time.Second).Should(gomega.Equal(2))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	childNode1 := nodes[0].EvhNodes[0]
	g.Expect(childNode1.Tenant).Should(gomega.Equal("nonadmin"))
	g.Expect(childNode1.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolGroupRefs[0].Tenant).Should(gomega.Equal("nonadmin"))
	g.Expect(childNode1.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolRefs[0].Tenant).Should(gomega.Equal("nonadmin"))
	g.Expect(len(childNode1.VHMatches)).To(gomega.Equal(2))

	childNode2 := nodes[0].EvhNodes[1]
	g.Expect(childNode2.Tenant).Should(gomega.Equal("nonadmin"))
	g.Expect(childNode2.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolGroupRefs[0].Tenant).Should(gomega.Equal("nonadmin"))
	g.Expect(childNode2.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolRefs[0].Tenant).Should(gomega.Equal("nonadmin"))
	g.Expect(len(childNode2.VHMatches)).To(gomega.Equal(2))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName2)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, DEFAULT_NAMESPACE)
}

func TestHTTPRouteAnnotateNamespaceWithNonAdminTenant(t *testing.T) {
	gatewayClassName := "gateway-class-02"
	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)

	gatewayName := "gateway-02"
	httpRouteName := "http-route-02"
	svcName1 := "avisvc-03"
	svcName2 := "avisvc-04"
	ports := []int32{8080, 8081}
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	modelName := lib.GetModelName("admin", akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName))

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName1, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName2, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.3")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule1 := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName1, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rule2 := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/bar"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName2, DEFAULT_NAMESPACE, "8081", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{rule1, rule2}

	hostnames := []gatewayv1.Hostname{"foo-8080.com", "foo-8081.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 50*time.Second, 5*time.Second).Should(gomega.Equal(2))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	childNode1 := nodes[0].EvhNodes[0]
	g.Expect(childNode1.Tenant).Should(gomega.Equal("admin"))
	g.Expect(childNode1.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolGroupRefs[0].Tenant).Should(gomega.Equal("admin"))
	g.Expect(childNode1.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolRefs[0].Tenant).Should(gomega.Equal("admin"))
	g.Expect(len(childNode1.VHMatches)).To(gomega.Equal(2))

	childNode2 := nodes[0].EvhNodes[1]
	g.Expect(childNode2.Tenant).Should(gomega.Equal("admin"))
	g.Expect(childNode2.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolGroupRefs[0].Tenant).Should(gomega.Equal("admin"))
	g.Expect(childNode2.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolRefs[0].Tenant).Should(gomega.Equal("admin"))
	g.Expect(len(childNode2.VHMatches)).To(gomega.Equal(2))

	integrationtest.AnnotateNamespaceWithTenant(t, DEFAULT_NAMESPACE, "nonadmin")

	g.Eventually(func() bool {
		_, model := objects.SharedAviGraphLister().Get(modelName)
		return model == nil
	}, 25*time.Second).Should(gomega.Equal(true))

	modelName = lib.GetModelName("nonadmin", akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName))

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 50*time.Second, 5*time.Second).Should(gomega.Equal(2))

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	childNode1 = nodes[0].EvhNodes[0]
	g.Expect(childNode1.Tenant).Should(gomega.Equal("nonadmin"))
	g.Expect(childNode1.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolGroupRefs[0].Tenant).Should(gomega.Equal("nonadmin"))
	g.Expect(childNode1.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolRefs[0].Tenant).Should(gomega.Equal("nonadmin"))
	g.Expect(len(childNode1.VHMatches)).To(gomega.Equal(2))

	childNode2 = nodes[0].EvhNodes[1]
	g.Expect(childNode2.Tenant).Should(gomega.Equal("nonadmin"))
	g.Expect(childNode2.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolGroupRefs[0].Tenant).Should(gomega.Equal("nonadmin"))
	g.Expect(childNode2.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolRefs[0].Tenant).Should(gomega.Equal("nonadmin"))
	g.Expect(len(childNode2.VHMatches)).To(gomega.Equal(2))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName2)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, DEFAULT_NAMESPACE)
}

func TestHTTPRouteDeannotateNamespaceWithNonAdminTenant(t *testing.T) {
	gatewayClassName := "gateway-class-03"
	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	integrationtest.AnnotateNamespaceWithTenant(t, DEFAULT_NAMESPACE, "nonadmin")

	gatewayName := "gateway-03"
	httpRouteName := "http-route-03"
	svcName1 := "avisvc-05"
	svcName2 := "avisvc-06"
	ports := []int32{8080, 8081}
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	modelName := lib.GetModelName("nonadmin", akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName))

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName1, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName2, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.3")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule1 := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName1, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rule2 := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/bar"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName2, DEFAULT_NAMESPACE, "8081", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{rule1, rule2}

	hostnames := []gatewayv1.Hostname{"foo-8080.com", "foo-8081.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 50*time.Second, 5*time.Second).Should(gomega.Equal(2))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	childNode1 := nodes[0].EvhNodes[0]
	g.Expect(childNode1.Tenant).Should(gomega.Equal("nonadmin"))
	g.Expect(childNode1.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolGroupRefs[0].Tenant).Should(gomega.Equal("nonadmin"))
	g.Expect(childNode1.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolRefs[0].Tenant).Should(gomega.Equal("nonadmin"))
	g.Expect(len(childNode1.VHMatches)).To(gomega.Equal(2))

	childNode2 := nodes[0].EvhNodes[1]
	g.Expect(childNode2.Tenant).Should(gomega.Equal("nonadmin"))
	g.Expect(childNode2.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolGroupRefs[0].Tenant).Should(gomega.Equal("nonadmin"))
	g.Expect(childNode2.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolRefs[0].Tenant).Should(gomega.Equal("nonadmin"))
	g.Expect(len(childNode2.VHMatches)).To(gomega.Equal(2))

	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, DEFAULT_NAMESPACE)

	g.Eventually(func() bool {
		_, model := objects.SharedAviGraphLister().Get(modelName)
		return model == nil
	}, 25*time.Second).Should(gomega.Equal(true))

	modelName = lib.GetModelName("admin", akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName))

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 50*time.Second, 5*time.Second).Should(gomega.Equal(2))

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	childNode1 = nodes[0].EvhNodes[0]
	g.Expect(childNode1.Tenant).Should(gomega.Equal("admin"))
	g.Expect(childNode1.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolGroupRefs[0].Tenant).Should(gomega.Equal("admin"))
	g.Expect(childNode1.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolRefs[0].Tenant).Should(gomega.Equal("admin"))
	g.Expect(len(childNode1.VHMatches)).To(gomega.Equal(2))

	childNode2 = nodes[0].EvhNodes[1]
	g.Expect(childNode2.Tenant).Should(gomega.Equal("admin"))
	g.Expect(childNode2.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolGroupRefs[0].Tenant).Should(gomega.Equal("admin"))
	g.Expect(childNode2.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolRefs[0].Tenant).Should(gomega.Equal("admin"))
	g.Expect(len(childNode2.VHMatches)).To(gomega.Equal(2))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName2)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, DEFAULT_NAMESPACE)
}
