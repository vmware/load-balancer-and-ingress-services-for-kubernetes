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

package ingestion

import (
	"testing"

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
	akogatewayapiobjects.GatewayApiLister().UpdateGatewayClassToGateway(gatewayClassName, true)

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
	akogatewayapiobjects.GatewayApiLister().UpdateGatewayClassToGateway(gatewayClassName, true)

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
	akogatewayapiobjects.GatewayApiLister().UpdateGatewayClassToGateway(gatewayClassName, true)

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
	akogatewayapiobjects.GatewayApiLister().UpdateGatewayClassToGateway(gatewayClassName, true)
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
	akogatewayapiobjects.GatewayApiLister().UpdateGatewayClassToGateway(gatewayClassName, true)
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
	akogatewayapiobjects.GatewayApiLister().UpdateGatewayClassToGateway(gatewayClassName, true)
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
	akogatewayapiobjects.GatewayApiLister().UpdateGatewayClassToGateway(gatewayClassName, true)
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
	akogatewayapiobjects.GatewayApiLister().UpdateGatewayClassToGateway(gatewayClassName, true)
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
