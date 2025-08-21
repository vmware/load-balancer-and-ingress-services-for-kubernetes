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

package graphlayer

import (
	"context"
	"fmt"
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
	integrationtest.CreateEPS(t, "default", "avisvc", false, false, "1.1.1")

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
	rule := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
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
	rule = akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
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
	integrationtest.CreateEPS(t, "default", svcName, false, false, "1.1.1")
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
	rule := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
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
	g.Expect(nodes[0].EvhNodes[0].AviMarkers.GatewayNamespace).To(gomega.Equal(namespace))
	g.Expect(nodes[0].EvhNodes[0].AviMarkers.HTTPRouteName).To(gomega.Equal(httpRouteName))
	g.Expect(nodes[0].EvhNodes[0].AviMarkers.HTTPRouteNamespace).To(gomega.Equal(namespace))
	g.Expect(nodes[0].EvhNodes[0].PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs).To(gomega.HaveLen(1))

	integrationtest.DelSVC(t, namespace, svcName)
	integrationtest.DelEPS(t, namespace, svcName)
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
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
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

	rule = akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
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
	rule = akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
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
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)
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
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
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
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return -1
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes[0].PoolGroupRefs)
	}, 30*time.Second).Should(gomega.Equal(0))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

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
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)
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
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
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
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.3")

	rule = akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
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
	rule = akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
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
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")
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
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName, true, true, "1.2.3")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
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
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)
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
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
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
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)
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
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")
	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)

	listeners := akogatewayapitests.GetListenersV1(ports, true, false)

	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 5*time.Second).Should(gomega.Equal(true))

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{}, nil,
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
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)
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
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	// httproute parent ref
	parentRefs := akogatewayapitests.GetParentReferencesV1WithGatewayNameOnly([]string{gatewayName}, DEFAULT_NAMESPACE)

	// httproute rule 1
	rule1 := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{}, nil,
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{rule1}

	// httproute rule2
	rule2 := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/bar"}, []string{}, nil,
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
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)
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
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")
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
	rule1 := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{}, nil,
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rule2 := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/bar"}, []string{}, nil,
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
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)
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
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")
	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersOnHostname([]string{"*.avi.internal", "hello.avi.internal"})

	akogatewayapitests.SetupGateway(t, gatewayName1, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 45*time.Second).Should(gomega.Equal(true))

	parentRefs := akogatewayapitests.GetParentReferencesFromListeners(listeners, gatewayName1, DEFAULT_NAMESPACE)
	rule1 := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{}, nil,
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
	rule1 = akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{}, nil,
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
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)
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
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName2, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.3")

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
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName2)
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
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	parentRefs1 := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, []int32{ports[0]})
	rule1 := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rules1 := []gatewayv1.HTTPRouteRule{rule1}
	hostnames1 := []gatewayv1.Hostname{"foo-8080.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName1, DEFAULT_NAMESPACE, parentRefs1, hostnames1, rules1)

	parentRefs2 := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, []int32{ports[1]})
	rule2 := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/bar"}, []string{},
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
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)
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
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName2, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.3")

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
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName2)
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
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName2, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.3")

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
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName2)
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
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName2, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.3")

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
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName2)
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
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName2, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.3")

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
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName2)
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
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName2, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.3")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, []int32{ports[0]})
	rule1 := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName1, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rule2 := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/bar"}, []string{},
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
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName2)
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
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName2, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.3")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, []int32{ports[0]})
	rule1 := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName1, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rule2 := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/bar"}, []string{},
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
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName2)
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
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, []int32{ports[0]})
	rule := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
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
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)
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
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName2, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.3")

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
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName2)
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
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
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
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")

	//reset the field
	os.Unsetenv("NSXT_T1_LR")
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteBackendServiceInvalidType(t *testing.T) {

	gatewayName := "gateway-hr-27"
	gatewayClassName := "gateway-class-hr-27"
	httpRouteName := "http-route-hr-27"
	svcName1 := "avisvc-hr-27"

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

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName1, "TCP", corev1.ServiceTypeExternalName, false)

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{},
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
	g.Expect(childNode.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolRefs[0].Servers).To(gomega.HaveLen(0))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName1)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestSecretCreateDeleteWithHTTPRoute(t *testing.T) {

	/*
		1. Create Secret, GW, HTTPRoute. Model should be present with all child values.
		2. Delete secret. model should be nil (Parent VS and child VS not present)
		3. Create secret again. Model should be present with all child values.
	*/
	gatewayName := "gateway-hr-27"
	gatewayClassName := "gateway-class-hr-27"
	httpRouteName := "http-route-hr-27"
	svcName1 := "avisvc-hr-27"

	ports := []int32{8080}
	secrets := []string{"secret-27"}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	g := gomega.NewGomegaWithT(t)
	integrationtest.AddSecret(secrets[0], DEFAULT_NAMESPACE, "cert", "key")
	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports, false, false, secrets...)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName1, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")
	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName1, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{rule}
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(1))

	childNode := nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode.DefaultPoolGroup).NotTo(gomega.Equal(""))

	// Delete secret
	integrationtest.DeleteSecret(secrets[0], DEFAULT_NAMESPACE)
	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if found {
			return aviModel != nil
		}
		return found
	}, 30*time.Second).Should(gomega.Equal(false))

	// again add the certificate
	integrationtest.AddSecret(secrets[0], DEFAULT_NAMESPACE, "cert", "key")
	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if found {
			return aviModel == nil
		}
		return found
	}, 30*time.Second).Should(gomega.Equal(false))

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	childNode = nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolGroupRefs[0].Members).To(gomega.HaveLen(1))
	g.Expect(childNode.DefaultPoolGroup).NotTo(gomega.Equal(""))

	// Delete secret
	integrationtest.DeleteSecret(secrets[0], DEFAULT_NAMESPACE)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")

	//reset the field
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHttpRouteWithDifferentGatewayController(t *testing.T) {
	gatewayName1 := "gateway-hr-28a"
	gatewayName2 := "gateway-hr-28b"
	gatewayName3 := "gateway-hr-28c"
	gatewayClassName1 := "gateway-class-hr-28a"
	gatewayClassName2 := "gateway-class-hr-28b"

	httpRouteName := "http-route-hr-28"
	svcName1 := "avisvc-hr-28-a"
	svcName2 := "avisvc-hr-28-b"
	ports := []int32{8080, 8081}

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName1, akogatewayapilib.GatewayController)
	akogatewayapitests.SetupGatewayClass(t, gatewayClassName2, "notavi-lb")

	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	listeners[0].Protocol = "TCP"
	akogatewayapitests.SetupGateway(t, gatewayName1, DEFAULT_NAMESPACE, gatewayClassName1, nil, listeners[:1])
	listeners[0].Protocol = "HTTP"
	akogatewayapitests.SetupGateway(t, gatewayName2, DEFAULT_NAMESPACE, gatewayClassName2, nil, listeners[:1])

	g := gomega.NewGomegaWithT(t)

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName1, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName2, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.3")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName1, gatewayName2, gatewayName3}, DEFAULT_NAMESPACE, ports)
	parentRefs = []gatewayv1.ParentReference{parentRefs[0], parentRefs[2], parentRefs[5]}
	rule1 := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName1, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rule2 := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/bar"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName2, DEFAULT_NAMESPACE, "8081", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{rule1, rule2}
	hostnames := []gatewayv1.Hostname{"foo-8081.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)
	time.Sleep(10 * time.Second)
	akogatewayapitests.SetupGateway(t, gatewayName3, DEFAULT_NAMESPACE, gatewayClassName1, nil, listeners[1:2])
	modelName3, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName3)
	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName3)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 50*time.Second, 5*time.Second).Should(gomega.Equal(2))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName3)
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
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName2)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName1, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName2, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName3, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName1)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName2)
}

func TestHTTPRouteCRUDWithRegexPath(t *testing.T) {

	gatewayName := "gateway-hr-29"
	gatewayClassName := "gateway-class-hr-29"
	httpRouteName := "http-route-hr-29"
	svcName := "svc-29"
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
		Name:         svcName,
		Namespace:    "default",
		Type:         corev1.ServiceTypeClusterIP,
		ServicePorts: []integrationtest.Serviceport{{PortName: "foo", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()

	_, err := akogatewayapitests.KubeClient.CoreV1().Services("default").Create(context.TODO(), svcExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	integrationtest.CreateEPS(t, "default", svcName, false, false, "1.1.1")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.REGULAREXPRESSION, []string{"/foo/[a-z]+/bar"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName, "default", "8080", "1"}}, nil)
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
	g.Expect(childNode.VHMatches[0].Rules[0].Matches.Path.MatchStr).To(gomega.ContainElement("/foo/[a-z]+/bar"))
	g.Expect(*childNode.VHMatches[0].Rules[0].Matches.Path.MatchCriteria).To(gomega.Equal("REGEX_MATCH"))

	rule = akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName, "default", "8080", "1"}}, nil)
	rules = []gatewayv1.HTTPRouteRule{rule}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return false
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes) != 1 {
			return false
		}
		childNode = nodes[0].EvhNodes[0]
		return *childNode.VHMatches[0].Rules[0].Matches.Path.MatchCriteria == "BEGINS_WITH"
	}, 25*time.Second).Should(gomega.Equal(true))

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
func TestHTTPRouteFilterWithUrlRewrite(t *testing.T) {
	gatewayName := "gateway-hr-30"
	gatewayClassName := "gateway-class-hr-30"
	httpRouteName := "http-route-hr-30"
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
		map[string][]string{"URLRewrite": {}},
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
		if len(nodes[0].EvhNodes) == 1 {
			childVS := nodes[0].EvhNodes[0]
			if len(childVS.HttpPolicyRefs) != 1 {
				return -1
			}
			return len(childVS.HttpPolicyRefs[0].RequestRules)
		}
		return -1

	}, 25*time.Second).Should(gomega.Equal(1))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childVS := nodes[0].EvhNodes[0]

	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].Name).To(gomega.Equal(childVS.Name))
	g.Expect(childVS.HttpPolicyRefs[0].RequestRules[0].RewriteURLAction).ShouldNot(gomega.BeNil())
	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].RewriteURLAction.HostHdr.Tokens[0].StrValue).To(gomega.Equal("rewrite.com"))
	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].RewriteURLAction.Path.Tokens[0].StrValue).To(gomega.Equal("bar"))

	//update httproute and combine requestRewrite with other requestHeader modifiers except requestRedirect
	rule = akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"URLRewrite": {}, "RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{"avisvc", "default", "8080", "1"}}, nil)
	rules = []gatewayv1.HTTPRouteRule{rule}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes) == 1 {
			childVS := nodes[0].EvhNodes[0]
			if len(childVS.HttpPolicyRefs) != 1 {
				return -1
			}
			return len(childVS.HttpPolicyRefs[0].RequestRules[0].HdrAction)
		}
		return -1
	}, 25*time.Second).Should(gomega.Equal(3))

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childVS = nodes[0].EvhNodes[0]

	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].Name).To(gomega.Equal(childVS.Name))
	g.Expect(childVS.HttpPolicyRefs[0].RequestRules[0].RewriteURLAction).ShouldNot(gomega.BeNil())
	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].RewriteURLAction.HostHdr.Tokens[0].StrValue).To(gomega.Equal("rewrite.com"))
	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].RewriteURLAction.Path.Tokens[0].StrValue).To(gomega.Equal("bar"))
	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].HdrAction[0].Action).To(gomega.Equal("HTTP_ADD_HDR"))
	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].HdrAction[0].Hdr.Name).To(gomega.Equal("new-header"))
	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].HdrAction[0].Hdr.Value.Val).To(gomega.Equal("any-value"))
	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].HdrAction[0].HdrIndex).To(gomega.Equal(uint32(0)))
	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].HdrAction[1].Action).To(gomega.Equal("HTTP_REPLACE_HDR"))
	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].HdrAction[1].Hdr.Name).To(gomega.Equal("my-header"))
	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].HdrAction[1].Hdr.Value.Val).To(gomega.Equal("any-value"))
	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].HdrAction[1].HdrIndex).To(gomega.Equal(uint32(1)))
	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].HdrAction[2].Action).To(gomega.Equal("HTTP_REMOVE_HDR"))
	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].HdrAction[2].Hdr.Name).To(gomega.Equal("old-header"))
	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].HdrAction[2].HdrIndex).To(gomega.Equal(uint32(2)))

	//update httproute to have only rewrite filter
	rule = akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"URLRewrite": {}},
		[][]string{{"avisvc", "default", "8080", "1"}}, nil)
	rules = []gatewayv1.HTTPRouteRule{rule}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes) == 1 {
			childVS := nodes[0].EvhNodes[0]
			if len(childVS.HttpPolicyRefs) != 1 {
				return -1
			}
			return len(childVS.HttpPolicyRefs[0].RequestRules[0].HdrAction)
		}
		return -1
	}, 25*time.Second).Should(gomega.Equal(0))

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childVS = nodes[0].EvhNodes[0]

	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].Name).To(gomega.Equal(childVS.Name))
	g.Expect(childVS.HttpPolicyRefs[0].RequestRules[0].HdrAction).Should(gomega.BeNil())
	g.Expect(childVS.HttpPolicyRefs[0].RequestRules[0].RewriteURLAction).ShouldNot(gomega.BeNil())
	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].RewriteURLAction.HostHdr.Tokens[0].StrValue).To(gomega.Equal("rewrite.com"))
	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].RewriteURLAction.Path.Tokens[0].StrValue).To(gomega.Equal("bar"))

	// update httproute to remove all the filters
	rule = akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
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

func TestHTTPRouteFilterWithUrlRewriteOnlyHostnameOrPath(t *testing.T) {
	gatewayName := "gateway-hr-30"
	gatewayClassName := "gateway-class-hr-30"
	httpRouteName := "http-route-hr-30"
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
		map[string][]string{"URLRewrite": {}},
		[][]string{{"avisvc", "default", "8080", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{rule}

	//Setting the hostname to be nil
	rules[0].Filters[0].URLRewrite.Hostname = nil
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found || aviModel == nil {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes) == 1 {
			childVS := nodes[0].EvhNodes[0]
			if len(childVS.HttpPolicyRefs) != 1 {
				return -1
			}
			return len(childVS.HttpPolicyRefs[0].RequestRules)

		}
		return -1
	}, 25*time.Second).Should(gomega.Equal(1))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childVS := nodes[0].EvhNodes[0]

	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].Name).To(gomega.Equal(childVS.Name))
	g.Expect(childVS.HttpPolicyRefs[0].RequestRules[0].RewriteURLAction).ShouldNot(gomega.BeNil())
	g.Expect(childVS.HttpPolicyRefs[0].RequestRules[0].RewriteURLAction.HostHdr).To(gomega.BeNil())
	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].RewriteURLAction.Path.Tokens[0].StrValue).To(gomega.Equal("bar"))

	//Setting the path to be nil
	rules[0].Filters[0].URLRewrite.Path = nil
	host := "rewrite.com"
	rules[0].Filters[0].URLRewrite.Hostname = (*gatewayv1.PreciseHostname)(&host)
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return false
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes) == 1 {
			childVS := nodes[0].EvhNodes[0]
			if len(childVS.HttpPolicyRefs) != 1 {
				return false
			}
			return childVS.HttpPolicyRefs[0].RequestRules[0].RewriteURLAction != nil && childVS.HttpPolicyRefs[0].RequestRules[0].RewriteURLAction.Path == nil
		}
		return false
	}, 25*time.Second).Should(gomega.Equal(true))

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childVS = nodes[0].EvhNodes[0]

	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].Name).To(gomega.Equal(childVS.Name))
	g.Expect(childVS.HttpPolicyRefs[0].RequestRules[0].RewriteURLAction).ShouldNot(gomega.BeNil())
	g.Expect(*childVS.HttpPolicyRefs[0].RequestRules[0].RewriteURLAction.HostHdr.Tokens[0].StrValue).To(gomega.Equal("rewrite.com"))
	g.Expect(childVS.HttpPolicyRefs[0].RequestRules[0].RewriteURLAction.Path).To(gomega.BeNil())
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	// update httproute to remove all the filters
	rule = akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
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

func TestHTTPRouteWithSessionPersistence(t *testing.T) {
	gatewayClassName := "gateway-class-hr-32"
	gatewayName := "gateway-hr-32"
	httpRouteName := "httproute-32"
	namespace := "default"
	svcName := "avisvc-hr-32"
	ports := []int32{8080}
	ruleName := "sticky-rule"
	cookieName := "sticky-cookie"
	var timeout gatewayv1.Duration = "10m"

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, "default", svcName, false, false, "1.1.1")
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
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}
	rule := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		nil,
		[][]string{{svcName, namespace, "8080", "1"}}, nil)

	// Add SessionPersistence
	rule.Name = (*gatewayv1.SectionName)(&ruleName)
	cookieType := gatewayv1.CookieBasedSessionPersistence
	rule.SessionPersistence = &gatewayv1.SessionPersistence{
		Type:            &cookieType,
		SessionName:     &cookieName,
		AbsoluteTimeout: &timeout,
	}

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
		return apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	modelName := lib.GetModelName(lib.GetTenant(), akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName))

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found || aviModel == nil {
			return false
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes) == 0 || len(nodes[0].EvhNodes) == 0 || len(nodes[0].EvhNodes[0].PoolRefs) == 0 {
			return false
		}
		return nodes[0].EvhNodes[0].PoolRefs[0].ApplicationPersistenceProfile != nil
	}, 25*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes).To(gomega.HaveLen(1))

	poolNode := nodes[0].EvhNodes[0].PoolRefs[0]
	g.Expect(poolNode.ApplicationPersistenceProfile).NotTo(gomega.BeNil())

	appPersistProfile := poolNode.ApplicationPersistenceProfile
	g.Expect(appPersistProfile.PersistenceType).To(gomega.Equal("PERSISTENCE_TYPE_HTTP_COOKIE"))
	g.Expect(appPersistProfile.HTTPCookiePersistenceProfile).NotTo(gomega.BeNil())
	g.Expect(appPersistProfile.HTTPCookiePersistenceProfile.CookieName).To(gomega.Equal(cookieName))
	g.Expect(*appPersistProfile.HTTPCookiePersistenceProfile.Timeout).To(gomega.Equal(int32(10))) // 10m -> 10 minutes
	g.Expect(*appPersistProfile.HTTPCookiePersistenceProfile.IsPersistentCookie).To(gomega.BeFalse())

	// Remove session persistence
	rule.SessionPersistence = nil
	rules = []gatewayv1.HTTPRouteRule{rule}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, rules)

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return nodes[0].EvhNodes[0].PoolRefs[0].ApplicationPersistenceProfile == nil
	}, 25*time.Second).Should(gomega.Equal(true))

	// Teardown
	integrationtest.DelSVC(t, namespace, svcName)
	integrationtest.DelEPS(t, namespace, svcName)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	akogatewayapitests.TeardownGateway(t, gatewayName, namespace)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteWithRouteRuleName(t *testing.T) {
	gatewayClassName := "gateway-class-hr-31"
	gatewayName := "gateway-hr-31"
	httpRouteName := "httproute-31"
	namespace := "default"
	svcName := "avisvc-hr-31"
	ports := []int32{8080, 8081}
	ruleName := "rulename-port-8080"

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, "default", svcName, false, false, "1.1.1")
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
	rule1 := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName, namespace, "8080", "1"}}, nil)
	rule2 := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/bar"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{svcName, namespace, "8081", "1"}}, nil)
	rule1.Name = (*gatewayv1.SectionName)(&ruleName)
	rules := []gatewayv1.HTTPRouteRule{rule1, rule2}
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
		return len(nodes[0].VSVIPRefs[0].FQDNs) > 1 && len(nodes[0].EvhNodes[0].PoolGroupRefs) > 0

	}, 25*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(2))
	g.Expect(nodes[0].SSLKeyCertRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].VSVIPRefs).To(gomega.HaveLen(1))

	g.Expect(nodes[0].VSVIPRefs[0].FQDNs).To(gomega.Equal([]string{
		"foo-8080.com", "foo-8081.com"}))
	g.Expect(nodes[0].EvhNodes).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].Name).To(gomega.Equal("ako-gw-cluster--79c80596b5e4aeb72b3ea5dcf831623412368b17"))
	g.Expect(nodes[0].EvhNodes[0].AviMarkers.GatewayName).To(gomega.Equal(gatewayName))
	g.Expect(nodes[0].EvhNodes[0].AviMarkers.GatewayNamespace).To(gomega.Equal(namespace))
	g.Expect(nodes[0].EvhNodes[0].AviMarkers.HTTPRouteName).To(gomega.Equal(httpRouteName))
	g.Expect(nodes[0].EvhNodes[0].AviMarkers.HTTPRouteNamespace).To(gomega.Equal(namespace))
	g.Expect(nodes[0].EvhNodes[0].AviMarkers.HTTPRouteRuleName).To(gomega.Equal(ruleName))

	g.Expect(nodes[0].EvhNodes[0].PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].PoolGroupRefs[0].Name).To(gomega.Equal("ako-gw-cluster--79c80596b5e4aeb72b3ea5dcf831623412368b17"))
	g.Expect(nodes[0].EvhNodes[0].PoolGroupRefs[0].AviMarkers.GatewayName).To(gomega.Equal(gatewayName))
	g.Expect(nodes[0].EvhNodes[0].PoolGroupRefs[0].AviMarkers.GatewayNamespace).To(gomega.Equal(namespace))
	g.Expect(nodes[0].EvhNodes[0].PoolGroupRefs[0].AviMarkers.HTTPRouteName).To(gomega.Equal(httpRouteName))
	g.Expect(nodes[0].EvhNodes[0].PoolGroupRefs[0].AviMarkers.HTTPRouteNamespace).To(gomega.Equal(namespace))
	g.Expect(nodes[0].EvhNodes[0].PoolGroupRefs[0].AviMarkers.HTTPRouteRuleName).To(gomega.Equal(ruleName))

	g.Expect(nodes[0].EvhNodes[0].PoolRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].Name).To(gomega.Equal("ako-gw-cluster--50036c244a7b7711b181d50d5ff1d7c8cc143d83"))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].AviMarkers.GatewayName).To(gomega.Equal(gatewayName))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].AviMarkers.GatewayNamespace).To(gomega.Equal(namespace))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].AviMarkers.HTTPRouteName).To(gomega.Equal(httpRouteName))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].AviMarkers.HTTPRouteNamespace).To(gomega.Equal(namespace))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].AviMarkers.HTTPRouteRuleName).To(gomega.Equal(ruleName))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].AviMarkers.BackendName).To(gomega.Equal(svcName))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].AviMarkers.BackendNs).To(gomega.Equal(namespace))

	g.Expect(nodes[0].EvhNodes[1].Name).To(gomega.Equal("ako-gw-cluster--20175aa4f7283a01de44b8bd0a39f0f2df07d44a"))
	g.Expect(nodes[0].EvhNodes[1].AviMarkers.GatewayName).To(gomega.Equal(gatewayName))
	g.Expect(nodes[0].EvhNodes[1].AviMarkers.GatewayNamespace).To(gomega.Equal(namespace))
	g.Expect(nodes[0].EvhNodes[1].AviMarkers.HTTPRouteName).To(gomega.Equal(httpRouteName))
	g.Expect(nodes[0].EvhNodes[1].AviMarkers.HTTPRouteNamespace).To(gomega.Equal(namespace))

	g.Expect(nodes[0].EvhNodes[1].PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[1].PoolGroupRefs[0].Name).To(gomega.Equal("ako-gw-cluster--20175aa4f7283a01de44b8bd0a39f0f2df07d44a"))
	g.Expect(nodes[0].EvhNodes[1].PoolGroupRefs[0].AviMarkers.GatewayName).To(gomega.Equal(gatewayName))
	g.Expect(nodes[0].EvhNodes[1].PoolGroupRefs[0].AviMarkers.GatewayNamespace).To(gomega.Equal(namespace))
	g.Expect(nodes[0].EvhNodes[1].PoolGroupRefs[0].AviMarkers.HTTPRouteName).To(gomega.Equal(httpRouteName))
	g.Expect(nodes[0].EvhNodes[1].PoolGroupRefs[0].AviMarkers.HTTPRouteNamespace).To(gomega.Equal(namespace))

	g.Expect(nodes[0].EvhNodes[1].PoolRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].EvhNodes[1].PoolRefs[0].Name).To(gomega.Equal("ako-gw-cluster--21618dd778e3c4fccaa3ae45ae085ecd39c19c22"))
	g.Expect(nodes[0].EvhNodes[1].PoolRefs[0].AviMarkers.GatewayName).To(gomega.Equal(gatewayName))
	g.Expect(nodes[0].EvhNodes[1].PoolRefs[0].AviMarkers.GatewayNamespace).To(gomega.Equal(namespace))
	g.Expect(nodes[0].EvhNodes[1].PoolRefs[0].AviMarkers.HTTPRouteName).To(gomega.Equal(httpRouteName))
	g.Expect(nodes[0].EvhNodes[1].PoolRefs[0].AviMarkers.HTTPRouteNamespace).To(gomega.Equal(namespace))
	g.Expect(nodes[0].EvhNodes[1].PoolRefs[0].AviMarkers.BackendName).To(gomega.Equal(svcName))
	g.Expect(nodes[0].EvhNodes[1].PoolRefs[0].AviMarkers.BackendNs).To(gomega.Equal(namespace))

	integrationtest.DelSVC(t, namespace, svcName)
	integrationtest.DelEPS(t, namespace, svcName)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	akogatewayapitests.TeardownGateway(t, gatewayName, namespace)
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
	}, 60*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

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
		return len(nodes[0].EvhNodes[0].PoolRefs[0].HealthMonitorRefs)
	}, 25*time.Second).Should(gomega.Equal(0))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)
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
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 60*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

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

	// Delete one HealthMonitor
	akogatewayapitests.DeleteHealthMonitorCRD(t, healthMonitorName1, DEFAULT_NAMESPACE)

	// this will remove the first healthmonitor from the poolref
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes[0].PoolRefs[0].HealthMonitorRefs)
	}, 25*time.Second).Should(gomega.Equal(1))

	g.Expect(childNode.PoolRefs[0].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm2"))

	akogatewayapitests.DeleteHealthMonitorCRD(t, healthMonitorName2, DEFAULT_NAMESPACE)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)
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
	}, 60*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

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
		return -1
	}, 25*time.Second).Should(gomega.Equal(0))

	akogatewayapitests.DeleteHealthMonitorCRD(t, healthMonitorName, DEFAULT_NAMESPACE)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)
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
	}, 60*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

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
			return -1
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(1))

	// this will remove the first healthmonitor from the poolref
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes[0].PoolRefs[0].HealthMonitorRefs)
	}, 25*time.Second).Should(gomega.Equal(0))

	akogatewayapitests.DeleteHealthMonitorCRD(t, healthMonitorName, DEFAULT_NAMESPACE)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteWithHealthMonitorMultipleRules(t *testing.T) {
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
	}, 60*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName1, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")
	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName2, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.4")

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
		return len(nodes[0].EvhNodes[0].PoolRefs[0].HealthMonitorRefs)
	}, 25*time.Second).Should(gomega.Equal(0))

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes[1].PoolRefs[0].HealthMonitorRefs)
	}, 25*time.Second).Should(gomega.Equal(1))

	g.Expect(childNode2.PoolRefs[0].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm2"))
	// Clean up
	akogatewayapitests.DeleteHealthMonitorCRD(t, healthMonitorName2, DEFAULT_NAMESPACE)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName2)
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
	}, 60*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName1, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")
	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName2, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.4")

	// Create HealthMonitor CRDs
	akogatewayapitests.CreateHealthMonitorCRD(t, healthMonitorName1, DEFAULT_NAMESPACE, "thisisaviref-hm1")
	akogatewayapitests.CreateHealthMonitorCRD(t, healthMonitorName2, DEFAULT_NAMESPACE, "thisisaviref-hm2")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)

	// Create HTTPRoute with two rules, each having different backends and HealthMonitors
	rule1 := akogatewayapitests.GetHTTPRouteRuleWithHealthMonitorFilters(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName1, DEFAULT_NAMESPACE, "8080", "1"}, {svcName2, DEFAULT_NAMESPACE, "8080", "1"}}, []string{healthMonitorName1, healthMonitorName2})

	rules := []gatewayv1.HTTPRouteRule{rule1}
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

	// Verify pools have their respective HealthMonitors
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	childNode1 := nodes[0].EvhNodes[0]
	g.Expect(childNode1.PoolRefs).To(gomega.HaveLen(2))

	g.Expect(childNode1.PoolRefs[0].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm1"))
	g.Expect(childNode1.PoolRefs[0].HealthMonitorRefs[1]).To(gomega.ContainSubstring("thisisaviref-hm2"))
	g.Expect(childNode1.PoolRefs[1].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm1"))
	g.Expect(childNode1.PoolRefs[1].HealthMonitorRefs[1]).To(gomega.ContainSubstring("thisisaviref-hm2"))

	// Delete one HealthMonitor and verify the system behavior
	akogatewayapitests.DeleteHealthMonitorCRD(t, healthMonitorName1, DEFAULT_NAMESPACE)

	// Wait for the model to be updated after HealthMonitor deletion
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes[0].PoolRefs[0].HealthMonitorRefs)
	}, 25*time.Second).Should(gomega.Equal(1))

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes[0].PoolRefs[1].HealthMonitorRefs)
	}, 25*time.Second).Should(gomega.Equal(1))

	g.Expect(childNode1.PoolRefs[0].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm2"))
	g.Expect(childNode1.PoolRefs[1].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm2"))

	// Clean up
	akogatewayapitests.DeleteHealthMonitorCRD(t, healthMonitorName2, DEFAULT_NAMESPACE)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName2)
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
	}, 60*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

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
		return len(nodes[0].EvhNodes[0].PoolRefs[0].HealthMonitorRefs)
	}, 60*time.Second).Should(gomega.Equal(0))

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
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteWithRouteBackendExtension(t *testing.T) {
	gatewayName := "gateway-rbe-01"
	gatewayClassName := "gateway-class-rbe-01"
	httpRouteName := "http-route-rbe-01"
	svcName := "avisvc-rbe-01"
	routeBackendExtensionName := "rbe-01"
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
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	// Create RouteBackendExtension CR
	rbe := akogatewayapitests.GetFakeDefaultRBEObj(routeBackendExtensionName, DEFAULT_NAMESPACE, "thisisaviref-hm1")
	rbe.CreateRouteBackendExtensionCRWithStatus(t)

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleWithRouteBackendExtensionAndHMFilters(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, []string{routeBackendExtensionName})

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

	// Verify RouteBackendExtension configured settings are present in graph layer
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childNode := nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolRefs[0].HealthMonitorRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolRefs[0].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm1"))
	g.Expect(*childNode.PoolRefs[0].LbAlgorithm).To(gomega.Equal("LB_ALGORITHM_ROUND_ROBIN"))

	// Delete routeBackendExtension and verify child VS is still present but with default settings
	rbe.DeleteRouteBackendExtensionCR(t)
	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return -1
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		childNode = nodes[0].EvhNodes[0]
		if len(childNode.PoolRefs) != 1 {
			return -1
		}
		// When RouteBackendExtension is rejected, HM should not be set in graph layer
		return len(childNode.PoolRefs[0].HealthMonitorRefs)
	}, 25*time.Second).Should(gomega.Equal(0))
	// When RouteBackendExtension is deleted, default settings should be applied
	g.Expect(childNode.PoolRefs[0].LbAlgorithm).To(gomega.BeNil())

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteWithRouteBackendExtensionsMultipleRules(t *testing.T) {
	gatewayName := "gateway-rbe-02"
	gatewayClassName := "gateway-class-rbe-02"
	httpRouteName := "http-route-rbe-02"
	svcName1 := "avisvc-rbe-02a"
	svcName2 := "avisvc-rbe-02b"
	routeBackendExtensionName1 := "rbe-02a"
	routeBackendExtensionName2 := "rbe-02b"
	ports := []int32{8080}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 60*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName1, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")
	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName2, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.4")

	// Create RouteBackendExtension CRs
	rbe1 := akogatewayapitests.GetFakeDefaultRBEObj(routeBackendExtensionName1, DEFAULT_NAMESPACE, "thisisaviref-hm1")
	rbe1.CreateRouteBackendExtensionCRWithStatus(t)
	rbe2 := akogatewayapitests.GetFakeDefaultRBEObj(routeBackendExtensionName2, DEFAULT_NAMESPACE, "thisisaviref-hm2")
	// Modifying the default object
	rbe2.LBAlgorithm = "LB_ALGORITHM_CONSISTENT_HASH"
	rbe2.LBAlgorithmHash = "LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER"
	rbe2.LBAlgorithmConsistentHashHdr = "test-header"
	rbe2.CreateRouteBackendExtensionCRWithStatus(t)

	// Create HTTPRoute with two rules, each having different backends with routeBackendExtensions
	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule1 := akogatewayapitests.GetHTTPRouteRuleWithRouteBackendExtensionAndHMFilters(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName1, DEFAULT_NAMESPACE, "8080", "1"}}, []string{routeBackendExtensionName1})
	rule2 := akogatewayapitests.GetHTTPRouteRuleWithRouteBackendExtensionAndHMFilters(integrationtest.PATHPREFIX, []string{"/bar"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName2, DEFAULT_NAMESPACE, "8080", "1"}}, []string{routeBackendExtensionName2})

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

	// Verify both pools have their respective RouteBackendExtension configured settings
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	childNode1 := nodes[0].EvhNodes[0]
	g.Expect(childNode1.PoolRefs).To(gomega.HaveLen(1))

	childNode2 := nodes[0].EvhNodes[1]
	g.Expect(childNode2.PoolRefs).To(gomega.HaveLen(1))

	g.Expect(childNode1.PoolRefs[0].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm1"))
	g.Expect(*childNode1.PoolRefs[0].LbAlgorithm).To(gomega.Equal("LB_ALGORITHM_ROUND_ROBIN"))

	g.Expect(childNode2.PoolRefs[0].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm2"))
	g.Expect(*childNode2.PoolRefs[0].LbAlgorithm).To(gomega.Equal("LB_ALGORITHM_CONSISTENT_HASH"))

	// Delete routeBackendExtension and verify it's settings are removed from graph layer
	rbe1.DeleteRouteBackendExtensionCR(t)

	// Both child nodes should still be present after one RouteBackendExtension deletion
	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return -1
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		childNode1 = nodes[0].EvhNodes[0]
		childNode2 = nodes[0].EvhNodes[1]
		if len(childNode1.PoolRefs) != 1 || len(childNode2.PoolRefs) != 1 {
			return -1
		}
		// When one RouteBackendExtension is deleted, HM should not be set in graph layer for childnode1
		return len(childNode1.PoolRefs[0].HealthMonitorRefs)
	}, 25*time.Second).Should(gomega.Equal(0))
	// When first RouteBackendExtension is deleted, default settings should be applied for childnode1
	g.Expect(childNode1.PoolRefs[0].LbAlgorithm).To(gomega.BeNil())
	// Verify childnode2 still has its settings
	g.Expect(childNode2.PoolRefs[0].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm2"))
	g.Expect(*childNode2.PoolRefs[0].LbAlgorithm).To(gomega.Equal("LB_ALGORITHM_CONSISTENT_HASH"))

	// Clean up
	rbe2.DeleteRouteBackendExtensionCR(t)
	// Both child nodes should still exist
	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(2))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName2)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteWithRouteBackendExtensionMultipleHMs(t *testing.T) {
	gatewayName := "gateway-rbe-03"
	gatewayClassName := "gateway-class-rbe-03"
	httpRouteName := "http-route-rbe-03"
	svcName := "avisvc-rbe-03"
	routeBackendExtensionName := "rbe-03"
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
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	// Create RouteBackendExtension CR
	rbe := akogatewayapitests.GetFakeDefaultRBEObj(routeBackendExtensionName, DEFAULT_NAMESPACE, "thisisaviref-hm1", "thisisaviref-hm2")
	rbe.CreateRouteBackendExtensionCRWithStatus(t)

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleWithRouteBackendExtensionAndHMFilters(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, []string{routeBackendExtensionName})

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

	// Verify RouteBackendExtension configured settings are present in graph layer
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childNode := nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolRefs[0].HealthMonitorRefs).To(gomega.HaveLen(2))
	g.Expect(childNode.PoolRefs[0].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm1"))
	g.Expect(childNode.PoolRefs[0].HealthMonitorRefs[1]).To(gomega.ContainSubstring("thisisaviref-hm2"))
	g.Expect(*childNode.PoolRefs[0].LbAlgorithm).To(gomega.Equal("LB_ALGORITHM_ROUND_ROBIN"))

	// Delete routeBackendExtension and verify it's settings are removed from graph layer
	rbe.DeleteRouteBackendExtensionCR(t)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return -1
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		childNode = nodes[0].EvhNodes[0]
		if len(childNode.PoolRefs) != 1 {
			return -1
		}
		// When RouteBackendExtension is deleted, HM should not be set in graph layer
		return len(childNode.PoolRefs[0].HealthMonitorRefs)
	}, 25*time.Second).Should(gomega.Equal(0))
	// When RouteBackendExtension is deleted, default settings should be applied
	g.Expect(childNode.PoolRefs[0].LbAlgorithm).To(gomega.BeNil())

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteWithHMAndRouteBackendExtensionMultipleRules(t *testing.T) {
	gatewayName := "gateway-rbe-04"
	gatewayClassName := "gateway-class-rbe-04"
	httpRouteName := "http-route-rbe-04"
	svcName1 := "avisvc-rbe-04a"
	svcName2 := "avisvc-rbe-04b"
	routeBackendExtensionName := "rbe-04"
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
	}, 60*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName1, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")
	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName2, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.4")

	// Create RouteBackendExtension and HealthMonitor CRs
	rbe := akogatewayapitests.GetFakeDefaultRBEObj(routeBackendExtensionName, DEFAULT_NAMESPACE, "thisisaviref-hm1")
	rbe.CreateRouteBackendExtensionCRWithStatus(t)
	akogatewayapitests.CreateHealthMonitorCRD(t, healthMonitorName, DEFAULT_NAMESPACE, "thisisaviref-hm2")

	// Create HTTPRoute with two rules, each having different backends, one with HM and the other with RBE CR specified in extensionRef
	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule1 := akogatewayapitests.GetHTTPRouteRuleWithRouteBackendExtensionAndHMFilters(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName1, DEFAULT_NAMESPACE, "8080", "1"}}, []string{routeBackendExtensionName})
	rule2 := akogatewayapitests.GetHTTPRouteRuleWithHealthMonitorFilters(integrationtest.PATHPREFIX, []string{"/bar"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName2, DEFAULT_NAMESPACE, "8080", "1"}}, []string{healthMonitorName})

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

	// Verify both pools have their respective HealthMonitor and RouteBackendExtension CR related settings
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	childNode1 := nodes[0].EvhNodes[0]
	g.Expect(childNode1.PoolRefs).To(gomega.HaveLen(1))

	childNode2 := nodes[0].EvhNodes[1]
	g.Expect(childNode2.PoolRefs).To(gomega.HaveLen(1))

	g.Expect(childNode1.PoolRefs[0].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm1"))
	g.Expect(*childNode1.PoolRefs[0].LbAlgorithm).To(gomega.Equal("LB_ALGORITHM_ROUND_ROBIN"))

	g.Expect(childNode2.PoolRefs[0].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm2"))
	g.Expect(childNode2.PoolRefs[0].LbAlgorithm).To(gomega.BeNil())

	// Delete routeBackendExtension and verify it's settings are removed from graph layer
	rbe.DeleteRouteBackendExtensionCR(t)

	// Both child nodes should still be present after RouteBackendExtension deletion
	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return -1
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		childNode1 = nodes[0].EvhNodes[0]
		childNode2 = nodes[0].EvhNodes[1]
		if len(childNode1.PoolRefs) != 1 || len(childNode2.PoolRefs) != 1 {
			return -1
		}
		// When RouteBackendExtension is deleted, HM should not be set in graph layer for childnode1
		return len(childNode1.PoolRefs[0].HealthMonitorRefs)
	}, 25*time.Second).Should(gomega.Equal(0))
	// When first RouteBackendExtension is deleted, default settings should be applied for childnode1
	g.Expect(childNode1.PoolRefs[0].LbAlgorithm).To(gomega.BeNil())
	// Verify childnode2 still has its settings
	g.Expect(childNode2.PoolRefs[0].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm2"))
	g.Expect(childNode2.PoolRefs[0].LbAlgorithm).To(gomega.BeNil())

	// Clean up
	akogatewayapitests.DeleteHealthMonitorCRD(t, healthMonitorName, DEFAULT_NAMESPACE)

	// Both child nodes should still exist
	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(2))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName2)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteWithHMAndRouteBackendExtensionSingleRule(t *testing.T) {
	gatewayName := "gateway-rbe-05"
	gatewayClassName := "gateway-class-rbe-05"
	httpRouteName := "http-route-rbe-05"
	svcName1 := "avisvc-rbe-05a"
	svcName2 := "avisvc-rbe-05b"
	routeBackendExtensionName := "rbe-05"
	healthMonitorName := "hm-05"
	ports := []int32{8080}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 60*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName1, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")
	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName2, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.4")

	// Create RouteBackendExtension and HealthMonitor CRs
	rbe := akogatewayapitests.GetFakeDefaultRBEObj(routeBackendExtensionName, DEFAULT_NAMESPACE, "thisisaviref-hm1")
	rbe.CreateRouteBackendExtensionCRWithStatus(t)
	akogatewayapitests.CreateHealthMonitorCRD(t, healthMonitorName, DEFAULT_NAMESPACE, "thisisaviref-hm2")

	// Create HTTPRoute with single rule, having single backend with HM and RBE CR specified in extensionRef
	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleWithRouteBackendExtensionAndHMFilters(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName1, DEFAULT_NAMESPACE, "8080", "1"}}, []string{routeBackendExtensionName}, healthMonitorName)

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

	// Verify the pool has the respective HealthMonitor and RouteBackendExtension settings. The hm will be set from the HM CR as it has higher priority.
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	childNode := nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolRefs).To(gomega.HaveLen(1))

	g.Expect(childNode.PoolRefs[0].HealthMonitorRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolRefs[0].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm2"))
	g.Expect(*childNode.PoolRefs[0].LbAlgorithm).To(gomega.Equal("LB_ALGORITHM_ROUND_ROBIN"))

	// Delete routeBackendExtension and verify it's settings are removed from graph layer
	rbe.DeleteRouteBackendExtensionCR(t)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return -1
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		childNode = nodes[0].EvhNodes[0]
		if len(childNode.PoolRefs) != 1 {
			return -1
		}
		// When RouteBackendExtension is deleted, LbAlgorithm should not be set in graph layer
		if childNode.PoolRefs[0].LbAlgorithm != nil {
			return -1
		}
		return 0
	}, 25*time.Second).Should(gomega.Equal(0))
	// Verify childnode still has HM set from HM CR
	g.Expect(childNode.PoolRefs[0].HealthMonitorRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolRefs[0].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm2"))

	// Clean up
	akogatewayapitests.DeleteHealthMonitorCRD(t, healthMonitorName, DEFAULT_NAMESPACE)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return -1
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		childNode = nodes[0].EvhNodes[0]
		if len(childNode.PoolRefs) != 1 {
			return -1
		}
		// When HM CR is also deleted, HM should not be set in graph layer
		return len(childNode.PoolRefs[0].HealthMonitorRefs)
	}, 25*time.Second).Should(gomega.Equal(0))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName2)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteWithRouteBackendExtensionMultipleBackends(t *testing.T) {
	gatewayName := "gateway-rbe-06"
	gatewayClassName := "gateway-class-rbe-06"
	httpRouteName := "http-route-rbe-06"
	svcName1 := "avisvc-rbe-06a"
	svcName2 := "avisvc-rbe-06b"
	routeBackendExtensionName := "rbe-06"
	ports := []int32{8080}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 60*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName1, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")
	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName2, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.4")

	// Create RouteBackendExtension CRs
	rbe := akogatewayapitests.GetFakeDefaultRBEObj(routeBackendExtensionName, DEFAULT_NAMESPACE, "thisisaviref-hm1")
	rbe.CreateRouteBackendExtensionCRWithStatus(t)

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)

	// Create HTTPRoute with single rule having two backends with routeBackendExtension specified
	rule := akogatewayapitests.GetHTTPRouteRuleWithRouteBackendExtensionAndHMFilters(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName1, DEFAULT_NAMESPACE, "8080", "1"}, {svcName2, DEFAULT_NAMESPACE, "8080", "1"}}, []string{routeBackendExtensionName})

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

	// Verify pools have their respective RouteBackendExtension settings
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	childNode := nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolRefs).To(gomega.HaveLen(2))

	g.Expect(childNode.PoolRefs[0].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm1"))
	g.Expect(*childNode.PoolRefs[0].LbAlgorithm).To(gomega.Equal("LB_ALGORITHM_ROUND_ROBIN"))
	g.Expect(childNode.PoolRefs[1].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm1"))
	g.Expect(*childNode.PoolRefs[1].LbAlgorithm).To(gomega.Equal("LB_ALGORITHM_ROUND_ROBIN"))

	// Delete routeBackendExtension and verify it's settings are removed from graph layer
	rbe.DeleteRouteBackendExtensionCR(t)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return -1
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		childNode = nodes[0].EvhNodes[0]
		if len(childNode.PoolRefs) != 2 {
			return -1
		}
		// When RouteBackendExtension is deleted, HM should not be set in graph layer
		if len(childNode.PoolRefs[0].HealthMonitorRefs) != 0 || len(childNode.PoolRefs[1].HealthMonitorRefs) != 0 {
			return -1
		}
		return 0
	}, 25*time.Second).Should(gomega.Equal(0))
	// When RouteBackendExtension is deleted, default settings should be applied
	g.Expect(childNode.PoolRefs[0].LbAlgorithm).To(gomega.BeNil())
	g.Expect(childNode.PoolRefs[1].LbAlgorithm).To(gomega.BeNil())

	// Clean up
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName2)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteWithRouteBackendExtensionCRUD(t *testing.T) {
	gatewayName := "gateway-rbe-08"
	gatewayClassName := "gateway-class-rbe-08"
	httpRouteName := "http-route-rbe-08"
	svcName := "avisvc-rbe-08"
	routeBackendExtensionName := "rbe-08"
	ports := []int32{8080}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 60*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	// Create HTTPRoute without RouteBackendExtension initially
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

	// Verify no RouteBackendExtension initially
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childNode := nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolRefs[0].HealthMonitorRefs).To(gomega.HaveLen(0))
	g.Expect(childNode.PoolRefs[0].LbAlgorithm).To(gomega.BeNil())

	// Create RouteBackendExtension CRs
	rbe := akogatewayapitests.GetFakeDefaultRBEObj(routeBackendExtensionName, DEFAULT_NAMESPACE, "thisisaviref-hm1")
	rbe.CreateRouteBackendExtensionCRWithStatus(t)

	// Update HTTPRoute with rule having routeBackendExtensionName
	rule = akogatewayapitests.GetHTTPRouteRuleWithRouteBackendExtensionAndHMFilters(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, []string{routeBackendExtensionName})
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

	// Verify RouteBackendExtension is now present
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childNode = nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolRefs[0].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm1"))
	g.Expect(*childNode.PoolRefs[0].LbAlgorithm).To(gomega.Equal("LB_ALGORITHM_ROUND_ROBIN"))

	// Remove RouteBackendExtension from HTTPRoute
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
		return -1
	}, 25*time.Second).Should(gomega.Equal(0))

	// Delete routeBackendExtension and verify it's settings are removed from graph layer
	rbe.DeleteRouteBackendExtensionCR(t)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

// TestHTTPRouteStatusWithRouteBackendExtensionStatusTransition tests the behavior when RouteBackendExtension
// status transitions between valid and invalid states.
func TestHTTPRouteStatusWithRouteBackendExtensionStatusTransition(t *testing.T) {
	gatewayName := "gateway-rbe-09"
	gatewayClassName := "gateway-class-rbe-09"
	httpRouteName := "http-route-rbe-09"
	svcName := "avisvc-rbe-09"
	routeBackendExtensionName := "rbe-09"
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
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	// Create RouteBackendExtension CR
	rbe := akogatewayapitests.GetFakeDefaultRBEObj(routeBackendExtensionName, DEFAULT_NAMESPACE, "thisisaviref-hm1")
	rbe.CreateRouteBackendExtensionCRWithStatus(t)

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleWithRouteBackendExtensionAndHMFilters(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, []string{routeBackendExtensionName})

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

	// Verify RouteBackendExtension configured settings are present in graph layer
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childNode := nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolRefs[0].HealthMonitorRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.PoolRefs[0].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm1"))
	g.Expect(*childNode.PoolRefs[0].LbAlgorithm).To(gomega.Equal("LB_ALGORITHM_ROUND_ROBIN"))

	// Test status transition from status accepted to rejected
	rbe.Status = "Rejected"
	rbe.UpdateRouteBackendExtensionStatus(t)

	// Child VS should still exist even when RouteBackendExtension is invalid
	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return -1
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		childNode = nodes[0].EvhNodes[0]
		if len(childNode.PoolRefs) != 1 {
			return -1
		}
		// When RouteBackendExtension is rejected, HM should not be set in graph layer
		return len(childNode.PoolRefs[0].HealthMonitorRefs)
	}, 25*time.Second).Should(gomega.Equal(0))

	// When RouteBackendExtension is rejected, default settings should be applied
	g.Expect(childNode.PoolRefs[0].LbAlgorithm).To(gomega.BeNil())

	// Test status transition from status rejected to accepted
	rbe.Status = "Accepted"
	rbe.UpdateRouteBackendExtensionStatus(t)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		childNode = nodes[0].EvhNodes[0]
		if len(childNode.PoolRefs) != 1 {
			return -1
		}
		// When RouteBackendExtension is accepted, HM should be set in graph layer
		return len(childNode.PoolRefs[0].HealthMonitorRefs)
	}, 25*time.Second).Should(gomega.Equal(1))

	// Verify RouteBackendExtension configured settings are present in graph layer
	g.Expect(childNode.PoolRefs[0].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm1"))
	g.Expect(*childNode.PoolRefs[0].LbAlgorithm).To(gomega.Equal("LB_ALGORITHM_ROUND_ROBIN"))

	// Test status transition from status controller valid to invalid
	rbe.Controller = "Invalid-Controller"
	rbe.UpdateRouteBackendExtensionStatus(t)

	// Child VS should still exist even when RouteBackendExtension has invalid controller
	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return -1
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		childNode = nodes[0].EvhNodes[0]
		if len(childNode.PoolRefs) != 1 {
			return -1
		}
		// When RouteBackendExtension is accepted, HM should not be set in graph layer
		return len(childNode.PoolRefs[0].HealthMonitorRefs)
	}, 25*time.Second).Should(gomega.Equal(0))

	// When RouteBackendExtension controller is invalid, default settings should be applied
	g.Expect(childNode.PoolRefs[0].LbAlgorithm).To(gomega.BeNil())

	// Test status transition from status controller invalid to valid
	rbe.Controller = "AKOCRDController"
	rbe.UpdateRouteBackendExtensionStatus(t)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		childNode = nodes[0].EvhNodes[0]
		if len(childNode.PoolRefs) != 1 {
			return -1
		}
		// When RouteBackendExtension is accepted, HM should be set in graph layer
		return len(childNode.PoolRefs[0].HealthMonitorRefs)
	}, 25*time.Second).Should(gomega.Equal(1))

	// Verify RouteBackendExtension configured settings are present in graph layer
	g.Expect(childNode.PoolRefs[0].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm1"))
	g.Expect(*childNode.PoolRefs[0].LbAlgorithm).To(gomega.Equal("LB_ALGORITHM_ROUND_ROBIN"))

	// Delete routeBackendExtension and verify it's settings are removed from graph layer
	rbe.DeleteRouteBackendExtensionCR(t)

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 60*time.Second).Should(gomega.Equal(1))

	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func verifyApplicationProfileRef(g *gomega.GomegaWithT, modelName, appProfileRef string) {
	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return false
		}

		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes) > 0 {
			childVS := nodes[0].EvhNodes[0]
			if childVS.ApplicationProfileRef == nil {
				return false
			}
			return *childVS.ApplicationProfileRef == appProfileRef
		}
		return false

	}, 25*time.Second).Should(gomega.Equal(true))
}

func TestHTTPRouteWithAppProfileExtensionRef(t *testing.T) {
	gatewayName := "gateway-hr-32"
	gatewayClassName := "gateway-class-hr-32"
	httpRouteName := "http-route-hr-32"
	svcName := "avisvc-hr-32"
	ports := []int32{8080}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)
	defaultHTTPAppProfile := "/api/applicationprofile/?name=System-HTTP"

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	// initial setup with no application profile ref
	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{rule}
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	// if no application profile ref is provided,
	// the default "System-HTTP" profile should be used.
	verifyApplicationProfileRef(g, modelName, defaultHTTPAppProfile)

	// check if invalid app profile is correctly handled post update
	rule = akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"ExtensionRef": {"test-app-profile"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rules = []gatewayv1.HTTPRouteRule{rule}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	// if an invalid application profile ref is provided,
	// the child vs should be absent
	verifyApplicationProfileRef(g, modelName, defaultHTTPAppProfile)

	// create the previous "test-app-profile" app profile and check if
	// http route was actually processed with child vs correctly getting
	// application profile ref applied
	akogatewayapitests.CreateApplicationProfileCRD(t, "test-app-profile", &akogatewayapitests.FakeApplicationProfileStatus{
		Status: "True",
	})
	// should be updated correctly with prefix <cluster-name>-<namespace>
	verifyApplicationProfileRef(g, modelName, "/api/applicationprofile/?name=cluster-default-test-app-profile")

	// delete the app profile ref and check if the child vs is removed
	akogatewayapitests.DeleteApplicationProfileCRD(t, "test-app-profile")
	verifyApplicationProfileRef(g, modelName, defaultHTTPAppProfile)

	// Delete the existing httproute and create a new one with existing app prof ref
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)

	akogatewayapitests.CreateApplicationProfileCRD(t, "test-app-profile-2", &akogatewayapitests.FakeApplicationProfileStatus{
		Status: "True",
	})

	rule = akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"ExtensionRef": {"test-app-profile-2"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rules = []gatewayv1.HTTPRouteRule{rule}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	// should be updated correctly
	verifyApplicationProfileRef(g, modelName, "/api/applicationprofile/?name=cluster-default-test-app-profile-2")

	// update the application profile to change condition status to "False"
	// and verify if existing application profile wasn't changed
	akogatewayapitests.UpdateApplicationProfileCRD(t, "test-app-profile-2", &akogatewayapitests.FakeApplicationProfileStatus{
		Status: "False",
	})
	verifyApplicationProfileRef(g, modelName, defaultHTTPAppProfile)

	// create a new application profile crd and update the http route ref to it
	// and it should update correctly
	akogatewayapitests.CreateApplicationProfileCRD(t, "test-app-profile-3", &akogatewayapitests.FakeApplicationProfileStatus{
		Status: "True",
	})
	rule = akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"ExtensionRef": {"test-app-profile-3"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rules = []gatewayv1.HTTPRouteRule{rule}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	verifyApplicationProfileRef(g, modelName, "/api/applicationprofile/?name=cluster-default-test-app-profile-3")

	// cleanup
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
	akogatewayapitests.DeleteApplicationProfileCRD(t, "test-app-profile-2")
	akogatewayapitests.DeleteApplicationProfileCRD(t, "test-app-profile-3")
}

func TestHTTPRouteWithAppProfileExtensionRefMultipleRules(t *testing.T) {
	gatewayName := "gateway-hr-33"
	gatewayClassName := "gateway-class-hr-33"
	httpRouteName := "http-route-hr-33"
	svcName1 := "avisvc-hr-33a"
	svcName2 := "avisvc-hr-33b"
	ports := []int32{8080}
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)
	defaultHTTPAppProfile := "/api/applicationprofile/?name=System-HTTP"

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)
	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName1, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName1, false, false, "1.2.3")
	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName2, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName2, false, false, "1.2.4")

	akogatewayapitests.CreateApplicationProfileCRD(t, "test-app-profile-a", &akogatewayapitests.FakeApplicationProfileStatus{
		Status: "True",
	})
	akogatewayapitests.CreateApplicationProfileCRD(t, "test-app-profile-b", &akogatewayapitests.FakeApplicationProfileStatus{
		Status: "True",
	})

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule1 := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"ExtensionRef": {"test-app-profile-a"}},
		[][]string{{svcName1, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rule2 := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/bar"}, []string{},
		map[string][]string{"ExtensionRef": {"test-app-profile-b"}},
		[][]string{{svcName2, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
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

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()

	childNode1 := nodes[0].EvhNodes[0]
	childNode2 := nodes[0].EvhNodes[1]

	g.Expect(*childNode1.ApplicationProfileRef).To(gomega.Equal("/api/applicationprofile/?name=cluster-default-test-app-profile-a"))
	g.Expect(*childNode2.ApplicationProfileRef).To(gomega.Equal("/api/applicationprofile/?name=cluster-default-test-app-profile-b"))

	// remove application profile from one rule and verify if it got reset to System-HTTP
	rule1 = akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{},
		[][]string{{svcName1, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rules = []gatewayv1.HTTPRouteRule{rule1, rule2}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return false
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes) != 2 {
			return false
		}
		childNode1 = nodes[0].EvhNodes[0]
		return *childNode1.ApplicationProfileRef == defaultHTTPAppProfile
	}, 25*time.Second).Should(gomega.Equal(true))

	// set invalid app profile to one of the rules and check vs has default application profile ref
	rule1 = akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"ExtensionRef": {"invalid-app-profile"}},
		[][]string{{svcName1, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rules = []gatewayv1.HTTPRouteRule{rule1, rule2}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	verifyApplicationProfileRef(g, modelName, defaultHTTPAppProfile)

	// update application profiles in both rules and verify if the refs are updated correctly
	rule1 = akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"ExtensionRef": {"test-app-profile-b"}},
		[][]string{{svcName1, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rule2 = akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/bar"}, []string{},
		map[string][]string{"ExtensionRef": {"test-app-profile-a"}},
		[][]string{{svcName2, DEFAULT_NAMESPACE, "8080", "1"}}, nil)
	rules = []gatewayv1.HTTPRouteRule{rule1, rule2}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return false
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes) != 2 {
			return false
		}
		childNode1 = nodes[0].EvhNodes[0]
		childNode2 = nodes[0].EvhNodes[1]
		return *childNode1.ApplicationProfileRef == "/api/applicationprofile/?name=cluster-default-test-app-profile-b" &&
			*childNode2.ApplicationProfileRef == "/api/applicationprofile/?name=cluster-default-test-app-profile-a"
	}, 25*time.Second).Should(gomega.Equal(true))

	// cleanup
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName1)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName2)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName2)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
	akogatewayapitests.DeleteApplicationProfileCRD(t, "test-app-profile-a")
	akogatewayapitests.DeleteApplicationProfileCRD(t, "test-app-profile-b")
}

func TestHTTPRouteWithL7Rule(t *testing.T) {
	gatewayName := "gateway-l7rule-01"
	gatewayClassName := "gateway-class-l7rule-01"
	httpRouteName := "http-route-l7rule-01"
	svcName := "avisvc-l7rule-01"
	l7RuleName := "l7rule-01"
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
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	// Create L7Rule CR
	l7Rule := akogatewayapitests.GetFakeDefaultL7RuleObj(l7RuleName, DEFAULT_NAMESPACE)
	l7Rule.CreateFakeL7RuleWithStatus(t)
	extensionRefCRDs := make(map[string][]string)
	extensionRefCRDs["L7Rule"] = []string{l7RuleName}

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleWithCustomCRDs(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, extensionRefCRDs)
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

	// Verify L7Rule configured settings are present in graph layer
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childNode := nodes[0].EvhNodes[0]
	g.Expect(*childNode.ApplicationProfileRef).To(gomega.ContainSubstring("thisisaviref-appprofile-l7"))
	g.Expect(*childNode.WafPolicyRef).To(gomega.ContainSubstring("thisisaviref-wafpolicy-l7"))
	g.Expect(*childNode.AnalyticsProfileRef).To(gomega.ContainSubstring("thisisaviref-analyticsprofile-l7"))
	g.Expect(childNode.ICAPProfileRefs[0]).To(gomega.ContainSubstring("thisisaviref-icaprofile-l7"))
	g.Expect(childNode.ErrorPageProfileRef).To(gomega.ContainSubstring("thisisaviref-errorpageprofile-l7"))
	// This is from HTTPRoute filter
	g.Expect(len(childNode.HttpPolicyRefs)).To(gomega.Equal(1))
	// This is from L7CRD
	g.Expect(len(childNode.HttpPolicySetRefs)).To(gomega.Equal(2))
	g.Expect(childNode.HttpPolicySetRefs[0]).To(gomega.ContainSubstring("policy1"))
	g.Expect(childNode.HttpPolicySetRefs[1]).To(gomega.ContainSubstring("policy2"))

	// Analytics Policy
	g.Expect(childNode.AnalyticsPolicy).NotTo(gomega.BeNil())
	g.Expect(childNode.AnalyticsPolicy.AllHeaders).NotTo(gomega.BeNil())
	g.Expect(*childNode.AnalyticsPolicy.AllHeaders).To(gomega.Equal(true))
	g.Expect(childNode.AnalyticsPolicy.FullClientLogs).NotTo(gomega.BeNil())
	g.Expect(childNode.AnalyticsPolicy.FullClientLogs.Enabled).NotTo(gomega.BeNil())
	g.Expect(*childNode.AnalyticsPolicy.FullClientLogs.Enabled).To(gomega.Equal(true))

	g.Expect(*childNode.CloseClientConnOnConfigUpdate).To(gomega.Equal(false))
	g.Expect(*childNode.AllowInvalidClientCert).To(gomega.Equal(true))
	g.Expect(*childNode.IgnPoolNetReach).To(gomega.Equal(true))
	g.Expect(*childNode.RemoveListeningPortOnVsDown).To(gomega.Equal(false))
	g.Expect(*childNode.BotPolicyRef).To(gomega.ContainSubstring("sample-bot"))

	g.Expect(childNode.HostNameXlate).To(gomega.BeNil())
	g.Expect(childNode.SecurityPolicyRef).To(gomega.BeNil())
	// Delete L7 and verify it's settings are removed from graph layer
	l7Rule.DeleteL7RuleCR(t)

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 20*time.Second).Should(gomega.Equal(1))

	g.Eventually(func() bool {
		return childNode.ApplicationProfileRef == nil
	}, 60*time.Second).Should(gomega.Equal(true))

	g.Expect(childNode.ApplicationProfileRef).To(gomega.BeNil())
	g.Expect(childNode.HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(childNode.AnalyticsPolicy).To(gomega.BeNil())
	g.Expect(childNode.ErrorPageProfileRef).To(gomega.BeEmpty())
	g.Expect(childNode.WafPolicyRef).To(gomega.BeNil())
	g.Expect(childNode.ICAPProfileRefs).To(gomega.HaveLen(0))
	g.Expect(childNode.CloseClientConnOnConfigUpdate).To(gomega.BeNil())
	g.Expect(childNode.AllowInvalidClientCert).To(gomega.BeNil())
	g.Expect(childNode.IgnPoolNetReach).To(gomega.BeNil())
	g.Expect(childNode.RemoveListeningPortOnVsDown).To(gomega.BeNil())
	g.Expect(childNode.BotPolicyRef).To(gomega.BeNil())

	// Teardown
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteWithL7RuleValidToInvalid(t *testing.T) {
	gatewayName := "gateway-l7rule-02"
	gatewayClassName := "gateway-class-l7rule-02"
	httpRouteName := "http-route-l7rule-02"
	svcName := "avisvc-l7rule-02"
	l7RuleName := "l7rule-02"
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
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	// Create L7Rule CR
	l7Rule := akogatewayapitests.GetFakeDefaultL7RuleObj(l7RuleName, DEFAULT_NAMESPACE)
	l7Rule.CreateFakeL7RuleWithStatus(t)
	extensionRefCRDs := make(map[string][]string)
	extensionRefCRDs["L7Rule"] = []string{l7RuleName}

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleWithCustomCRDs(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, extensionRefCRDs)
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

	// Verify L7Rule configured settings are present in graph layer
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childNode := nodes[0].EvhNodes[0]
	g.Expect(*childNode.ApplicationProfileRef).To(gomega.ContainSubstring("thisisaviref-appprofile-l7"))
	g.Expect(*childNode.WafPolicyRef).To(gomega.ContainSubstring("thisisaviref-wafpolicy-l7"))
	g.Expect(*childNode.AnalyticsProfileRef).To(gomega.ContainSubstring("thisisaviref-analyticsprofile-l7"))
	g.Expect(childNode.ICAPProfileRefs[0]).To(gomega.ContainSubstring("thisisaviref-icaprofile-l7"))
	g.Expect(childNode.ErrorPageProfileRef).To(gomega.ContainSubstring("thisisaviref-errorpageprofile-l7"))
	// This is from HTTPRoute filter
	g.Expect(len(childNode.HttpPolicyRefs)).To(gomega.Equal(1))
	// This is from L7CRD
	g.Expect(len(childNode.HttpPolicySetRefs)).To(gomega.Equal(2))
	g.Expect(childNode.HttpPolicySetRefs[0]).To(gomega.ContainSubstring("policy1"))
	g.Expect(childNode.HttpPolicySetRefs[1]).To(gomega.ContainSubstring("policy2"))

	// Analytics Policy
	g.Expect(childNode.AnalyticsPolicy).NotTo(gomega.BeNil())
	g.Expect(childNode.AnalyticsPolicy.AllHeaders).NotTo(gomega.BeNil())
	g.Expect(*childNode.AnalyticsPolicy.AllHeaders).To(gomega.Equal(true))
	g.Expect(childNode.AnalyticsPolicy.FullClientLogs).NotTo(gomega.BeNil())
	g.Expect(childNode.AnalyticsPolicy.FullClientLogs.Enabled).NotTo(gomega.BeNil())
	g.Expect(*childNode.AnalyticsPolicy.FullClientLogs.Enabled).To(gomega.Equal(true))

	g.Expect(*childNode.CloseClientConnOnConfigUpdate).To(gomega.Equal(false))
	g.Expect(*childNode.AllowInvalidClientCert).To(gomega.Equal(true))
	g.Expect(*childNode.IgnPoolNetReach).To(gomega.Equal(true))
	g.Expect(*childNode.RemoveListeningPortOnVsDown).To(gomega.Equal(false))
	g.Expect(*childNode.BotPolicyRef).To(gomega.ContainSubstring("sample-bot"))

	g.Expect(childNode.HostNameXlate).To(gomega.BeNil())
	g.Expect(childNode.SecurityPolicyRef).To(gomega.BeNil())

	// Reject
	l7Rule.Status = "Rejected"
	l7Rule.Error = "Invalid Application Profile ref"
	l7Rule.UpdateL7RuleStatus(t)

	// L7Rule fields should be removed.
	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes) > 0 {
			return nodes[0].EvhNodes[0].ApplicationProfileRef == nil
		}
		return false
	}, 25*time.Second, 1*time.Second).Should(gomega.Equal(true))
	g.Expect(childNode.HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(childNode.AnalyticsPolicy).To(gomega.BeNil())
	g.Expect(childNode.ErrorPageProfileRef).To(gomega.BeEmpty())
	g.Expect(childNode.WafPolicyRef).To(gomega.BeNil())
	g.Expect(childNode.ICAPProfileRefs).To(gomega.HaveLen(0))
	g.Expect(childNode.CloseClientConnOnConfigUpdate).To(gomega.BeNil())
	g.Expect(childNode.AllowInvalidClientCert).To(gomega.BeNil())
	g.Expect(childNode.IgnPoolNetReach).To(gomega.BeNil())
	g.Expect(childNode.RemoveListeningPortOnVsDown).To(gomega.BeNil())
	g.Expect(childNode.BotPolicyRef).To(gomega.BeNil())

	// Teardown
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteWithL7RuleWithApplicationProfileCRD(t *testing.T) {
	gatewayName := "gateway-l7rule-03"
	gatewayClassName := "gateway-class-l7rule-03"
	httpRouteName := "http-route-l7rule-03"
	svcName := "avisvc-l7rule-03"
	l7RuleName := "l7rule-03"
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
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	// Create L7Rule CR
	l7Rule := akogatewayapitests.GetFakeDefaultL7RuleObj(l7RuleName, DEFAULT_NAMESPACE)
	l7Rule.CreateFakeL7RuleWithStatus(t)
	extensionRefCRDs := make(map[string][]string)
	extensionRefCRDs["L7Rule"] = []string{l7RuleName}
	extensionRefCRDs["ApplicationProfile"] = []string{"AppProfile1"}

	akogatewayapitests.CreateApplicationProfileCRD(t, "AppProfile1", &akogatewayapitests.FakeApplicationProfileStatus{
		Status: "True",
	})

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleWithCustomCRDs(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, extensionRefCRDs)
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

	// Verify L7Rule configured settings are present in graph layer
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childNode := nodes[0].EvhNodes[0]
	// Application Profile CRD object
	g.Expect(*childNode.ApplicationProfileRef).To(gomega.ContainSubstring("AppProfile1"))
	g.Expect(*childNode.WafPolicyRef).To(gomega.ContainSubstring("thisisaviref-wafpolicy-l7"))
	g.Expect(*childNode.AnalyticsProfileRef).To(gomega.ContainSubstring("thisisaviref-analyticsprofile-l7"))
	g.Expect(childNode.ICAPProfileRefs[0]).To(gomega.ContainSubstring("thisisaviref-icaprofile-l7"))
	g.Expect(childNode.ErrorPageProfileRef).To(gomega.ContainSubstring("thisisaviref-errorpageprofile-l7"))
	// This is from HTTPRoute filter
	g.Expect(len(childNode.HttpPolicyRefs)).To(gomega.Equal(1))
	// This is from L7CRD
	g.Expect(len(childNode.HttpPolicySetRefs)).To(gomega.Equal(2))
	g.Expect(childNode.HttpPolicySetRefs[0]).To(gomega.ContainSubstring("policy1"))
	g.Expect(childNode.HttpPolicySetRefs[1]).To(gomega.ContainSubstring("policy2"))

	// Analytics Policy
	g.Expect(childNode.AnalyticsPolicy).NotTo(gomega.BeNil())
	g.Expect(childNode.AnalyticsPolicy.AllHeaders).NotTo(gomega.BeNil())
	g.Expect(*childNode.AnalyticsPolicy.AllHeaders).To(gomega.Equal(true))
	g.Expect(childNode.AnalyticsPolicy.FullClientLogs).NotTo(gomega.BeNil())
	g.Expect(childNode.AnalyticsPolicy.FullClientLogs.Enabled).NotTo(gomega.BeNil())
	g.Expect(*childNode.AnalyticsPolicy.FullClientLogs.Enabled).To(gomega.Equal(true))

	g.Expect(*childNode.CloseClientConnOnConfigUpdate).To(gomega.Equal(false))
	g.Expect(*childNode.AllowInvalidClientCert).To(gomega.Equal(true))
	g.Expect(*childNode.IgnPoolNetReach).To(gomega.Equal(true))
	g.Expect(*childNode.RemoveListeningPortOnVsDown).To(gomega.Equal(false))
	g.Expect(*childNode.BotPolicyRef).To(gomega.ContainSubstring("sample-bot"))

	g.Expect(childNode.HostNameXlate).To(gomega.BeNil())
	g.Expect(childNode.SecurityPolicyRef).To(gomega.BeNil())

	// Reject
	l7Rule.Status = "Rejected"
	l7Rule.Error = "Invalid Application Profile ref"
	l7Rule.UpdateL7RuleStatus(t)

	// L7Rule fields should be removed.
	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes) > 0 {
			return len(nodes[0].EvhNodes[0].HttpPolicySetRefs) == 0
		}
		return false
	}, 25*time.Second, 1*time.Second).Should(gomega.Equal(true))
	// Application Profile CRD object
	g.Expect(*childNode.ApplicationProfileRef).To(gomega.ContainSubstring("AppProfile1"))
	g.Expect(childNode.HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(childNode.HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(childNode.AnalyticsPolicy).To(gomega.BeNil())
	g.Expect(childNode.ErrorPageProfileRef).To(gomega.BeEmpty())
	g.Expect(childNode.WafPolicyRef).To(gomega.BeNil())
	g.Expect(childNode.ICAPProfileRefs).To(gomega.HaveLen(0))
	g.Expect(childNode.CloseClientConnOnConfigUpdate).To(gomega.BeNil())
	g.Expect(childNode.AllowInvalidClientCert).To(gomega.BeNil())
	g.Expect(childNode.IgnPoolNetReach).To(gomega.BeNil())
	g.Expect(childNode.RemoveListeningPortOnVsDown).To(gomega.BeNil())
	g.Expect(childNode.BotPolicyRef).To(gomega.BeNil())

	// Teardown
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

// @AI-Generated
// [Generated by Cursor claude-4-sonnet]
func TestHTTPRouteWithRouteBackendExtensionPersistenceProfile(t *testing.T) {
	gatewayName := "gateway-rbe-persistence-01"
	gatewayClassName := "gateway-class-rbe-persistence-01"
	httpRouteName := "http-route-rbe-persistence-01"
	svcName := "avisvc-rbe-persistence-01"
	routeBackendExtensionName := "rbe-persistence-01"
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
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	// Create RouteBackendExtension CR with persistenceProfile
	rbe := akogatewayapitests.GetFakeDefaultRBEObj(routeBackendExtensionName, DEFAULT_NAMESPACE)
	rbe.PersistenceProfile = "System-Persistence-Client-IP"
	rbe.CreateRouteBackendExtensionCRWithStatus(t)

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleWithRouteBackendExtensionAndHMFilters(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, []string{routeBackendExtensionName})
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

	// Verify RouteBackendExtension with persistenceProfile configured settings are present in graph layer
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childNode := nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolRefs).To(gomega.HaveLen(1))

	// Verify persistence profile reference is set
	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes) > 0 && len(nodes[0].EvhNodes[0].PoolRefs) > 0 {
			poolNode := nodes[0].EvhNodes[0].PoolRefs[0]
			if poolNode.ApplicationPersistenceProfileRef != nil {
				return *poolNode.ApplicationPersistenceProfileRef == "/api/applicationpersistenceprofile?name=System-Persistence-Client-IP"
			}
		}
		return false
	}, 25*time.Second).Should(gomega.Equal(true))

	// Test different persistence profile types
	testCases := []struct {
		persistenceType string
		expectedRef     string
	}{
		{"System-Persistence-Http-Cookie", "System-Persistence-Http-Cookie"},
		{"System-Persistence-TLS", "System-Persistence-TLS"},
		{"System-Persistence-App-Cookie", "System-Persistence-App-Cookie"},
	}

	for _, tc := range testCases {
		// Update RouteBackendExtension with different persistence profile
		rbe.PersistenceProfile = tc.persistenceType
		rbe.UpdateRouteBackendExtensionCR(t)

		g.Eventually(func() bool {
			_, aviModel := objects.SharedAviGraphLister().Get(modelName)
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			if len(nodes[0].EvhNodes) > 0 && len(nodes[0].EvhNodes[0].PoolRefs) > 0 {
				poolNode := nodes[0].EvhNodes[0].PoolRefs[0]
				if poolNode.ApplicationPersistenceProfileRef != nil {
					return *poolNode.ApplicationPersistenceProfileRef == fmt.Sprintf("/api/applicationpersistenceprofile?name=%s", tc.expectedRef)
				}
			}
			return false
		}, 25*time.Second).Should(gomega.Equal(true))
	}

	// Update RouteBackendExtension to remove persistenceProfile
	rbe.PersistenceProfile = ""
	rbe.UpdateRouteBackendExtensionCR(t)

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes) > 0 && len(nodes[0].EvhNodes[0].PoolRefs) > 0 {
			poolNode := nodes[0].EvhNodes[0].PoolRefs[0]
			return poolNode.ApplicationPersistenceProfileRef == nil
		}
		return false
	}, 25*time.Second).Should(gomega.Equal(true))

	// Clean up
	rbe.DeleteRouteBackendExtensionCR(t)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

// @AI-Generated
// [Generated by Cursor claude-4-sonnet]
func TestHTTPRouteWithSessionPersistenceAndRouteBackendExtensionPersistenceProfile(t *testing.T) {
	gatewayName := "gateway-dual-persistence-01"
	gatewayClassName := "gateway-class-dual-persistence-01"
	httpRouteName := "http-route-dual-persistence-01"
	svcName := "avisvc-dual-persistence-01"
	routeBackendExtensionName := "rbe-dual-persistence-01"
	ports := []int32{8080}
	ruleName := "dual-persistence-rule"
	cookieName := "session-cookie"
	var timeout gatewayv1.Duration = "10m"
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
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	// Create RouteBackendExtension CR with persistenceProfile
	rbe := akogatewayapitests.GetFakeDefaultRBEObj(routeBackendExtensionName, DEFAULT_NAMESPACE)
	rbe.PersistenceProfile = "System-Persistence-Client-IP"
	rbe.CreateRouteBackendExtensionCRWithStatus(t)

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}

	// Create rule with both SessionPersistence and RouteBackendExtension reference
	rule := akogatewayapitests.GetHTTPRouteRuleWithRouteBackendExtensionAndHMFilters(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, []string{routeBackendExtensionName})

	// Add SessionPersistence to the rule
	rule.Name = (*gatewayv1.SectionName)(&ruleName)
	cookieType := gatewayv1.CookieBasedSessionPersistence
	rule.SessionPersistence = &gatewayv1.SessionPersistence{
		Type:            &cookieType,
		SessionName:     &cookieName,
		AbsoluteTimeout: &timeout,
	}

	rules := []gatewayv1.HTTPRouteRule{rule}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(1))

	// Verify that SessionPersistence takes precedence over RouteBackendExtension persistenceProfile
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childNode := nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolRefs).To(gomega.HaveLen(1))
	poolNode := childNode.PoolRefs[0]

	// Verify SessionPersistence configuration is applied (takes precedence)
	g.Expect(poolNode.ApplicationPersistenceProfile).NotTo(gomega.BeNil())
	appPersistProfile := poolNode.ApplicationPersistenceProfile
	g.Expect(appPersistProfile.PersistenceType).To(gomega.Equal("PERSISTENCE_TYPE_HTTP_COOKIE"))
	g.Expect(appPersistProfile.HTTPCookiePersistenceProfile).NotTo(gomega.BeNil())
	g.Expect(appPersistProfile.HTTPCookiePersistenceProfile.CookieName).To(gomega.Equal(cookieName))
	g.Expect(*appPersistProfile.HTTPCookiePersistenceProfile.Timeout).To(gomega.Equal(int32(10))) // 10m -> 10 minutes
	g.Expect(*appPersistProfile.HTTPCookiePersistenceProfile.IsPersistentCookie).To(gomega.BeFalse())

	// Verify RouteBackendExtension persistenceProfile is NOT applied (due to precedence)
	g.Expect(poolNode.ApplicationPersistenceProfileRef).To(gomega.BeNil())

	// Test case: Remove SessionPersistence, RouteBackendExtension persistenceProfile should now be applied
	rule.SessionPersistence = nil
	rules = []gatewayv1.HTTPRouteRule{rule}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes) > 0 && len(nodes[0].EvhNodes[0].PoolRefs) > 0 {
			poolNode := nodes[0].EvhNodes[0].PoolRefs[0]
			// SessionPersistence should be gone
			if poolNode.ApplicationPersistenceProfile != nil {
				return false
			}
			// RouteBackendExtension persistenceProfile should now be applied
			if poolNode.ApplicationPersistenceProfileRef != nil {
				return *poolNode.ApplicationPersistenceProfileRef == "/api/applicationpersistenceprofile?name=System-Persistence-Client-IP"
			}
		}
		return false
	}, 25*time.Second).Should(gomega.Equal(true))

	// Verify RouteBackendExtension persistenceProfile is now applied
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	poolNode = nodes[0].EvhNodes[0].PoolRefs[0]
	g.Expect(poolNode.ApplicationPersistenceProfile).To(gomega.BeNil())
	g.Expect(poolNode.ApplicationPersistenceProfileRef).NotTo(gomega.BeNil())
	g.Expect(*poolNode.ApplicationPersistenceProfileRef).To(gomega.Equal("/api/applicationpersistenceprofile?name=System-Persistence-Client-IP"))

	// Test case: Re-add SessionPersistence, it should take precedence again
	rule.SessionPersistence = &gatewayv1.SessionPersistence{
		Type:            &cookieType,
		SessionName:     &cookieName,
		AbsoluteTimeout: &timeout,
	}
	rules = []gatewayv1.HTTPRouteRule{rule}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes) > 0 && len(nodes[0].EvhNodes[0].PoolRefs) > 0 {
			poolNode := nodes[0].EvhNodes[0].PoolRefs[0]
			// SessionPersistence should be back
			if poolNode.ApplicationPersistenceProfile == nil {
				return false
			}
			if poolNode.ApplicationPersistenceProfile.PersistenceType == "PERSISTENCE_TYPE_HTTP_COOKIE" {
				// RouteBackendExtension persistenceProfile should be ignored
				return poolNode.ApplicationPersistenceProfileRef == nil
			}
		}
		return false
	}, 25*time.Second).Should(gomega.Equal(true))

	// Final verification: SessionPersistence takes precedence
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	poolNode = nodes[0].EvhNodes[0].PoolRefs[0]
	g.Expect(poolNode.ApplicationPersistenceProfile).NotTo(gomega.BeNil())
	g.Expect(poolNode.ApplicationPersistenceProfile.PersistenceType).To(gomega.Equal("PERSISTENCE_TYPE_HTTP_COOKIE"))
	g.Expect(poolNode.ApplicationPersistenceProfileRef).To(gomega.BeNil())

	// Clean up
	rbe.DeleteRouteBackendExtensionCR(t)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

// @AI-Generated
// [Generated by Cursor claude-4-sonnet]
func TestHTTPRouteWithRouteBackendExtensionBackendTLS(t *testing.T) {
	gatewayName := "gateway-rbe-tls-01"
	gatewayClassName := "gateway-class-rbe-tls-01"
	httpRouteName := "http-route-rbe-tls-01"
	svcName := "avisvc-rbe-tls-01"
	routeBackendExtensionName := "rbe-tls-01"
	pkiProfileName := "pki-profile-tls-01"
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersOnHostname([]string{"backend-tls-8080.com"})
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		gateway, err := akogatewayapitests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			return false
		}
		return apimeta.IsStatusConditionTrue(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted))
	}, 25*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	// Create PKIProfile CR
	caCerts := []string{
		"-----BEGIN CERTIFICATE-----\nMIIC...test...cert\n-----END CERTIFICATE-----",
	}
	akogatewayapitests.CreatePKIProfileCR(t, pkiProfileName, DEFAULT_NAMESPACE, caCerts)

	// Create RouteBackendExtension CR with BackendTLS configuration
	rbe := akogatewayapitests.GetFakeRBEObjWithBackendTLS(
		routeBackendExtensionName,
		DEFAULT_NAMESPACE,
		true,                             // enableBackendSSL
		akogatewayapitests.BoolPtr(true), // hostCheckEnabled
		akogatewayapitests.StringPtr("backend.example.com"), // domainName
		akogatewayapitests.StringPtr(pkiProfileName),        // pkiProfile
		"thisisaviref-hm1", // healthMonitor
	)
	rbe.CreateRouteBackendExtensionCRWithStatus(t)

	parentRefs := akogatewayapitests.GetParentReferencesFromListeners(listeners, gatewayName, DEFAULT_NAMESPACE)
	rule := akogatewayapitests.GetHTTPRouteRuleWithRouteBackendExtensionAndHMFilters(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, []string{routeBackendExtensionName})

	rules := []gatewayv1.HTTPRouteRule{rule}
	hostnames := []gatewayv1.Hostname{"backend-tls-8080.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if aviModel == nil {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(1))

	// Verify BackendTLS configured settings are present in graph layer
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childNode := nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolRefs).To(gomega.HaveLen(1))
	poolNode := childNode.PoolRefs[0]

	// Verify SSL/TLS settings
	g.Expect(poolNode.SslProfileRef).NotTo(gomega.BeNil())
	g.Expect(*poolNode.SslProfileRef).To(gomega.ContainSubstring("System-Standard"))
	g.Expect(poolNode.PkiProfileRef).NotTo(gomega.BeNil())
	g.Expect(*poolNode.PkiProfileRef).To(gomega.ContainSubstring(fmt.Sprintf("cluster-%s-%s", DEFAULT_NAMESPACE, pkiProfileName)))
	g.Expect(poolNode.HostCheckEnabled).NotTo(gomega.BeNil())
	g.Expect(*poolNode.HostCheckEnabled).To(gomega.Equal(true))
	g.Expect(poolNode.DomainName).NotTo(gomega.BeNil())
	g.Expect(poolNode.DomainName[0]).To(gomega.Equal("backend.example.com"))

	// Verify health monitor is still configured
	g.Expect(poolNode.HealthMonitorRefs).To(gomega.HaveLen(1))
	g.Expect(poolNode.HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm1"))

	// Clean up
	rbe.DeleteRouteBackendExtensionCR(t)
	akogatewayapitests.DeletePKIProfileCR(t, pkiProfileName, DEFAULT_NAMESPACE)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

// @AI-Generated
// [Generated by Cursor claude-4-sonnet]
func TestHTTPRouteWithRouteBackendExtensionBackendTLSMinimal(t *testing.T) {
	gatewayName := "gateway-rbe-tls-02"
	gatewayClassName := "gateway-class-rbe-tls-02"
	httpRouteName := "http-route-rbe-tls-02"
	svcName := "avisvc-rbe-tls-02"
	routeBackendExtensionName := "rbe-tls-02"
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersOnHostname([]string{"backend-minimal-8080.com"})
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		gateway, err := akogatewayapitests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			return false
		}
		return apimeta.IsStatusConditionTrue(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted))
	}, 25*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	// Create RouteBackendExtension CR with minimal BackendTLS configuration (only enableBackendSSL)
	rbe := akogatewayapitests.GetFakeRBEObjWithBackendTLS(
		routeBackendExtensionName,
		DEFAULT_NAMESPACE,
		true, // enableBackendSSL
		nil,  // hostCheckEnabled not set
		nil,  // domainName not set
		nil,  // pkiProfile not set
	)
	rbe.CreateRouteBackendExtensionCRWithStatus(t)

	parentRefs := akogatewayapitests.GetParentReferencesFromListeners(listeners, gatewayName, DEFAULT_NAMESPACE)
	rule := akogatewayapitests.GetHTTPRouteRuleWithRouteBackendExtensionAndHMFilters(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, []string{routeBackendExtensionName})

	rules := []gatewayv1.HTTPRouteRule{rule}
	hostnames := []gatewayv1.Hostname{"backend-minimal-8080.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if aviModel == nil {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(1))

	// Verify minimal BackendTLS configured settings are present in graph layer
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childNode := nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolRefs).To(gomega.HaveLen(1))
	poolNode := childNode.PoolRefs[0]

	// Verify only SSL profile is set, other TLS fields should be nil
	g.Expect(poolNode.SslProfileRef).NotTo(gomega.BeNil())
	g.Expect(*poolNode.SslProfileRef).To(gomega.ContainSubstring("System-Standard"))
	g.Expect(poolNode.PkiProfileRef).To(gomega.BeNil())
	g.Expect(poolNode.HostCheckEnabled).To(gomega.BeNil())
	g.Expect(poolNode.DomainName).To(gomega.BeNil())

	// Clean up
	rbe.DeleteRouteBackendExtensionCR(t)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

// @AI-Generated
// [Generated by Cursor claude-4-sonnet]
func TestHTTPRouteWithRouteBackendExtensionHostCheckEnabledOnly(t *testing.T) {
	gatewayName := "gateway-rbe-tls-03"
	gatewayClassName := "gateway-class-rbe-tls-03"
	httpRouteName := "http-route-rbe-tls-03"
	svcName := "avisvc-rbe-tls-03"
	routeBackendExtensionName := "rbe-tls-03"
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersOnHostname([]string{"hostcheck-8080.com"})
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		gateway, err := akogatewayapitests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			return false
		}
		return apimeta.IsStatusConditionTrue(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted))
	}, 25*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	// Create RouteBackendExtension CR with BackendTLS and hostCheckEnabled
	rbe := akogatewayapitests.GetFakeRBEObjWithBackendTLS(
		routeBackendExtensionName,
		DEFAULT_NAMESPACE,
		true,                             // enableBackendSSL
		akogatewayapitests.BoolPtr(true), // hostCheckEnabled
		nil,                              // domainName not set
		nil,                              // pkiProfile not set
	)
	rbe.CreateRouteBackendExtensionCRWithStatus(t)

	parentRefs := akogatewayapitests.GetParentReferencesFromListeners(listeners, gatewayName, DEFAULT_NAMESPACE)
	rule := akogatewayapitests.GetHTTPRouteRuleWithRouteBackendExtensionAndHMFilters(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, []string{routeBackendExtensionName})

	rules := []gatewayv1.HTTPRouteRule{rule}
	hostnames := []gatewayv1.Hostname{"hostcheck-8080.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if aviModel == nil {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(1))

	// Verify BackendTLS with hostCheckEnabled configured settings are present in graph layer
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childNode := nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolRefs).To(gomega.HaveLen(1))
	poolNode := childNode.PoolRefs[0]

	// Verify SSL profile and hostCheckEnabled are set
	g.Expect(poolNode.SslProfileRef).NotTo(gomega.BeNil())
	g.Expect(*poolNode.SslProfileRef).To(gomega.ContainSubstring("System-Standard"))
	g.Expect(poolNode.HostCheckEnabled).NotTo(gomega.BeNil())
	g.Expect(*poolNode.HostCheckEnabled).To(gomega.Equal(true))
	g.Expect(poolNode.PkiProfileRef).To(gomega.BeNil())
	g.Expect(poolNode.DomainName).To(gomega.BeNil())

	// Clean up
	rbe.DeleteRouteBackendExtensionCR(t)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

// @AI-Generated
// [Generated by Cursor claude-4-sonnet]
func TestHTTPRouteWithRouteBackendExtensionPKIProfileUpdate(t *testing.T) {
	gatewayName := "gateway-rbe-tls-04"
	gatewayClassName := "gateway-class-rbe-tls-04"
	httpRouteName := "http-route-rbe-tls-04"
	svcName := "avisvc-rbe-tls-04"
	routeBackendExtensionName := "rbe-tls-04"
	pkiProfileName := "pki-profile-tls-04"
	modelName, _ := akogatewayapitests.GetModelName(DEFAULT_NAMESPACE, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersOnHostname([]string{"pki-update-8080.com"})
	akogatewayapitests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		gateway, err := akogatewayapitests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			return false
		}
		return apimeta.IsStatusConditionTrue(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted))
	}, 25*time.Second).Should(gomega.Equal(true))

	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	// Create PKIProfile CR
	caCerts := []string{
		"-----BEGIN CERTIFICATE-----\nMIIC...test...cert\n-----END CERTIFICATE-----",
	}
	akogatewayapitests.CreatePKIProfileCR(t, pkiProfileName, DEFAULT_NAMESPACE, caCerts)

	// Create RouteBackendExtension CR without PKIProfile initially
	rbe := akogatewayapitests.GetFakeRBEObjWithBackendTLS(
		routeBackendExtensionName,
		DEFAULT_NAMESPACE,
		true, // enableBackendSSL
		nil,  // hostCheckEnabled not set
		nil,  // domainName not set
		nil,  // pkiProfile not set initially
	)
	rbe.CreateRouteBackendExtensionCRWithStatus(t)

	parentRefs := akogatewayapitests.GetParentReferencesFromListeners(listeners, gatewayName, DEFAULT_NAMESPACE)
	rule := akogatewayapitests.GetHTTPRouteRuleWithRouteBackendExtensionAndHMFilters(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, []string{routeBackendExtensionName})

	rules := []gatewayv1.HTTPRouteRule{rule}
	hostnames := []gatewayv1.Hostname{"pki-update-8080.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE, parentRefs, hostnames, rules)

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if aviModel == nil {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(1))

	// Verify initial state without PKIProfile
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	childNode := nodes[0].EvhNodes[0]
	g.Expect(childNode.PoolRefs).To(gomega.HaveLen(1))
	poolNode := childNode.PoolRefs[0]
	g.Expect(poolNode.SslProfileRef).NotTo(gomega.BeNil())
	g.Expect(poolNode.PkiProfileRef).To(gomega.BeNil())

	// Update RouteBackendExtension to add PKIProfile
	rbe.PKIProfile = &akogatewayapitests.FakeRouteBackendExtensionPKIProfile{
		Kind: "CRD",
		Name: pkiProfileName,
	}
	rbe.UpdateRouteBackendExtensionCR(t)

	// Verify PKIProfile is now configured
	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if aviModel == nil {
			return false
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes) == 0 || len(nodes[0].EvhNodes) == 0 || len(nodes[0].EvhNodes[0].PoolRefs) == 0 {
			return false
		}
		poolNode := nodes[0].EvhNodes[0].PoolRefs[0]
		return poolNode.PkiProfileRef != nil && *poolNode.PkiProfileRef != ""
	}, 25*time.Second).Should(gomega.Equal(true))

	// Final verification
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	poolNode = nodes[0].EvhNodes[0].PoolRefs[0]
	g.Expect(poolNode.PkiProfileRef).NotTo(gomega.BeNil())
	g.Expect(*poolNode.PkiProfileRef).To(gomega.ContainSubstring(fmt.Sprintf("cluster-%s-%s", DEFAULT_NAMESPACE, pkiProfileName)))

	// Clean up
	rbe.DeleteRouteBackendExtensionCR(t)
	akogatewayapitests.DeletePKIProfileCR(t, pkiProfileName, DEFAULT_NAMESPACE)
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}
