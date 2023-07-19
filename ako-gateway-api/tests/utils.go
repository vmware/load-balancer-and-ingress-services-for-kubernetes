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
	"testing"

	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	gatewayfake "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned/fake"

	akogatewayapik8s "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/k8s"
	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
)

var KubeClient *k8sfake.Clientset
var GatewayClient *gatewayfake.Clientset
var keyChan chan string
var ctrl *akogatewayapik8s.GatewayController

func SetGatewayName(gw *gatewayv1beta1.Gateway, name string) {
	gw.Name = name
}
func UnsetGatewayName(gw *gatewayv1beta1.Gateway) {
	gw.Name = ""
}

func SetGatewayGatewayClass(gw *gatewayv1beta1.Gateway, name string) {
	gw.Spec.GatewayClassName = gatewayv1beta1.ObjectName(name)
}
func UnsetGatewayGatewayClass(gw *gatewayv1beta1.Gateway) {
	gw.Spec.GatewayClassName = ""
}

func AddGatewayListener(gw *gatewayv1beta1.Gateway, name string, port int32, protocol gatewayv1beta1.ProtocolType, isTLS bool) {

	listner := gatewayv1beta1.Listener{
		Name:     gatewayv1beta1.SectionName(name),
		Port:     gatewayv1beta1.PortNumber(port),
		Protocol: protocol,
	}
	if isTLS {
		SetListenerTLS(&listner, gatewayv1beta1.TLSModeTerminate, "secret-example", "default")
	}
	gw.Spec.Listeners = append(gw.Spec.Listeners, listner)
}

func SetListenerTLS(l *gatewayv1beta1.Listener, tlsMode gatewayv1beta1.TLSModeType, secretName, secretNS string) {
	l.TLS = &gatewayv1beta1.GatewayTLSConfig{Mode: &tlsMode}
	namespace := gatewayv1beta1.Namespace(secretNS)
	kind := gatewayv1beta1.Kind("Secret")
	l.TLS.CertificateRefs = []gatewayv1beta1.SecretObjectReference{
		{
			Name:      gatewayv1beta1.ObjectName(secretName),
			Namespace: &namespace,
			Kind:      &kind,
		},
	}
}
func UnsetListenerTLS(l *gatewayv1beta1.Listener) {
	l.TLS = &gatewayv1beta1.GatewayTLSConfig{}
}

func SetListenerHostname(l *gatewayv1beta1.Listener, hostname string) {
	l.Hostname = (*gatewayv1beta1.Hostname)(&hostname)
}
func UnsetListenerHostname(l *gatewayv1beta1.Listener) {
	var hname gatewayv1beta1.Hostname
	l.Hostname = &hname
}

func GetListenersV1Beta1(ports []int32, secrets ...string) []gatewayv1beta1.Listener {
	listeners := make([]gatewayv1beta1.Listener, 0, len(ports))
	for _, port := range ports {
		hostname := fmt.Sprintf("foo-%d.com", port)
		listener := gatewayv1beta1.Listener{
			Name:     gatewayv1beta1.SectionName(fmt.Sprintf("listener-%d", port)),
			Port:     gatewayv1beta1.PortNumber(port),
			Protocol: gatewayv1beta1.ProtocolType("HTTPS"),
			Hostname: (*gatewayv1beta1.Hostname)(&hostname),
		}
		if len(secrets) > 0 {
			certRefs := make([]gatewayv1beta1.SecretObjectReference, 0, len(secrets))
			for _, secret := range secrets {
				secretRef := gatewayv1beta1.SecretObjectReference{
					Name: gatewayv1beta1.ObjectName(secret),
				}
				certRefs = append(certRefs, secretRef)
			}
			tlsMode := "Terminate"
			listener.TLS = &gatewayv1beta1.GatewayTLSConfig{
				Mode:            (*gatewayv1beta1.TLSModeType)(&tlsMode),
				CertificateRefs: certRefs,
			}
		}
		listeners = append(listeners, listener)
	}
	return listeners
}

func GetListenerStatusV1Beta1(ports []int32, attachedRoutes []int32) []gatewayv1beta1.ListenerStatus {
	listeners := make([]gatewayv1beta1.ListenerStatus, 0, len(ports))
	for i, port := range ports {
		listener := gatewayv1beta1.ListenerStatus{
			Name:           gatewayv1beta1.SectionName(fmt.Sprintf("listener-%d", port)),
			SupportedKinds: akogatewayapilib.SupportedKinds["HTTPS"],
			AttachedRoutes: attachedRoutes[i],
			Conditions: []metav1.Condition{
				{
					Type:               string(gatewayv1beta1.GatewayConditionAccepted),
					Status:             metav1.ConditionTrue,
					Message:            "Listener is valid",
					ObservedGeneration: 1,
					Reason:             string(gatewayv1beta1.GatewayReasonAccepted),
				},
			},
		}
		listeners = append(listeners, listener)
	}
	return listeners
}

type Gateway struct {
	*gatewayv1beta1.Gateway
}

func (g *Gateway) GatewayV1Beta1(name, namespace, gatewayClass string, address []gatewayv1beta1.GatewayAddress, listeners []gatewayv1beta1.Listener) *gatewayv1beta1.Gateway {
	gateway := &gatewayv1beta1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: gatewayv1beta1.GatewaySpec{
			GatewayClassName: gatewayv1beta1.ObjectName(gatewayClass),
			Addresses:        address,
		},
	}

	gateway.Spec.Listeners = listeners
	return gateway
}

func (g *Gateway) Create(t *testing.T) {
	_, err := GatewayClient.GatewayV1beta1().Gateways(g.Namespace).Create(context.TODO(), g.Gateway, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Couldn't create the gateway, err: %+v", err)
	}
}

func (g *Gateway) Update(t *testing.T) {
	_, err := GatewayClient.GatewayV1beta1().Gateways(g.Namespace).Update(context.TODO(), g.Gateway, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Couldn't update the gateway, err: %+v", err)
	}
}

func (g *Gateway) Delete(t *testing.T) {
	err := GatewayClient.GatewayV1beta1().Gateways(g.Namespace).Delete(context.TODO(), g.Name, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't delete the gateway, err: %+v", err)
	}
}

func SetupGateway(t *testing.T, name, namespace, gatewayClass string, ipAddress []gatewayv1beta1.GatewayAddress, listeners []gatewayv1beta1.Listener) {
	g := &Gateway{}
	g.Gateway = g.GatewayV1Beta1(name, namespace, gatewayClass, ipAddress, listeners)
	g.Create(t)
}

func UpdateGateway(t *testing.T, name, namespace, gatewayClass string, ipAddress []gatewayv1beta1.GatewayAddress, listeners []gatewayv1beta1.Listener) {
	g := &Gateway{}
	g.Gateway = g.GatewayV1Beta1(name, namespace, gatewayClass, ipAddress, listeners)
	g.Update(t)
}

func TeardownGateway(t *testing.T, name, namespace string) {
	g := &Gateway{}
	g.Gateway = g.GatewayV1Beta1(name, namespace, "", nil, nil)
	g.Delete(t)
}

type FakeGatewayClass struct {
	Name           string
	ControllerName string
}

func (gc *FakeGatewayClass) GatewayClassV1Beta1() *gatewayv1beta1.GatewayClass {
	return &gatewayv1beta1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: gc.Name,
		},
		Spec: gatewayv1beta1.GatewayClassSpec{
			ControllerName: gatewayv1beta1.GatewayController(gc.ControllerName),
		},
	}
}

func (gc *FakeGatewayClass) Create(t *testing.T) {
	gatewayClass := gc.GatewayClassV1Beta1()
	_, err := GatewayClient.GatewayV1beta1().GatewayClasses().Create(context.TODO(), gatewayClass, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Couldn't create the gateway class, err: %+v", err)
	}
}

func (gc *FakeGatewayClass) Update(t *testing.T) {
	gatewayClass := gc.GatewayClassV1Beta1()
	_, err := GatewayClient.GatewayV1beta1().GatewayClasses().Update(context.TODO(), gatewayClass, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Couldn't update the gateway class, err: %+v", err)
	}
}

func (gc *FakeGatewayClass) Delete(t *testing.T) {
	err := GatewayClient.GatewayV1beta1().GatewayClasses().Delete(context.TODO(), gc.Name, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't delete the gateway class, err: %+v", err)
	}
}

func SetupGatewayClass(t *testing.T, name, controllerName string) {
	gc := &FakeGatewayClass{
		Name:           name,
		ControllerName: controllerName,
	}
	gc.Create(t)
}

func TeardownGatewayClass(t *testing.T, name string) {
	gc := &FakeGatewayClass{
		Name: name,
	}
	gc.Delete(t)
}

func ValidateGatewayStatus(t *testing.T, actualStatus, expectedStatus *gatewayv1beta1.GatewayStatus) {

	g := gomega.NewGomegaWithT(t)

	// validate the ip address
	if len(expectedStatus.Addresses) > 0 {
		g.Expect(actualStatus.Addresses).To(gomega.HaveLen(1))
		g.Expect(actualStatus.Addresses[0]).Should(gomega.Equal(expectedStatus.Addresses[0]))
	}

	ValidateConditions(t, actualStatus.Conditions, expectedStatus.Conditions)

	g.Expect(actualStatus.Listeners).To(gomega.HaveLen(len(expectedStatus.Listeners)))
	for _, actualListenerStatus := range actualStatus.Listeners {
		for _, expectedListenerStatus := range expectedStatus.Listeners {
			if actualListenerStatus.Name == expectedListenerStatus.Name {
				ValidateGatewayListeners(t, &actualListenerStatus, &expectedListenerStatus)
			}
		}
	}
}

func ValidateGatewayListeners(t *testing.T, actual, expected *gatewayv1beta1.ListenerStatus) {
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
