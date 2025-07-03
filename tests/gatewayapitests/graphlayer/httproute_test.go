/*
 * Copyright 2023-2024 VMware, Inc.
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

package graphlayer

import (
	"context"
	"testing"
	"time"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	akogatewayapitests "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/gatewayapitests"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"
)

/* Test cases
 * - HTTPRoute CRUD
 * - HTTPRouteRule CRUD
 * - HTTPRouteFilter CRUD
 * - HTTPRouteFilter with Request Header Modifier
 * - HTTPRouteFilter with Response Header Modifier
 * - HTTPRouteFilter with Request Redirect
 * - HTTPRouteBackendRef CRUD (TODO)
 */
func TestHTTPRouteCRUD(t *testing.T) {

	gatewayName := "gateway-hr-01"
	gatewayClassName := "gateway-class-hr-01"
	httpRouteName := "http-route-hr-01"
	ports := []int32{8080}
	modelName, parentVSName := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))
	svcExample := (integrationtest.FakeService{
		Name:         "avisvc",
		Namespace:    "default",
		Type:         corev1.ServiceTypeClusterIP,
		ServicePorts: []integrationtest.Serviceport{{PortName: "foo", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()

	_, err := akogatewayapitests.KubeClient.CoreV1().Services("default").Create(context.TODO(), svcExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	integrationtest.CreateEPorEPS(t, "default", "avisvc", false, false, "1.1.1")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{"avisvc", "default", "8080", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{rule}
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)

		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(1))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	childNode := nodes[0].EvhNodes[0]
	g.Expect(childNode.VHParentName).To(gomega.Equal(parentVSName))
	g.Expect(*childNode.VHMatches[0].Host).To(gomega.Equal("foo-8080.com"))
	g.Expect(childNode.VHMatches[0].Rules[0].Matches.Path.MatchStr).To(gomega.ContainElement("/foo"))
	g.Expect(*childNode.VHMatches[0].Rules[0].Matches.Path.MatchCriteria).To(gomega.Equal("BEGINS_WITH"))

	rule = akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{"avisvc", "default", "8080", "1"}}, nil)
	rules = []gatewayv1.HTTPRouteRule{rule}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(1))

	childNode = nodes[0].EvhNodes[0]
	g.Expect(childNode.VHParentName).To(gomega.Equal(parentVSName))
	g.Expect(*childNode.VHMatches[0].Host).To(gomega.Equal("foo-8080.com"))
	g.Expect(childNode.VHMatches[0].Rules[0].Matches.Path.MatchStr).To(gomega.ContainElement("/foo"))
	g.Expect(*childNode.VHMatches[0].Rules[0].Matches.Path.MatchCriteria).To(gomega.Equal("BEGINS_WITH"))

	// delete httproute
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)

	// verifies the child deletion
	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return -1
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(0))

	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteRuleCRUD(t *testing.T) {

	gatewayName := "gateway-hr-02"
	gatewayClassName := "gateway-class-hr-02"
	httpRouteName := "http-route-hr-02"
	ports := []int32{8080}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	ruleWithoutCanary := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{"avisvc", "default", "8080", "1"}}, nil)
	ruleWithCanary := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/bar"}, []string{"canary"},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{"avisvc", "default", "8080", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{ruleWithCanary, ruleWithoutCanary}
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(2))

	// update httproute
	rules = []gatewayv1.HTTPRouteRule{ruleWithCanary}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)
	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(1))

	// update httproute
	rules = []gatewayv1.HTTPRouteRule{ruleWithCanary, ruleWithoutCanary}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)
	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(2))

	// delete httproute
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteFilterCRUD(t *testing.T) {

	gatewayName := "gateway-hr-03"
	gatewayClassName := "gateway-class-hr-03"
	httpRouteName := "http-route-hr-03"
	ports := []int32{8080}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{"avisvc", "default", "8080", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{rule}
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(1))

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		childVS := nodes[0].EvhNodes[0]
		if len(childVS.HttpPolicyRefs) != 1 {
			return -1
		}
		return len(childVS.HttpPolicyRefs[0].RequestRules)
	}, 25*time.Second).Should(gomega.Equal(1))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childVS := nodes[0].EvhNodes[0]

	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].Name).To(gomega.Equal(childVS.Name))
	g.Expect(childVS.HttpPolicyRefs[0].RequestRules[0].HdrAction).To(gomega.HaveLen(1))
	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].HdrAction[0].Action).To(gomega.Equal("HTTP_ADD_HDR"))

	// update httproute
	rule = akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"replace"}},
		[][]string{{"avisvc", "default", "8080", "1"}}, nil)
	rules = []gatewayv1.HTTPRouteRule{rule}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() string {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		childVS := nodes[0].EvhNodes[0]
		if childVS.HttpPolicyRefs[0].RequestRules[0].HdrAction[0].Action == nil {
			return ""
		}
		return *childVS.HttpPolicyRefs[0].RequestRules[0].HdrAction[0].Action
	}, 25*time.Second).Should(gomega.Equal("HTTP_REPLACE_HDR"))

	// delete httproute
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteFilterWithRequestHeaderModifier(t *testing.T) {

	gatewayName := "gateway-hr-04"
	gatewayClassName := "gateway-class-hr-04"
	httpRouteName := "http-route-hr-04"
	ports := []int32{8080}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "replace", "remove"}},
		[][]string{{"avisvc", "default", "8080", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{rule}
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(1))

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		childVS := nodes[0].EvhNodes[0]
		if len(childVS.HttpPolicyRefs) != 1 {
			return -1
		}
		return len(childVS.HttpPolicyRefs[0].RequestRules)
	}, 25*time.Second).Should(gomega.Equal(1))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childVS := nodes[0].EvhNodes[0]

	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].Name).To(gomega.Equal(childVS.Name))
	g.Expect(childVS.HttpPolicyRefs[0].RequestRules[0].HdrAction).To(gomega.HaveLen(3))
	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].HdrAction[0].Action).To(gomega.Equal("HTTP_ADD_HDR"))
	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].HdrAction[1].Action).To(gomega.Equal("HTTP_REPLACE_HDR"))
	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].HdrAction[2].Action).To(gomega.Equal("HTTP_REMOVE_HDR"))

	// update httproute
	rule = akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"replace", "remove"}},
		[][]string{{"avisvc", "default", "8080", "1"}}, nil)
	rules = []gatewayv1.HTTPRouteRule{rule}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		childVS := nodes[0].EvhNodes[0]
		if childVS.HttpPolicyRefs[0].RequestRules[0].HdrAction[0].Action == nil ||
			childVS.HttpPolicyRefs[0].RequestRules[0].HdrAction[1].Action == nil {
			return false
		}
		return *childVS.HttpPolicyRefs[0].RequestRules[0].HdrAction[0].Action == "HTTP_REPLACE_HDR" &&
			*childVS.HttpPolicyRefs[0].RequestRules[0].HdrAction[1].Action == "HTTP_REMOVE_HDR"
	}, 25*time.Second).Should(gomega.BeTrue())

	// delete httproute
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteFilterWithResponseHeaderModifier(t *testing.T) {

	gatewayName := "gateway-hr-05"
	gatewayClassName := "gateway-class-hr-05"
	httpRouteName := "http-route-hr-05"
	ports := []int32{8080}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"ResponseHeaderModifier": {"add", "replace", "remove"}},
		[][]string{{"avisvc", "default", "8080", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{rule}
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(1))

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		childVS := nodes[0].EvhNodes[0]
		if len(childVS.HttpPolicyRefs) != 1 {
			return -1
		}
		return len(childVS.HttpPolicyRefs[0].ResponseRules)
	}, 25*time.Second).Should(gomega.Equal(1))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childVS := nodes[0].EvhNodes[0]

	g.Expect(*childVS.HttpPolicyRefs[0].ResponseRules[0].Name).To(gomega.Equal(childVS.Name))
	g.Expect(childVS.HttpPolicyRefs[0].ResponseRules[0].HdrAction).To(gomega.HaveLen(3))
	g.Expect(*childVS.HttpPolicyRefs[0].ResponseRules[0].HdrAction[0].Action).To(gomega.Equal("HTTP_ADD_HDR"))
	g.Expect(*childVS.HttpPolicyRefs[0].ResponseRules[0].HdrAction[1].Action).To(gomega.Equal("HTTP_REPLACE_HDR"))
	g.Expect(*childVS.HttpPolicyRefs[0].ResponseRules[0].HdrAction[2].Action).To(gomega.Equal("HTTP_REMOVE_HDR"))

	// update httproute
	rule = akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"ResponseHeaderModifier": {"replace", "remove"}},
		[][]string{{"avisvc", "default", "8080", "1"}}, nil)
	rules = []gatewayv1.HTTPRouteRule{rule}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		childVS := nodes[0].EvhNodes[0]
		if childVS.HttpPolicyRefs[0].ResponseRules[0].HdrAction[0].Action == nil ||
			childVS.HttpPolicyRefs[0].ResponseRules[0].HdrAction[1].Action == nil {
			return false
		}
		return *childVS.HttpPolicyRefs[0].ResponseRules[0].HdrAction[0].Action == "HTTP_REPLACE_HDR" &&
			*childVS.HttpPolicyRefs[0].ResponseRules[0].HdrAction[1].Action == "HTTP_REMOVE_HDR"
	}, 25*time.Second).Should(gomega.BeTrue())

	// delete httproute
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteWithSingleHealthMonitor(t *testing.T) {
	gatewayName := "gateway-hm-01"
	gatewayClassName := "gateway-class-hm-01"
	httpRouteName := "http-route-hm-01"
	svcName := "avisvc-hm-01"
	healthMonitorName := "hm-01"
	ports := []int32{8080}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	// Create HealthMonitor CRD
	akogatewayapitests.CreateHealthMonitorCRD(t, healthMonitorName, DEFAULT_NAMESPACE, "thisisaviref-hm1")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleWithHealthMonitorFilters(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, []string{healthMonitorName})

	rules := []gatewayv1.HTTPRouteRule{rule}
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(1))

	// Verify HealthMonitor is present in graph layer
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childNode := nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolRefs[0].HealthMonitorRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolRefs[0].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm1"))

	// Delete HealthMonitor and verify it's removed from graph layer
	akogatewayapitests.DeleteHealthMonitorCRD(t, healthMonitorName, DEFAULT_NAMESPACE)

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes) > 0 && len(nodes[0].EvhNodes[0].PoolRefs) > 0 {
			return len(nodes[0].EvhNodes[0].PoolRefs[0].HealthMonitorRefs)
		}
		return 0
	}, 25*time.Second).Should(gomega.Equal(0))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteWithMultipleHealthMonitors(t *testing.T) {
	gatewayName := "gateway-hm-02"
	gatewayClassName := "gateway-class-hm-02"
	httpRouteName := "http-route-hm-02"
	svcName := "avisvc-hm-02"
	healthMonitorName1 := "hm-02a"
	healthMonitorName2 := "hm-02b"
	ports := []int32{8080}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports, true, false)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	// Create multiple HealthMonitor CRDs
	akogatewayapitests.CreateHealthMonitorCRD(t, healthMonitorName1, DEFAULT_NAMESPACE, "thisisaviref-hm1")
	akogatewayapitests.CreateHealthMonitorCRD(t, healthMonitorName2, DEFAULT_NAMESPACE, "thisisaviref-hm2")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleWithHealthMonitorFilters(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, []string{healthMonitorName1, healthMonitorName2})

	rules := []gatewayv1.HTTPRouteRule{rule}
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes) > 0 && len(nodes[0].EvhNodes[0].PoolRefs) > 0 {
			return len(nodes[0].EvhNodes[0].PoolRefs)
		}
		return 0
	}, 25*time.Second).Should(gomega.Equal(1))

	// Verify both HealthMonitors are present in graph layer
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childNode := nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolRefs[0].HealthMonitorRefs).To(gomega.HaveLen(2))
	g.Expect(childNode.PoolRefs[0].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm1"))
	g.Expect(childNode.PoolRefs[0].HealthMonitorRefs[1]).To(gomega.ContainSubstring("thisisaviref-hm2"))

	// Delete one HealthMonitor and update HTTPRoute to remove its reference
	akogatewayapitests.DeleteHealthMonitorCRD(t, healthMonitorName1, DEFAULT_NAMESPACE)

	// this will make the HTTPRoute invalid and remove all poolrefs
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes[0].PoolRefs)
	}, 25*time.Second).Should(gomega.Equal(0))

	akogatewayapitests.DeleteHealthMonitorCRD(t, healthMonitorName2, DEFAULT_NAMESPACE)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteWithHealthMonitorCRUD(t *testing.T) {
	gatewayName := "gateway-hm-03"
	gatewayClassName := "gateway-class-hm-03"
	httpRouteName := "http-route-hm-03"
	svcName := "avisvc-hm-03"
	healthMonitorName := "hm-03"
	ports := []int32{8080}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	// Create HTTPRoute without HealthMonitor initially
	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{rule}
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(1))

	// Verify no HealthMonitor initially
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childNode := nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolRefs[0].HealthMonitorRefs).To(gomega.HaveLen(0))

	// Create HealthMonitor and update HTTPRoute
	akogatewayapitests.CreateHealthMonitorCRD(t, healthMonitorName, DEFAULT_NAMESPACE, "thisisaviref-hm1")
	rule = akogatewayapitests.GetHTTPRouteRuleWithHealthMonitorFilters(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, []string{healthMonitorName})
	rules = []gatewayv1.HTTPRouteRule{rule}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes) > 0 && len(nodes[0].EvhNodes[0].PoolRefs) > 0 {
			return len(nodes[0].EvhNodes[0].PoolRefs[0].HealthMonitorRefs)
		}
		return 0
	}, 25*time.Second).Should(gomega.Equal(1))

	// Verify HealthMonitor is now present
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childNode = nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolRefs[0].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm1"))

	// Remove HealthMonitor from HTTPRoute
	rule = akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rules = []gatewayv1.HTTPRouteRule{rule}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes) > 0 && len(nodes[0].EvhNodes[0].PoolRefs) > 0 {
			return len(nodes[0].EvhNodes[0].PoolRefs[0].HealthMonitorRefs)
		}
		return 0
	}, 25*time.Second).Should(gomega.Equal(0))

	akogatewayapitests.DeleteHealthMonitorCRD(t, healthMonitorName, DEFAULT_NAMESPACE)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteWithInvalidHealthMonitor(t *testing.T) {
	gatewayName := "gateway-hm-04"
	gatewayClassName := "gateway-class-hm-04"
	httpRouteName := "http-route-hm-04"
	svcName := "avisvc-hm-04"
	healthMonitorName := "hm-04"
	ports := []int32{8080}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	// Create HealthMonitor CRD with Ready=False
	akogatewayapitests.CreateHealthMonitorCRDWithStatus(t, healthMonitorName, DEFAULT_NAMESPACE, "", false, "ValidationError", "HealthMonitor configuration is invalid")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleWithHealthMonitorFilters(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, []string{healthMonitorName})
	rules := []gatewayv1.HTTPRouteRule{rule}
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(1))

	// Verify that route is rejected
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childNode := nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolRefs).To(gomega.HaveLen(0))

	akogatewayapitests.DeleteHealthMonitorCRD(t, healthMonitorName, DEFAULT_NAMESPACE)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteWithHealthMonitorMultipleBackends(t *testing.T) {
	gatewayName := "gateway-hm-05"
	gatewayClassName := "gateway-class-hm-05"
	httpRouteName := "http-route-hm-05"
	svcName1 := "avisvc-hm-05a"
	svcName2 := "avisvc-hm-05b"
	healthMonitorName1 := "hm-05a"
	healthMonitorName2 := "hm-05b"
	ports := []int32{8080}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName1, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")
	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName2, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.4")

	// Create HealthMonitor CRDs
	akogatewayapitests.CreateHealthMonitorCRD(t, healthMonitorName1, DEFAULT_NAMESPACE, "thisisaviref-hm1")
	akogatewayapitests.CreateHealthMonitorCRD(t, healthMonitorName2, DEFAULT_NAMESPACE, "thisisaviref-hm2")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)

	// Create HTTPRoute with two rules, each having different backends and HealthMonitors
	rule1 := akogatewayapitests.GetHTTPRouteRuleWithHealthMonitorFilters(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName1, DEFAULT_NAMESPACE, "8080", "1"}}, []string{healthMonitorName1})
	rule2 := akogatewayapitests.GetHTTPRouteRuleWithHealthMonitorFilters(integrationtest.PATHPREFIX, []string{"/bar"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName2, DEFAULT_NAMESPACE, "8080", "1"}}, []string{healthMonitorName2})
	rules := []gatewayv1.HTTPRouteRule{rule1, rule2}
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(2))

	// Verify both pools have their respective HealthMonitors
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	childNode1 := nodes[0].EvhNodes[0]
	g.Expect(childNode1.PoolRefs).To(gomega.HaveLen(1))

	childNode2 := nodes[0].EvhNodes[1]
	g.Expect(childNode2.PoolRefs).To(gomega.HaveLen(1))

	g.Expect(childNode1.PoolRefs[0].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm1"))

	g.Expect(childNode2.PoolRefs[0].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm2"))

	// Delete one HealthMonitor and verify the system behavior
	akogatewayapitests.DeleteHealthMonitorCRD(t, healthMonitorName1, DEFAULT_NAMESPACE)

	// Wait for the model to be updated after HealthMonitor deletion
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes) > 0 {
			return len(nodes[0].EvhNodes[0].PoolRefs)
		}
		return 0
	}, 25*time.Second).Should(gomega.Equal(0))

	// Verify that only one pool remains (the one with valid HealthMonitor)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childNode := nodes[0].EvhNodes[1]
	g.Expect(childNode.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolRefs[0].HealthMonitorRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolRefs[0].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm2"))

	// Clean up
	akogatewayapitests.DeleteHealthMonitorCRD(t, healthMonitorName2, DEFAULT_NAMESPACE)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName2)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteWithHealthMonitorStatusTransition(t *testing.T) {
	gatewayName := "gateway-hm-06"
	gatewayClassName := "gateway-class-hm-06"
	httpRouteName := "http-route-hm-06"
	svcName := "avisvc-hm-06"
	healthMonitorName := "hm-06"
	ports := []int32{8080}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	// Create HealthMonitor CRD with Ready=True
	akogatewayapitests.CreateHealthMonitorCRD(t, healthMonitorName, DEFAULT_NAMESPACE, "thisisaviref-hm1")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleWithHealthMonitorFilters(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, []string{healthMonitorName})
	rules := []gatewayv1.HTTPRouteRule{rule}
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(1))

	// Verify HealthMonitor is initially present in graph layer
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childNode := nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolRefs[0].HealthMonitorRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolRefs[0].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm1"))

	// Update HealthMonitor status from Ready=True to Ready=False
	akogatewayapitests.UpdateHealthMonitorStatus(t, healthMonitorName, DEFAULT_NAMESPACE, false, "ValidationError", "HealthMonitor configuration is invalid")

	// Verify HealthMonitor is removed from graph layer after status change
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes[0].PoolRefs)
	}, 25*time.Second).Should(gomega.Equal(0))

	// Update HealthMonitor status back to Ready=True
	akogatewayapitests.UpdateHealthMonitorStatus(t, healthMonitorName, DEFAULT_NAMESPACE, true, "Accepted", "HealthMonitor has been successfully processed")

	// Verify HealthMonitor is added back to graph layer
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes) > 0 && len(nodes[0].EvhNodes[0].PoolRefs) > 0 {
			return len(nodes[0].EvhNodes[0].PoolRefs[0].HealthMonitorRefs)
		}
		return -1
	}, 25*time.Second).Should(gomega.Equal(1))

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childNode = nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolRefs[0].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm1"))

	akogatewayapitests.DeleteHealthMonitorCRD(t, healthMonitorName, DEFAULT_NAMESPACE)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}
