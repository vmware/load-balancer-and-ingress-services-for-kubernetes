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
	"testing"
	"time"

	"github.com/onsi/gomega"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	akogatewayapitests "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/tests"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
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
	listeners := akogatewayapitests.GetListenersV1Beta1(ports)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	parentRefs := akogatewayapitests.GetParentReferencesV1Beta1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1Beta1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}})
	rules := []gatewayv1beta1.HTTPRouteRule{rule}
	hostnames := []gatewayv1beta1.Hostname{"foo-8080.com"}
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

	rule = akogatewayapitests.GetHTTPRouteRuleV1Beta1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}})
	rules = []gatewayv1beta1.HTTPRouteRule{rule}
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

	gatewayName := "gateway-hrr-01"
	gatewayClassName := "gateway-class-hrr-01"
	httpRouteName := "http-route-hrr-01"
	ports := []int32{8080}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1Beta1(ports)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	parentRefs := akogatewayapitests.GetParentReferencesV1Beta1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	ruleWithoutCanary := akogatewayapitests.GetHTTPRouteRuleV1Beta1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}})
	ruleWithCanary := akogatewayapitests.GetHTTPRouteRuleV1Beta1([]string{"/bar"}, []string{"canary"},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}})
	rules := []gatewayv1beta1.HTTPRouteRule{ruleWithCanary, ruleWithoutCanary}
	hostnames := []gatewayv1beta1.Hostname{"foo-8080.com"}
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
	rules = []gatewayv1beta1.HTTPRouteRule{ruleWithCanary}
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
	rules = []gatewayv1beta1.HTTPRouteRule{ruleWithCanary, ruleWithoutCanary}
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

	gatewayName := "gateway-hrf-01"
	gatewayClassName := "gateway-class-hrf-01"
	httpRouteName := "http-route-hrf-01"
	ports := []int32{8080}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1Beta1(ports)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	parentRefs := akogatewayapitests.GetParentReferencesV1Beta1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1Beta1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}})
	rules := []gatewayv1beta1.HTTPRouteRule{rule}
	hostnames := []gatewayv1beta1.Hostname{"foo-8080.com"}
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
	rule = akogatewayapitests.GetHTTPRouteRuleV1Beta1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"replace"}})
	rules = []gatewayv1beta1.HTTPRouteRule{rule}
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

	gatewayName := "gateway-hrf-02"
	gatewayClassName := "gateway-class-hrf-02"
	httpRouteName := "http-route-hrf-02"
	ports := []int32{8080}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1Beta1(ports)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	parentRefs := akogatewayapitests.GetParentReferencesV1Beta1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1Beta1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "replace", "remove"}})
	rules := []gatewayv1beta1.HTTPRouteRule{rule}
	hostnames := []gatewayv1beta1.Hostname{"foo-8080.com"}
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
	rule = akogatewayapitests.GetHTTPRouteRuleV1Beta1([]string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"replace", "remove"}})
	rules = []gatewayv1beta1.HTTPRouteRule{rule}
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

	gatewayName := "gateway-hrf-03"
	gatewayClassName := "gateway-class-hrf-03"
	httpRouteName := "http-route-hrf-03"
	ports := []int32{8080}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1Beta1(ports)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	parentRefs := akogatewayapitests.GetParentReferencesV1Beta1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1Beta1([]string{"/foo"}, []string{},
		map[string][]string{"ResponseHeaderModifier": {"add", "replace", "remove"}})
	rules := []gatewayv1beta1.HTTPRouteRule{rule}
	hostnames := []gatewayv1beta1.Hostname{"foo-8080.com"}
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
	rule = akogatewayapitests.GetHTTPRouteRuleV1Beta1([]string{"/foo"}, []string{},
		map[string][]string{"ResponseHeaderModifier": {"replace", "remove"}})
	rules = []gatewayv1beta1.HTTPRouteRule{rule}
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

	gatewayName := "gateway-hrf-04"
	gatewayClassName := "gateway-class-hrf-04"
	httpRouteName := "http-route-hrf-04"
	ports := []int32{8080}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1Beta1(ports)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	parentRefs := akogatewayapitests.GetParentReferencesV1Beta1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1Beta1([]string{"/foo"}, []string{},
		map[string][]string{"RequestRedirect": {}})
	rules := []gatewayv1beta1.HTTPRouteRule{rule}
	hostnames := []gatewayv1beta1.Hostname{"foo-8080.com"}
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
	rule = akogatewayapitests.GetHTTPRouteRuleV1Beta1([]string{"/foo"}, []string{},
		map[string][]string{})
	rules = []gatewayv1beta1.HTTPRouteRule{rule}
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
