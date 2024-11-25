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
	"os"
	"testing"
	"time"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
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
	rule := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
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

	rule = akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
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
	ruleWithoutCanary := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{"avisvc", "default", "8080", "1"}}, nil)
	ruleWithCanary := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/bar"}, []string{"canary"},
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
	rule := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
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
	rule = akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
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
	rule := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
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
	rule = akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
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
	rule := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
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
	rule = akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
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

func TestHTTPRouteFilterWithRequestRedirect(t *testing.T) {

	gatewayName := "gateway-hr-06"
	gatewayClassName := "gateway-class-hr-06"
	httpRouteName := "http-route-hr-06"
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
	rule := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
		map[string][]string{"RequestRedirect": {}},
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
	g.Expect(childVS.HttpPolicyRefs[0].RequestRules[0].RedirectAction).ShouldNot(gomega.BeNil())
	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].RedirectAction.Host.Tokens[0].StrValue).To(gomega.Equal("redirect.com"))
	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].RedirectAction.StatusCode).To(gomega.Equal("HTTP_REDIRECT_STATUS_CODE_302"))

	// update httproute
	rule = akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
		map[string][]string{},
		[][]string{{"avisvc", "default", "8080", "1"}}, nil)
	rules = []gatewayv1.HTTPRouteRule{rule}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		childVS := nodes[0].EvhNodes[0]
		return len(childVS.HttpPolicyRefs)
	}, 25*time.Second).Should(gomega.Equal(0))

	// delete httproute
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteWithValidConfig(t *testing.T) {
	gatewayClassName := "gateway-class-hr-07"
	gatewayName := "gateway-hr-07"
	httpRouteName := "httproute-07"
	namespace := "default"
	svcName := "avisvc-hr-07"
	ports := []int32{8080, 8081}

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, "default", svcName, false, false, "1.1.1")
	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)

	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, namespace, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		gateway, err := akogatewayapitests.GatewayClient.GatewayV1().Gateways(namespace).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, ports)
	hostnames := []gatewayv1.Hostname{"foo-8080.com", "foo-8081.com"}
	rule := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName, namespace, "8080", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{rule}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, rules)
	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			t.Logf("Couldn't get the HTTPRoute, err: %+v", err)
			return false
		}
		if len(httpRoute.Status.Parents) != len(ports) {
			return false
		}
		return apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil &&
			apimeta.FindStatusCondition(httpRoute.Status.Parents[1].Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	modelName := lib.GetModelName(lib.GetTenant(), akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName))

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return false
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].VSVIPRefs[0].FQDNs) > 0 && len(nodes[0].EvhNodes[0].PoolGroupRefs) > 0

	}, 25*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(2))
	g.Expect(nodes[0].SSLKeyCertRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].VSVIPRefs).To(gomega.HaveLen(1))

	g.Expect(nodes[0].VSVIPRefs[0].FQDNs).To(gomega.Equal([]string{
		"foo-8080.com", "foo-8081.com"}))
	g.Expect(nodes[0].EvhNodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].AviMarkers.GatewayName).To(gomega.Equal(gatewayName))
	g.Expect(nodes[0].EvhNodes[0].AviMarkers.Namespace).To(gomega.Equal(namespace))
	g.Expect(nodes[0].EvhNodes[0].PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs).To(gomega.HaveLen(1))

	integrationtest.DelSVC(t, namespace, svcName)
	integrationtest.DelEPorEPS(t, namespace, svcName)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	akogatewayapitests.TeardownGateway(t, gatewayName, namespace)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteBackendRefCRUD(t *testing.T) {

	gatewayName := "gateway-hr-08"
	gatewayClassName := "gateway-class-hr-08"
	httpRouteName := "http-route-hr-08"
	svcName := "avisvc-hr-08"
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

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
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

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	childNode := nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode.DefaultPoolGroup).NotTo(gomega.Equal(""))

	rule = akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}}, nil, nil)
	rules = []gatewayv1.HTTPRouteRule{rule}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return -1
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes[0].PoolGroupRefs)
	}, 25*time.Second).Should(gomega.Equal(0))

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childNode = nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolRefs).To(gomega.HaveLen(0))
	g.Expect(childNode.PoolGroupRefs).To(gomega.HaveLen(0))
	g.Expect(childNode.DefaultPoolGroup).To(gomega.Equal(""))

	// update the backend service
	rule = akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rules = []gatewayv1.HTTPRouteRule{rule}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes[0].PoolGroupRefs)
	}, 25*time.Second).Should(gomega.Equal(1))

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childNode = nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode.DefaultPoolGroup).NotTo(gomega.Equal(""))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteBackendServiceCDC(t *testing.T) {

	gatewayName := "gateway-hr-09"
	gatewayClassName := "gateway-class-hr-09"
	httpRouteName := "http-route-hr-09"
	svcName := "avisvc-hr-09"
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

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
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
	}, 30*time.Second).Should(gomega.Equal(1))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	childNode := nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode.DefaultPoolGroup).NotTo(gomega.Equal(""))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return -1
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes[0].PoolGroupRefs)
	}, 30*time.Second).Should(gomega.Equal(0))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes[0].PoolGroupRefs[0].Members)
	}, 30*time.Second).Should(gomega.Equal(1))

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childNode = nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode.DefaultPoolGroup).NotTo(gomega.Equal(""))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteBackendServiceUpdate(t *testing.T) {

	gatewayName := "gateway-hr-10"
	gatewayClassName := "gateway-class-hr-10"
	httpRouteName := "http-route-hr-10"
	svcName1 := "avisvc-hr-04-10a"
	svcName2 := "avisvc-hr-04-10b"
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

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName1, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
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
	g.Expect(childNode.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode.DefaultPoolGroup).NotTo(gomega.Equal(""))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName2, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.3")

	rule = akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName1, DEFAULT_NAMESPACE, "8080", "1"}, {svcName2, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rules = []gatewayv1.HTTPRouteRule{rule}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return -1
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes[0].PoolGroupRefs[0].Members)
	}, 25*time.Second).Should(gomega.Equal(2))

	// update the backend service
	rule = akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName2, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rules = []gatewayv1.HTTPRouteRule{rule}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes[0].PoolGroupRefs[0].Members)
	}, 25*time.Second).Should(gomega.Equal(1))

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childNode = nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode.DefaultPoolGroup).NotTo(gomega.Equal(""))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteMultiportBackendSvc(t *testing.T) {

	gatewayName := "gateway-hr-11"
	gatewayClassName := "gateway-class-hr-11"
	httpRouteName := "http-route-hr-11"
	svcName := "avisvc-hr-11"
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

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, true)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName, true, true, "1.2.3")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
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

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	childNode := nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode.DefaultPoolGroup).NotTo(gomega.Equal(""))
	g.Expect(childNode.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolRefs[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(len(childNode.PoolRefs[0].Servers)).To(gomega.Equal(3))
	//g.Expect(childNode.PoolRefs).To(gomega.HaveLen(1))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteInvalidHostname(t *testing.T) {

	gatewayName := "gateway-hr-12"
	gatewayClassName := "gateway-class-hr-12"
	httpRouteName := "http-route-hr-12"
	svcName := "avisvc-hr-12"
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

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{rule}
	hostnames := []gatewayv1.Hostname{"foo-8080.com", "test-failure.com"}
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
	g.Expect(childNode.VHMatches).To(gomega.HaveLen(1))
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

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteGatewayWithEmptyHostnameInGatewayHTTPRoute(t *testing.T) {
	gatewayName := "gateway-hr-13"
	gatewayClassName := "gateway-class-hr-13"
	httpRouteName := "http-route-hr-13"
	svcName := "avisvc-hr-13"
	ports := []int32{8080}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")
	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)

	listeners := akogatewayapitests.GetListenersV1(ports, true, false)

	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 5*time.Second).Should(gomega.Equal(true))

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{}, nil,
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{rule}
	hostnames := []gatewayv1.Hostname{}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	// no child vs
	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return -1
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 5*time.Second).Should(gomega.Equal(0))

	// Check Parent Properties
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Eventually(func() int {
		return len(nodes[0].HttpPolicyRefs)
	}, 20*time.Second).Should(gomega.Equal(1))
	g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(0))

	// delete httproute
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Eventually(func() int {
		return len(nodes[0].HttpPolicyRefs)
	}, 10*time.Second).Should(gomega.Equal(1))
	g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(0))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteWithMultipleListenerGateway(t *testing.T) {
	gatewayName := "gateway-hr-14"
	gatewayClassName := "gateway-class-hr-14"
	httpRouteName := "http-route-hr-14"
	svcName := "avisvc-hr-14"
	ports := []int32{8080, 8082}
	modelName, parentVSName := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	// create a gateway with listener with same hostname and different port
	listeners := akogatewayapitests.GetListenersV1(ports, false, true)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	// httproute parent ref
	parentRefs := akogatewayapitests.GetParentReferencesV1WithGatewayNameOnly([]string{gatewayName}, DEFAULT_NAMESPACE)

	// httproute rule 1
	rule1 := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{}, nil,
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{rule1}

	// httproute rule2
	rule2 := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/bar"}, []string{}, nil,
		[][]string{{svcName, DEFAULT_NAMESPACE, "8082", "1"}}, nil)
	rules = append(rules, rule2)

	hostnames := []gatewayv1.Hostname{"foo.com"}

	//create httproute
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(2))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	// childe node 1
	childNode := nodes[0].EvhNodes[0]
	g.Expect(childNode.VHParentName).To(gomega.Equal(parentVSName))
	g.Expect(childNode.VHMatches).To(gomega.HaveLen(1))
	g.Expect(*childNode.VHMatches[0].Host).To(gomega.Equal("foo.com"))
	// path foo
	g.Expect(childNode.VHMatches[0].Rules[0].Matches.Path.MatchStr).To(gomega.ContainElement("/foo"))
	g.Expect(*childNode.VHMatches[0].Rules[0].Matches.Path.MatchCriteria).To(gomega.Equal("BEGINS_WITH"))
	g.Expect(len(childNode.VHMatches[0].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(2))
	g.Expect(childNode.VHMatches[0].Rules[0].Matches.VsPort.Ports).Should(gomega.ConsistOf([]int64{8080, 8082}))

	// child node 2
	childNode = nodes[0].EvhNodes[1]
	g.Expect(childNode.VHParentName).To(gomega.Equal(parentVSName))
	g.Expect(childNode.VHMatches).To(gomega.HaveLen(1))
	g.Expect(*childNode.VHMatches[0].Host).To(gomega.Equal("foo.com"))
	// Path bar
	g.Expect(childNode.VHMatches[0].Rules[0].Matches.Path.MatchStr).To(gomega.ContainElement("/bar"))
	g.Expect(*childNode.VHMatches[0].Rules[0].Matches.Path.MatchCriteria).To(gomega.Equal("BEGINS_WITH"))
	g.Expect(len(childNode.VHMatches[0].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(2))
	g.Expect(childNode.VHMatches[0].Rules[0].Matches.VsPort.Ports).Should(gomega.ConsistOf([]int64{8080, 8082}))

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
	}, 30*time.Second).Should(gomega.Equal(0))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}
func TestHTTPRouteWithMultipleGateways(t *testing.T) {
	gatewayName1 := "gateway-hr-15a"
	gatewayName2 := "gateway-hr-15b"
	gatewayClassName := "gateway-class-hr-15"
	httpRouteName := "http-route-hr-15"
	svcName := "avisvc-hr-15"
	ports1 := []int32{8080}
	ports2 := []int32{8081}
	modelName1, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName1)
	modelName2, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName2)

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")
	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports1, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName1, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName1)
		return found
	}, 45*time.Second).Should(gomega.Equal(true))

	listeners = akogatewayapitests.GetListenersV1(ports2, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName2, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName2)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName1, gatewayName2}, DEFAULT_NAMESPACE, []int32{8080, 8081})
	parentRefs = []gatewayv1.ParentReference{parentRefs[0], parentRefs[3]}
	rule1 := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{}, nil,
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rule2 := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/bar"}, []string{}, nil,
		[][]string{{svcName, DEFAULT_NAMESPACE, "8081", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{rule1, rule2}
	hostnames := []gatewayv1.Hostname{"foo-8080.com", "foo-8081.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)
	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName1)
		if !found {
			return -1
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(2))

	// Check Parent Properties
	_, aviModel1 := objects.SharedAviGraphLister().Get(modelName1)
	nodes := aviModel1.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(0))
	g.Expect(len(nodes[0].VSVIPRefs[0].FQDNs)).To(gomega.Equal(1))
	g.Expect(nodes[0].VSVIPRefs[0].FQDNs[0]).To(gomega.Equal("foo-8080.com"))

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName2)
		if !found {
			return -1
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(2))

	// Check Parent Properties
	_, aviModel2 := objects.SharedAviGraphLister().Get(modelName2)
	nodes = aviModel2.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(0))
	g.Expect(len(nodes[0].VSVIPRefs[0].FQDNs)).To(gomega.Equal(1))
	g.Expect(nodes[0].VSVIPRefs[0].FQDNs[0]).To(gomega.Equal("foo-8081.com"))

	// delete httproute
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName1)
		if !found {
			return -1
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(0))

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName2)
		if !found {
			return -1
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(0))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownGateway(t, gatewayName1, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName2, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteParentFQDN(t *testing.T) {
	// create a gateway with two listener *.avi.internal and specific hello.avi.internal
	// create httproute1 with hostname abc.avi.internal
	// validate parent FQDN
	// add hostname efg.avi.internal
	// validate parent FQDN
	// update hostname abc.avi.internal to abcde.avi.internal
	// validate parent FQDN
	// create another httproute2 with hostname hello.avi.internal mapping to 2nd listener
	// validate parent FQDN
	// delete httproutes
	// validate parent FQDN

	gatewayName1 := "gateway-hr-16"

	gatewayClassName := "gateway-class-hr-16"
	httpRouteName1 := "http-route-hr-16a"
	httpRouteName2 := "http-route-hr-16b"
	svcName := "avisvc-hr-16"

	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName1)

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")
	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersOnHostname([]string{"*.avi.internal", "hello.avi.internal"})

	akogatewayapitests.SetupGateway(t, gatewayName1, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 45*time.Second).Should(gomega.Equal(true))

	parentRefs := akogatewayapitests.GetParentReferencesFromListeners(listeners, gatewayName1, DEFAULT_NAMESPACE)
	rule1 := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{}, nil,
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, nil)

	rules := []gatewayv1.HTTPRouteRule{rule1}
	hostnames := []gatewayv1.Hostname{"abc.avi.internal"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName1, DEFAULT_NAMESPACE, []gatewayv1.ParentReference{parentRefs[0]}, hostnames, rules)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return -1
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(1))

	// Check Parent Properties
	_, aviModel1 := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel1.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(0))
	g.Expect(len(nodes[0].VSVIPRefs[0].FQDNs)).To(gomega.Equal(1))
	g.Expect(nodes[0].VSVIPRefs[0].FQDNs[0]).To(gomega.Equal("abc.avi.internal"))

	// update httproute to add one more hostname
	hostnames = []gatewayv1.Hostname{"abc.avi.internal", "efg.avi.internal"}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName1, DEFAULT_NAMESPACE, []gatewayv1.ParentReference{parentRefs[0]}, hostnames, rules)

	g.Eventually(func() int {
		_, aviModel1 = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel1.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].VSVIPRefs[0].FQDNs)
	}, 25*time.Second, 1*time.Second).Should(gomega.Equal(2))

	// validate parent
	_, aviModel1 = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel1.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes[0].VSVIPRefs[0].FQDNs)).To(gomega.Equal(2))
	g.Expect(nodes[0].VSVIPRefs[0].FQDNs[0]).To(gomega.Equal("abc.avi.internal"))
	g.Expect(nodes[0].VSVIPRefs[0].FQDNs[1]).To(gomega.Equal("efg.avi.internal"))

	// update httproute with new hostname
	hostnames = []gatewayv1.Hostname{"abcdef.avi.internal", "efg.avi.internal"}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName1, DEFAULT_NAMESPACE, []gatewayv1.ParentReference{parentRefs[0]}, hostnames, rules)

	g.Eventually(func() string {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return ""
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return nodes[0].VSVIPRefs[0].FQDNs[0]
	}, 25*time.Second).Should(gomega.Equal("abcdef.avi.internal"))

	// check other parent fqdns
	_, aviModel1 = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel1.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes[0].VSVIPRefs[0].FQDNs)).To(gomega.Equal(2))
	g.Expect(nodes[0].VSVIPRefs[0].FQDNs[0]).To(gomega.Equal("abcdef.avi.internal"))
	g.Expect(nodes[0].VSVIPRefs[0].FQDNs[1]).To(gomega.Equal("efg.avi.internal"))

	// create httproute 2 and attach it to listener 2
	rule1 = akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{}, nil,
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, nil)

	rules = []gatewayv1.HTTPRouteRule{rule1}
	hostnames = []gatewayv1.Hostname{"hello.avi.internal"}

	akogatewayapitests.SetupHTTPRoute(t, httpRouteName2, DEFAULT_NAMESPACE, []gatewayv1.ParentReference{parentRefs[1]}, hostnames, rules)
	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return -1
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(2))

	// validate parent
	_, aviModel1 = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel1.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(0))
	g.Expect(len(nodes[0].VSVIPRefs[0].FQDNs)).To(gomega.Equal(3))
	g.Expect(nodes[0].VSVIPRefs[0].FQDNs[0]).To(gomega.Equal("abcdef.avi.internal"))
	g.Expect(nodes[0].VSVIPRefs[0].FQDNs[1]).To(gomega.Equal("efg.avi.internal"))
	g.Expect(nodes[0].VSVIPRefs[0].FQDNs[2]).To(gomega.Equal("hello.avi.internal"))

	// update httproute after deleting one hostname
	hostnames = []gatewayv1.Hostname{"efg.avi.internal"}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName1, DEFAULT_NAMESPACE, []gatewayv1.ParentReference{parentRefs[0]}, hostnames, rules)

	g.Eventually(func() int {
		_, aviModel1 = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel1.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].VSVIPRefs[0].FQDNs)
	}, 25*time.Second, 1*time.Second).Should(gomega.Equal(2))

	// delete httproutes
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName1, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName2, DEFAULT_NAMESPACE)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return -1
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(0))

	// validate parent
	_, aviModel1 = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel1.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(0))
	g.Expect(len(nodes[0].VSVIPRefs[0].FQDNs)).To(gomega.Equal(0))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownGateway(t, gatewayName1, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}
func TestHttpRouteWithValidAndInvalidGatewayListeners(t *testing.T) {
	gatewayName := "gateway-hr-17"
	gatewayClassName := "gateway-class-hr-17"
	httpRouteName := "http-route-hr-17"
	svcName1 := "avisvc-hr-17-a"
	svcName2 := "avisvc-hr-17-b"
	ports := []int32{8080, 8081}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	listeners[1].Protocol = "TCP"
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName1, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName2, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.3")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule1 := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName1, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rule2 := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/bar"}, []string{},
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
	g.Expect(childNode1.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode1.DefaultPoolGroup).NotTo(gomega.Equal(""))
	g.Expect(childNode1.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolRefs[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(len(childNode1.PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(childNode1.VHMatches)).To(gomega.Equal(1))
	g.Expect(*childNode1.VHMatches[0].Host).To(gomega.Equal("foo-8080.com"))
	g.Expect(len(childNode1.VHMatches[0].Rules)).To(gomega.Equal(1))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/foo"))
	g.Expect(len(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(1))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))

	childNode2 := nodes[0].EvhNodes[1]
	g.Expect(childNode2.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode2.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode2.DefaultPoolGroup).NotTo(gomega.Equal(""))
	g.Expect(childNode2.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode2.PoolRefs[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(len(childNode2.PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(childNode2.VHMatches)).To(gomega.Equal(1))
	g.Expect(*childNode2.VHMatches[0].Host).To(gomega.Equal("foo-8080.com"))
	g.Expect(len(childNode2.VHMatches[0].Rules)).To(gomega.Equal(1))
	g.Expect(childNode2.VHMatches[0].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/bar"))
	g.Expect(len(childNode2.VHMatches[0].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(1))
	g.Expect(childNode2.VHMatches[0].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName2)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestMultipleHttpRoutesWithValidAndInvalidGatewayListeners(t *testing.T) {
	gatewayName := "gateway-hr-18"
	gatewayClassName := "gateway-class-hr-18"
	httpRouteName1 := "http-route-hr-18-a"
	httpRouteName2 := "http-route-hr-18-b"
	svcName := "avisvc-hr-18"
	ports := []int32{8080, 8081}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	listeners[1].Protocol = "TCP"
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	parentRefs1 := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, []int32{ports[0]})
	rule1 := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rules1 := []gatewayv1.HTTPRouteRule{rule1}
	hostnames1 := []gatewayv1.Hostname{"foo-8080.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName1, DEFAULT_NAMESPACE, parentRefs1, hostnames1, rules1)

	parentRefs2 := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, []int32{ports[1]})
	rule2 := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/bar"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8081", "1"}}, nil)
	rules2 := []gatewayv1.HTTPRouteRule{rule2}
	hostnames2 := []gatewayv1.Hostname{"foo-8081.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName2, DEFAULT_NAMESPACE, parentRefs2, hostnames2, rules2)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 30*time.Second).Should(gomega.Equal(1))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	childNode1 := nodes[0].EvhNodes[0]
	g.Expect(childNode1.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode1.DefaultPoolGroup).NotTo(gomega.Equal(""))
	g.Expect(childNode1.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolRefs[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(len(childNode1.PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(childNode1.VHMatches)).To(gomega.Equal(1))
	g.Expect(*childNode1.VHMatches[0].Host).To(gomega.Equal("foo-8080.com"))
	g.Expect(len(childNode1.VHMatches[0].Rules)).To(gomega.Equal(1))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/foo"))
	g.Expect(len(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(1))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName1, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName2, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}
func TestTransitionsHttpRouteWithPartiallyValidGatewayToValidGateway(t *testing.T) {
	//1: One HTTPRoute with partially valid gateway
	gatewayName := "gateway-hr-19"
	gatewayClassName := "gateway-class-hr-19"
	httpRouteName := "http-route-hr-19"
	svcName1 := "avisvc-hr-19-a"
	svcName2 := "avisvc-hr-19-b"
	ports := []int32{8080, 8081}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	listeners[1].Protocol = "TCP"
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName1, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName2, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.3")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule1 := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName1, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rule2 := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/bar"}, []string{},
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
	g.Expect(childNode1.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode1.DefaultPoolGroup).NotTo(gomega.Equal(""))
	g.Expect(childNode1.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolRefs[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(len(childNode1.PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(childNode1.VHMatches)).To(gomega.Equal(1))
	g.Expect(*childNode1.VHMatches[0].Host).To(gomega.Equal("foo-8080.com"))
	g.Expect(len(childNode1.VHMatches[0].Rules)).To(gomega.Equal(1))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/foo"))
	g.Expect(len(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(1))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))

	childNode2 := nodes[0].EvhNodes[1]
	g.Expect(childNode2.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode2.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode2.DefaultPoolGroup).NotTo(gomega.Equal(""))
	g.Expect(childNode2.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode2.PoolRefs[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(len(childNode2.PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(childNode2.VHMatches)).To(gomega.Equal(1))
	g.Expect(*childNode2.VHMatches[0].Host).To(gomega.Equal("foo-8080.com"))
	g.Expect(len(childNode2.VHMatches[0].Rules)).To(gomega.Equal(1))
	g.Expect(childNode2.VHMatches[0].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/bar"))
	g.Expect(len(childNode2.VHMatches[0].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(1))
	g.Expect(childNode2.VHMatches[0].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))

	listeners[1].Protocol = "HTTPS"
	akogatewayapitests.UpdateGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g.Eventually(func() bool {
		gateway, err := akogatewayapitests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		if len(gateway.Status.Listeners) < 2 {
			return false
		}
		return gateway.Status.Listeners[1].Conditions[0].Status == metav1.ConditionTrue
	}, 30*time.Second).Should(gomega.Equal(true))

	gateway, _ := akogatewayapitests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	g.Expect(gateway.Status.Listeners[1].Conditions[0].Status).To(gomega.Equal(metav1.ConditionTrue))
	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes) != 2 {
			return 0
		}
		return len(nodes[0].EvhNodes[0].VHMatches)
	}, 30*time.Second).Should(gomega.Equal(2))

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes) != 2 {
			return 0
		}
		return len(nodes[0].EvhNodes[1].VHMatches)
	}, 30*time.Second).Should(gomega.Equal(2))

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	childNode1 = nodes[0].EvhNodes[0]
	g.Expect(childNode1.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode1.DefaultPoolGroup).NotTo(gomega.Equal(""))
	g.Expect(childNode1.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolRefs[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(len(childNode1.PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(childNode1.VHMatches)).To(gomega.Equal(2))
	g.Expect(*childNode1.VHMatches[0].Host).To(gomega.Equal("foo-8080.com"))
	g.Expect(len(childNode1.VHMatches[0].Rules)).To(gomega.Equal(1))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/foo"))
	g.Expect(len(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(2))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports[1]).To(gomega.Equal(int64(8081)))
	g.Expect(*childNode1.VHMatches[1].Host).To(gomega.Equal("foo-8081.com"))
	g.Expect(len(childNode1.VHMatches[1].Rules)).To(gomega.Equal(1))
	g.Expect(childNode1.VHMatches[1].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/foo"))
	g.Expect(len(childNode1.VHMatches[1].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(2))
	g.Expect(childNode1.VHMatches[1].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))
	g.Expect(childNode1.VHMatches[1].Rules[0].Matches.VsPort.Ports[1]).To(gomega.Equal(int64(8081)))

	childNode2 = nodes[0].EvhNodes[1]
	g.Expect(childNode2.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode2.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode2.DefaultPoolGroup).NotTo(gomega.Equal(""))
	g.Expect(childNode2.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode2.PoolRefs[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(len(childNode2.PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(childNode2.VHMatches)).To(gomega.Equal(2))
	g.Expect(*childNode2.VHMatches[0].Host).To(gomega.Equal("foo-8080.com"))
	g.Expect(len(childNode2.VHMatches[0].Rules)).To(gomega.Equal(1))
	g.Expect(childNode2.VHMatches[0].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/bar"))
	g.Expect(len(childNode2.VHMatches[0].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(2))
	g.Expect(childNode2.VHMatches[0].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))
	g.Expect(childNode2.VHMatches[0].Rules[0].Matches.VsPort.Ports[1]).To(gomega.Equal(int64(8081)))
	g.Expect(*childNode2.VHMatches[1].Host).To(gomega.Equal("foo-8081.com"))
	g.Expect(len(childNode2.VHMatches[1].Rules)).To(gomega.Equal(1))
	g.Expect(childNode2.VHMatches[1].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/bar"))
	g.Expect(len(childNode2.VHMatches[1].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(2))
	g.Expect(childNode2.VHMatches[1].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))
	g.Expect(childNode2.VHMatches[1].Rules[0].Matches.VsPort.Ports[1]).To(gomega.Equal(int64(8081)))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName2)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestTransitionsHttpRouteWithPartiallyValidGatewayToInvalidGateway(t *testing.T) {
	t.Skip("Skipping since current implementation is not supporting partially Valid to Invalid gateway transition")
	gatewayName := "gateway-hr-20"
	gatewayClassName := "gateway-class-hr-20"
	httpRouteName := "http-route-hr-20"
	svcName1 := "avisvc-hr-20-a"
	svcName2 := "avisvc-hr-20-b"
	ports := []int32{8080, 8081}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	listeners[1].Protocol = "TCP"
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName1, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName2, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.3")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule1 := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName1, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rule2 := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/bar"}, []string{},
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
	g.Expect(childNode1.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode1.DefaultPoolGroup).NotTo(gomega.Equal(""))
	g.Expect(childNode1.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolRefs[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(len(childNode1.PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(childNode1.VHMatches)).To(gomega.Equal(1))
	g.Expect(*childNode1.VHMatches[0].Host).To(gomega.Equal("foo-8080.com"))
	g.Expect(len(childNode1.VHMatches[0].Rules)).To(gomega.Equal(1))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/foo"))
	g.Expect(len(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(1))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))

	childNode2 := nodes[0].EvhNodes[1]
	g.Expect(childNode2.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode2.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode2.DefaultPoolGroup).NotTo(gomega.Equal(""))
	g.Expect(childNode2.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode2.PoolRefs[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(len(childNode2.PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(childNode2.VHMatches)).To(gomega.Equal(1))
	g.Expect(*childNode2.VHMatches[0].Host).To(gomega.Equal("foo-8080.com"))
	g.Expect(len(childNode2.VHMatches[0].Rules)).To(gomega.Equal(1))
	g.Expect(childNode2.VHMatches[0].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/bar"))
	g.Expect(len(childNode2.VHMatches[0].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(1))
	g.Expect(childNode2.VHMatches[0].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))

	listeners[0].Protocol = "TCP"
	akogatewayapitests.UpdateGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g.Eventually(func() bool {
		gateway, err := akogatewayapitests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	gateway, _ := akogatewayapitests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	g.Expect(gateway.Status.Listeners[1].Conditions[0].Status).To(gomega.Equal(metav1.ConditionTrue))

	g.Eventually(func() int {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes[0].VHMatches)
	}, 50*time.Second, 5*time.Second).Should(gomega.Equal(2))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName2)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)

}

func TestTransitionsHttpRouteWithInvalidGatewayToPartiallyValidGateway(t *testing.T) {
	gatewayName := "gateway-hr-21"
	gatewayClassName := "gateway-class-hr-21"
	httpRouteName := "http-route-hr-21"
	svcName1 := "avisvc-hr-21-a"
	svcName2 := "avisvc-hr-21-b"
	ports := []int32{8080, 8081}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	listeners[1].Protocol = "TCP"
	listeners[0].Protocol = "TCP"
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(false))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName1, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName2, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.3")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule1 := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName1, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rule2 := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/bar"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName2, DEFAULT_NAMESPACE, "8081", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{rule1, rule2}

	hostnames := []gatewayv1.Hostname{"foo-8080.com", "foo-8081.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)
	time.Sleep(10 * time.Second)
	listeners[1].Protocol = "HTTPS"
	akogatewayapitests.UpdateGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)
	g.Eventually(func() bool {
		gateway, err := akogatewayapitests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))
	gateway, _ := akogatewayapitests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	g.Expect(gateway.Status.Listeners[1].Conditions[0].Status).To(gomega.Equal(metav1.ConditionTrue))
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
	g.Expect(childNode1.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode1.DefaultPoolGroup).NotTo(gomega.Equal(""))
	g.Expect(childNode1.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolRefs[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(len(childNode1.PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(childNode1.VHMatches)).To(gomega.Equal(1))
	g.Expect(*childNode1.VHMatches[0].Host).To(gomega.Equal("foo-8081.com"))
	g.Expect(len(childNode1.VHMatches[0].Rules)).To(gomega.Equal(1))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/foo"))
	g.Expect(len(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(1))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8081)))

	childNode2 := nodes[0].EvhNodes[1]
	g.Expect(childNode2.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode2.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode2.DefaultPoolGroup).NotTo(gomega.Equal(""))
	g.Expect(childNode2.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode2.PoolRefs[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(len(childNode2.PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(childNode2.VHMatches)).To(gomega.Equal(1))
	g.Expect(*childNode2.VHMatches[0].Host).To(gomega.Equal("foo-8081.com"))
	g.Expect(len(childNode2.VHMatches[0].Rules)).To(gomega.Equal(1))
	g.Expect(childNode2.VHMatches[0].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/bar"))
	g.Expect(len(childNode2.VHMatches[0].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(1))
	g.Expect(childNode2.VHMatches[0].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8081)))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName2)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)

}

func TestTransitionsHttpRouteWithValidGatewayToPartiallyValidGateway(t *testing.T) {
	t.Skip("Skipping since current implementation is not supporting  Valid to partially valid gateway transition")
	gatewayName := "gateway-hr-22"
	gatewayClassName := "gateway-class-hr-22"
	httpRouteName := "http-route-hr-22"
	svcName1 := "avisvc-hr-22-a"
	svcName2 := "avisvc-hr-22-b"
	ports := []int32{8080, 8081}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName1, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName2, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.3")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule1 := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName1, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rule2 := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/bar"}, []string{},
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
	g.Expect(childNode1.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode1.DefaultPoolGroup).NotTo(gomega.Equal(""))
	g.Expect(childNode1.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolRefs[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(len(childNode1.PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(childNode1.VHMatches)).To(gomega.Equal(2))
	g.Expect(*childNode1.VHMatches[0].Host).To(gomega.Equal("foo-8080.com"))
	g.Expect(len(childNode1.VHMatches[0].Rules)).To(gomega.Equal(1))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/foo"))
	g.Expect(len(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(2))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports[1]).To(gomega.Equal(int64(8081)))
	g.Expect(*childNode1.VHMatches[1].Host).To(gomega.Equal("foo-8081.com"))
	g.Expect(len(childNode1.VHMatches[1].Rules)).To(gomega.Equal(1))
	g.Expect(childNode1.VHMatches[1].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/foo"))
	g.Expect(len(childNode1.VHMatches[1].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(2))
	g.Expect(childNode1.VHMatches[1].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))
	g.Expect(childNode1.VHMatches[1].Rules[0].Matches.VsPort.Ports[1]).To(gomega.Equal(int64(8081)))

	childNode2 := nodes[0].EvhNodes[1]
	g.Expect(childNode2.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode2.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode2.DefaultPoolGroup).NotTo(gomega.Equal(""))
	g.Expect(childNode2.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode2.PoolRefs[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(len(childNode2.PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(childNode2.VHMatches)).To(gomega.Equal(2))
	g.Expect(*childNode2.VHMatches[0].Host).To(gomega.Equal("foo-8080.com"))
	g.Expect(len(childNode2.VHMatches[0].Rules)).To(gomega.Equal(1))
	g.Expect(childNode2.VHMatches[0].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/bar"))
	g.Expect(len(childNode2.VHMatches[0].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(2))
	g.Expect(childNode2.VHMatches[0].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))
	g.Expect(childNode2.VHMatches[0].Rules[0].Matches.VsPort.Ports[1]).To(gomega.Equal(int64(8081)))
	g.Expect(*childNode2.VHMatches[1].Host).To(gomega.Equal("foo-8081.com"))
	g.Expect(len(childNode2.VHMatches[1].Rules)).To(gomega.Equal(1))
	g.Expect(childNode2.VHMatches[1].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/bar"))
	g.Expect(len(childNode2.VHMatches[1].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(2))
	g.Expect(childNode2.VHMatches[1].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))
	g.Expect(childNode2.VHMatches[1].Rules[0].Matches.VsPort.Ports[1]).To(gomega.Equal(int64(8081)))

	listeners[1].Protocol = "TCP"
	akogatewayapitests.UpdateGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g.Eventually(func() bool {
		gateway, err := akogatewayapitests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	gateway, _ := akogatewayapitests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	g.Expect(gateway.Status.Listeners[1].Conditions[0].Status).To(gomega.Equal(metav1.ConditionFalse))

	g.Eventually(func() int {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes[0].VHMatches)
	}, 50*time.Second, 5*time.Second).Should(gomega.Equal(1))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	childNode1 = nodes[0].EvhNodes[0]
	g.Expect(childNode1.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode1.DefaultPoolGroup).NotTo(gomega.Equal(""))
	g.Expect(childNode1.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolRefs[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(len(childNode1.PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(childNode1.VHMatches)).To(gomega.Equal(2))
	g.Expect(*childNode1.VHMatches[0].Host).To(gomega.Equal("foo-8080.com"))
	g.Expect(len(childNode1.VHMatches[0].Rules)).To(gomega.Equal(1))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/foo"))
	g.Expect(len(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(2))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports[1]).To(gomega.Equal(int64(8081)))
	g.Expect(*childNode1.VHMatches[1].Host).To(gomega.Equal("foo-8081.com"))
	g.Expect(len(childNode1.VHMatches[1].Rules)).To(gomega.Equal(1))
	g.Expect(childNode1.VHMatches[1].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/foo"))
	g.Expect(len(childNode1.VHMatches[1].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(2))
	g.Expect(childNode1.VHMatches[1].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))
	g.Expect(childNode1.VHMatches[1].Rules[0].Matches.VsPort.Ports[1]).To(gomega.Equal(int64(8081)))

	childNode2 = nodes[0].EvhNodes[1]
	g.Expect(childNode2.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode2.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode2.DefaultPoolGroup).NotTo(gomega.Equal(""))
	g.Expect(childNode2.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode2.PoolRefs[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(len(childNode2.PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(childNode2.VHMatches)).To(gomega.Equal(2))
	g.Expect(*childNode2.VHMatches[0].Host).To(gomega.Equal("foo-8080.com"))
	g.Expect(len(childNode2.VHMatches[0].Rules)).To(gomega.Equal(1))
	g.Expect(childNode2.VHMatches[0].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/bar"))
	g.Expect(len(childNode2.VHMatches[0].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(2))
	g.Expect(childNode2.VHMatches[0].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))
	g.Expect(childNode2.VHMatches[0].Rules[0].Matches.VsPort.Ports[1]).To(gomega.Equal(int64(8081)))
	g.Expect(*childNode2.VHMatches[1].Host).To(gomega.Equal("foo-8081.com"))
	g.Expect(len(childNode2.VHMatches[1].Rules)).To(gomega.Equal(1))
	g.Expect(childNode2.VHMatches[1].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/bar"))
	g.Expect(len(childNode2.VHMatches[1].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(2))
	g.Expect(childNode2.VHMatches[1].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))
	g.Expect(childNode2.VHMatches[1].Rules[0].Matches.VsPort.Ports[1]).To(gomega.Equal(int64(8081)))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName2)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestTransitionsMultipleHttpRoutesWithPartiallyValidGatewayToValidGateway(t *testing.T) {
	// 1: Two HTTPRoute with partially valid gateway
	gatewayName := "gateway-hr-23"
	gatewayClassName := "gateway-class-hr-23"
	httpRoute1Name := "http-route-hr-23a"
	httpRoute2Name := "http-route-hr-23b"
	svcName1 := "avisvc-hr-23-a"
	svcName2 := "avisvc-hr-23-b"
	ports := []int32{8080, 8081}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	listeners[1].Protocol = "TCP"
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName1, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName2, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.3")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, []int32{ports[0]})
	rule1 := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName1, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rule2 := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/bar"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName2, DEFAULT_NAMESPACE, "8081", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{rule1, rule2}

	hostnames := []gatewayv1.Hostname{"foo-8080.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRoute1Name, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)
	hostnames = []gatewayv1.Hostname{"foo-8081.com"}
	parentRefs = akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, []int32{ports[1]})
	akogatewayapitests.SetupHTTPRoute(t, httpRoute2Name, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

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
	g.Expect(childNode1.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode1.DefaultPoolGroup).NotTo(gomega.Equal(""))
	g.Expect(childNode1.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolRefs[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(len(childNode1.PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(childNode1.VHMatches)).To(gomega.Equal(1))
	g.Expect(*childNode1.VHMatches[0].Host).To(gomega.Equal("foo-8080.com"))
	g.Expect(len(childNode1.VHMatches[0].Rules)).To(gomega.Equal(1))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/foo"))
	g.Expect(len(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(1))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))

	childNode2 := nodes[0].EvhNodes[1]
	g.Expect(childNode2.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode2.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode2.DefaultPoolGroup).NotTo(gomega.Equal(""))
	g.Expect(childNode2.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode2.PoolRefs[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(len(childNode2.PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(childNode2.VHMatches)).To(gomega.Equal(1))
	g.Expect(*childNode2.VHMatches[0].Host).To(gomega.Equal("foo-8080.com"))
	g.Expect(len(childNode2.VHMatches[0].Rules)).To(gomega.Equal(1))
	g.Expect(childNode2.VHMatches[0].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/bar"))
	g.Expect(len(childNode2.VHMatches[0].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(1))
	g.Expect(childNode2.VHMatches[0].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))

	listeners[1].Protocol = "HTTPS"
	akogatewayapitests.UpdateGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g.Eventually(func() bool {
		gateway, err := akogatewayapitests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	gateway, _ := akogatewayapitests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	g.Expect(gateway.Status.Listeners[1].Conditions[0].Status).To(gomega.Equal(metav1.ConditionTrue))
	g.Eventually(func() int {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 50*time.Second, 5*time.Second).Should(gomega.Equal(4))

	childNode1 = nodes[0].EvhNodes[0]
	g.Expect(childNode1.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode1.DefaultPoolGroup).NotTo(gomega.Equal(""))
	g.Expect(childNode1.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolRefs[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(len(childNode1.PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(childNode1.VHMatches)).To(gomega.Equal(1))
	g.Expect(*childNode1.VHMatches[0].Host).To(gomega.Equal("foo-8080.com"))
	g.Expect(len(childNode1.VHMatches[0].Rules)).To(gomega.Equal(1))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/foo"))
	g.Expect(len(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(1))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))

	childNode2 = nodes[0].EvhNodes[1]
	g.Expect(childNode2.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode2.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode2.DefaultPoolGroup).NotTo(gomega.Equal(""))
	g.Expect(childNode2.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode2.PoolRefs[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(len(childNode2.PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(childNode2.VHMatches)).To(gomega.Equal(1))
	g.Expect(*childNode2.VHMatches[0].Host).To(gomega.Equal("foo-8080.com"))
	g.Expect(len(childNode2.VHMatches[0].Rules)).To(gomega.Equal(1))
	g.Expect(childNode2.VHMatches[0].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/bar"))
	g.Expect(len(childNode2.VHMatches[0].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(1))
	g.Expect(childNode2.VHMatches[0].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))

	childNode3 := nodes[0].EvhNodes[2]
	g.Expect(childNode3.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode3.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode3.DefaultPoolGroup).NotTo(gomega.Equal(""))
	g.Expect(childNode3.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode3.PoolRefs[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(len(childNode3.PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(childNode3.VHMatches)).To(gomega.Equal(1))
	g.Expect(*childNode3.VHMatches[0].Host).To(gomega.Equal("foo-8081.com"))
	g.Expect(len(childNode3.VHMatches[0].Rules)).To(gomega.Equal(1))
	g.Expect(childNode3.VHMatches[0].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/foo"))
	g.Expect(len(childNode3.VHMatches[0].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(1))
	g.Expect(childNode3.VHMatches[0].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8081)))

	childNode4 := nodes[0].EvhNodes[3]
	g.Expect(childNode4.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode4.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode4.DefaultPoolGroup).NotTo(gomega.Equal(""))
	g.Expect(childNode4.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode4.PoolRefs[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(len(childNode4.PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(childNode4.VHMatches)).To(gomega.Equal(1))
	g.Expect(*childNode4.VHMatches[0].Host).To(gomega.Equal("foo-8081.com"))
	g.Expect(len(childNode4.VHMatches[0].Rules)).To(gomega.Equal(1))
	g.Expect(childNode4.VHMatches[0].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/bar"))
	g.Expect(len(childNode4.VHMatches[0].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(1))
	g.Expect(childNode4.VHMatches[0].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8081)))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName2)
	akogatewayapitests.TeardownHTTPRoute(t, httpRoute1Name, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownHTTPRoute(t, httpRoute2Name, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}
func TestTransitionsMultipleHttpRouteWithInvalidGatewayToPartiallyValidGateway(t *testing.T) {
	// 1: Two HTTPRoute with  invalid gateway
	gatewayName := "gateway-hr-24"
	gatewayClassName := "gateway-class-hr-24"
	httpRoute1Name := "http-route-hr-24a"
	httpRoute2Name := "http-route-hr-24b"
	svcName1 := "avisvc-hr-24-a"
	svcName2 := "avisvc-hr-24-b"
	ports := []int32{8080, 8081}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	listeners[0].Protocol = "TCP"
	listeners[1].Protocol = "TCP"
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName1, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName2, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.3")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, []int32{ports[0]})
	rule1 := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName1, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rule2 := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/bar"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName2, DEFAULT_NAMESPACE, "8081", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{rule1, rule2}
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRoute1Name, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)
	hostnames = []gatewayv1.Hostname{"foo-8081.com"}
	parentRefs = akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, []int32{ports[1]})
	akogatewayapitests.SetupHTTPRoute(t, httpRoute2Name, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)
	listeners[0].Protocol = "HTTPS"
	time.Sleep(10 * time.Second)
	akogatewayapitests.UpdateGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g.Eventually(func() bool {
		gateway, err := akogatewayapitests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	gateway, _ := akogatewayapitests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	g.Expect(gateway.Status.Listeners[0].Conditions[0].Status).To(gomega.Equal(metav1.ConditionTrue))
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
	g.Expect(childNode1.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode1.DefaultPoolGroup).NotTo(gomega.Equal(""))
	g.Expect(childNode1.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolRefs[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(len(childNode1.PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(childNode1.VHMatches)).To(gomega.Equal(1))
	g.Expect(*childNode1.VHMatches[0].Host).To(gomega.Equal("foo-8080.com"))
	g.Expect(len(childNode1.VHMatches[0].Rules)).To(gomega.Equal(1))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/foo"))
	g.Expect(len(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(1))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))

	childNode2 := nodes[0].EvhNodes[1]
	g.Expect(childNode2.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode2.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode2.DefaultPoolGroup).NotTo(gomega.Equal(""))
	g.Expect(childNode2.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode2.PoolRefs[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(len(childNode2.PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(childNode2.VHMatches)).To(gomega.Equal(1))
	g.Expect(*childNode2.VHMatches[0].Host).To(gomega.Equal("foo-8080.com"))
	g.Expect(len(childNode2.VHMatches[0].Rules)).To(gomega.Equal(1))
	g.Expect(childNode2.VHMatches[0].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/bar"))
	g.Expect(len(childNode2.VHMatches[0].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(1))
	g.Expect(childNode2.VHMatches[0].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName2)
	akogatewayapitests.TeardownHTTPRoute(t, httpRoute1Name, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownHTTPRoute(t, httpRoute2Name, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}
func TestHttpRouteCreationBeforeGateway(t *testing.T) {
	// 1: Two HTTPRoute with  invalid gateway
	gatewayName := "gateway-hr-25"
	gatewayClassName := "gateway-class-hr-25"
	httpRouteName := "http-route-hr-25"
	svcName := "avisvc-hr-25"
	ports := []int32{8080}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)

	g := gomega.NewGomegaWithT(t)
	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, []int32{ports[0]})
	rule := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, nil)

	rules := []gatewayv1.HTTPRouteRule{rule}
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)
	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(DEFAULT_NAMESPACE).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			t.Logf("Couldn't get the HTTPRoute, err: %+v", err)
			return false
		}
		return httpRoute.Status.RouteStatus.Parents != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g.Eventually(func() bool {
		gateway, err := akogatewayapitests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	gateway, _ := akogatewayapitests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	g.Expect(gateway.Status.Listeners[0].Conditions[0].Status).To(gomega.Equal(metav1.ConditionTrue))
	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 300*time.Second, 5*time.Second).Should(gomega.Equal(1))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childNode := nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode.DefaultPoolGroup).NotTo(gomega.Equal(""))
	g.Expect(childNode.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolRefs[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(len(childNode.PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(childNode.VHMatches)).To(gomega.Equal(1))
	g.Expect(*childNode.VHMatches[0].Host).To(gomega.Equal("foo-8080.com"))
	g.Expect(len(childNode.VHMatches[0].Rules)).To(gomega.Equal(1))
	g.Expect(childNode.VHMatches[0].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/foo"))
	g.Expect(len(childNode.VHMatches[0].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(1))
	g.Expect(childNode.VHMatches[0].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHttpRouteWithPortUpdateInGateway(t *testing.T) {
	gatewayName := "gateway-hr-26"
	gatewayClassName := "gateway-class-hr-26"
	httpRouteName := "http-route-hr-26"
	svcName1 := "avisvc-hr-26-a"
	svcName2 := "avisvc-hr-26-b"
	ports := []int32{8080, 8081}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName1, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName2, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.3")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule1 := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName1, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rule2 := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/bar"}, []string{},
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
	g.Expect(childNode1.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode1.DefaultPoolGroup).NotTo(gomega.Equal(""))
	g.Expect(childNode1.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolRefs[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(len(childNode1.PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(childNode1.VHMatches)).To(gomega.Equal(2))
	g.Expect(*childNode1.VHMatches[0].Host).To(gomega.Equal("foo-8080.com"))
	g.Expect(len(childNode1.VHMatches[0].Rules)).To(gomega.Equal(1))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/foo"))
	g.Expect(len(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(2))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports[1]).To(gomega.Equal(int64(8081)))
	g.Expect(*childNode1.VHMatches[1].Host).To(gomega.Equal("foo-8081.com"))
	g.Expect(len(childNode1.VHMatches[1].Rules)).To(gomega.Equal(1))
	g.Expect(childNode1.VHMatches[1].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/foo"))
	g.Expect(len(childNode1.VHMatches[1].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(2))
	g.Expect(childNode1.VHMatches[1].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))
	g.Expect(childNode1.VHMatches[1].Rules[0].Matches.VsPort.Ports[1]).To(gomega.Equal(int64(8081)))

	childNode2 := nodes[0].EvhNodes[1]
	g.Expect(childNode2.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode2.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode2.DefaultPoolGroup).NotTo(gomega.Equal(""))
	g.Expect(childNode2.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode2.PoolRefs[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(len(childNode2.PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(childNode2.VHMatches)).To(gomega.Equal(2))
	g.Expect(*childNode2.VHMatches[0].Host).To(gomega.Equal("foo-8080.com"))
	g.Expect(len(childNode2.VHMatches[0].Rules)).To(gomega.Equal(1))
	g.Expect(childNode2.VHMatches[0].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/bar"))
	g.Expect(len(childNode2.VHMatches[0].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(2))
	g.Expect(childNode2.VHMatches[0].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))
	g.Expect(childNode2.VHMatches[0].Rules[0].Matches.VsPort.Ports[1]).To(gomega.Equal(int64(8081)))
	g.Expect(*childNode2.VHMatches[1].Host).To(gomega.Equal("foo-8081.com"))
	g.Expect(len(childNode2.VHMatches[1].Rules)).To(gomega.Equal(1))
	g.Expect(childNode2.VHMatches[1].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/bar"))
	g.Expect(len(childNode2.VHMatches[1].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(2))
	g.Expect(childNode2.VHMatches[1].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))
	g.Expect(childNode2.VHMatches[1].Rules[0].Matches.VsPort.Ports[1]).To(gomega.Equal(int64(8081)))

	listeners[1].Port = 8082
	akogatewayapitests.UpdateGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)
	g.Eventually(func() int64 {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return nodes[0].EvhNodes[0].VHMatches[1].Rules[0].Matches.VsPort.Ports[1]
	}, 50*time.Second, 5*time.Second).Should(gomega.Equal(int64(8082)))
	_, aviModel1 := objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel1.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childNode1 = nodes[0].EvhNodes[0]
	g.Expect(childNode1.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode1.DefaultPoolGroup).NotTo(gomega.Equal(""))
	g.Expect(childNode1.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode1.PoolRefs[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(len(childNode1.PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(childNode1.VHMatches)).To(gomega.Equal(2))
	g.Expect(*childNode1.VHMatches[0].Host).To(gomega.Equal("foo-8080.com"))
	g.Expect(len(childNode1.VHMatches[0].Rules)).To(gomega.Equal(1))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/foo"))
	g.Expect(len(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(2))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))
	g.Expect(childNode1.VHMatches[0].Rules[0].Matches.VsPort.Ports[1]).To(gomega.Equal(int64(8082)))
	g.Expect(*childNode1.VHMatches[1].Host).To(gomega.Equal("foo-8081.com"))
	g.Expect(len(childNode1.VHMatches[1].Rules)).To(gomega.Equal(1))
	g.Expect(childNode1.VHMatches[1].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/foo"))
	g.Expect(len(childNode1.VHMatches[1].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(2))
	g.Expect(childNode1.VHMatches[1].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))
	g.Expect(childNode1.VHMatches[1].Rules[0].Matches.VsPort.Ports[1]).To(gomega.Equal(int64(8082)))

	childNode2 = nodes[0].EvhNodes[1]
	g.Expect(childNode2.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode2.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode2.DefaultPoolGroup).NotTo(gomega.Equal(""))
	g.Expect(childNode2.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode2.PoolRefs[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(len(childNode2.PoolRefs[0].Servers)).To(gomega.Equal(1))
	g.Expect(len(childNode2.VHMatches)).To(gomega.Equal(2))
	g.Expect(*childNode2.VHMatches[0].Host).To(gomega.Equal("foo-8080.com"))
	g.Expect(len(childNode2.VHMatches[0].Rules)).To(gomega.Equal(1))
	g.Expect(childNode2.VHMatches[0].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/bar"))
	g.Expect(len(childNode2.VHMatches[0].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(2))
	g.Expect(childNode2.VHMatches[0].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))
	g.Expect(childNode2.VHMatches[0].Rules[0].Matches.VsPort.Ports[1]).To(gomega.Equal(int64(8082)))
	g.Expect(*childNode2.VHMatches[1].Host).To(gomega.Equal("foo-8081.com"))
	g.Expect(len(childNode2.VHMatches[1].Rules)).To(gomega.Equal(1))
	g.Expect(childNode2.VHMatches[1].Rules[0].Matches.Path.MatchStr[0]).To(gomega.Equal("/bar"))
	g.Expect(len(childNode2.VHMatches[1].Rules[0].Matches.VsPort.Ports)).To(gomega.Equal(2))
	g.Expect(childNode2.VHMatches[1].Rules[0].Matches.VsPort.Ports[0]).To(gomega.Equal(int64(8080)))
	g.Expect(childNode2.VHMatches[1].Rules[0].Matches.VsPort.Ports[1]).To(gomega.Equal(int64(8082)))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName2)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteFilterCRUDWithNSXT(t *testing.T) {

	// simulate nsx-t cloud by setting up t1lr
	os.Setenv("NSXT_T1_LR", "/infra/t1lr/sample")
	gatewayName := "gateway-hr-26"
	gatewayClassName := "gateway-class-hr-26"
	httpRouteName := "http-route-hr-26"
	svcName1 := "avisvc-hr-26"

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

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName1, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
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
	g.Expect(childNode.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode.DefaultPoolGroup).NotTo(gomega.Equal(""))

	// Check t1lr set at vsvip, pool. vrf empty at vsvip, pool and vs
	g.Expect(nodes[0].VSVIPRefs[0].T1Lr).Should(gomega.Equal("/infra/t1lr/sample"))
	g.Expect(nodes[0].VSVIPRefs[0].VrfContext).Should(gomega.Equal(""))
	g.Expect(childNode.PoolRefs[0].T1Lr).Should(gomega.Equal("/infra/t1lr/sample"))
	g.Expect(childNode.PoolRefs[0].VrfContext).Should(gomega.Equal(""))
	g.Expect(nodes[0].VrfContext).Should(gomega.Equal(""))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")

	//reset the field
	os.Unsetenv("NSXT_T1_LR")
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}
