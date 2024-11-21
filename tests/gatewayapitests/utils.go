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

package tests

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayfake "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned/fake"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"
)

var KubeClient *k8sfake.Clientset
var GatewayClient *gatewayfake.Clientset

func NewAviFakeClientInstance(kubeclient *k8sfake.Clientset, skipCachePopulation ...bool) {
	if integrationtest.AviFakeClientInstance == nil {
		integrationtest.AviFakeClientInstance = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			utils.AviLog.Infof("[fakeAPI]: %s %s", r.Method, r.URL)

			if integrationtest.FakeServerMiddleware != nil {
				integrationtest.FakeServerMiddleware(w, r)
				return
			}

			integrationtest.NormalControllerServer(w, r, "../../avimockobjectsgw")
		}))

		url := strings.Split(integrationtest.AviFakeClientInstance.URL, "https://")[1]
		os.Setenv("CTRL_IPADDRESS", url)
		os.Setenv("FULL_SYNC_INTERVAL", "600")
		// resets avi client pool instance, allows to connect with the new `ts` server
		//cache.AviClientInstanceMap = nil
		k8s.PopulateControllerProperties(kubeclient)
		if len(skipCachePopulation) == 0 || !skipCachePopulation[0] {
			k8s.PopulateCache()
		}
	}
}

func GetModelName(namespace, name string) (string, string) {
	vsName := akogatewayapilib.Prefix + "cluster--" + namespace + "-" + name + "-EVH"
	return "admin/" + vsName, vsName
}

func SetGatewayName(gw *gatewayv1.Gateway, name string) {
	gw.Name = name
}
func UnsetGatewayName(gw *gatewayv1.Gateway) {
	gw.Name = ""
}

func SetGatewayGatewayClass(gw *gatewayv1.Gateway, name string) {
	gw.Spec.GatewayClassName = gatewayv1.ObjectName(name)
}
func UnsetGatewayGatewayClass(gw *gatewayv1.Gateway) {
	gw.Spec.GatewayClassName = ""
}

func AddGatewayListener(gw *gatewayv1.Gateway, name string, port int32, protocol gatewayv1.ProtocolType, isTLS bool) {

	listner := gatewayv1.Listener{
		Name:     gatewayv1.SectionName(name),
		Port:     gatewayv1.PortNumber(port),
		Protocol: protocol,
	}
	if isTLS {
		SetListenerTLS(&listner, gatewayv1.TLSModeTerminate, "secret-example", "default")
	}
	gw.Spec.Listeners = append(gw.Spec.Listeners, listner)
}

func SetListenerTLS(l *gatewayv1.Listener, tlsMode gatewayv1.TLSModeType, secretName, secretNS string) {
	l.TLS = &gatewayv1.GatewayTLSConfig{Mode: &tlsMode}
	namespace := gatewayv1.Namespace(secretNS)
	kind := gatewayv1.Kind("Secret")
	l.TLS.CertificateRefs = []gatewayv1.SecretObjectReference{
		{
			Name:      gatewayv1.ObjectName(secretName),
			Namespace: &namespace,
			Kind:      &kind,
		},
	}
}
func UnsetListenerTLS(l *gatewayv1.Listener) {
	l.TLS = &gatewayv1.GatewayTLSConfig{}
}

func SetListenerHostname(l *gatewayv1.Listener, hostname string) {
	if hostname == "" {
		l.Hostname = nil
	} else {
		l.Hostname = (*gatewayv1.Hostname)(&hostname)
	}
}
func UnsetListenerHostname(l *gatewayv1.Listener) {
	var hname gatewayv1.Hostname
	l.Hostname = &hname
}

func GetListenersV1(ports []int32, emptyHostName, samehost bool, secrets ...string) []gatewayv1.Listener {
	listeners := make([]gatewayv1.Listener, 0, len(ports))
	for _, port := range ports {
		listener := gatewayv1.Listener{
			Name:     gatewayv1.SectionName(fmt.Sprintf("listener-%d", port)),
			Port:     gatewayv1.PortNumber(port),
			Protocol: gatewayv1.ProtocolType("HTTPS"),
		}
		if !samehost && !emptyHostName {
			hostname := fmt.Sprintf("foo-%d.com", port)
			listener.Hostname = (*gatewayv1.Hostname)(&hostname)
		} else if samehost {
			hostname := "foo.com"
			listener.Hostname = (*gatewayv1.Hostname)(&hostname)
		}

		if len(secrets) > 0 {
			certRefs := make([]gatewayv1.SecretObjectReference, 0, len(secrets))
			for _, secret := range secrets {
				secretRef := gatewayv1.SecretObjectReference{
					Name: gatewayv1.ObjectName(secret),
				}
				certRefs = append(certRefs, secretRef)
			}
			tlsMode := "Terminate"
			listener.TLS = &gatewayv1.GatewayTLSConfig{
				Mode:            (*gatewayv1.TLSModeType)(&tlsMode),
				CertificateRefs: certRefs,
			}
		}
		listeners = append(listeners, listener)
	}
	return listeners
}

func GetListenersOnHostname(hostnames []string) []gatewayv1.Listener {
	listeners := make([]gatewayv1.Listener, 0, len(hostnames))
	for i, hostname := range hostnames {
		hn := hostname
		listener := gatewayv1.Listener{
			Name:     gatewayv1.SectionName(fmt.Sprintf("listener-%d", i)),
			Port:     gatewayv1.PortNumber(8080),
			Hostname: (*gatewayv1.Hostname)(&hn),
			Protocol: gatewayv1.ProtocolType("HTTP"),
		}
		listeners = append(listeners, listener)
	}
	return listeners
}
func GetNegativeConditions(ports []int32) *gatewayv1.GatewayStatus {
	expectedStatus := &gatewayv1.GatewayStatus{
		Conditions: []metav1.Condition{
			{
				Type:               string(gatewayv1.GatewayConditionAccepted),
				Status:             metav1.ConditionFalse,
				Message:            "Gateway does not contain any valid listener",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonListenersNotValid),
			},
		},
		Listeners: GetListenerStatusV1(ports, []int32{0, 0}, true, false),
	}
	expectedStatus.Listeners[0].Conditions[0].Reason = string(gatewayv1.ListenerReasonInvalid)
	expectedStatus.Listeners[0].Conditions[0].Status = metav1.ConditionFalse
	expectedStatus.Listeners[0].Conditions[0].Message = "Listener is Invalid"

	expectedStatus.Listeners[0].Conditions[1].Reason = string(gatewayv1.ListenerReasonInvalidCertificateRef)
	expectedStatus.Listeners[0].Conditions[1].Status = metav1.ConditionFalse
	expectedStatus.Listeners[0].Conditions[1].Message = "Secret does not exist"

	return expectedStatus
}

func GetPositiveConditions(ports []int32) *gatewayv1.GatewayStatus {
	expectedStatus := &gatewayv1.GatewayStatus{
		Conditions: []metav1.Condition{
			{
				Type:               string(gatewayv1.GatewayConditionAccepted),
				Status:             metav1.ConditionTrue,
				Message:            "Gateway configuration is valid",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayConditionAccepted),
			},
		},
		Listeners: GetListenerStatusV1(ports, []int32{0, 0}, true, false),
	}
	expectedStatus.Listeners[0].Conditions[0].Reason = string(gatewayv1.ListenerReasonAccepted)
	expectedStatus.Listeners[0].Conditions[0].Status = metav1.ConditionTrue
	expectedStatus.Listeners[0].Conditions[0].Message = "Listener is valid"

	expectedStatus.Listeners[0].Conditions[1].Reason = string(gatewayv1.ListenerReasonResolvedRefs)
	expectedStatus.Listeners[0].Conditions[1].Status = metav1.ConditionTrue
	expectedStatus.Listeners[0].Conditions[1].Message = "Reference is valid"
	return expectedStatus
}
func GetListenerStatusV1(ports []int32, attachedRoutes []int32, getResolvedRefCondition bool, getProgrammedCondition bool) []gatewayv1.ListenerStatus {
	listeners := make([]gatewayv1.ListenerStatus, 0, len(ports))
	for i, port := range ports {
		listener := gatewayv1.ListenerStatus{
			Name:           gatewayv1.SectionName(fmt.Sprintf("listener-%d", port)),
			SupportedKinds: akogatewayapilib.SupportedKinds["HTTPS"],
			AttachedRoutes: attachedRoutes[i],
			Conditions: []metav1.Condition{
				{
					Type:               string(gatewayv1.ListenerConditionAccepted),
					Status:             metav1.ConditionTrue,
					Message:            "Listener is valid",
					ObservedGeneration: 1,
					Reason:             string(gatewayv1.ListenerReasonAccepted),
				},
			},
		}
		if getResolvedRefCondition {
			resolvedRefCondition := &metav1.Condition{
				Type:               string(gatewayv1.ListenerConditionResolvedRefs),
				Status:             metav1.ConditionTrue,
				Message:            "All the references are valid",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.ListenerReasonResolvedRefs),
			}
			listener.Conditions = append(listener.Conditions, *resolvedRefCondition)
		}
		if getProgrammedCondition {
			programmedCondition := &metav1.Condition{
				Type:               string(gatewayv1.ListenerConditionProgrammed),
				Status:             metav1.ConditionTrue,
				Message:            "Virtual service configured/updated",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.ListenerReasonProgrammed),
			}
			listener.Conditions = append(listener.Conditions, *programmedCondition)
		}
		listeners = append(listeners, listener)
	}
	return listeners
}

type Gateway struct {
	*gatewayv1.Gateway
}

func (g *Gateway) GatewayV1(name, namespace, gatewayClass string, address []gatewayv1.GatewayAddress, listeners []gatewayv1.Listener) *gatewayv1.Gateway {
	gateway := &gatewayv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			ResourceVersion: time.Now().Local().String(),
		},
		Spec: gatewayv1.GatewaySpec{
			GatewayClassName: gatewayv1.ObjectName(gatewayClass),
			Addresses:        address,
		},
	}

	gateway.Spec.Listeners = listeners
	return gateway
}

func (g *Gateway) Create(t *testing.T) {
	_, err := GatewayClient.GatewayV1().Gateways(g.Namespace).Create(context.TODO(), g.Gateway, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Couldn't create the gateway, err: %+v", err)
	}
	t.Logf("Created Gateway %s", g.Gateway.Name)
}

func (g *Gateway) Update(t *testing.T) {
	_, err := GatewayClient.GatewayV1().Gateways(g.Namespace).Update(context.TODO(), g.Gateway, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Couldn't update the gateway, err: %+v", err)
	}
	t.Logf("Updated Gateway %s", g.Gateway.Name)
}

func (g *Gateway) Delete(t *testing.T) {
	err := GatewayClient.GatewayV1().Gateways(g.Namespace).Delete(context.TODO(), g.Name, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't delete the gateway, err: %+v", err)
	}
	t.Logf("Deleted Gateway %s", g.Gateway.Name)
}

func SetupGateway(t *testing.T, name, namespace, gatewayClass string, ipAddress []gatewayv1.GatewayAddress, listeners []gatewayv1.Listener) {
	g := &Gateway{}
	g.Gateway = g.GatewayV1(name, namespace, gatewayClass, ipAddress, listeners)
	g.Create(t)
}

func UpdateGateway(t *testing.T, name, namespace, gatewayClass string, ipAddress []gatewayv1.GatewayAddress, listeners []gatewayv1.Listener) {
	g := &Gateway{}
	g.Gateway = g.GatewayV1(name, namespace, gatewayClass, ipAddress, listeners)
	g.Update(t)
}

func TeardownGateway(t *testing.T, name, namespace string) {
	g := &Gateway{}
	g.Gateway = g.GatewayV1(name, namespace, "", nil, nil)
	g.Delete(t)
}

type FakeGatewayClass struct {
	Name           string
	ControllerName string
}

func (gc *FakeGatewayClass) GatewayClassV1() *gatewayv1.GatewayClass {
	return &gatewayv1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: gc.Name,
		},
		Spec: gatewayv1.GatewayClassSpec{
			ControllerName: gatewayv1.GatewayController(gc.ControllerName),
		},
	}
}

func (gc *FakeGatewayClass) Create(t *testing.T) {
	gatewayClass := gc.GatewayClassV1()
	_, err := GatewayClient.GatewayV1().GatewayClasses().Create(context.TODO(), gatewayClass, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Couldn't create the gateway class, err: %+v", err)
	}
	t.Logf("Created GatewayClass %s", gatewayClass.Name)
}

func (gc *FakeGatewayClass) Update(t *testing.T) {
	gatewayClass := gc.GatewayClassV1()
	_, err := GatewayClient.GatewayV1().GatewayClasses().Update(context.TODO(), gatewayClass, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Couldn't update the gateway class, err: %+v", err)
	}
	t.Logf("Updated GatewayClass %s", gatewayClass.Name)
}

func (gc *FakeGatewayClass) Delete(t *testing.T) {
	err := GatewayClient.GatewayV1().GatewayClasses().Delete(context.TODO(), gc.Name, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't delete the gateway class, err: %+v", err)
	}
	t.Logf("Deleted GatewayClass %s", gc.Name)
}

func SetupGatewayClass(t *testing.T, name, controllerName string) {
	gc := &FakeGatewayClass{
		Name:           name,
		ControllerName: controllerName,
	}
	gc.Create(t)
	time.Sleep(10 * time.Second)
}

func TeardownGatewayClass(t *testing.T, name string) {
	gc := &FakeGatewayClass{
		Name: name,
	}
	gc.Delete(t)
}

type HTTPRoute struct {
	*gatewayv1.HTTPRoute
}

func (hr *HTTPRoute) HTTPRouteV1(name, namespace string, parentRefs []gatewayv1.ParentReference, hostnames []gatewayv1.Hostname, rules []gatewayv1.HTTPRouteRule) *gatewayv1.HTTPRoute {
	httpRoute := &gatewayv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			ResourceVersion: time.Now().Local().String(),
		},
		Spec: gatewayv1.HTTPRouteSpec{
			CommonRouteSpec: gatewayv1.CommonRouteSpec{
				ParentRefs: parentRefs,
			},
			Hostnames: hostnames,
			Rules:     rules,
		},
	}
	return httpRoute
}

func GetParentReferencesV1(gatewayNames []string, namespace string, ports []int32) []gatewayv1.ParentReference {
	parentRefs := make([]gatewayv1.ParentReference, 0)
	for _, gwName := range gatewayNames {
		for _, port := range ports {
			sectionName := gatewayv1.SectionName(fmt.Sprintf("listener-%d", port))
			parentRef := gatewayv1.ParentReference{
				Name:        gatewayv1.ObjectName(gwName),
				Namespace:   (*gatewayv1.Namespace)(&namespace),
				SectionName: &sectionName,
			}
			parentRefs = append(parentRefs, parentRef)
		}
	}
	return parentRefs
}

func GetParentReferencesFromListeners(listeners []gatewayv1.Listener, gwName, namespace string) []gatewayv1.ParentReference {
	parentRefs := make([]gatewayv1.ParentReference, 0)
	for i := range listeners {
		sectionName := gatewayv1.SectionName(fmt.Sprintf("listener-%d", i))
		parentRef := gatewayv1.ParentReference{
			Name:        gatewayv1.ObjectName(gwName),
			Namespace:   (*gatewayv1.Namespace)(&namespace),
			SectionName: &sectionName,
		}
		parentRefs = append(parentRefs, parentRef)

	}
	return parentRefs
}

// created new function to avoid confusion
func GetParentReferencesV1WithGatewayNameOnly(gatewayNames []string, namespace string) []gatewayv1.ParentReference {
	parentRefs := make([]gatewayv1.ParentReference, 0)
	for _, gwName := range gatewayNames {

		parentRef := gatewayv1.ParentReference{
			Name:      gatewayv1.ObjectName(gwName),
			Namespace: (*gatewayv1.Namespace)(&namespace),
		}
		parentRefs = append(parentRefs, parentRef)

	}
	return parentRefs
}

func GetRouteStatusV1(gatewayNames []string, namespace string, ports []int32, conditions map[string][]metav1.Condition) *gatewayv1.RouteStatus {
	routeStatus := &gatewayv1.RouteStatus{}
	routeStatus.Parents = make([]gatewayv1.RouteParentStatus, 0, len(gatewayNames)+len(ports))
	for _, gatewayName := range gatewayNames {
		for _, port := range ports {
			parent := gatewayv1.RouteParentStatus{}
			parent.ControllerName = akogatewayapilib.GatewayController
			parent.Conditions = conditions[fmt.Sprintf("%s-%d", gatewayName, port)]
			sectionName := gatewayv1.SectionName(fmt.Sprintf("listener-%d", port))
			parent.ParentRef = gatewayv1.ParentReference{
				Name:        gatewayv1.ObjectName(gatewayName),
				Namespace:   (*gatewayv1.Namespace)(&namespace),
				SectionName: &sectionName,
			}
			routeStatus.Parents = append(routeStatus.Parents, parent)
		}
	}
	return routeStatus
}

func GetHTTPRouteMatchV1(path string, pathMatchType string, headers []string) gatewayv1.HTTPRouteMatch {
	routeMatch := gatewayv1.HTTPRouteMatch{}
	routeMatch.Path = &gatewayv1.HTTPPathMatch{}
	routeMatch.Path.Type = (*gatewayv1.PathMatchType)(proto.String(pathMatchType))
	routeMatch.Path.Value = &path
	for _, header := range headers {
		headerMatch := gatewayv1.HTTPHeaderMatch{}
		headerMatch.Type = (*gatewayv1.HeaderMatchType)(proto.String("Exact"))
		headerMatch.Name = gatewayv1.HTTPHeaderName(header)
		headerMatch.Value = "some-value"
		routeMatch.Headers = append(routeMatch.Headers, headerMatch)
	}
	return routeMatch
}

func GetHTTPHeaderFilterV1(actions []string) *gatewayv1.HTTPHeaderFilter {
	headerFilter := &gatewayv1.HTTPHeaderFilter{}
	for _, action := range actions {
		switch action {
		case "add":
			headerFilter.Add =
				append(headerFilter.Add,
					gatewayv1.HTTPHeader{
						Name:  gatewayv1.HTTPHeaderName("new-header"),
						Value: "any-value",
					},
				)
		case "remove":
			headerFilter.Remove = append(headerFilter.Remove, "old-header")
		case "replace":
			headerFilter.Set =
				append(headerFilter.Set,
					gatewayv1.HTTPHeader{
						Name:  gatewayv1.HTTPHeaderName("my-header"),
						Value: "any-value",
					},
				)
		}
	}
	return headerFilter
}

func GetHTTPRouteFilterV1(filterType string, actions []string) gatewayv1.HTTPRouteFilter {
	routeFilter := gatewayv1.HTTPRouteFilter{}
	routeFilter.Type = gatewayv1.HTTPRouteFilterType(filterType)
	switch filterType {
	case "RequestHeaderModifier":
		routeFilter.RequestHeaderModifier = GetHTTPHeaderFilterV1(actions)
	case "ResponseHeaderModifier":
		routeFilter.ResponseHeaderModifier = GetHTTPHeaderFilterV1(actions)
	case "RequestRedirect":
		statusCode302 := 302
		host := "redirect.com"
		routeFilter.RequestRedirect = &gatewayv1.HTTPRequestRedirectFilter{
			Hostname:   (*gatewayv1.PreciseHostname)(&host),
			StatusCode: &statusCode302,
		}
	}
	return routeFilter
}

func GetHTTPRouteBackendV1(backendRefs []string) gatewayv1.HTTPBackendRef {
	serviceKind := gatewayv1.Kind("Service")
	port, _ := strconv.Atoi(backendRefs[2])
	servicePort := gatewayv1.PortNumber(port)
	backendRef := gatewayv1.HTTPBackendRef{}
	backendRef.BackendObjectReference = gatewayv1.BackendObjectReference{
		Kind:      &serviceKind,
		Name:      gatewayv1.ObjectName(backendRefs[0]),
		Namespace: (*gatewayv1.Namespace)(&backendRefs[1]),
		Port:      &servicePort,
	}
	weight, _ := strconv.Atoi(backendRefs[3])
	weight32 := int32(weight)
	backendRef.Weight = &weight32
	return backendRef

}

func GetHTTPRouteRuleV1(paths []string, matchHeaders []string, filterActionMap map[string][]string, backendRefs [][]string, backendRefFilters map[string][]string) gatewayv1.HTTPRouteRule {
	matches := make([]gatewayv1.HTTPRouteMatch, 0, len(paths))
	for _, path := range paths {
		match := GetHTTPRouteMatchV1(path, "PathPrefix", matchHeaders)
		matches = append(matches, match)
	}

	filters := make([]gatewayv1.HTTPRouteFilter, 0, len(filterActionMap))
	for filterType, actions := range filterActionMap {
		filter := GetHTTPRouteFilterV1(filterType, actions)
		filters = append(filters, filter)
	}
	backends := make([]gatewayv1.HTTPBackendRef, 0, len(backendRefs))
	for _, backendRef := range backendRefs {
		httpBackend := GetHTTPRouteBackendV1(backendRef)
		backendFilters := make([]gatewayv1.HTTPRouteFilter, 0, len(filterActionMap))
		for filterType, actions := range backendRefFilters {
			filter := GetHTTPRouteFilterV1(filterType, actions)
			backendFilters = append(backendFilters, filter)
		}
		httpBackend.Filters = backendFilters
		backends = append(backends, httpBackend)
	}
	rule := gatewayv1.HTTPRouteRule{}
	rule.Matches = matches
	rule.Filters = filters
	rule.BackendRefs = backends
	return rule
}
func GetHTTPRouteRulesV1Login() []gatewayv1.HTTPRouteRule {
	rules := make([]gatewayv1.HTTPRouteRule, 0)
	// TODO: add few rules

	//login rule
	var serviceKind gatewayv1.Kind
	var servicePort gatewayv1.PortNumber
	serviceKind = "Service"
	servicePort = 8080
	pathPrefix := gatewayv1.PathMatchPathPrefix
	path := "/login"
	rules = append(rules, gatewayv1.HTTPRouteRule{
		Matches: []gatewayv1.HTTPRouteMatch{
			{
				Path: &gatewayv1.HTTPPathMatch{
					Type:  &pathPrefix,
					Value: &path,
				},
			},
		},
		BackendRefs: []gatewayv1.HTTPBackendRef{
			{
				BackendRef: gatewayv1.BackendRef{
					BackendObjectReference: gatewayv1.BackendObjectReference{
						Kind: &serviceKind,
						Name: "avisvc",
						Port: &servicePort,
					},
				},
			},
		},
	})
	return rules
}

func (hr *HTTPRoute) Create(t *testing.T) {
	_, err := GatewayClient.GatewayV1().HTTPRoutes(hr.Namespace).Create(context.TODO(), hr.HTTPRoute, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Couldn't create the HTTPRoute, err: %+v", err)
	}
	t.Logf("Created HTTPRoute %s", hr.Name)
}

func (hr *HTTPRoute) Update(t *testing.T) {
	_, err := GatewayClient.GatewayV1().HTTPRoutes(hr.Namespace).Update(context.TODO(), hr.HTTPRoute, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Couldn't update the HTTPRoute, err: %+v", err)
	}
	t.Logf("Updated HTTPRoute %s", hr.Name)
}

func (hr *HTTPRoute) Delete(t *testing.T) {
	err := GatewayClient.GatewayV1().HTTPRoutes(hr.Namespace).Delete(context.TODO(), hr.Name, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't delete the HTTPRoute, err: %+v", err)
	}
	t.Logf("Deleted HTTPRoute %s", hr.Name)
}

func SetupHTTPRoute(t *testing.T, name, namespace string, parentRefs []gatewayv1.ParentReference, hostnames []gatewayv1.Hostname, rules []gatewayv1.HTTPRouteRule) {
	hr := &HTTPRoute{}
	hr.HTTPRoute = hr.HTTPRouteV1(name, namespace, parentRefs, hostnames, rules)
	hr.Create(t)
}

func UpdateHTTPRoute(t *testing.T, name, namespace string, parentRefs []gatewayv1.ParentReference, hostnames []gatewayv1.Hostname, rules []gatewayv1.HTTPRouteRule) {
	hr := &HTTPRoute{}
	hr.HTTPRoute = hr.HTTPRouteV1(name, namespace, parentRefs, hostnames, rules)
	hr.Update(t)
}

func TeardownHTTPRoute(t *testing.T, name, namespace string) {
	hr := &HTTPRoute{}
	hr.HTTPRoute = hr.HTTPRouteV1(name, namespace, nil, nil, nil)
	hr.Delete(t)
}

func ValidateGatewayStatus(t *testing.T, actualStatus, expectedStatus *gatewayv1.GatewayStatus) {

	g := gomega.NewGomegaWithT(t)

	// validate the ip address
	if len(expectedStatus.Addresses) > 0 {
		g.Expect(actualStatus.Addresses).To(gomega.HaveLen(1))
		g.Expect(actualStatus.Addresses[0]).Should(gomega.Equal(expectedStatus.Addresses[0]))
	}

	g.Expect(actualStatus.Listeners).To(gomega.HaveLen(len(expectedStatus.Listeners)))
	ValidateConditions(t, actualStatus.Conditions, expectedStatus.Conditions)

	for _, actualListenerStatus := range actualStatus.Listeners {
		for _, expectedListenerStatus := range expectedStatus.Listeners {
			if actualListenerStatus.Name == expectedListenerStatus.Name {
				ValidateGatewayListeners(t, &actualListenerStatus, &expectedListenerStatus)
			}
		}
	}
}

func ValidateGatewayListeners(t *testing.T, actual, expected *gatewayv1.ListenerStatus) {
	g := gomega.NewGomegaWithT(t)
	g.Expect(actual.Name).Should(gomega.Equal(expected.Name))
	g.Expect(actual.AttachedRoutes).Should(gomega.Equal(expected.AttachedRoutes))
	g.Expect(actual.SupportedKinds).Should(gomega.Equal(expected.SupportedKinds))
	ValidateConditions(t, actual.Conditions, expected.Conditions)
}

func ValidateConditions(t *testing.T, actualConditions, expectedConditions []metav1.Condition) {
	g := gomega.NewGomegaWithT(t)

	for _, actualCondition := range actualConditions {
		for _, expectedCondition := range expectedConditions {
			if actualCondition.Type == expectedCondition.Type {
				g.Expect(actualCondition.Message).Should(gomega.Equal(expectedCondition.Message))
				g.Expect(actualCondition.Reason).Should(gomega.Equal(expectedCondition.Reason))
				g.Expect(actualCondition.Status).Should(gomega.Equal(expectedCondition.Status))
			}
		}
	}
}

func ValidateHTTPRouteStatus(t *testing.T, actualStatus, expectedStatus *gatewayv1.HTTPRouteStatus) {
	g := gomega.NewGomegaWithT(t)
	g.Expect(actualStatus.Parents).To(gomega.HaveLen(len(expectedStatus.Parents)))
	for i := 0; i < len(actualStatus.Parents); i++ {
		actualRouteParentStatus := actualStatus.Parents[i]
		expectedRouteParentStatus := expectedStatus.Parents[i]
		g.Expect(actualRouteParentStatus.ControllerName).To(gomega.Equal(expectedRouteParentStatus.ControllerName))
		g.Expect(actualRouteParentStatus.ParentRef).To(gomega.Equal(expectedRouteParentStatus.ParentRef))
		ValidateConditions(t, actualRouteParentStatus.Conditions, expectedRouteParentStatus.Conditions)
	}
}
