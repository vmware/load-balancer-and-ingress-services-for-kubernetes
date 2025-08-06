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
	integrationtest.CreateEPS(t, "default", svcName, false, false, "1.1.1")

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
	integrationtest.CreateEPS(t, "default", svcName, false, false, "1.1.1")

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

func TestHTTPRouteMultpleL7Rules(t *testing.T) {
	gatewayName := "gateway-l7Rule-01"
	gatewayClassName := "gateway-class-l7rule-01"
	httpRouteName := "http-route-l7rule-01"
	svcName := "avisvc-l7rule-01"
	l7RuleName1 := "l7rule-01a"
	l7RuleName2 := "l7rule-01b"

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

	l7CRDObj1 := akogatewayapitests.GetFakeDefaultL7RuleObj(l7RuleName1, DEFAULT_NAMESPACE)
	l7CRDObj1.CreateFakeL7RuleWithStatus(t)
	l7CRDObj2 := akogatewayapitests.GetFakeDefaultL7RuleObj(l7RuleName2, DEFAULT_NAMESPACE)
	l7CRDObj2.CreateFakeL7RuleWithStatus(t)

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, DEFAULT_NAMESPACE, ports)

	extensionRefCRDs := make(map[string][]string)
	extensionRefCRDs["L7Rule"] = []string{l7RuleName1, l7RuleName2}
	rule := akogatewayapitests.GetHTTPRouteRuleWithCustomCRDs(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"RequestHeaderModifier": {"add"}},
		[][]string{{svcName, DEFAULT_NAMESPACE, "8080", "1"}}, extensionRefCRDs)

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

	// HTTPRoute should have unresolved refs condition due to non-existent L7Rule
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
			condition.Reason == string(gatewayv1.RouteReasonInvalidKind)
	}, 30*time.Second).Should(gomega.Equal(true))

	// Verify error message mentions MultipleExtensionRef of same kind defined
	httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(DEFAULT_NAMESPACE).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
	if err != nil || httpRoute == nil {
		t.Fatalf("Couldn't get the HTTPRoute, err: %+v", err)
	}
	condition := apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs))
	g.Expect(condition.Message).To(gomega.ContainSubstring("MultipleExtensionRef of same kind defined on HTTPRoute-Rule"))

	// Delete L7Rule and verify it's removed from graph layer
	l7CRDObj1.DeleteL7RuleCR(t)
	l7CRDObj2.DeleteL7RuleCR(t)

	g.Eventually(func() int {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return -1
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 25*time.Second).Should(gomega.Equal(1))

	// Clean up
	integrationtest.DelSVC(t, DEFAULT_NAMESPACE, svcName)
	integrationtest.DelEPS(t, DEFAULT_NAMESPACE, svcName)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestHTTPRouteStatusWithL7RuleLifecycle(t *testing.T) {
	gatewayClassName := "gateway-class-l7rule-lifecycle"
	gatewayName := "gateway-l7rule-lifecycle"
	httpRouteName := "httproute-l7rule-lifecycle"
	namespace := "default"
	svcName := "avisvc-l7rule-lifecycle"
	ports := []int32{8080}
	l7RuleName1 := "l7rule-02a"

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	integrationtest.CreateSVC(t, namespace, svcName, "TCP", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, namespace, svcName, false, false, "1.1.1")

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

	extensionRefCRDs := make(map[string][]string)
	extensionRefCRDs["L7Rule"] = []string{l7RuleName1}
	// Create HTTPRoute with reference to non-existent L7Rule CRD
	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, ports)
	hostnames := []gatewayv1.Hostname{"foo.avi.internal"}
	rules := []gatewayv1.HTTPRouteRule{
		akogatewayapitests.GetHTTPRouteRuleWithCustomCRDs(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
			map[string][]string{"RequestHeaderModifier": {"add"}},
			[][]string{{svcName, namespace, "8080", "1"}}, extensionRefCRDs),
	}

	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, rules)

	// HTTPRoute should have unresolved refs condition due to non-existent L7Rule
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

	// Verify error message mentions L7Rule
	httpRoute, err := akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
	if err != nil || httpRoute == nil {
		t.Fatalf("Couldn't get the HTTPRoute, err: %+v", err)
	}
	condition := apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs))
	g.Expect(condition.Message).To(gomega.ContainSubstring("L7Rule CRD default/l7rule-02a not found"))

	// Create L7Rule CRD with status as rejected
	l7RuleObj := akogatewayapitests.GetFakeDefaultL7RuleObj(l7RuleName1, DEFAULT_NAMESPACE)
	l7RuleObj.Status = "Rejected"
	l7RuleObj.CreateFakeL7RuleWithStatus(t)

	// HTTPRoute should still have unresolved refs condition due to rejected L7Rule
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
			condition.Reason == string(gatewayv1.RouteReasonBackendNotFound) &&
			strings.Contains(condition.Message, "L7Rule CRD default/l7rule-02a is not accepted")
	}, 30*time.Second).Should(gomega.Equal(true))

	// Update L7Rule status to Accepted
	l7RuleObj.Status = "Accepted"
	l7RuleObj.UpdateL7RuleStatus(t)

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
	l7RuleObj.Status = "Rejected"
	l7RuleObj.UpdateL7RuleStatus(t)

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
	l7RuleObj.Status = "Accepted"
	l7RuleObj.UpdateL7RuleStatus(t)

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

	//Test L7Rule deletion
	l7RuleObj.DeleteL7RuleCR(t)

	// HTTPRoute should now have unresolved refs condition due to deleted L7Rule
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

	// Verify the condition message mentions the L7Rule
	httpRoute, err = akogatewayapitests.GatewayClient.GatewayV1().HTTPRoutes(namespace).Get(context.TODO(), httpRouteName, metav1.GetOptions{})
	if err != nil || httpRoute == nil {
		t.Fatalf("Couldn't get the HTTPRoute, err: %+v", err)
	}
	condition = apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.RouteConditionResolvedRefs))
	g.Expect(condition.Message).To(gomega.ContainSubstring("L7Rule CRD default/l7rule-02a not found"))

	// Recreate L7Rule to verify HTTPRoute status recovers
	l7RuleObj.CreateFakeL7RuleWithStatus(t)

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
	l7RuleObj.DeleteL7RuleCR(t)
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	akogatewayapitests.TeardownGateway(t, gatewayName, namespace)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}
