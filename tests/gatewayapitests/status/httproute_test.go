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

package status

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	akogatewayapitests "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/gatewayapitests"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"
)

/* Positive test cases
 * - HTTPRoute with valid configurations (both parent reference and hostnames)
 * - HTTPRoute with valid rules (TODO: end-to-end code is required to check this)
 * - HTTPRoute update with new parent reference (adding one more parent reference)
 */
func TestHTTPRouteWithValidConfig(t *testing.T) {
	gatewayClassName := "gateway-class-hr-01"
	gatewayName := "gateway-hr-01"
	httpRouteName := "httproute-01"
	namespace := "default"
	svcName := "avisvc-hr-01"
	ports := []int32{8080, 8081}

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, "default", svcName, false, false, "1.1.1")

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
		return apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs)) != nil &&
			apimeta.FindStatusCondition(httpRoute.Status.Parents[1].Conditions, string(gatewayv1.RouteConditionResolvedRefs)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	conditionMap := make(map[string][]metav1.Condition)

	for _, port := range ports {
		conditions := []metav1.Condition{{
			Type:    string(gatewayv1.RouteConditionAccepted),
			Reason:  string(gatewayv1.RouteReasonAccepted),
			Status:  metav1.ConditionTrue,
			Message: "Parent reference is valid",
		}, {
			Type:   string(gatewayv1.RouteConditionResolvedRefs),
			Reason: string(gatewayv1.RouteReasonResolvedRefs),
			Status: metav1.ConditionTrue,
		}}
		conditionMap[fmt.Sprintf("%s-%d", gatewayName, port)] = conditions
	}
	expectedRouteStatus := akogatewayapitests.GetRouteStatusV1([]string{gatewayName}, namespace, ports, conditionMap)

	httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
	if err != nil || httpRoute == nil {
		t.Fatalf("Couldn't get the HTTPRoute, err: %+v", err)
	}
	akogatewayapitests.ValidateHTTPRouteStatus(t, &httpRoute.Status, &gatewayv1.HTTPRouteStatus{RouteStatus: *expectedRouteStatus})

	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	akogatewayapitests.TeardownGateway(t, gatewayName, namespace)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteWithAtleastOneParentReferenceValid(t *testing.T) {
	gatewayClassName := "gateway-class-hr-02"
	gatewayName := "gateway-hr-02"
	httpRouteName := "httproute-02"
	namespace := "default"
	ports := []int32{8080, 8081}

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)

	// creates a gateway with listeners 8080 and 8082
	listeners := akogatewayapitests.GetListenersV1([]int32{8080, 8082}, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, namespace, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		gateway, err := akogatewayapitests.GatewayClient.GatewayV1().Gateways(namespace).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.IsStatusConditionTrue(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted))
	}, 30*time.Second).Should(gomega.Equal(true))

	// creates a httproute with parent which has listeners 8080, 8081
	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, ports)
	hostnames := []gatewayv1.Hostname{"foo-8080.com", "foo-8081.com"}

	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, nil)

	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			t.Logf("Couldn't get the HTTPRoute, err: %+v", err)
			return false
		}
		if len(httpRoute.Status.Parents) != len(ports) {
			return false
		}
		return apimeta.IsStatusConditionTrue(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionAccepted)) &&
			apimeta.IsStatusConditionFalse(httpRoute.Status.Parents[1].Conditions, string(gatewayv1.RouteConditionAccepted))
	}, 30*time.Second).Should(gomega.Equal(true))

	conditionMap := make(map[string][]metav1.Condition)
	conditionMap[fmt.Sprintf("%s-%d", gatewayName, 8080)] = []metav1.Condition{
		{
			Type:    string(gatewayv1.RouteConditionAccepted),
			Reason:  string(gatewayv1.RouteReasonAccepted),
			Status:  metav1.ConditionTrue,
			Message: "Parent reference is valid",
		},
	}
	conditionMap[fmt.Sprintf("%s-%d", gatewayName, 8081)] = []metav1.Condition{
		{
			Type:    string(gatewayv1.RouteConditionAccepted),
			Reason:  string(gatewayv1.RouteReasonNoMatchingParent),
			Status:  metav1.ConditionFalse,
			Message: "Invalid listener name provided",
		},
	}

	expectedRouteStatus := akogatewayapitests.GetRouteStatusV1([]string{gatewayName}, namespace, ports, conditionMap)

	httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
	if err != nil || httpRoute == nil {
		t.Fatalf("Couldn't get the HTTPRoute, err: %+v", err)
	}
	akogatewayapitests.ValidateHTTPRouteStatus(t, &httpRoute.Status, &gatewayv1.HTTPRouteStatus{RouteStatus: *expectedRouteStatus})

	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	akogatewayapitests.TeardownGateway(t, gatewayName, namespace)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

/* Transition test cases
 * - HTTPRoute transition from invalid to valid
 * - HTTPRoute transition from valid to invalid
 */
func TestHTTPRouteTransitionFromInvalidToValid(t *testing.T) {
	gatewayClassName := "gateway-class-hr-03"
	gatewayName := "gateway-hr-03"
	httpRouteName := "httproute-03"
	namespace := "default"
	ports := []int32{8080}

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)

	// creates a gateway with listeners 8080
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, namespace, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		gateway, err := akogatewayapitests.GatewayClient.GatewayV1().Gateways(namespace).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.IsStatusConditionTrue(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted))
	}, 30*time.Second).Should(gomega.Equal(true))

	// creates an invalid httproute with parent which has listeners 8081
	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, []int32{8081})
	hostnames := []gatewayv1.Hostname{"foo-8081.com"}

	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, nil)

	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			t.Logf("Couldn't get the HTTPRoute, err: %+v", err)
			return false
		}
		if len(httpRoute.Status.Parents) != len(ports) {
			return false
		}
		return apimeta.IsStatusConditionFalse(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionAccepted))
	}, 30*time.Second).Should(gomega.Equal(true))

	conditionMap := make(map[string][]metav1.Condition)
	conditionMap[fmt.Sprintf("%s-%d", gatewayName, 8081)] = []metav1.Condition{
		{
			Type:    string(gatewayv1.RouteConditionAccepted),
			Reason:  string(gatewayv1.RouteReasonNoMatchingParent),
			Status:  metav1.ConditionFalse,
			Message: "Invalid listener name provided",
		},
	}

	expectedRouteStatus := akogatewayapitests.GetRouteStatusV1([]string{gatewayName}, namespace, []int32{8081}, conditionMap)

	httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
	if err != nil || httpRoute == nil {
		t.Fatalf("Couldn't get the HTTPRoute, err: %+v", err)
	}
	akogatewayapitests.ValidateHTTPRouteStatus(t, &httpRoute.Status, &gatewayv1.HTTPRouteStatus{RouteStatus: *expectedRouteStatus})

	// update the httproute with valid configuration
	parentRefs = akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, ports)
	hostnames = []gatewayv1.Hostname{"foo-8080.com"}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, nil)

	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			t.Logf("Couldn't get the HTTPRoute, err: %+v", err)
			return false
		}
		if len(httpRoute.Status.Parents) != len(ports) {
			return false
		}
		return apimeta.IsStatusConditionTrue(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionAccepted))
	}, 30*time.Second).Should(gomega.Equal(true))

	conditionMap[fmt.Sprintf("%s-%d", gatewayName, 8080)] = []metav1.Condition{
		{
			Type:    string(gatewayv1.RouteConditionAccepted),
			Reason:  string(gatewayv1.RouteReasonAccepted),
			Status:  metav1.ConditionTrue,
			Message: "Parent reference is valid",
		},
	}
	expectedRouteStatus = akogatewayapitests.GetRouteStatusV1([]string{gatewayName}, namespace, ports, conditionMap)
	httpRoute, err = akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
	if err != nil || httpRoute == nil {
		t.Fatalf("Couldn't get the HTTPRoute, err: %+v", err)
	}
	akogatewayapitests.ValidateHTTPRouteStatus(t, &httpRoute.Status, &gatewayv1.HTTPRouteStatus{RouteStatus: *expectedRouteStatus})

	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	akogatewayapitests.TeardownGateway(t, gatewayName, namespace)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteTransitionFromValidToInvalid(t *testing.T) {
	gatewayClassName := "gateway-class-hr-04"
	gatewayName := "gateway-hr-04"
	httpRouteName := "httproute-04"
	namespace := "default"
	ports := []int32{8080}

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)

	// creates a gateway with listeners 8080
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

	// creates an invalid httproute with parent which has listeners 8080
	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, ports)
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}

	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, nil)

	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			t.Logf("Couldn't get the HTTPRoute, err: %+v", err)
			return false
		}
		if len(httpRoute.Status.Parents) != len(ports) {
			return false
		}
		return apimeta.IsStatusConditionTrue(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionAccepted))
	}, 30*time.Second).Should(gomega.Equal(true))

	conditionMap := make(map[string][]metav1.Condition)
	conditionMap[fmt.Sprintf("%s-%d", gatewayName, 8080)] = []metav1.Condition{
		{
			Type:    string(gatewayv1.RouteConditionAccepted),
			Reason:  string(gatewayv1.RouteReasonAccepted),
			Status:  metav1.ConditionTrue,
			Message: "Parent reference is valid",
		},
	}

	expectedRouteStatus := akogatewayapitests.GetRouteStatusV1([]string{gatewayName}, namespace, ports, conditionMap)

	httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
	if err != nil || httpRoute == nil {
		t.Fatalf("Couldn't get the HTTPRoute, err: %+v", err)
	}
	akogatewayapitests.ValidateHTTPRouteStatus(t, &httpRoute.Status, &gatewayv1.HTTPRouteStatus{RouteStatus: *expectedRouteStatus})

	parentRefs = akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, []int32{8081})
	hostnames = []gatewayv1.Hostname{"foo-8080.com"}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, nil)

	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			t.Logf("Couldn't get the HTTPRoute, err: %+v", err)
			return false
		}
		if len(httpRoute.Status.Parents) != len(ports) {
			return false
		}
		return apimeta.IsStatusConditionFalse(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionAccepted))
	}, 30*time.Second).Should(gomega.Equal(true))

	conditionMap[fmt.Sprintf("%s-%d", gatewayName, 8081)] = []metav1.Condition{
		{
			Type:    string(gatewayv1.RouteConditionAccepted),
			Reason:  string(gatewayv1.RouteReasonNoMatchingParent),
			Status:  metav1.ConditionFalse,
			Message: "Invalid listener name provided",
		},
	}
	expectedRouteStatus = akogatewayapitests.GetRouteStatusV1([]string{gatewayName}, namespace, []int32{8081}, conditionMap)
	httpRoute, err = akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
	if err != nil || httpRoute == nil {
		t.Fatalf("Couldn't get the HTTPRoute, err: %+v", err)
	}
	akogatewayapitests.ValidateHTTPRouteStatus(t, &httpRoute.Status, &gatewayv1.HTTPRouteStatus{RouteStatus: *expectedRouteStatus})

	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	akogatewayapitests.TeardownGateway(t, gatewayName, namespace)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

/* Negative test cases
 * - HTTPRoute with no parent reference
 * - HTTPRoute with all parent reference invalid
 * - HTTPRoute with non existing gateway reference
 * - HTTPRoute with non existing listener reference
 * - HTTPRoute with non AKO gateway controller reference (TODO: transition case need to be taken care)
 * - HTTPRoute with no hostnames
 */
func TestHTTPRouteWithNoParentReference(t *testing.T) {
	gatewayClassName := "gateway-class-hr-05"
	gatewayName := "gateway-hr-05"
	httpRouteName := "httproute-05"
	namespace := "default"
	ports := []int32{8080}

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)

	// creates a gateway with listeners 8080
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

	// creates a httproute with no parent reference which has listeners 8080
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}

	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, nil, hostnames, nil)

	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			t.Logf("Couldn't get the HTTPRoute, err: %+v", err)
			return false
		}
		return len(httpRoute.Status.Parents) == 0
	}, 30*time.Second).Should(gomega.Equal(true))

	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	akogatewayapitests.TeardownGateway(t, gatewayName, namespace)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteWithAllParentReferenceInvalid(t *testing.T) {
	gatewayClassName := "gateway-class-hr-06"
	gatewayName := "gateway-hr-06"
	httpRouteName := "httproute-06"
	namespace := "default"
	ports := []int32{8080, 8081}

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)

	// creates a gateway with listeners 8082 and 8083
	listeners := akogatewayapitests.GetListenersV1([]int32{8082, 8083}, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, namespace, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		gateway, err := akogatewayapitests.GatewayClient.GatewayV1().Gateways(namespace).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.IsStatusConditionTrue(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted))
	}, 30*time.Second).Should(gomega.Equal(true))

	// creates a httproute with parent which has listeners 8080, 8081
	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, ports)
	hostnames := []gatewayv1.Hostname{"foo-8080.com", "foo-8081.com"}

	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, nil)

	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			t.Logf("Couldn't get the HTTPRoute, err: %+v", err)
			return false
		}
		if len(httpRoute.Status.Parents) != len(ports) {
			return false
		}
		return apimeta.IsStatusConditionFalse(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionAccepted)) &&
			apimeta.IsStatusConditionFalse(httpRoute.Status.Parents[1].Conditions, string(gatewayv1.RouteConditionAccepted))
	}, 30*time.Second).Should(gomega.Equal(true))

	conditionMap := map[string][]metav1.Condition{
		fmt.Sprintf("%s-%d", gatewayName, 8080): {
			{
				Type:    string(gatewayv1.RouteConditionAccepted),
				Reason:  string(gatewayv1.RouteReasonNoMatchingParent),
				Status:  metav1.ConditionFalse,
				Message: "Invalid listener name provided",
			},
		},
		fmt.Sprintf("%s-%d", gatewayName, 8081): {
			{
				Type:    string(gatewayv1.RouteConditionAccepted),
				Reason:  string(gatewayv1.RouteReasonNoMatchingParent),
				Status:  metav1.ConditionFalse,
				Message: "Invalid listener name provided",
			},
		},
	}

	expectedRouteStatus := akogatewayapitests.GetRouteStatusV1([]string{gatewayName}, namespace, ports, conditionMap)

	httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
	if err != nil || httpRoute == nil {
		t.Fatalf("Couldn't get the HTTPRoute, err: %+v", err)
	}
	akogatewayapitests.ValidateHTTPRouteStatus(t, &httpRoute.Status, &gatewayv1.HTTPRouteStatus{RouteStatus: *expectedRouteStatus})

	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	akogatewayapitests.TeardownGateway(t, gatewayName, namespace)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteWithNonExistingGatewayReference(t *testing.T) {
	gatewayName := "gateway-hr-07"
	httpRouteName := "httproute-07"
	namespace := "default"
	ports := []int32{8080}

	// creates a httproute with no parent reference which has listeners 8080
	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, ports)
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}

	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, nil)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			t.Logf("Couldn't get the HTTPRoute, err: %+v", err)
			return false
		}
		return len(httpRoute.Status.Parents) == 0
	}, 30*time.Second).Should(gomega.Equal(true))

	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
}

func TestHTTPRouteWithNonExistingListenerReference(t *testing.T) {
	gatewayClassName := "gateway-class-hr-08"
	gatewayName := "gateway-hr-08"
	httpRouteName := "httproute-08"
	namespace := "default"
	ports := []int32{8080, 8081}

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)

	// creates a gateway with listeners 8082 and 8083
	listeners := akogatewayapitests.GetListenersV1([]int32{8082}, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, namespace, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		gateway, err := akogatewayapitests.GatewayClient.GatewayV1().Gateways(namespace).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.IsStatusConditionTrue(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted))
	}, 30*time.Second).Should(gomega.Equal(true))

	// creates a httproute with parent which has listeners 8080, 8081
	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, ports)
	hostnames := []gatewayv1.Hostname{"foo-8080.com", "foo-8081.com"}

	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, nil)

	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			t.Logf("Couldn't get the HTTPRoute, err: %+v", err)
			return false
		}
		if len(httpRoute.Status.Parents) != len(ports) {
			return false
		}
		return apimeta.IsStatusConditionFalse(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionAccepted)) &&
			apimeta.IsStatusConditionFalse(httpRoute.Status.Parents[1].Conditions, string(gatewayv1.RouteConditionAccepted))
	}, 30*time.Second).Should(gomega.Equal(true))

	conditionMap := map[string][]metav1.Condition{
		fmt.Sprintf("%s-%d", gatewayName, 8080): {
			{
				Type:    string(gatewayv1.RouteConditionAccepted),
				Reason:  string(gatewayv1.RouteReasonNoMatchingParent),
				Status:  metav1.ConditionFalse,
				Message: "Invalid listener name provided",
			},
		},
		fmt.Sprintf("%s-%d", gatewayName, 8081): {
			{
				Type:    string(gatewayv1.RouteConditionAccepted),
				Reason:  string(gatewayv1.RouteReasonNoMatchingParent),
				Status:  metav1.ConditionFalse,
				Message: "Invalid listener name provided",
			},
		},
	}

	expectedRouteStatus := akogatewayapitests.GetRouteStatusV1([]string{gatewayName}, namespace, ports, conditionMap)

	httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
	if err != nil || httpRoute == nil {
		t.Fatalf("Couldn't get the HTTPRoute, err: %+v", err)
	}
	akogatewayapitests.ValidateHTTPRouteStatus(t, &httpRoute.Status, &gatewayv1.HTTPRouteStatus{RouteStatus: *expectedRouteStatus})

	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	akogatewayapitests.TeardownGateway(t, gatewayName, namespace)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteWithNoHostnames(t *testing.T) {
	gatewayClassName := "gateway-class-hr-10"
	gatewayName := "gateway-hr-10"
	httpRouteName := "httproute-10"
	namespace := "default"
	ports := []int32{8080, 8081}

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

	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, nil, nil)

	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			t.Logf("Couldn't get the HTTPRoute, err: %+v", err)
			return false
		}
		if len(httpRoute.Status.Parents) != len(ports) {
			return false
		}
		return apimeta.IsStatusConditionTrue(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.GatewayConditionAccepted)) &&
			apimeta.IsStatusConditionTrue(httpRoute.Status.Parents[1].Conditions, string(gatewayv1.GatewayConditionAccepted))
	}, 30*time.Second).Should(gomega.Equal(true))

	conditionMap := map[string][]metav1.Condition{
		fmt.Sprintf("%s-%d", gatewayName, 8080): {
			{
				Type:    string(gatewayv1.GatewayConditionAccepted),
				Reason:  string(gatewayv1.GatewayReasonAccepted),
				Status:  metav1.ConditionTrue,
				Message: "Parent reference is valid",
			},
		},
		fmt.Sprintf("%s-%d", gatewayName, 8081): {
			{
				Type:    string(gatewayv1.GatewayConditionAccepted),
				Reason:  string(gatewayv1.GatewayReasonAccepted),
				Status:  metav1.ConditionTrue,
				Message: "Parent reference is valid",
			},
		},
	}
	expectedRouteStatus := akogatewayapitests.GetRouteStatusV1([]string{gatewayName}, namespace, ports, conditionMap)

	httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
	if err != nil || httpRoute == nil {
		t.Fatalf("Couldn't get the HTTPRoute, err: %+v", err)
	}
	akogatewayapitests.ValidateHTTPRouteStatus(t, &httpRoute.Status, &gatewayv1.HTTPRouteStatus{RouteStatus: *expectedRouteStatus})

	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	akogatewayapitests.TeardownGateway(t, gatewayName, namespace)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteUnprocessedGateway(t *testing.T) {
	t.Skip("Skipping this test case as we are unable to enforce the expected race condition")
	gatewayClassName := "gateway-class-hr-11"
	gatewayName := "gateway-hr-11"
	httpRouteName := "httproute-11"
	namespace := "default"
	ports := []int32{8080}

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)

	listeners := akogatewayapitests.GetListenersV1(ports, false, false)

	g := gomega.NewGomegaWithT(t)

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, ports)
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}

	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, nil)
	akogatewayapitests.SetupGateway(t, gatewayName, namespace, gatewayClassName, nil, listeners)

	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			t.Logf("Couldn't get the HTTPRoute, err: %+v", err)
			return false
		}
		if len(httpRoute.Status.Parents) != len(ports) {
			return false
		}
		return apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	conditionMap := make(map[string][]metav1.Condition)

	for _, port := range ports {
		conditions := make([]metav1.Condition, 0, 1)
		condition := metav1.Condition{
			Type:    string(gatewayv1.RouteConditionAccepted),
			Reason:  string(gatewayv1.RouteReasonPending),
			Status:  metav1.ConditionFalse,
			Message: "AKO is yet to process Gateway gateway-hr-11 for parent reference gateway-hr-11",
		}
		conditions = append(conditions, condition)
		conditionMap[fmt.Sprintf("%s-%d", gatewayName, port)] = conditions
	}
	expectedRouteStatus := akogatewayapitests.GetRouteStatusV1([]string{gatewayName}, namespace, ports, conditionMap)

	httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
	if err != nil || httpRoute == nil {
		t.Fatalf("Couldn't get the HTTPRoute, err: %+v", err)
	}
	akogatewayapitests.ValidateHTTPRouteStatus(t, &httpRoute.Status, &gatewayv1.HTTPRouteStatus{RouteStatus: *expectedRouteStatus})

	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	akogatewayapitests.TeardownGateway(t, gatewayName, namespace)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteWithInvalidGatewayListener(t *testing.T) {
	gatewayClassName := "gateway-class-hr-12"
	gatewayName := "gateway-hr-12"
	httpRouteName := "httproute-12"
	namespace := "default"
	ports := []int32{8080}

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)

	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	listeners[0].Hostname = nil

	g := gomega.NewGomegaWithT(t)
	akogatewayapitests.SetupGateway(t, gatewayName, namespace, gatewayClassName, nil, listeners)
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

	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, nil)
	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			t.Logf("Couldn't get the HTTPRoute, err: %+v", err)
			return false
		}
		if len(httpRoute.Status.Parents) != len(ports) {
			return false
		}
		return apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	conditionMap := make(map[string][]metav1.Condition)

	for _, port := range ports {
		conditions := make([]metav1.Condition, 0, 1)
		condition := metav1.Condition{
			Type:    string(gatewayv1.GatewayConditionAccepted),
			Reason:  string(gatewayv1.GatewayReasonAccepted),
			Status:  metav1.ConditionTrue,
			Message: "Parent reference is valid",
		}
		conditions = append(conditions, condition)
		conditionMap[fmt.Sprintf("%s-%d", gatewayName, port)] = conditions
	}
	expectedRouteStatus := akogatewayapitests.GetRouteStatusV1([]string{gatewayName}, namespace, ports, conditionMap)

	httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
	if err != nil || httpRoute == nil {
		t.Fatalf("Couldn't get the HTTPRoute, err: %+v", err)
	}
	akogatewayapitests.ValidateHTTPRouteStatus(t, &httpRoute.Status, &gatewayv1.HTTPRouteStatus{RouteStatus: *expectedRouteStatus})

	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	akogatewayapitests.TeardownGateway(t, gatewayName, namespace)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}
func TestHTTPRouteWithOneExistingAndOneNonExistingGateway(t *testing.T) {
	gatewayClassName := "gateway-class-hr-13"
	gatewayName1 := "Non-Existing-Gateway"
	gatewayName2 := "gateway-hr-13"
	httpRouteName := "httproute-13"
	namespace := "default"
	ports := []int32{8080}

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)

	listeners := akogatewayapitests.GetListenersV1(ports, false, false)

	g := gomega.NewGomegaWithT(t)
	akogatewayapitests.SetupGateway(t, gatewayName2, namespace, gatewayClassName, nil, listeners)
	g.Eventually(func() bool {
		gateway, err := akogatewayapitests.GatewayClient.GatewayV1().Gateways(namespace).Get(context.TODO(), gatewayName2, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName1, gatewayName2}, namespace, ports)
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}

	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, nil)
	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			t.Logf("Couldn't get the HTTPRoute, err: %+v", err)
			return false
		}
		if len(httpRoute.Status.Parents) != 1 {
			return false
		}
		return apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	conditionMap := make(map[string][]metav1.Condition)

	for _, port := range ports {
		conditions := make([]metav1.Condition, 0, 1)
		condition := metav1.Condition{
			Type:    string(gatewayv1.RouteConditionAccepted),
			Reason:  string(gatewayv1.RouteReasonAccepted),
			Status:  metav1.ConditionTrue,
			Message: "Parent reference is valid",
		}
		conditions = append(conditions, condition)
		conditionMap[fmt.Sprintf("%s-%d", gatewayName2, port)] = conditions
	}
	expectedRouteStatus := akogatewayapitests.GetRouteStatusV1([]string{gatewayName2}, namespace, ports, conditionMap)

	httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
	if err != nil || httpRoute == nil {
		t.Fatalf("Couldn't get the HTTPRoute, err: %+v", err)
	}
	akogatewayapitests.ValidateHTTPRouteStatus(t, &httpRoute.Status, &gatewayv1.HTTPRouteStatus{RouteStatus: *expectedRouteStatus})

	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	akogatewayapitests.TeardownGateway(t, gatewayName2, namespace)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestMultipleHttpRoutesWithValidAndInvalidGatewayListeners(t *testing.T) {
	gatewayClassName := "gateway-class-hr-14"
	gatewayName := "gateway-hr-14"
	httpRouteName1 := "httproute-14a"
	httpRouteName2 := "httproute-14b"
	namespace := "default"
	ports := []int32{8080, 8081}

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)

	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	listeners[1].Protocol = "TCP"

	g := gomega.NewGomegaWithT(t)
	akogatewayapitests.SetupGateway(t, gatewayName, namespace, gatewayClassName, nil, listeners)
	g.Eventually(func() bool {
		gateway, err := akogatewayapitests.GatewayClient.GatewayV1().Gateways(namespace).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, []int32{ports[0]})
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}

	akogatewayapitests.SetupHTTPRoute(t, httpRouteName1, namespace, parentRefs, hostnames, nil)
	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName1, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			t.Logf("Couldn't get the HTTPRoute, err: %+v", err)
			return false
		}
		if len(httpRoute.Status.Parents) != 1 {
			return false
		}
		return apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName1, metav1.GetOptions{})
	if err != nil || httpRoute == nil {
		t.Fatalf("Couldn't get the HTTPRoute, err: %+v", err)
	}

	conditionMap := make(map[string][]metav1.Condition)

	conditions := make([]metav1.Condition, 0)
	condition := metav1.Condition{
		Type:    string(gatewayv1.RouteConditionAccepted),
		Reason:  string(gatewayv1.RouteReasonAccepted),
		Status:  metav1.ConditionTrue,
		Message: "Parent reference is valid",
	}
	conditions = append(conditions, condition)
	conditionMap[fmt.Sprintf("%s-%d", gatewayName, ports[0])] = conditions

	expectedRouteStatus := akogatewayapitests.GetRouteStatusV1([]string{gatewayName}, namespace, []int32{ports[0]}, conditionMap)
	akogatewayapitests.ValidateHTTPRouteStatus(t, &httpRoute.Status, &gatewayv1.HTTPRouteStatus{RouteStatus: *expectedRouteStatus})

	parentRefs = akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, []int32{ports[1]})
	hostnames = []gatewayv1.Hostname{"foo-8081.com"}

	akogatewayapitests.SetupHTTPRoute(t, httpRouteName2, namespace, parentRefs, hostnames, nil)
	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName2, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			t.Logf("Couldn't get the HTTPRoute, err: %+v", err)
			return false
		}
		if len(httpRoute.Status.Parents) != 1 {
			return false
		}
		return apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	httpRoute, err = akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName2, metav1.GetOptions{})
	if err != nil || httpRoute == nil {
		t.Fatalf("Couldn't get the HTTPRoute, err: %+v", err)
	}

	conditionMap = make(map[string][]metav1.Condition)
	conditions = make([]metav1.Condition, 0, 1)
	condition = metav1.Condition{
		Type:    string(gatewayv1.RouteConditionAccepted),
		Reason:  string(gatewayv1.RouteReasonAccepted),
		Status:  metav1.ConditionTrue,
		Message: "Parent reference is valid",
	}

	conditions = append(conditions, condition)
	conditionMap[fmt.Sprintf("%s-%d", gatewayName, ports[1])] = conditions

	conditionMap["gateway-hr-14-8081"][0].Message = "Matching gateway listener is in Invalid state"
	conditionMap["gateway-hr-14-8081"][0].Status = metav1.ConditionFalse
	conditionMap["gateway-hr-14-8081"][0].Reason = string(gatewayv1.RouteReasonPending)

	expectedRouteStatus = akogatewayapitests.GetRouteStatusV1([]string{gatewayName}, namespace, []int32{ports[1]}, conditionMap)
	akogatewayapitests.ValidateHTTPRouteStatus(t, &httpRoute.Status, &gatewayv1.HTTPRouteStatus{RouteStatus: *expectedRouteStatus})

	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName2, namespace)
	akogatewayapitests.TeardownGateway(t, gatewayName, namespace)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHttpRouteWithValidAndInvalidGatewayListeners(t *testing.T) {
	gatewayClassName := "gateway-class-hr-15"
	gatewayName := "gateway-hr-15"
	httpRouteName := "httproute-15"
	namespace := "default"
	ports := []int32{8080, 8081}

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)

	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	listeners[1].Protocol = "TCP"

	g := gomega.NewGomegaWithT(t)
	akogatewayapitests.SetupGateway(t, gatewayName, namespace, gatewayClassName, nil, listeners)
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

	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, nil)
	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			t.Logf("Couldn't get the HTTPRoute, err: %+v", err)
			return false
		}
		if len(httpRoute.Status.Parents) != 2 {
			return false
		}
		return apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	conditionMap := make(map[string][]metav1.Condition)

	for _, port := range ports {
		conditions := []metav1.Condition{{
			Type:    string(gatewayv1.RouteConditionAccepted),
			Reason:  string(gatewayv1.RouteReasonAccepted),
			Status:  metav1.ConditionTrue,
			Message: "Parent reference is valid",
		}}
		conditionMap[fmt.Sprintf("%s-%d", gatewayName, port)] = conditions
	}
	conditionMap["gateway-hr-15-8081"][0].Message = "Matching gateway listener is in Invalid state"
	conditionMap["gateway-hr-15-8081"][0].Status = metav1.ConditionFalse
	conditionMap["gateway-hr-15-8081"][0].Reason = string(gatewayv1.RouteReasonPending)

	expectedRouteStatus := akogatewayapitests.GetRouteStatusV1([]string{gatewayName}, namespace, ports, conditionMap)

	httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
	if err != nil || httpRoute == nil {
		t.Fatalf("Couldn't get the HTTPRoute, err: %+v", err)
	}
	akogatewayapitests.ValidateHTTPRouteStatus(t, &httpRoute.Status, &gatewayv1.HTTPRouteStatus{RouteStatus: *expectedRouteStatus})

	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	akogatewayapitests.TeardownGateway(t, gatewayName, namespace)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteWithInvalidBackendKind(t *testing.T) {
	gatewayClassName := "gateway-class-hr-16"
	gatewayName := "gateway-hr-16"
	httpRouteName := "httproute-16"
	namespace := "default"
	ports := []int32{8080}

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
	rule := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{"avisvc", "default", "8080", "1"}}, nil)
	kind := gatewayv1.Kind("InvalidKind")
	rule.BackendRefs[0].BackendRef.Kind = &kind
	rules := []gatewayv1.HTTPRouteRule{rule}
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}

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
		return apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	conditionMap := make(map[string][]metav1.Condition)

	for _, port := range ports {
		conditions := []metav1.Condition{{
			Type:    string(gatewayv1.RouteConditionAccepted),
			Reason:  string(gatewayv1.RouteReasonAccepted),
			Status:  metav1.ConditionTrue,
			Message: "Parent reference is valid",
		}, {
			Type:    string(gatewayv1.RouteConditionResolvedRefs),
			Reason:  string(gatewayv1.RouteReasonInvalidKind),
			Status:  metav1.ConditionFalse,
			Message: "backendRef avisvc has invalid kind InvalidKind",
		},
		}
		conditionMap[fmt.Sprintf("%s-%d", gatewayName, port)] = conditions
	}
	expectedRouteStatus := akogatewayapitests.GetRouteStatusV1([]string{gatewayName}, namespace, ports, conditionMap)

	httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
	if err != nil || httpRoute == nil {
		t.Fatalf("Couldn't get the HTTPRoute, err: %+v", err)
	}
	akogatewayapitests.ValidateHTTPRouteStatus(t, &httpRoute.Status, &gatewayv1.HTTPRouteStatus{RouteStatus: *expectedRouteStatus})

	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	akogatewayapitests.TeardownGateway(t, gatewayName, namespace)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}
func TestHTTPRouteWithValidAndInvalidBackendKind(t *testing.T) {
	gatewayClassName := "gateway-class-hr-17"
	gatewayName := "gateway-hr-17"
	httpRouteName := "httproute-17"
	namespace := "default"
	ports := []int32{8080}
	svcName := "avisvc-hr-17"

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	integrationtest.CreateSVC(t, DEFAULT_NAMESPACE, svcName, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, "default", svcName, false, false, "1.1.1")

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
	rule := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add", "remove", "replace"}},
		[][]string{{"avisvc", "default", "8080", "1"}, {svcName, "default", "8080", "1"}}, nil)
	kind := gatewayv1.Kind("InvalidKind")
	rule.BackendRefs[0].BackendRef.Kind = &kind
	rules := []gatewayv1.HTTPRouteRule{rule}
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}

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
		return apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	conditionMap := make(map[string][]metav1.Condition)

	for _, port := range ports {
		conditions := []metav1.Condition{{
			Type:    string(gatewayv1.RouteConditionAccepted),
			Reason:  string(gatewayv1.RouteReasonAccepted),
			Status:  metav1.ConditionTrue,
			Message: "Parent reference is valid",
		}, {
			Type:    string(gatewayv1.RouteConditionResolvedRefs),
			Reason:  string(gatewayv1.RouteReasonInvalidKind),
			Status:  metav1.ConditionFalse,
			Message: "backendRef avisvc has invalid kind InvalidKind",
		},
		}
		conditionMap[fmt.Sprintf("%s-%d", gatewayName, port)] = conditions
	}
	expectedRouteStatus := akogatewayapitests.GetRouteStatusV1([]string{gatewayName}, namespace, ports, conditionMap)

	httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
	if err != nil || httpRoute == nil {
		t.Fatalf("Couldn't get the HTTPRoute, err: %+v", err)
	}
	akogatewayapitests.ValidateHTTPRouteStatus(t, &httpRoute.Status, &gatewayv1.HTTPRouteStatus{RouteStatus: *expectedRouteStatus})

	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	akogatewayapitests.TeardownGateway(t, gatewayName, namespace)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}
func TestHTTPRouteGatewayWithEmptyHostnameInGatewayHTTPRoute(t *testing.T) {
	gatewayName := "gateway-hr-18"
	gatewayClassName := "gateway-class-hr-18"
	httpRouteName := "http-route-hr-18"
	svcName := "avisvc-hr-18"
	namespace := "default"
	ports := []int32{8080}
	modelName, _ := akogatewayapitests.GetModelName(namespace, gatewayName)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)

	listeners := akogatewayapitests.GetListenersV1(ports, true, false)

	akogatewayapitests.SetupGateway(t, gatewayName, namespace, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 5*time.Second).Should(gomega.Equal(true))

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{}, nil,
		[][]string{{svcName, namespace, "8080", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{rule}
	hostnames := []gatewayv1.Hostname{}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, rules)

	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			t.Logf("Couldn't get the HTTPRoute, err: %+v", err)
			return false
		}
		if len(httpRoute.Status.Parents) != 1 {
			return false
		}
		return apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	conditionMap := make(map[string][]metav1.Condition)

	for _, port := range ports {
		conditions := make([]metav1.Condition, 0, 1)
		condition := metav1.Condition{
			Type:    string(gatewayv1.RouteConditionAccepted),
			Reason:  string(gatewayv1.RouteReasonNoMatchingListenerHostname),
			Status:  metav1.ConditionFalse,
			Message: "Hostname in Gateway Listener doesn't match with any of the hostnames in HTTPRoute",
		}
		conditions = append(conditions, condition)
		conditionMap[fmt.Sprintf("%s-%d", gatewayName, port)] = conditions
	}
	expectedRouteStatus := akogatewayapitests.GetRouteStatusV1([]string{gatewayName}, namespace, ports, conditionMap)

	httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
	if err != nil || httpRoute == nil {
		t.Fatalf("Couldn't get the HTTPRoute, err: %+v", err)
	}
	akogatewayapitests.ValidateHTTPRouteStatus(t, &httpRoute.Status, &gatewayv1.HTTPRouteStatus{RouteStatus: *expectedRouteStatus})

	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}
func TestHTTPRouteFilterWithUnsupportedUrlRewritePathType(t *testing.T) {

	gatewayName := "gateway-hr-19"
	gatewayClassName := "gateway-class-hr-19"
	httpRouteName := "http-route-hr-19"
	svcName := "avisvc-hr-19"
	namespace := "default"
	ports := []int32{8080}

	modelName, _ := akogatewayapitests.GetModelName(namespace, gatewayName)
	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, namespace, gatewayClassName, nil, listeners)
	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 5*time.Second).Should(gomega.Equal(true))

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"URLRewrite": {}},
		[][]string{{svcName, "default", "8080", "1"}}, nil)
	rule.Filters[0].URLRewrite.Path.Type = gatewayv1.PrefixMatchHTTPPathModifier
	rules := []gatewayv1.HTTPRouteRule{rule}
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, rules)

	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			t.Logf("Couldn't get the HTTPRoute, err: %+v", err)
			return false
		}
		if len(httpRoute.Status.Parents) != 1 {
			return false
		}
		return apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	conditionMap := make(map[string][]metav1.Condition)

	for _, port := range ports {
		conditions := make([]metav1.Condition, 0, 1)
		condition := metav1.Condition{
			Type:    string(gatewayv1.RouteConditionAccepted),
			Reason:  string(gatewayv1.RouteReasonUnsupportedValue),
			Status:  metav1.ConditionFalse,
			Message: "HTTPUrlRewrite PathType has Unsupported value",
		}
		conditions = append(conditions, condition)
		conditionMap[fmt.Sprintf("%s-%d", gatewayName, port)] = conditions
	}
	expectedRouteStatus := akogatewayapitests.GetRouteStatusV1([]string{gatewayName}, namespace, ports, conditionMap)

	httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
	if err != nil || httpRoute == nil {
		t.Fatalf("Couldn't get the HTTPRoute, err: %+v", err)
	}
	akogatewayapitests.ValidateHTTPRouteStatus(t, &httpRoute.Status, &gatewayv1.HTTPRouteStatus{RouteStatus: *expectedRouteStatus})

	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteStatusWithHealthMonitorLifecycle(t *testing.T) {
	gatewayClassName := "gateway-class-hm-lifecycle"
	gatewayName := "gateway-hm-lifecycle"
	httpRouteName := "httproute-hm-lifecycle"
	healthMonitorName := "hm-lifecycle"
	namespace := "default"
	svcName := "avisvc-hm-lifecycle"
	ports := []int32{8080}

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	integrationtest.CreateSVC(t, namespace, svcName, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, namespace, svcName, false, false, "1.1.1")

	listeners := akogatewayapitests.GetListenersV1(ports, true, false)
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

	// Create HTTPRoute with reference to non-existent HealthMonitor
	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, ports)
	hostnames := []gatewayv1.Hostname{"foo-hm-lifecycle.com"}
	rules := []gatewayv1.HTTPRouteRule{
		akogatewayapitests.GetHTTPRouteRuleWithHealthMonitorFilters(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
			map[string][]string{"RequestHeaderModifier": {"add"}},
			[][]string{{svcName, namespace, "8080", "1"}}, []string{healthMonitorName}),
	}

	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, rules)

	// HTTPRoute should have unresolved refs condition due to non-existent HealthMonitor
	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			t.Logf("Couldn't get the HTTPRoute, err: %+v", err)
			return false
		}
		if len(httpRoute.Status.Parents) != len(ports) {
			return false
		}
		condition := apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs))
		return condition != nil && condition.Status == metav1.ConditionFalse &&
			condition.Reason == string(gatewayv1.RouteReasonBackendNotFound)
	}, 30*time.Second).Should(gomega.Equal(true))

	// Verify error message mentions HealthMonitor
	httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
	if err != nil || httpRoute == nil {
		t.Fatalf("Couldn't get the HTTPRoute, err: %+v", err)
	}
	condition := apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs))
	g.Expect(condition.Message).To(gomega.ContainSubstring("healthMonitor default/hm-lifecycle not found"))

	// Create HealthMonitor with Ready=False status
	akogatewayapitests.CreateHealthMonitorCRD(t, healthMonitorName, namespace, "thisisaviref-hm-lifecycle")
	akogatewayapitests.UpdateHealthMonitorStatus(t, healthMonitorName, namespace, false, "ValidationError", "HealthMonitor configuration is invalid")

	// HTTPRoute should still have unresolved refs condition due to unready HealthMonitor
	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			return false
		}
		if len(httpRoute.Status.Parents) != len(ports) {
			return false
		}
		condition := apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs))
		return condition != nil && condition.Status == metav1.ConditionFalse &&
			condition.Reason == string(gatewayv1.RouteReasonBackendNotFound)
	}, 30*time.Second).Should(gomega.Equal(true))

	// Update HealthMonitor to Ready=True
	akogatewayapitests.UpdateHealthMonitorStatus(t, healthMonitorName, namespace, true, "Accepted", "HealthMonitor has been successfully processed")

	// HTTPRoute should now have resolved refs condition
	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			return false
		}
		if len(httpRoute.Status.Parents) != len(ports) {
			return false
		}
		condition := apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs))
		return condition != nil && condition.Status == metav1.ConditionTrue
	}, 30*time.Second).Should(gomega.Equal(true))

	// Test status transition from Ready=True to Ready=False
	akogatewayapitests.UpdateHealthMonitorStatus(t, healthMonitorName, namespace, false, "ValidationError", "HealthMonitor configuration became invalid")

	// HTTPRoute should now have unresolved refs condition
	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			return false
		}
		if len(httpRoute.Status.Parents) != len(ports) {
			return false
		}
		condition := apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs))
		return condition != nil && condition.Status == metav1.ConditionFalse &&
			condition.Reason == string(gatewayv1.RouteReasonBackendNotFound)
	}, 30*time.Second).Should(gomega.Equal(true))

	// Transition back to Ready=True
	akogatewayapitests.UpdateHealthMonitorStatus(t, healthMonitorName, namespace, true, "Accepted", "HealthMonitor has been successfully processed again")

	// HTTPRoute should have resolved refs condition again
	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			return false
		}
		if len(httpRoute.Status.Parents) != len(ports) {
			return false
		}
		condition := apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs))
		return condition != nil && condition.Status == metav1.ConditionTrue
	}, 30*time.Second).Should(gomega.Equal(true))

	//Test HealthMonitor deletion
	akogatewayapitests.DeleteHealthMonitorCRD(t, healthMonitorName, namespace)

	// HTTPRoute should now have unresolved refs condition due to deleted HealthMonitor
	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			return false
		}
		if len(httpRoute.Status.Parents) != len(ports) {
			return false
		}
		condition := apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs))
		return condition != nil && condition.Status == metav1.ConditionFalse &&
			condition.Reason == string(gatewayv1.RouteReasonBackendNotFound)
	}, 30*time.Second).Should(gomega.Equal(true))

	// Verify the condition message mentions the HealthMonitor
	httpRoute, err = akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
	if err != nil || httpRoute == nil {
		t.Fatalf("Couldn't get the HTTPRoute, err: %+v", err)
	}
	condition = apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs))
	g.Expect(condition.Message).To(gomega.ContainSubstring("healthMonitor default/hm-lifecycle not found"))

	// Recreate HealthMonitor to verify HTTPRoute status recovers
	akogatewayapitests.CreateHealthMonitorCRD(t, healthMonitorName, namespace, "thisisaviref-hm-lifecycle-recreated")
	akogatewayapitests.UpdateHealthMonitorStatus(t, healthMonitorName, namespace, true, "Accepted", "HealthMonitor has been successfully processed")

	// HTTPRoute should have resolved refs condition again
	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			return false
		}
		if len(httpRoute.Status.Parents) != len(ports) {
			return false
		}
		condition := apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs))
		return condition != nil && condition.Status == metav1.ConditionTrue
	}, 30*time.Second).Should(gomega.Equal(true))

	// Cleanup
	akogatewayapitests.DeleteHealthMonitorCRD(t, healthMonitorName, namespace)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	akogatewayapitests.TeardownGateway(t, gatewayName, namespace)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteStatusWithCrossNamespaceHealthMonitorDifferentTenant(t *testing.T) {
	gatewayClassName := "gateway-class-hm-status-04"
	gatewayName := "gateway-hm-status-04"
	httpRouteName := "httproute-hm-status-04"
	healthMonitorName := "hm-status-04-cross-ns"
	namespace := "default"
	healthMonitorNamespace := "hm-namespace-04"
	svcName := "avisvc-hm-status-04"
	ports := []int32{8080}

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	integrationtest.CreateSVC(t, namespace, svcName, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, namespace, svcName, false, false, "1.1.1")

	// Create the HealthMonitor namespace and annotate it with a different tenant
	integrationtest.AddNamespace(t, healthMonitorNamespace, map[string]string{})
	integrationtest.AnnotateNamespaceWithTenant(t, healthMonitorNamespace, "tenant-01")

	listeners := akogatewayapitests.GetListenersV1(ports, true, false)
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

	// Create HealthMonitor CRD in different namespace with different tenant
	akogatewayapitests.CreateHealthMonitorCRDWithStatus(t, healthMonitorName, healthMonitorNamespace, "thisisaviref-hm-status-04", true, "Accepted", "HealthMonitor has been successfully processed")

	// Create HTTPRoute with reference to HealthMonitor in different namespace
	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, ports)
	hostnames := []gatewayv1.Hostname{"foo-hm-status-04.com"}

	// Create rule with cross-namespace HealthMonitor reference
	rule := akogatewayapitests.GetHTTPRouteRuleWithHealthMonitorFilters(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName, healthMonitorNamespace, "8080", "1"}}, []string{healthMonitorName})

	rules := []gatewayv1.HTTPRouteRule{rule}

	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, rules)

	// HTTPRoute should have unresolved refs condition due to cross-namespace access with different tenant
	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			t.Logf("Couldn't get the HTTPRoute, err: %+v", err)
			return false
		}
		if len(httpRoute.Status.Parents) != len(ports) {
			return false
		}
		// Should have unresolved refs condition due to tenant isolation
		condition := apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs))
		return condition != nil && condition.Status == metav1.ConditionFalse &&
			condition.Reason == string(gatewayv1.RouteReasonRefNotPermitted)
	}, 30*time.Second).Should(gomega.Equal(true))

	// Verify the condition message mentions tenant isolation
	httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
	if err != nil || httpRoute == nil {
		t.Fatalf("Couldn't get the HTTPRoute, err: %+v", err)
	}
	condition := apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs))
	g.Expect(condition.Message).To(gomega.ContainSubstring("tenant tenant-01 is not equal to HTTPRoute tenant admin"))

	// Cleanup
	akogatewayapitests.DeleteHealthMonitorCRD(t, healthMonitorName, healthMonitorNamespace)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	akogatewayapitests.TeardownGateway(t, gatewayName, namespace)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteMultpleRouteBackendExtensionSingleBackend(t *testing.T) {
	gatewayName := "gateway-rbe-01"
	gatewayClassName := "gateway-class-rbe-01"
	httpRouteName := "http-route-rbe-01"
	svcName := "avisvc-rbe-01"
	routeBackendExtensionName1 := "rbe-01a"
	routeBackendExtensionName2 := "rbe-01b"

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
	integrationtest.CreateEPorEPS(t, DEFAULT_NAMESPACE, svcName, false, false, "1.2.3")

	// Create RouteBackendExtension CRs
	rbe1 := akogatewayapitests.GetFakeDefaultRBEObj(routeBackendExtensionName1, DEFAULT_NAMESPACE, "thisisaviref-hm1")
	rbe1.CreateRouteBackendExtensionCRWithStatus(t)
	rbe2 := akogatewayapitests.GetFakeDefaultRBEObj(routeBackendExtensionName2, DEFAULT_NAMESPACE, "thisisaviref-hm2")
	rbe2.CreateRouteBackendExtensionCRWithStatus(t)

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)

	// Create HTTPRoute with single rule, with single backend that specifies multiple routeBackendExtensions
	rule := akogatewayapitests.GetHTTPRouteRuleWithRouteBackendExtensionAndHMFilters(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, []string{routeBackendExtensionName1, routeBackendExtensionName2})

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
	}, 25*time.Second).Should(gomega.Equal(0))

	// HTTPRoute should have unresolved refs condition due to non-existent RouteBackendExtension
	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(DEFAULT_NAMESPACE).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			t.Logf("Couldn't get the HTTPRoute, err: %+v", err)
			return false
		}
		if len(httpRoute.Status.Parents) != len(ports) {
			return false
		}
		condition := apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs))
		return condition != nil && condition.Status == metav1.ConditionFalse &&
			condition.Reason == string(gatewayv1.RouteReasonIncompatibleFilters)
	}, 30*time.Second).Should(gomega.Equal(true))

	// Verify error message mentions MultipleExtensionRef of same kind defined
	httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(DEFAULT_NAMESPACE).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
	if err != nil || httpRoute == nil {
		t.Fatalf("Couldn't get the HTTPRoute, err: %+v", err)
	}
	condition := apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs))
	g.Expect(condition.Message).To(gomega.ContainSubstring("MultipleExtensionRef of same kind defined on HTTPRoute-Rule-BackendRef"))

	// Delete routeBackendExtension and verify it's removed from graph layer
	rbe1.DeleteRouteBackendExtensionCR(t)
	rbe2.DeleteRouteBackendExtensionCR(t)

	// Wait for the model to be updated after RouteBackendExtension deletion
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(0))

	// Clean up
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPorEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteStatusWithRouteBackendExtensionLifecycle(t *testing.T) {
	gatewayClassName := "gateway-class-rbe-lifecycle"
	gatewayName := "gateway-rbe-lifecycle"
	httpRouteName := "httproute-rbe-lifecycle"
	routeBackendExtensionName := "rbe-lifecycle"
	namespace := "default"
	svcName := "avisvc-rbe-lifecycle"
	ports := []int32{8080}

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	integrationtest.CreateSVC(t, namespace, svcName, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, namespace, svcName, false, false, "1.1.1")

	listeners := akogatewayapitests.GetListenersV1(ports, true, false)
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

	// Create HTTPRoute with reference to non-existent RouteBackendExtension
	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, ports)
	hostnames := []gatewayv1.Hostname{"foo-hm-lifecycle.com"}
	rules := []gatewayv1.HTTPRouteRule{
		akogatewayapitests.GetHTTPRouteRuleWithRouteBackendExtensionAndHMFilters(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
			map[string][]string{"RequestHeaderModifier": {"add"}},
			[][]string{{svcName, namespace, "8080", "1"}}, []string{routeBackendExtensionName}),
	}

	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, rules)

	// HTTPRoute should have unresolved refs condition due to non-existent RouteBackendExtensionName
	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			t.Logf("Couldn't get the HTTPRoute, err: %+v", err)
			return false
		}
		if len(httpRoute.Status.Parents) != len(ports) {
			return false
		}
		condition := apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs))
		return condition != nil && condition.Status == metav1.ConditionFalse &&
			condition.Reason == string(gatewayv1.RouteReasonBackendNotFound)
	}, 30*time.Second).Should(gomega.Equal(true))

	// Verify error message mentions RouteBackendExtension
	httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
	if err != nil || httpRoute == nil {
		t.Fatalf("Couldn't get the HTTPRoute, err: %+v", err)
	}
	condition := apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs))
	g.Expect(condition.Message).To(gomega.ContainSubstring("RouteBackendExtension object default/rbe-lifecycle not found"))

	// Create RouteBackendExtension CR with status as rejected
	rbe := akogatewayapitests.GetFakeDefaultRBEObj(routeBackendExtensionName, DEFAULT_NAMESPACE, "thisisaviref-hm1")
	rbe.Status = "Rejected"
	rbe.CreateRouteBackendExtensionCRWithStatus(t)

	// HTTPRoute should still have unresolved refs condition due to rejected RouteBackendExtension
	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			return false
		}
		if len(httpRoute.Status.Parents) != len(ports) {
			return false
		}
		condition := apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs))
		return condition != nil && condition.Status == metav1.ConditionFalse &&
			condition.Reason == string(gatewayv1.RouteReasonBackendNotFound) && strings.Contains(condition.Message, "RouteBackendExtension object default/rbe-lifecycle is not in Accepted state")
	}, 30*time.Second).Should(gomega.Equal(true))

	// Update RouteBackendExtension status to Accepted
	rbe.Status = "Accepted"
	rbe.UpdateRouteBackendExtensionStatus(t)

	// HTTPRoute should now have resolved refs condition
	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			return false
		}
		if len(httpRoute.Status.Parents) != len(ports) {
			return false
		}
		condition := apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs))
		return condition != nil && condition.Status == metav1.ConditionTrue
	}, 30*time.Second).Should(gomega.Equal(true))

	// Test status transition from status accepted to rejected
	rbe.Status = "Rejected"
	rbe.UpdateRouteBackendExtensionStatus(t)

	// HTTPRoute should now have unresolved refs condition
	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			return false
		}
		if len(httpRoute.Status.Parents) != len(ports) {
			return false
		}
		condition := apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs))
		return condition != nil && condition.Status == metav1.ConditionFalse &&
			condition.Reason == string(gatewayv1.RouteReasonBackendNotFound)
	}, 30*time.Second).Should(gomega.Equal(true))

	// Transition back to status accepted
	rbe.Status = "Accepted"
	rbe.UpdateRouteBackendExtensionStatus(t)

	// HTTPRoute should have resolved refs condition again
	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			return false
		}
		if len(httpRoute.Status.Parents) != len(ports) {
			return false
		}
		condition := apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs))
		return condition != nil && condition.Status == metav1.ConditionTrue
	}, 30*time.Second).Should(gomega.Equal(true))

	//Test RouteBackendExtension deletion
	rbe.DeleteRouteBackendExtensionCR(t)

	// HTTPRoute should now have unresolved refs condition due to deleted RouteBackendExtension
	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			return false
		}
		if len(httpRoute.Status.Parents) != len(ports) {
			return false
		}
		condition := apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs))
		return condition != nil && condition.Status == metav1.ConditionFalse &&
			condition.Reason == string(gatewayv1.RouteReasonBackendNotFound)
	}, 30*time.Second).Should(gomega.Equal(true))

	// Verify the condition message mentions the RouteBackendExtension
	httpRoute, err = akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
	if err != nil || httpRoute == nil {
		t.Fatalf("Couldn't get the HTTPRoute, err: %+v", err)
	}
	condition = apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs))
	g.Expect(condition.Message).To(gomega.ContainSubstring("RouteBackendExtension object default/rbe-lifecycle not found"))

	// Recreate RouteBackendExtension to verify HTTPRoute status recovers
	rbe.CreateRouteBackendExtensionCRWithStatus(t)

	// HTTPRoute should have resolved refs condition again
	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			return false
		}
		if len(httpRoute.Status.Parents) != len(ports) {
			return false
		}
		condition := apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs))
		return condition != nil && condition.Status == metav1.ConditionTrue
	}, 30*time.Second).Should(gomega.Equal(true))

	// Transition status controller from valid AKOCRDController to invalid controller
	rbe.Controller = "Invalid-Controller"
	rbe.UpdateRouteBackendExtensionStatus(t)

	// HTTPRoute should now have unresolved refs condition
	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			return false
		}
		if len(httpRoute.Status.Parents) != len(ports) {
			return false
		}
		condition := apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs))
		return condition != nil && condition.Status == metav1.ConditionFalse &&
			condition.Reason == string(gatewayv1.RouteReasonBackendNotFound)
	}, 30*time.Second).Should(gomega.Equal(true))

	// Verify error message mentions RouteBackendExtension
	httpRoute, err = akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
	if err != nil || httpRoute == nil {
		t.Fatalf("Couldn't get the HTTPRoute, err: %+v", err)
	}
	condition = apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs))
	g.Expect(condition.Message).To(gomega.ContainSubstring("RouteBackendExtension CR default/rbe-lifecycle is not handled by AKO CRD Operator"))

	// Transition status controller back to valid AKOCRDController
	rbe.Controller = "AKOCRDController"
	rbe.UpdateRouteBackendExtensionStatus(t)

	// HTTPRoute should have resolved refs condition again
	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			return false
		}
		if len(httpRoute.Status.Parents) != len(ports) {
			return false
		}
		condition := apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs))
		return condition != nil && condition.Status == metav1.ConditionTrue
	}, 30*time.Second).Should(gomega.Equal(true))

	// Cleanup
	rbe.DeleteRouteBackendExtensionCR(t)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	akogatewayapitests.TeardownGateway(t, gatewayName, namespace)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteStatusWithCrossNamespaceRouteBackendExtensionDifferentTenant(t *testing.T) {
	gatewayClassName := "gateway-class-rbe-status-02"
	gatewayName := "gateway-rbe-status-02"
	httpRouteName := "httproute-rbe-status-02"
	routeBackendExtensionName := "rbe-status-02-cross-ns"
	namespace := "default"
	routeBackendExtensionNamespace := "rbe-namespace-02"
	svcName := "avisvc-rbe-status-02"
	ports := []int32{8080}

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	integrationtest.CreateSVC(t, namespace, svcName, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, namespace, svcName, false, false, "1.1.1")

	// Create the RouteBackendExtension namespace and annotate it with a different tenant
	integrationtest.AddNamespace(t, routeBackendExtensionNamespace, map[string]string{})
	integrationtest.AnnotateNamespaceWithTenant(t, routeBackendExtensionNamespace, "tenant-01")

	listeners := akogatewayapitests.GetListenersV1(ports, true, false)
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

	// Create RouteBackendExtension CR in different namespace with different tenant
	rbe := akogatewayapitests.GetFakeDefaultRBEObj(routeBackendExtensionName, routeBackendExtensionNamespace, "thisisaviref-rbe-status-02")
	rbe.CreateRouteBackendExtensionCRWithStatus(t)

	// Create HTTPRoute with reference to RouteBackendExtension in different namespace
	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, ports)
	hostnames := []gatewayv1.Hostname{"foo-rbe-status-02.com"}

	// Create rule with cross-namespace RouteBackendExtension reference
	rule := akogatewayapitests.GetHTTPRouteRuleWithRouteBackendExtensionAndHMFilters(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName, routeBackendExtensionNamespace, "8080", "1"}}, []string{routeBackendExtensionName})
	rules := []gatewayv1.HTTPRouteRule{rule}

	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, rules)

	// HTTPRoute should have unresolved refs condition due to cross-namespace access with different tenant
	g.Eventually(func() bool {
		httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
		if err != nil || httpRoute == nil {
			t.Logf("Couldn't get the HTTPRoute, err: %+v", err)
			return false
		}
		if len(httpRoute.Status.Parents) != len(ports) {
			return false
		}
		// Should have unresolved refs condition due to tenant isolation
		condition := apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs))
		return condition != nil && condition.Status == metav1.ConditionFalse &&
			condition.Reason == string(gatewayv1.RouteReasonRefNotPermitted)
	}, 30*time.Second).Should(gomega.Equal(true))

	// Verify the condition message mentions tenant isolation
	httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
	if err != nil || httpRoute == nil {
		t.Fatalf("Couldn't get the HTTPRoute, err: %+v", err)
	}
	condition := apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs))
	g.Expect(condition.Message).To(gomega.ContainSubstring("tenant tenant-01 is not equal to HTTPRoute tenant admin"))

	// Cleanup
	rbe.DeleteRouteBackendExtensionCR(t)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	akogatewayapitests.TeardownGateway(t, gatewayName, namespace)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}
