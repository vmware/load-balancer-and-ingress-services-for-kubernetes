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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"

	k8sfake "k8s.io/client-go/kubernetes/fake"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayfake "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned/fake"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"
)

var KubeClient *k8sfake.Clientset
var GatewayClient *gatewayfake.Clientset
var DynamicClient *dynamicfake.FakeDynamicClient
var GvrToKind = map[schema.GroupVersionResource]string{
	akogatewayapilib.L7CRDGVR:                    "l7rulesList",
	akogatewayapilib.HealthMonitorGVR:            "healthmonitorsList",
	akogatewayapilib.RouteBackendExtensionCRDGVR: "routebackendextensionsList",
	akogatewayapilib.AppProfileCRDGVR:            "applicationProfileList",
}
var testData unstructured.Unstructured

func GetL7RuleFakeData() unstructured.Unstructured {
	testData.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "ako.vmware.com",
		"kind":       "l7rules",
		"metadata": map[string]interface{}{
			"name":      "testL7Rule",
			"namespace": "default",
		},
		"spec": map[string]interface{}{},
	})
	return testData
}
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
	l.TLS = &gatewayv1.ListenerTLSConfig{Mode: &tlsMode}
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
	l.TLS = &gatewayv1.ListenerTLSConfig{}
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
			listener.TLS = &gatewayv1.ListenerTLSConfig{
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
func GetListenerStatusV1(ports []int32, attachedRoutes []int32, getResolvedRefCondition bool, getProgrammedCondition bool, programmedConditionMessage ...string) []gatewayv1.ListenerStatus {
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
			message := "Virtual service configured/updated"
			if len(programmedConditionMessage) > 0 {
				message = programmedConditionMessage[0]
			}

			programmedCondition := &metav1.Condition{
				Type:               string(gatewayv1.ListenerConditionProgrammed),
				Status:             metav1.ConditionTrue,
				Message:            message,
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

func (g *Gateway) GatewayV1(name, namespace, gatewayClass string, address []gatewayv1.GatewaySpecAddress, listeners []gatewayv1.Listener, vipType ...string) *gatewayv1.Gateway {
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
	if len(vipType) > 0 {
		gateway.Annotations = map[string]string{
			akogatewayapilib.LBVipTypeAnnotation: vipType[0],
		}
	}
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

func SetupGateway(t *testing.T, name, namespace, gatewayClass string, ipAddress []gatewayv1.GatewaySpecAddress, listeners []gatewayv1.Listener) {
	g := &Gateway{}
	g.Gateway = g.GatewayV1(name, namespace, gatewayClass, ipAddress, listeners)
	g.Create(t)
}

func UpdateGateway(t *testing.T, name, namespace, gatewayClass string, ipAddress []gatewayv1.GatewaySpecAddress, listeners []gatewayv1.Listener, vipType ...string) {
	g := &Gateway{}
	g.Gateway = g.GatewayV1(name, namespace, gatewayClass, ipAddress, listeners, vipType...)
	g.Update(t)
}
func SetupGatewayWithAnnotation(t *testing.T, name, namespace, gatewayClass string, ipAddress []gatewayv1.GatewaySpecAddress, listeners []gatewayv1.Listener) {
	g := &Gateway{}
	g.Gateway = g.GatewayV1(name, namespace, gatewayClass, ipAddress, listeners)
	g.Gateway.Annotations = map[string]string{lib.VSTrafficDisabled: "true"}
	g.Create(t)
}

func UpdateGatewayWithAnnotation(t *testing.T, name, namespace, gatewayClass, annotation string, ipAddress []gatewayv1.GatewaySpecAddress, listeners []gatewayv1.Listener, vipType ...string) {
	g := &Gateway{}
	g.Gateway = g.GatewayV1(name, namespace, gatewayClass, ipAddress, listeners, vipType...)
	if annotation == "nil" {
		g.Gateway.Annotations = nil
	} else {
		g.Gateway.Annotations = map[string]string{lib.VSTrafficDisabled: annotation}
	}
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
	switch routeFilter.Type {
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
	case "URLRewrite":
		replaceFullPath := "/bar"
		host := "rewrite.com"
		routeFilter.URLRewrite = &gatewayv1.HTTPURLRewriteFilter{
			Hostname: (*gatewayv1.PreciseHostname)(&host),
			Path: &gatewayv1.HTTPPathModifier{
				Type:            gatewayv1.FullPathHTTPPathModifier,
				ReplaceFullPath: &replaceFullPath,
			},
		}
	case gatewayv1.HTTPRouteFilterExtensionRef:
		if len(actions) > 0 {
			routeFilter.ExtensionRef = &gatewayv1.LocalObjectReference{
				Group: "ako.vmware.com",
				Kind:  "ApplicationProfile",
				Name:  gatewayv1.ObjectName(actions[0]),
			}
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

func GetHTTPRouteRuleV1(pathMatchType string, paths []string, matchHeaders []string, filterActionMap map[string][]string, backendRefs [][]string, backendRefFilters map[string][]string) gatewayv1.HTTPRouteRule {
	matches := make([]gatewayv1.HTTPRouteMatch, 0, len(paths))
	for _, path := range paths {
		match := GetHTTPRouteMatchV1(path, pathMatchType, matchHeaders)
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

func ValidateGatewayStatusWithRetry(t *testing.T, actualStatus, expectedStatus *gatewayv1.GatewayStatus) bool {
	// validate the ip address
	if len(expectedStatus.Addresses) > 0 {
		if len(actualStatus.Addresses) != 1 {
			return false
		}
		if actualStatus.Addresses[0] != expectedStatus.Addresses[0] {
			return false
		}
	}

	// validate listeners length
	if len(actualStatus.Listeners) != len(expectedStatus.Listeners) {
		return false
	}

	// validate gateway conditions
	if !ValidateConditionsWithRetry(t, actualStatus.Conditions, expectedStatus.Conditions) {
		return false
	}

	// validate each listener
	for _, actualListenerStatus := range actualStatus.Listeners {
		matched := false
		for _, expectedListenerStatus := range expectedStatus.Listeners {
			if actualListenerStatus.Name == expectedListenerStatus.Name {
				if !ValidateGatewayListenersWithRetry(t, &actualListenerStatus, &expectedListenerStatus) {
					return false
				}
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

func ValidateGatewayListenersWithRetry(t *testing.T, actual, expected *gatewayv1.ListenerStatus) bool {
	if actual.Name != expected.Name {
		return false
	}
	if actual.AttachedRoutes != expected.AttachedRoutes {
		return false
	}
	// Compare SupportedKinds manually
	if len(actual.SupportedKinds) != len(expected.SupportedKinds) {
		return false
	}
	for i, actualKind := range actual.SupportedKinds {
		if actualKind.Group != expected.SupportedKinds[i].Group {
			return false
		}
		if actualKind.Kind != expected.SupportedKinds[i].Kind {
			return false
		}
	}
	return ValidateConditionsWithRetry(t, actual.Conditions, expected.Conditions)
}

func ValidateConditionsWithRetry(t *testing.T, actualConditions, expectedConditions []metav1.Condition) bool {
	for _, actualCondition := range actualConditions {
		for _, expectedCondition := range expectedConditions {
			if actualCondition.Type == expectedCondition.Type {
				if actualCondition.Message != expectedCondition.Message {
					return false
				}
				if actualCondition.Reason != expectedCondition.Reason {
					return false
				}
				if actualCondition.Status != expectedCondition.Status {
					return false
				}
			}
		}
	}
	return true
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

func ValidateHTTPRouteStatusWithRetry(t *testing.T, actualStatus, expectedStatus *gatewayv1.HTTPRouteStatus) bool {
	if len(actualStatus.Parents) != len(expectedStatus.Parents) {
		return false
	}
	for i := range actualStatus.Parents {
		if actualStatus.Parents[i].ControllerName != expectedStatus.Parents[i].ControllerName {
			return false
		}
		if actualStatus.Parents[i].ParentRef.Name != expectedStatus.Parents[i].ParentRef.Name ||
			*actualStatus.Parents[i].ParentRef.Namespace != *expectedStatus.Parents[i].ParentRef.Namespace ||
			*actualStatus.Parents[i].ParentRef.SectionName != *expectedStatus.Parents[i].ParentRef.SectionName {
			return false
		}
		for _, actualCondition := range actualStatus.Parents[i].Conditions {
			for _, expectedCondition := range expectedStatus.Parents[i].Conditions {
				if actualCondition.Type != expectedCondition.Type {
					continue
				}
				if actualCondition.Status != expectedCondition.Status {
					return false
				}
				if actualCondition.Reason != expectedCondition.Reason {
					return false
				}
				if actualCondition.Message != expectedCondition.Message {
					return false
				}
			}
		}
	}
	return true
}

func CreateHealthMonitorCRD(t *testing.T, name, namespace, uuid string) {
	CreateHealthMonitorCRDWithStatus(t, name, namespace, uuid, true, "Accepted", "HealthMonitor has been successfully processed")
}

func CreateHealthMonitorCRDWithStatus(t *testing.T, name, namespace, uuid string, ready bool, reason, message string) {
	status := "True"
	if !ready {
		status = "False"
	}

	healthMonitor := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "ako.vmware.com/v1alpha1",
			"kind":       "HealthMonitor",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"type": "HTTP",
				"httpMonitor": map[string]interface{}{
					"requestHeader": "GET /health HTTP/1.1",
					"responseCode":  []interface{}{"HTTP_2XX", "HTTP_3XX"},
				},
			},
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":    "Programmed",
						"status":  status,
						"reason":  reason,
						"message": message,
					},
				},
				"tenant": "admin",
			},
		},
	}

	// Only add uuid to status if it's provided (for ready HealthMonitors)
	if ready && uuid != "" {
		statusObj := healthMonitor.Object["status"].(map[string]interface{})
		statusObj["uuid"] = uuid
	}

	_, err := DynamicClient.Resource(akogatewayapilib.HealthMonitorGVR).Namespace(namespace).Create(context.TODO(), healthMonitor, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in creating HealthMonitor: %v", err)
	}
	t.Logf("Created HealthMonitor %s/%s with Ready=%v", namespace, name, ready)
}

func DeleteHealthMonitorCRD(t *testing.T, name, namespace string) {
	err := DynamicClient.Resource(akogatewayapilib.HealthMonitorGVR).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error in deleting HealthMonitor: %v", err)
	}
	t.Logf("Deleted HealthMonitor %s/%s", namespace, name)
}

// TO DO : Replace with a generic function GetHTTPRouteRuleWithExtensionRef which can generate a rule with any backendRef filter
func GetHTTPRouteRuleWithHealthMonitorFilters(pathType string, paths []string, headers []string, filters map[string][]string, backends [][]string, healthMonitors []string) gatewayv1.HTTPRouteRule {
	rule := GetHTTPRouteRuleV1(pathType, paths, headers, filters, backends, nil)

	for i := range rule.BackendRefs {
		for _, healthMonitor := range healthMonitors {
			filter := gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterExtensionRef,
				ExtensionRef: &gatewayv1.LocalObjectReference{
					Group: gatewayv1.Group("ako.vmware.com"),
					Kind:  gatewayv1.Kind("HealthMonitor"),
					Name:  gatewayv1.ObjectName(healthMonitor),
				},
			}
			rule.BackendRefs[i].Filters = append(rule.BackendRefs[i].Filters, filter)
		}
	}
	return rule
}

func UpdateHealthMonitorStatus(t *testing.T, name, namespace string, ready bool, reason, message string) {
	healthMonitor, err := DynamicClient.Resource(akogatewayapilib.HealthMonitorGVR).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("error in getting HealthMonitor: %v", err)
	}

	// Update the status condition
	status := "True"
	if !ready {
		status = "False"
	}

	conditions := []interface{}{
		map[string]interface{}{
			"type":    "Programmed",
			"status":  status,
			"reason":  reason,
			"message": message,
		},
	}

	// Update the status section
	statusObj := healthMonitor.Object["status"].(map[string]interface{})
	statusObj["conditions"] = conditions

	_, err = DynamicClient.Resource(akogatewayapilib.HealthMonitorGVR).Namespace(namespace).UpdateStatus(context.TODO(), healthMonitor, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HealthMonitor status: %v", err)
	}
	t.Logf("Updated HealthMonitor %s/%s status to Ready=%v", namespace, name, ready)
}

// TO DO : Replace with a generic function GetHTTPRouteRuleWithExtensionRef which can generate a rule with any backendRef filter
func GetHTTPRouteRuleWithRouteBackendExtensionAndHMFilters(pathType string, paths []string, headers []string, filters map[string][]string, backends [][]string, routeBackendExtensions []string, healthMonitors ...string) gatewayv1.HTTPRouteRule {
	rule := GetHTTPRouteRuleV1(pathType, paths, headers, filters, backends, nil)

	for i := range rule.BackendRefs {
		for _, routeBackendExtension := range routeBackendExtensions {
			filter := gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterExtensionRef,
				ExtensionRef: &gatewayv1.LocalObjectReference{
					Group: gatewayv1.Group("ako.vmware.com"),
					Kind:  gatewayv1.Kind("RouteBackendExtension"),
					Name:  gatewayv1.ObjectName(routeBackendExtension),
				},
			}
			rule.BackendRefs[i].Filters = append(rule.BackendRefs[i].Filters, filter)
		}
		for _, healthMonitor := range healthMonitors {
			filter := gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterExtensionRef,
				ExtensionRef: &gatewayv1.LocalObjectReference{
					Group: gatewayv1.Group("ako.vmware.com"),
					Kind:  gatewayv1.Kind("HealthMonitor"),
					Name:  gatewayv1.ObjectName(healthMonitor),
				},
			}
			rule.BackendRefs[i].Filters = append(rule.BackendRefs[i].Filters, filter)
		}
	}
	return rule
}

type FakeRouteBackendExtensionHM struct {
	Kind string
	Name string
}

type FakeRouteBackendExtension struct {
	Name                         string
	Namespace                    string
	LBAlgorithm                  string
	LBAlgorithmHash              string
	LBAlgorithmConsistentHashHdr string
	PersistenceProfile           string
	Hm                           []FakeRouteBackendExtensionHM
	HostCheckEnabled             *bool
	DomainName                   *string
	PKIProfile                   *FakeRouteBackendExtensionPKIProfile
	Status                       string
	Controller                   string
}

type FakeRouteBackendExtensionPKIProfile struct {
	Kind string
	Name string
}

// GetFakeDefaultRBEObj returns a fake RBE object that will be frequently used for testing.
// The returned object can be modified in the caller to make specific modifications as required by a test
func GetFakeDefaultRBEObj(name, namespace string, healthMonitorNames ...string) *FakeRouteBackendExtension {
	var hms []FakeRouteBackendExtensionHM
	for _, hmName := range healthMonitorNames {
		hms = append(hms, FakeRouteBackendExtensionHM{Kind: "AVIREF", Name: hmName})
	}
	rbe := FakeRouteBackendExtension{
		Name:        name,
		Namespace:   namespace,
		Hm:          hms,
		LBAlgorithm: "LB_ALGORITHM_ROUND_ROBIN",
		Status:      "Accepted",
		Controller:  akogatewayapilib.AKOCRDController,
	}
	return &rbe
}

// GetFakeRBEObjWithBackendTLS returns a fake RBE object with BackendTLS configuration for testing
func GetFakeRBEObjWithBackendTLS(name, namespace string, hostCheckEnabled *bool, domainName *string, pkiProfileName *string, healthMonitorNames ...string) *FakeRouteBackendExtension {
	rbe := GetFakeDefaultRBEObj(name, namespace, healthMonitorNames...)

	// Set BackendTLS fields
	rbe.HostCheckEnabled = hostCheckEnabled
	rbe.DomainName = domainName

	if pkiProfileName != nil {
		rbe.PKIProfile = &FakeRouteBackendExtensionPKIProfile{
			Kind: "CRD",
			Name: *pkiProfileName,
		}
	}

	return rbe
}

// Helper functions to create pointers for optional fields
func BoolPtr(b bool) *bool {
	return &b
}

func StringPtr(s string) *string {
	return &s
}

// CreatePKIProfileCR creates a PKIProfile CRD for testing
func CreatePKIProfileCR(t *testing.T, name, namespace string, caCerts []string) {
	caCertObjects := make([]interface{}, len(caCerts))
	for i, cert := range caCerts {
		caCertObjects[i] = map[string]interface{}{
			"certificate": cert,
		}
	}

	pkiProfile := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "ako.vmware.com/v1alpha1",
			"kind":       "PKIProfile",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"ca_certs": caCertObjects,
			},
			"status": map[string]interface{}{
				"controller": akogatewayapilib.AKOCRDController,
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Programmed",
						"status": "True",
						"reason": "ValidationSucceeded",
					},
				},
			},
		},
	}

	_, err := DynamicClient.Resource(akogatewayapilib.PKIProfileCRDGVR).Namespace(namespace).Create(context.TODO(), pkiProfile, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in creating PKIProfile: %v", err)
	}
	t.Logf("Created PKIProfile %s/%s", namespace, name)
}

// DeletePKIProfileCR deletes a PKIProfile CRD
func DeletePKIProfileCR(t *testing.T, name, namespace string) {
	err := DynamicClient.Resource(akogatewayapilib.PKIProfileCRDGVR).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error in deleting PKIProfile: %v", err)
	}
	t.Logf("Deleted PKIProfile %s/%s", namespace, name)
}

func (rbe *FakeRouteBackendExtension) CreateRouteBackendExtensionCRWithStatus(t *testing.T) {
	hms := make([]interface{}, len(rbe.Hm))
	for _, hm := range rbe.Hm {
		hms = append(hms, map[string]interface{}{
			"kind": hm.Kind,
			"name": hm.Name,
		})

	}
	spec := map[string]interface{}{
		"lbAlgorithm":                  rbe.LBAlgorithm,
		"lbAlgorithmHash":              rbe.LBAlgorithmHash,
		"lbAlgorithmConsistentHashHdr": rbe.LBAlgorithmConsistentHashHdr,
		"persistenceProfile":           rbe.PersistenceProfile,
		"healthMonitor":                hms,
	}

	// Add BackendTLS related fields - always create backendTLS section for BackendTLS-enabled RBEs
	backendTLS := map[string]interface{}{}

	if rbe.HostCheckEnabled != nil {
		backendTLS["hostCheckEnabled"] = *rbe.HostCheckEnabled
	}
	if rbe.DomainName != nil {
		backendTLS["domainName"] = []interface{}{*rbe.DomainName}
	}
	if rbe.PKIProfile != nil {
		backendTLS["pkiProfile"] = map[string]interface{}{
			"kind": rbe.PKIProfile.Kind,
			"name": rbe.PKIProfile.Name,
		}
	}

	// Always add backendTLS section since this function is specifically for BackendTLS RBEs
	spec["backendTLS"] = backendTLS

	routeBackendExtension := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "ako.vmware.com/v1alpha1",
			"kind":       "RouteBackendExtension",
			"metadata": map[string]interface{}{
				"name":      rbe.Name,
				"namespace": rbe.Namespace,
			},
			"spec": spec,
			"status": map[string]interface{}{
				"controller": rbe.Controller,
				"status":     rbe.Status,
			},
		},
	}

	_, err := DynamicClient.Resource(akogatewayapilib.RouteBackendExtensionCRDGVR).Namespace(rbe.Namespace).Create(context.TODO(), routeBackendExtension, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in creating RouteBackendExtension: %v", err)
	}
	t.Logf("Created RouteBackendExtension %s/%s with status=%s", rbe.Namespace, rbe.Name, rbe.Status)
}

// @AI-Generated
// [Generated by Cursor claude-4-sonnet]
func (rbe *FakeRouteBackendExtension) UpdateRouteBackendExtensionCR(t *testing.T) {
	hms := make([]interface{}, len(rbe.Hm))
	for _, hm := range rbe.Hm {
		hms = append(hms, map[string]interface{}{
			"kind": hm.Kind,
			"name": hm.Name,
		})
	}

	// Get the existing RouteBackendExtension
	existingRBE, err := DynamicClient.Resource(akogatewayapilib.RouteBackendExtensionCRDGVR).Namespace(rbe.Namespace).Get(context.TODO(), rbe.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("error in getting RouteBackendExtension: %v", err)
	}

	// Update the spec with new values
	specObj := existingRBE.Object["spec"].(map[string]interface{})
	specObj["lbAlgorithm"] = rbe.LBAlgorithm
	specObj["lbAlgorithmHash"] = rbe.LBAlgorithmHash
	specObj["lbAlgorithmConsistentHashHdr"] = rbe.LBAlgorithmConsistentHashHdr
	if rbe.PersistenceProfile != "" {
		specObj["persistenceProfile"] = rbe.PersistenceProfile
	} else {
		delete(specObj, "persistenceProfile")
	}
	specObj["healthMonitor"] = hms

	// Update BackendTLS related fields - always create backendTLS section for BackendTLS-enabled RBEs
	backendTLS := map[string]interface{}{}

	if rbe.HostCheckEnabled != nil {
		backendTLS["hostCheckEnabled"] = *rbe.HostCheckEnabled
	}
	if rbe.DomainName != nil {
		backendTLS["domainName"] = []interface{}{*rbe.DomainName}
	}
	if rbe.PKIProfile != nil {
		backendTLS["pkiProfile"] = map[string]interface{}{
			"kind": rbe.PKIProfile.Kind,
			"name": rbe.PKIProfile.Name,
		}
	}

	// Always add backendTLS section since this function is specifically for BackendTLS RBEs
	specObj["backendTLS"] = backendTLS

	// Update the status section
	statusObj := existingRBE.Object["status"].(map[string]interface{})
	statusObj["status"] = rbe.Status
	statusObj["controller"] = rbe.Controller

	_, err = DynamicClient.Resource(akogatewayapilib.RouteBackendExtensionCRDGVR).Namespace(rbe.Namespace).Update(context.TODO(), existingRBE, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating RouteBackendExtension: %v", err)
	}
	t.Logf("Updated RouteBackendExtension %s/%s with persistenceProfile=%s", rbe.Namespace, rbe.Name, rbe.PersistenceProfile)
}

func (rbe *FakeRouteBackendExtension) DeleteRouteBackendExtensionCR(t *testing.T) {
	err := DynamicClient.Resource(akogatewayapilib.RouteBackendExtensionCRDGVR).Namespace(rbe.Namespace).Delete(context.TODO(), rbe.Name, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error in deleting RouteBackendExtension: %v", err)
	}
	t.Logf("Deleted RouteBackendExtension %s/%s", rbe.Namespace, rbe.Name)
}

func (rbe *FakeRouteBackendExtension) UpdateRouteBackendExtensionStatus(t *testing.T) {
	routeBackendExtension, err := DynamicClient.Resource(akogatewayapilib.RouteBackendExtensionCRDGVR).Namespace(rbe.Namespace).Get(context.TODO(), rbe.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("error in getting RouteBackendExtension: %v", err)
	}

	// Update the status section
	statusObj := routeBackendExtension.Object["status"].(map[string]interface{})
	statusObj["status"] = rbe.Status
	statusObj["controller"] = rbe.Controller

	_, err = DynamicClient.Resource(akogatewayapilib.RouteBackendExtensionCRDGVR).Namespace(rbe.Namespace).UpdateStatus(context.TODO(), routeBackendExtension, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating RouteBackendExtension status: %v", err)
	}
	t.Logf("Updated RouteBackendExtension %s/%s status to Status=%s, Controller=%s", rbe.Namespace, rbe.Name, rbe.Status, rbe.Controller)
}

type FullClientLogsL7 struct {
	Enabled  bool
	Throttle string
	Duration uint32
}

type L7RuleAnalyticsPolicy struct {
	FullClientLogs FullClientLogsL7
	LogAllHeaders  bool
}
type KindNameNamespace struct {
	Kind string
	Name string
}
type L7RuleHTTPPolicy struct {
	PolicySets []string
	Overwrite  bool
}
type FakeL7Rule struct {
	AllowInvalidClientCert        bool
	BotPolicyRef                  string
	CloseClientConnOnConfigUpdate bool
	HostNameXlate                 string
	IgnPoolNetReach               bool
	MinPoolsUp                    int
	RemoveListeningPortOnVsDown   bool
	SecurityPolicyRef             string
	SslSessCacheAvgSize           int
	Name                          string
	Namespace                     string
	AnalyticsProfile              KindNameNamespace
	ApplicationProfile            KindNameNamespace
	WafPolicy                     KindNameNamespace
	IcapProfile                   KindNameNamespace
	ErrorPageProfile              KindNameNamespace
	HTTPPolicy                    L7RuleHTTPPolicy
	AnalyticsPolicy               L7RuleAnalyticsPolicy
	Status                        string
	Error                         string
}

func GetFakeDefaultL7RuleObj(name, namespace string) *FakeL7Rule {

	l7Rule := FakeL7Rule{
		Name:      name,
		Namespace: namespace,
		Status:    "Accepted",
		Error:     "",
	}
	return &l7Rule
}

func (l7rule *FakeL7Rule) CreateFakeL7RuleWithStatus(t *testing.T) {
	l7RuleNew := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "ako.vmware.com/v1alpha2",
			"kind":       "L7Rule",
			"metadata": map[string]interface{}{
				"name":      l7rule.Name,
				"namespace": l7rule.Namespace,
			},
			"spec": map[string]interface{}{
				"allowInvalidClientCert":        true,
				"botPolicyRef":                  "sample-bot",
				"closeClientConnOnConfigUpdate": false,
				"hostNameXlate":                 "foo.com",
				"ignPoolNetReach":               true,
				"minPoolsUp":                    "2",
				"removeListeningPortOnVsDown":   false,
				"securityPolicyRef":             "thisisaviref-secpolicy",
				"sslSessCacheAvgSize":           "1024",
				"analyticsProfile": map[string]interface{}{
					"kind": "AviRef",
					"name": "thisisaviref-analyticsprofile-l7",
				},
				"applicationProfile": map[string]interface{}{
					"kind": "AviRef",
					"name": "thisisaviref-appprofile-l7",
				},
				"wafPolicy": map[string]interface{}{
					"kind": "AviRef",
					"name": "thisisaviref-wafpolicy-l7",
				},
				"icapProfile": map[string]interface{}{
					"kind": "AviRef",
					"name": "thisisaviref-icaprofile-l7",
				},
				"errorPageProfile": map[string]interface{}{
					"kind": "AviRef",
					"name": "thisisaviref-errorpageprofile-l7",
				},
				"httpPolicy": map[string]interface{}{
					"overwrite":  l7rule.HTTPPolicy.Overwrite,
					"policySets": []interface{}{"policy1", "policy2"},
				},
				"analyticsPolicy": map[string]interface{}{
					"logAllHeaders": true,
					"fullClientLogs": map[string]interface{}{
						"enabled": true,
					},
				},
			},
			"status": map[string]interface{}{
				"status": l7rule.Status,
				"error":  l7rule.Error,
			},
		},
	}
	_, err := DynamicClient.Resource(akogatewayapilib.L7CRDGVR).Namespace(l7rule.Namespace).Create(context.TODO(), l7RuleNew, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in creating L7Rule: %v", err)
	}
	t.Logf("Created L7Rule %s/%s with status=%s", l7rule.Namespace, l7rule.Name, l7rule.Status)
}

func (l7rule *FakeL7Rule) DeleteL7RuleCR(t *testing.T) {
	err := DynamicClient.Resource(akogatewayapilib.L7CRDGVR).Namespace(l7rule.Namespace).Delete(context.TODO(), l7rule.Name, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error in deleting RouteBackendExtension: %v", err)
	}
	t.Logf("Deleted RouteBackendExtension %s/%s", l7rule.Namespace, l7rule.Name)
}

func (l7rule *FakeL7Rule) UpdateL7RuleStatus(t *testing.T) {
	l7RuleResource, err := DynamicClient.Resource(akogatewayapilib.L7CRDGVR).Namespace(l7rule.Namespace).Get(context.TODO(), l7rule.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("error in getting L7Rule: %v", err)
	}

	// Update the status section
	statusObj := l7RuleResource.Object["status"].(map[string]interface{})
	statusObj["status"] = l7rule.Status
	statusObj["error"] = l7rule.Error

	_, err = DynamicClient.Resource(akogatewayapilib.L7CRDGVR).Namespace(l7rule.Namespace).UpdateStatus(context.TODO(), l7RuleResource, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating L7Rule status: %v", err)
	}
	t.Logf("Updated L7Rule %s/%s status to Status=%s, Error=%s", l7rule.Namespace, l7rule.Name, l7rule.Status, l7rule.Error)
}

// @AI-Generated
// [Generated by Cursor claude-4.5-sonnet]
func (l7rule *FakeL7Rule) CreateFakeL7RuleWithCustomDuration(t *testing.T, duration int64, enabled bool, throttle string) {
	l7RuleNew := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "ako.vmware.com/v1alpha2",
			"kind":       "L7Rule",
			"metadata": map[string]interface{}{
				"name":      l7rule.Name,
				"namespace": l7rule.Namespace,
			},
			"spec": map[string]interface{}{
				"allowInvalidClientCert":        true,
				"botPolicyRef":                  "sample-bot",
				"closeClientConnOnConfigUpdate": false,
				"hostNameXlate":                 "foo.com",
				"ignPoolNetReach":               true,
				"minPoolsUp":                    "2",
				"removeListeningPortOnVsDown":   false,
				"securityPolicyRef":             "thisisaviref-secpolicy",
				"sslSessCacheAvgSize":           "1024",
				"analyticsProfile": map[string]interface{}{
					"kind": "AviRef",
					"name": "thisisaviref-analyticsprofile-l7",
				},
				"applicationProfile": map[string]interface{}{
					"kind": "AviRef",
					"name": "thisisaviref-appprofile-l7",
				},
				"wafPolicy": map[string]interface{}{
					"kind": "AviRef",
					"name": "thisisaviref-wafpolicy-l7",
				},
				"icapProfile": map[string]interface{}{
					"kind": "AviRef",
					"name": "thisisaviref-icaprofile-l7",
				},
				"errorPageProfile": map[string]interface{}{
					"kind": "AviRef",
					"name": "thisisaviref-errorpageprofile-l7",
				},
				"httpPolicy": map[string]interface{}{
					"overwrite":  l7rule.HTTPPolicy.Overwrite,
					"policySets": []interface{}{"policy1", "policy2"},
				},
				"analyticsPolicy": map[string]interface{}{
					"logAllHeaders": true,
					"fullClientLogs": map[string]interface{}{
						"duration": duration,
						"enabled":  enabled,
						"throttle": throttle,
					},
				},
			},
			"status": map[string]interface{}{
				"status": l7rule.Status,
				"error":  l7rule.Error,
			},
		},
	}
	_, err := DynamicClient.Resource(akogatewayapilib.L7CRDGVR).Namespace(l7rule.Namespace).Create(context.TODO(), l7RuleNew, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in creating L7Rule: %v", err)
	}
	t.Logf("Created L7Rule %s/%s with duration=%d, enabled=%t, throttle=%s", l7rule.Namespace, l7rule.Name, duration, enabled, throttle)
}

type FakeApplicationProfileStatus struct {
	Status  string
	Reason  string
	Message string
	UUID    string
}

func GetFakeApplicationProfile(name string, status *FakeApplicationProfileStatus) unstructured.Unstructured {
	appProfile := unstructured.Unstructured{}

	ob := map[string]interface{}{
		"apiVersion": "ako.vmware.com/v1alpha1",
		"kind":       "ApplicationProfile",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"type": "APPLICATION_PROFILE_TYPE_HTTP",
		},
	}
	if status != nil {
		statusMap := map[string]interface{}{
			"backendObjectName": name,
			"conditions": []interface{}{
				map[string]interface{}{
					"type":    "Programmed",
					"status":  status.Status,
					"reason":  status.Reason,
					"message": status.Message,
				},
			},
			"tenant": "admin",
		}
		if status.UUID != "" {
			statusMap["uuid"] = status.UUID
		}
		ob["status"] = statusMap
	}

	appProfile.SetUnstructuredContent(ob)
	return appProfile
}

func CreateApplicationProfileCRD(t *testing.T, name string, status *FakeApplicationProfileStatus) {
	appProfile := GetFakeApplicationProfile(name, status)
	_, err := DynamicClient.Resource(akogatewayapilib.AppProfileCRDGVR).Namespace("default").Create(context.TODO(), &appProfile, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in creating ApplicationProfile CRD: %v", err)
	}
}

func DeleteApplicationProfileCRD(t *testing.T, name string) {
	err := DynamicClient.Resource(akogatewayapilib.AppProfileCRDGVR).Namespace("default").Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error in deleting ApplicationProfile CRD: %v", err)
	}
}

func GetApplicationProfileCRD(t *testing.T, name string) *unstructured.Unstructured {
	crd, err := DynamicClient.Resource(akogatewayapilib.AppProfileCRDGVR).Namespace("default").Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("error in getting ApplicationProfile CRD: %v", err)
	}
	return crd
}

func UpdateApplicationProfileCRD(t *testing.T, name string, status *FakeApplicationProfileStatus) {
	appProfile := GetFakeApplicationProfile(name, status)
	_, err := DynamicClient.Resource(akogatewayapilib.AppProfileCRDGVR).Namespace("default").Update(context.TODO(), &appProfile, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating ApplicationProfile CRD: %v", err)
	}
}

func GetHTTPRouteRuleWithCustomCRDs(pathType string, paths []string, headers []string, filters map[string][]string, backends [][]string, filterExtensionCrds map[string][]string) gatewayv1.HTTPRouteRule {
	rule := GetHTTPRouteRuleV1(pathType, paths, headers, filters, backends, nil)
	// Frontend - L7 Rule
	if l7Rules, ok := filterExtensionCrds["L7Rule"]; ok {
		for _, l7Rule := range l7Rules {
			filter := gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterExtensionRef,
				ExtensionRef: &gatewayv1.LocalObjectReference{
					Group: gatewayv1.Group("ako.vmware.com"),
					Kind:  gatewayv1.Kind("L7Rule"),
					Name:  gatewayv1.ObjectName(l7Rule),
				},
			}
			rule.Filters = append(rule.Filters, filter)
		}
	}
	if appProfileRules, ok := filterExtensionCrds["ApplicationProfile"]; ok {
		for _, appProfile := range appProfileRules {
			filter := gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterExtensionRef,
				ExtensionRef: &gatewayv1.LocalObjectReference{
					Group: gatewayv1.Group("ako.vmware.com"),
					Kind:  gatewayv1.Kind("ApplicationProfile"),
					Name:  gatewayv1.ObjectName(appProfile),
				},
			}
			rule.Filters = append(rule.Filters, filter)
		}
	}
	return rule
}

// SetupDedicatedGateway creates a Gateway with dedicated mode annotation
func SetupDedicatedGateway(t *testing.T, name, namespace, gatewayClass string, ipAddress []gatewayv1.GatewaySpecAddress, listeners []gatewayv1.Listener) {
	g := &Gateway{}
	g.Gateway = g.GatewayV1(name, namespace, gatewayClass, ipAddress, listeners)
	// Add dedicated mode annotation
	if g.Gateway.Annotations == nil {
		g.Gateway.Annotations = make(map[string]string)
	}
	g.Gateway.Annotations[akogatewayapilib.DedicatedGatewayModeAnnotation] = "true"
	g.Create(t)
}

// UpdateDedicatedGateway updates a Gateway with dedicated mode annotation
func UpdateDedicatedGateway(t *testing.T, name, namespace, gatewayClass string, ipAddress []gatewayv1.GatewaySpecAddress, listeners []gatewayv1.Listener) {
	g := &Gateway{}
	g.Gateway = g.GatewayV1(name, namespace, gatewayClass, ipAddress, listeners)
	// Add dedicated mode annotation
	if g.Gateway.Annotations == nil {
		g.Gateway.Annotations = make(map[string]string)
	}
	g.Gateway.Annotations[akogatewayapilib.DedicatedGatewayModeAnnotation] = "true"
	g.Update(t)
}

// GetDedicatedListenersV1 creates listeners for dedicated mode gateways
func GetDedicatedListenersV1(ports []int32, secrets ...string) []gatewayv1.Listener {
	listeners := make([]gatewayv1.Listener, 0, len(ports))
	for _, port := range ports {
		listener := gatewayv1.Listener{
			Name:     gatewayv1.SectionName(fmt.Sprintf("listener-%d", port)),
			Port:     gatewayv1.PortNumber(port),
			Protocol: gatewayv1.ProtocolType("HTTP"), // Default to HTTP
		}

		if len(secrets) > 0 {
			certRefs := make([]gatewayv1.SecretObjectReference, 0, len(secrets))
			for _, secret := range secrets {
				secretRef := gatewayv1.SecretObjectReference{
					Name: gatewayv1.ObjectName(secret),
				}
				certRefs = append(certRefs, secretRef)
			}
			tlsMode := gatewayv1.TLSModeTerminate
			listener.TLS = &gatewayv1.ListenerTLSConfig{
				Mode:            &tlsMode,
				CertificateRefs: certRefs,
			}
			listener.Protocol = gatewayv1.ProtocolType("HTTPS")
		}
		listeners = append(listeners, listener)
	}
	return listeners
}
