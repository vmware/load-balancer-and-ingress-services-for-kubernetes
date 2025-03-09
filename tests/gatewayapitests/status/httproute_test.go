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
	"testing"
	"time"

	"github.com/onsi/gomega"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	akogatewayapitests "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/gatewayapitests"
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
	ports := []int32{8080, 8081}

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)

	listeners := akogatewayapitests.GetListenersV1(ports, false)
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
		return apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil &&
			apimeta.FindStatusCondition(httpRoute.Status.Parents[1].Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
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

func TestHTTPRouteWithAtleastOneParentReferenceValid(t *testing.T) {
	gatewayClassName := "gateway-class-hr-02"
	gatewayName := "gateway-hr-02"
	httpRouteName := "httproute-02"
	namespace := "default"
	ports := []int32{8080, 8081}

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)

	// creates a gateway with listeners 8080 and 8082
	listeners := akogatewayapitests.GetListenersV1([]int32{8080, 8082}, false)
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
		return apimeta.IsStatusConditionTrue(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.GatewayConditionAccepted)) &&
			apimeta.IsStatusConditionFalse(httpRoute.Status.Parents[1].Conditions, string(gatewayv1.GatewayConditionAccepted))
	}, 30*time.Second).Should(gomega.Equal(true))

	conditionMap := make(map[string][]metav1.Condition)
	conditionMap[fmt.Sprintf("%s-%d", gatewayName, 8080)] = []metav1.Condition{
		{
			Type:    string(gatewayv1.GatewayConditionAccepted),
			Reason:  string(gatewayv1.GatewayReasonAccepted),
			Status:  metav1.ConditionTrue,
			Message: "Parent reference is valid",
		},
	}
	conditionMap[fmt.Sprintf("%s-%d", gatewayName, 8081)] = []metav1.Condition{
		{
			Type:    string(gatewayv1.GatewayConditionAccepted),
			Reason:  string(gatewayv1.GatewayReasonInvalid),
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
	listeners := akogatewayapitests.GetListenersV1(ports, false)
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
		return apimeta.IsStatusConditionFalse(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.GatewayConditionAccepted))
	}, 30*time.Second).Should(gomega.Equal(true))

	conditionMap := make(map[string][]metav1.Condition)
	conditionMap[fmt.Sprintf("%s-%d", gatewayName, 8081)] = []metav1.Condition{
		{
			Type:    string(gatewayv1.GatewayConditionAccepted),
			Reason:  string(gatewayv1.GatewayReasonInvalid),
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
		return apimeta.IsStatusConditionTrue(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.GatewayConditionAccepted))
	}, 30*time.Second).Should(gomega.Equal(true))

	conditionMap[fmt.Sprintf("%s-%d", gatewayName, 8080)] = []metav1.Condition{
		{
			Type:    string(gatewayv1.GatewayConditionAccepted),
			Reason:  string(gatewayv1.GatewayReasonAccepted),
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
	listeners := akogatewayapitests.GetListenersV1(ports, false)
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
		return apimeta.IsStatusConditionTrue(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.GatewayConditionAccepted))
	}, 30*time.Second).Should(gomega.Equal(true))

	conditionMap := make(map[string][]metav1.Condition)
	conditionMap[fmt.Sprintf("%s-%d", gatewayName, 8081)] = []metav1.Condition{
		{
			Type:    string(gatewayv1.GatewayConditionAccepted),
			Reason:  string(gatewayv1.GatewayReasonAccepted),
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
		return apimeta.IsStatusConditionFalse(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.GatewayConditionAccepted))
	}, 30*time.Second).Should(gomega.Equal(true))

	conditionMap[fmt.Sprintf("%s-%d", gatewayName, 8081)] = []metav1.Condition{
		{
			Type:    string(gatewayv1.GatewayConditionAccepted),
			Reason:  string(gatewayv1.GatewayReasonInvalid),
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
	listeners := akogatewayapitests.GetListenersV1(ports, false)
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

	time.Sleep(10 * time.Second)

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
	listeners := akogatewayapitests.GetListenersV1([]int32{8082, 8083}, false)
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
		return apimeta.IsStatusConditionFalse(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.GatewayConditionAccepted)) &&
			apimeta.IsStatusConditionFalse(httpRoute.Status.Parents[1].Conditions, string(gatewayv1.GatewayConditionAccepted))
	}, 30*time.Second).Should(gomega.Equal(true))

	conditionMap := map[string][]metav1.Condition{
		fmt.Sprintf("%s-%d", gatewayName, 8080): {
			{
				Type:    string(gatewayv1.GatewayConditionAccepted),
				Reason:  string(gatewayv1.GatewayReasonInvalid),
				Status:  metav1.ConditionFalse,
				Message: "Invalid listener name provided",
			},
		},
		fmt.Sprintf("%s-%d", gatewayName, 8081): {
			{
				Type:    string(gatewayv1.GatewayConditionAccepted),
				Reason:  string(gatewayv1.GatewayReasonInvalid),
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

	time.Sleep(10 * time.Second)

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
	listeners := akogatewayapitests.GetListenersV1([]int32{8082}, false)
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
		return apimeta.IsStatusConditionFalse(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.GatewayConditionAccepted)) &&
			apimeta.IsStatusConditionFalse(httpRoute.Status.Parents[1].Conditions, string(gatewayv1.GatewayConditionAccepted))
	}, 30*time.Second).Should(gomega.Equal(true))

	conditionMap := map[string][]metav1.Condition{
		fmt.Sprintf("%s-%d", gatewayName, 8080): {
			{
				Type:    string(gatewayv1.GatewayConditionAccepted),
				Reason:  string(gatewayv1.GatewayReasonInvalid),
				Status:  metav1.ConditionFalse,
				Message: "Invalid listener name provided",
			},
		},
		fmt.Sprintf("%s-%d", gatewayName, 8081): {
			{
				Type:    string(gatewayv1.GatewayConditionAccepted),
				Reason:  string(gatewayv1.GatewayReasonInvalid),
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

	listeners := akogatewayapitests.GetListenersV1(ports, false)
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
		return apimeta.IsStatusConditionFalse(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.GatewayConditionAccepted)) &&
			apimeta.IsStatusConditionFalse(httpRoute.Status.Parents[1].Conditions, string(gatewayv1.GatewayConditionAccepted))
	}, 30*time.Second).Should(gomega.Equal(true))

	conditionMap := map[string][]metav1.Condition{
		fmt.Sprintf("%s-%d", gatewayName, 8080): {
			{
				Type:    string(gatewayv1.GatewayConditionAccepted),
				Reason:  string(gatewayv1.GatewayReasonInvalid),
				Status:  metav1.ConditionFalse,
				Message: "Hostname in Gateway Listener doesn't match with any of the hostnames in HTTPRoute",
			},
		},
		fmt.Sprintf("%s-%d", gatewayName, 8081): {
			{
				Type:    string(gatewayv1.GatewayConditionAccepted),
				Reason:  string(gatewayv1.GatewayReasonInvalid),
				Status:  metav1.ConditionFalse,
				Message: "Hostname in Gateway Listener doesn't match with any of the hostnames in HTTPRoute",
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

	listeners := akogatewayapitests.GetListenersV1(ports, false)

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
		return apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	conditionMap := make(map[string][]metav1.Condition)

	for _, port := range ports {
		conditions := make([]metav1.Condition, 0, 1)
		condition := metav1.Condition{
			Type:    string(gatewayv1.GatewayConditionAccepted),
			Reason:  string(gatewayv1.GatewayReasonInvalid),
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

	listeners := akogatewayapitests.GetListenersV1(ports, false)
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
		return apimeta.FindStatusCondition(httpRoute.Status.Parents[0].Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	conditionMap := make(map[string][]metav1.Condition)

	for _, port := range ports {
		conditions := make([]metav1.Condition, 0, 1)
		condition := metav1.Condition{
			Type:    string(gatewayv1.GatewayConditionAccepted),
			Reason:  string(gatewayv1.GatewayReasonInvalid),
			Status:  metav1.ConditionFalse,
			Message: "Gateway gateway-hr-12 is in Invalid State",
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
func TestHTTPRouteFilterWithUnsupportedUrlRewritePathType(t *testing.T) {

	gatewayName := "gateway-hr-19"
	gatewayClassName := "gateway-class-hr-19"
	httpRouteName := "http-route-hr-19"
	svcName := "avisvc-hr-19"
	namespace := "default"
	ports := []int32{8080}

	modelName, _ := akogatewayapitests.GetModelName(namespace, gatewayName)
	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := akogatewayapitests.GetListenersV1(ports, false)
	akogatewayapitests.SetupGateway(t, gatewayName, namespace, gatewayClassName, nil, listeners)
	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 5*time.Second).Should(gomega.Equal(true))

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, ports)
	rule := akogatewayapitests.GetHTTPRouteRuleV1([]string{"/foo"}, []string{},
		map[string][]string{"URLRewrite": {}},
		[][]string{{svcName, "default", "8080", "1"}})
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
