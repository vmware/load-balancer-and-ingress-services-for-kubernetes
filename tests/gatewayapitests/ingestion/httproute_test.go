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

package ingestion

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	akogatewayapiobjects "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/objects"
	akogatewayapitests "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/gatewayapitests"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"
)

func TestHTTPRouteCUD(t *testing.T) {
	gatewayClassName := "gateway-class-01"
	gatewayName := "gateway-01"
	httpRouteName := "httproute-01"
	namespace := "default"
	ports := []int32{8080, 8081}
	key := "HTTPRoute" + "/" + namespace + "/" + httpRouteName
	gwkey := "Gateway/" + DEFAULT_NAMESPACE + "/" + gatewayName
	akogatewayapiobjects.GatewayApiLister().UpdateGatewayClass(gatewayClassName, true)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	t.Logf("Created GatewayClass %s", gatewayClassName)
	waitAndverify(t, "GatewayClass/gateway-class-01")

	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, namespace, gatewayClassName, nil, listeners)
	t.Logf("Created Gateway %s", gatewayName)
	waitAndverify(t, "Gateway/default/gateway-01")

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, ports)
	hostnames := []gatewayv1.Hostname{"foo-8080.com", "foo-8081.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, nil)
	waitAndverify(t, key)

	// update
	hostnames = []gatewayv1.Hostname{"foo-8080.com"}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, nil)
	waitAndverify(t, key)

	// delete
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	waitAndverify(t, key)
	akogatewayapitests.TeardownGateway(t, gatewayName, namespace)
	waitAndverify(t, gwkey)
}

func TestHTTPRouteHostnameInvalid(t *testing.T) {
	gatewayClassName := "gateway-class-02"
	gatewayName := "gateway-02"
	httpRouteName := "httproute-02"
	gwKey := "Gateway/" + DEFAULT_NAMESPACE + "/" + gatewayName
	gwClassKey := "GatewayClass/" + gatewayClassName
	namespace := "default"
	ports := []int32{8080}
	key := "HTTPRoute" + "/" + namespace + "/" + httpRouteName
	akogatewayapiobjects.GatewayApiLister().UpdateGatewayClass(gatewayClassName, true)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	t.Logf("Created GatewayClass %s", gatewayClassName)
	waitAndverify(t, gwClassKey)

	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, namespace, gatewayClassName, nil, listeners)
	t.Logf("Created Gateway %s", gatewayName)
	waitAndverify(t, gwKey)

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, ports)
	hostnames := []gatewayv1.Hostname{"*.example.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, nil)
	waitAndverify(t, key)

	// update
	hostnames = []gatewayv1.Hostname{"foo-8080.com"}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, nil)
	waitAndverify(t, key)

	// delete
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	waitAndverify(t, key)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	waitAndverify(t, gwKey)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
	waitAndverify(t, gwClassKey)
}

func TestHTTPRouteGatewayNotPresent(t *testing.T) {
	gatewayClassName := "gateway-class-03"
	gatewayName := "gateway-03"
	httpRouteName := "httproute-03"
	gwKey := "Gateway/" + DEFAULT_NAMESPACE + "/" + gatewayName
	gwClassKey := "GatewayClass/" + gatewayClassName
	namespace := "default"
	ports := []int32{8080, 8081}
	key := "HTTPRoute" + "/" + namespace + "/" + httpRouteName
	akogatewayapiobjects.GatewayApiLister().UpdateGatewayClass(gatewayClassName, true)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	t.Logf("Created GatewayClass %s", gatewayClassName)
	waitAndverify(t, gwClassKey)

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, ports)
	hostnames := []gatewayv1.Hostname{"foo-8080.com", "foo-8081.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, nil)
	waitAndverify(t, key)

	// update
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, namespace, gatewayClassName, nil, listeners)
	t.Logf("Created Gateway %s", gatewayName)
	waitAndverify(t, gwKey)
	hostnames = []gatewayv1.Hostname{"foo-8080.com"}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, nil)
	waitAndverify(t, key)

	// delete
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	waitAndverify(t, key)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	waitAndverify(t, gwKey)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
	waitAndverify(t, gwClassKey)
}

func TestHTTPRouteGatewayWithEmptyHostnameInGateway(t *testing.T) {
	gatewayClassName := "gateway-class-07"
	gatewayName := "gateway-07"
	httpRouteName := "httproute-07"
	gwKey := "Gateway/" + DEFAULT_NAMESPACE + "/" + gatewayName
	gwClassKey := "GatewayClass/" + gatewayClassName
	namespace := "default"
	ports := []int32{8080}
	key := "HTTPRoute" + "/" + namespace + "/" + httpRouteName

	// gatewayclass
	akogatewayapiobjects.GatewayApiLister().UpdateGatewayClass(gatewayClassName, true)
	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	t.Logf("Created GatewayClass %s", gatewayClassName)
	waitAndverify(t, gwClassKey)

	// Gateway with empty hostname
	listeners := akogatewayapitests.GetListenersV1(ports, true, false)
	akogatewayapitests.SetupGateway(t, gatewayName, namespace, gatewayClassName, nil, listeners)
	t.Logf("Created Gateway %s without hostname", gatewayName)
	waitAndverify(t, gwKey)

	t.Logf("Now creating httproute")
	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, ports)
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, nil)
	waitAndverify(t, key)

	// delete
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	waitAndverify(t, key)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	waitAndverify(t, gwKey)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
	waitAndverify(t, gwClassKey)
}

func TestHTTPRouteGatewayWithEmptyHostnameInHTTPRoute(t *testing.T) {
	gatewayClassName := "gateway-class-05"
	gatewayName := "gateway-05"
	httpRouteName := "httproute-05"
	gwKey := "Gateway/" + DEFAULT_NAMESPACE + "/" + gatewayName
	gwClassKey := "GatewayClass/" + gatewayClassName
	namespace := "default"
	ports := []int32{8080}
	key := "HTTPRoute" + "/" + namespace + "/" + httpRouteName

	// gatewayclass
	akogatewayapiobjects.GatewayApiLister().UpdateGatewayClass(gatewayClassName, true)
	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	t.Logf("Created GatewayClass %s", gatewayClassName)
	waitAndverify(t, gwClassKey)

	// Gateway
	listeners := akogatewayapitests.GetListenersV1(ports, false, false)
	akogatewayapitests.SetupGateway(t, gatewayName, namespace, gatewayClassName, nil, listeners)
	t.Logf("Created Gateway %s", gatewayName)
	waitAndverify(t, gwKey)

	// httproute without hostname
	t.Logf("Now creating httproute without hostname")
	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, ports)
	hostnames := []gatewayv1.Hostname{}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, nil)
	waitAndverify(t, key)

	// delete
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	waitAndverify(t, key)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	waitAndverify(t, gwKey)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
	waitAndverify(t, gwClassKey)
}

func TestHTTPRouteGatewayWithEmptyHostname(t *testing.T) {
	gatewayClassName := "gateway-class-06"
	gatewayName := "gateway-06"
	httpRouteName := "httproute-06"
	gwKey := "Gateway/" + DEFAULT_NAMESPACE + "/" + gatewayName
	gwClassKey := "GatewayClass/" + gatewayClassName
	namespace := "default"
	ports := []int32{8080}
	key := "HTTPRoute" + "/" + namespace + "/" + httpRouteName

	// gatewayclass
	akogatewayapiobjects.GatewayApiLister().UpdateGatewayClass(gatewayClassName, true)
	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	t.Logf("Created GatewayClass %s", gatewayClassName)
	waitAndverify(t, gwClassKey)

	// Gateway without hostname
	listeners := akogatewayapitests.GetListenersV1(ports, true, false)
	akogatewayapitests.SetupGateway(t, gatewayName, namespace, gatewayClassName, nil, listeners)
	t.Logf("Created Gateway %s", gatewayName)
	waitAndverify(t, gwKey)

	// httproute without hostname
	t.Logf("Now creating httproute without hostname")
	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, ports)
	hostnames := []gatewayv1.Hostname{}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, nil)
	waitAndverify(t, key)

	// delete
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	waitAndverify(t, key)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	waitAndverify(t, gwKey)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
	waitAndverify(t, gwClassKey)
}

func TestHTTPRouteGatewayWithRegexPath(t *testing.T) {
	gatewayClassName := "gateway-class-07"
	gatewayName := "gateway-07"
	httpRouteName := "httproute-07"
	gwKey := "Gateway/" + DEFAULT_NAMESPACE + "/" + gatewayName
	gwClassKey := "GatewayClass/" + gatewayClassName
	namespace := "default"
	ports := []int32{8080}
	key := "HTTPRoute" + "/" + namespace + "/" + httpRouteName

	// gatewayclass
	akogatewayapiobjects.GatewayApiLister().UpdateGatewayClass(gatewayClassName, true)
	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	t.Logf("Created GatewayClass %s", gatewayClassName)
	waitAndverify(t, gwClassKey)

	// Gateway without hostname
	listeners := akogatewayapitests.GetListenersV1(ports, true, false)
	akogatewayapitests.SetupGateway(t, gatewayName, namespace, gatewayClassName, nil, listeners)
	t.Logf("Created Gateway %s", gatewayName)
	waitAndverify(t, gwKey)

	t.Logf("Now creating httproute")
	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, ports)
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}
	rule := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.REGULAREXPRESSION, []string{"/foo/[a-z]+/bar"}, []string{}, nil,
		nil, nil)
	rules := []gatewayv1.HTTPRouteRule{rule}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, rules)
	waitAndverify(t, key)

	// delete
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	waitAndverify(t, key)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	waitAndverify(t, gwKey)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
	waitAndverify(t, gwClassKey)
}

func TestHTTPRouteFilterWithUrlRewrite(t *testing.T) {
	gatewayClassName := "gateway-class-08"
	gatewayName := "gateway-08"
	httpRouteName := "httproute-08"
	gwKey := "Gateway/" + DEFAULT_NAMESPACE + "/" + gatewayName
	gwClassKey := "GatewayClass/" + gatewayClassName
	namespace := "default"
	ports := []int32{8080}
	key := "HTTPRoute" + "/" + namespace + "/" + httpRouteName

	// gatewayclass
	akogatewayapiobjects.GatewayApiLister().UpdateGatewayClass(gatewayClassName, true)
	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	t.Logf("Created GatewayClass %s", gatewayClassName)
	waitAndverify(t, gwClassKey)

	// Gateway without hostname
	listeners := akogatewayapitests.GetListenersV1(ports, true, false)
	akogatewayapitests.SetupGateway(t, gatewayName, namespace, gatewayClassName, nil, listeners)
	t.Logf("Created Gateway %s", gatewayName)
	waitAndverify(t, gwKey)

	t.Logf("Now creating httproute")
	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, ports)
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}

	rule := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"URLRewrite": {}},
		[][]string{{"avisvc", "default", "8080", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{rule}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, rules)
	waitAndverify(t, key)

	//httproute with invalid rewrite filter configuration
	rules[0].Filters[0].URLRewrite.Path.Type = gatewayv1.PrefixMatchHTTPPathModifier
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, rules)
	waitAndverify(t, key)

	// delete
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	waitAndverify(t, key)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	waitAndverify(t, gwKey)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
	waitAndverify(t, gwClassKey)
}

// TestHTTPRouteWithAppProfileExtensionRef validates ingestion of http route with
// extension ref filter
func TestHTTPRouteWithAppProfileExtensionRef(t *testing.T) {
	gatewayClassName := "gateway-class-09"
	gatewayName := "gateway-09"
	httpRouteName := "httproute-09"
	gwKey := "Gateway/" + DEFAULT_NAMESPACE + "/" + gatewayName

	gwClassKey := "GatewayClass/" + gatewayClassName
	namespace := "default"
	ports := []int32{8080}
	key := "HTTPRoute" + "/" + namespace + "/" + httpRouteName

	// setup gatewayclass
	akogatewayapiobjects.GatewayApiLister().UpdateGatewayClass(gatewayClassName, true)
	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	t.Logf("Created GatewayClass %s", gatewayClassName)
	waitAndverify(t, gwClassKey)

	// setup gateway
	listeners := akogatewayapitests.GetListenersV1(ports, true, false)
	akogatewayapitests.SetupGateway(t, gatewayName, namespace, gatewayClassName, nil, listeners)
	t.Logf("Created Gateway %s", gatewayName)
	waitAndverify(t, gwKey)

	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, ports)
	hostnames := []gatewayv1.Hostname{"foo-8080.com"}

	// setup http route with app profile extension ref
	rule := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/foo"}, []string{},
		map[string][]string{"ExtensionRef": {"app-profile-ref1"}},
		[][]string{{"avisvc", "default", "8080", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{rule}
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, rules)
	t.Logf("Created HTTPRoute %s", httpRouteName)
	waitAndverify(t, key)

	// update httproute with another  app profile name
	rules[0].Filters[0].ExtensionRef.Name = "app-profile-ref2"
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, rules)
	waitAndverify(t, key)

	// cleanup
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	waitAndverify(t, key)
	akogatewayapitests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	waitAndverify(t, gwKey)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
	waitAndverify(t, gwClassKey)
}

// Helper function to setup a dedicated gateway for HTTPRoute tests
func setupDedicatedGatewayForHTTPRoute(t *testing.T, name, namespace, gatewayClass string, listeners []gatewayv1.Listener) {
	gateway := gatewayv1.Gateway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "gateway.networking.k8s.io/v1",
			Kind:       "Gateway",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: map[string]string{
				"ako.vmware.com/dedicated-gateway-mode": "true",
			},
		},
		Spec: gatewayv1.GatewaySpec{
			GatewayClassName: gatewayv1.ObjectName(gatewayClass),
			Listeners:        listeners,
		},
		Status: gatewayv1.GatewayStatus{},
	}

	gw, err := akogatewayapitests.GatewayClient.GatewayV1().Gateways(namespace).Create(context.TODO(), &gateway, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Couldn't create dedicated gateway for HTTPRoute test, err: %+v", err)
	}
	t.Logf("Created dedicated gateway %+v for HTTPRoute test", gw.Name)
}

func TestHTTPRouteWithDedicatedGatewayCUD(t *testing.T) {
	gatewayClassName := "dedicated-gateway-class-httproute-01"
	gatewayName := "dedicated-gateway-httproute-01"
	httpRouteName := "httproute-dedicated-01"
	namespace := "default"
	ports := []int32{8080}

	gwKey := "Gateway/" + namespace + "/" + gatewayName
	gwClassKey := "GatewayClass/" + gatewayClassName
	hrKey := "HTTPRoute/" + namespace + "/" + httpRouteName

	// Setup Gateway Class
	akogatewayapiobjects.GatewayApiLister().UpdateGatewayClass(gatewayClassName, true)
	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	t.Logf("Created GatewayClass %s", gatewayClassName)
	waitAndverify(t, gwClassKey)

	// Setup Dedicated Gateway without hostnames (dedicated mode doesn't support hostnames)
	listeners := []gatewayv1.Listener{
		{
			Name:     "listener-http",
			Port:     gatewayv1.PortNumber(ports[0]),
			Protocol: gatewayv1.HTTPProtocolType,
		},
	}
	setupDedicatedGatewayForHTTPRoute(t, gatewayName, namespace, gatewayClassName, listeners)
	waitAndverify(t, gwKey)

	// Create HTTPRoute (no hostnames since dedicated gateway doesn't support them)
	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, ports)
	hostnames := []gatewayv1.Hostname{} // Empty hostnames for dedicated mode
	rule := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/api"}, []string{},
		map[string][]string{}, [][]string{{"avisvc", "default", "8080", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{rule}

	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, rules)
	t.Logf("Created HTTPRoute %s with dedicated gateway", httpRouteName)
	waitAndverify(t, hrKey)

	// Update HTTPRoute - add another rule
	rule2 := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/admin"}, []string{},
		map[string][]string{}, [][]string{{"avisvc", "default", "8080", "1"}}, nil)
	rules = append(rules, rule2)
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, rules)
	t.Logf("Updated HTTPRoute %s with additional rule", httpRouteName)
	waitAndverify(t, hrKey)

	// Update HTTPRoute - change path match type
	rules[0].Matches[0].Path.Type = func() *gatewayv1.PathMatchType {
		exactType := gatewayv1.PathMatchExact
		return &exactType
	}()
	rules[0].Matches[0].Path.Value = func() *string {
		path := "/api/v1"
		return &path
	}()
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, rules)
	t.Logf("Updated HTTPRoute %s with exact path match", httpRouteName)
	waitAndverify(t, hrKey)

	// Delete HTTPRoute
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	t.Logf("Deleted HTTPRoute %s", httpRouteName)
	waitAndverify(t, hrKey)

	// Cleanup
	akogatewayapitests.TeardownGateway(t, gatewayName, namespace)
	waitAndverify(t, gwKey)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
	waitAndverify(t, gwClassKey)
}

func TestHTTPRouteWithDedicatedGatewayAndFilters(t *testing.T) {
	gatewayClassName := "dedicated-gateway-class-httproute-02"
	gatewayName := "dedicated-gateway-httproute-02"
	httpRouteName := "httproute-dedicated-filters-01"
	namespace := "default"
	ports := []int32{8080}

	gwKey := "Gateway/" + namespace + "/" + gatewayName
	gwClassKey := "GatewayClass/" + gatewayClassName
	hrKey := "HTTPRoute/" + namespace + "/" + httpRouteName

	// Setup Gateway Class
	akogatewayapiobjects.GatewayApiLister().UpdateGatewayClass(gatewayClassName, true)
	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	t.Logf("Created GatewayClass %s", gatewayClassName)
	waitAndverify(t, gwClassKey)

	// Setup Dedicated Gateway
	listeners := []gatewayv1.Listener{
		{
			Name:     "listener-http",
			Port:     gatewayv1.PortNumber(ports[0]),
			Protocol: gatewayv1.HTTPProtocolType,
		},
	}
	setupDedicatedGatewayForHTTPRoute(t, gatewayName, namespace, gatewayClassName, listeners)
	waitAndverify(t, gwKey)

	// Create HTTPRoute with URL rewrite filter
	parentRefs := akogatewayapitests.GetParentReferencesV1([]string{gatewayName}, namespace, ports)
	hostnames := []gatewayv1.Hostname{} // Empty hostnames for dedicated mode
	rule := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/old-api"}, []string{},
		map[string][]string{"URLRewrite": {}}, [][]string{{"avisvc", "default", "8080", "1"}}, nil)
	rules := []gatewayv1.HTTPRouteRule{rule}

	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, rules)
	t.Logf("Created HTTPRoute %s with URL rewrite filter", httpRouteName)
	waitAndverify(t, hrKey)

	// Update HTTPRoute - change filter type to RequestHeaderModifier
	rule.Filters = []gatewayv1.HTTPRouteFilter{
		{
			Type: gatewayv1.HTTPRouteFilterRequestHeaderModifier,
			RequestHeaderModifier: &gatewayv1.HTTPHeaderFilter{
				Add: []gatewayv1.HTTPHeader{
					{Name: "X-Custom-Header", Value: "custom-value"},
				},
			},
		},
	}
	rules[0] = rule
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, rules)
	t.Logf("Updated HTTPRoute %s with header modifier filter", httpRouteName)
	waitAndverify(t, hrKey)

	// Update HTTPRoute - add extension ref filter
	extRule := akogatewayapitests.GetHTTPRouteRuleV1(integrationtest.PATHPREFIX, []string{"/ext"}, []string{},
		map[string][]string{"ExtensionRef": {"app-profile-dedicated"}}, [][]string{{"avisvc", "default", "8080", "1"}}, nil)
	rules = append(rules, extRule)
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, rules)
	t.Logf("Updated HTTPRoute %s with extension ref filter", httpRouteName)
	waitAndverify(t, hrKey)

	// Delete HTTPRoute
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	t.Logf("Deleted HTTPRoute %s", httpRouteName)
	waitAndverify(t, hrKey)

	// Cleanup
	akogatewayapitests.TeardownGateway(t, gatewayName, namespace)
	waitAndverify(t, gwKey)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
	waitAndverify(t, gwClassKey)
}
