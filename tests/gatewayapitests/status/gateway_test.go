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
	"os"
	"sync"
	"testing"
	"time"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayfake "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned/fake"

	akogatewayapik8s "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/k8s"
	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	tests "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/gatewayapitests"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"
)

var ctrl *akogatewayapik8s.GatewayController
var endpointSliceEnabled bool

const (
	DEFAULT_NAMESPACE = "default"
)

func TestMain(m *testing.M) {
	tests.KubeClient = k8sfake.NewSimpleClientset()
	tests.GatewayClient = gatewayfake.NewSimpleClientset()
	integrationtest.KubeClient = tests.KubeClient

	// Sets the environment variables
	os.Setenv("CLUSTER_NAME", "cluster")
	os.Setenv("CLOUD_NAME", "CLOUD_VCENTER")
	os.Setenv("SEG_NAME", "Default-Group")
	os.Setenv("POD_NAMESPACE", utils.AKO_DEFAULT_NS)
	os.Setenv("POD_NAME", "ako-0")
	os.Setenv("ENABLE_EVH", "true")

	utils.AviLog.SetLevel("DEBUG")
	// Set the user with prefix
	_ = lib.AKOControlConfig()
	lib.SetAKOUser(akogatewayapilib.Prefix)
	lib.SetNamePrefix(akogatewayapilib.Prefix)
	endpointSliceEnabled = lib.GetEndpointSliceEnabled()
	lib.AKOControlConfig().SetEndpointSlicesEnabled(endpointSliceEnabled)
	lib.AKOControlConfig().SetIsLeaderFlag(true)
	akoControlConfig := akogatewayapilib.AKOControlConfig()
	lib.AKOControlConfig().SetIsLeaderFlag(true)
	akoControlConfig.SetEventRecorder(lib.AKOGatewayEventComponent, tests.KubeClient, true)
	registeredInformers := []string{
		utils.ServiceInformer,
		utils.SecretInformer,
		utils.NSInformer,
		utils.ConfigMapInformer,
	}
	if lib.AKOControlConfig().GetEndpointSlicesEnabled() {
		registeredInformers = append(registeredInformers, utils.EndpointSlicesInformer)
	} else {
		registeredInformers = append(registeredInformers, utils.EndpointInformer)
	}
	utils.NewInformers(utils.KubeClientIntf{ClientSet: tests.KubeClient}, registeredInformers, make(map[string]interface{}))
	data := map[string][]byte{
		"username": []byte("admin"),
		"password": []byte("admin"),
	}
	object := metav1.ObjectMeta{Name: "avi-secret", Namespace: utils.GetAKONamespace()}
	secret := &corev1.Secret{Data: data, ObjectMeta: object}
	tests.KubeClient.CoreV1().Secrets(utils.GetAKONamespace()).Create(context.TODO(), secret, metav1.CreateOptions{})

	akoApi := integrationtest.InitializeFakeAKOAPIServer()
	defer akoApi.ShutDown()

	tests.NewAviFakeClientInstance(tests.KubeClient)
	defer integrationtest.AviFakeClientInstance.Close()

	ctrl = akogatewayapik8s.SharedGatewayController()
	ctrl.DisableSync = false
	ctrl.InitGatewayAPIInformers(tests.GatewayClient)
	akoControlConfig.SetGatewayAPIClientset(tests.GatewayClient)

	stopCh := utils.SetupSignalHandler()
	ctrlCh := make(chan struct{})
	quickSyncCh := make(chan struct{})

	waitGroupMap := make(map[string]*sync.WaitGroup)
	wgIngestion := &sync.WaitGroup{}
	waitGroupMap["ingestion"] = wgIngestion
	wgFastRetry := &sync.WaitGroup{}
	waitGroupMap["fastretry"] = wgFastRetry
	wgSlowRetry := &sync.WaitGroup{}
	waitGroupMap["slowretry"] = wgSlowRetry
	wgGraph := &sync.WaitGroup{}
	waitGroupMap["graph"] = wgGraph
	wgStatus := &sync.WaitGroup{}
	waitGroupMap["status"] = wgStatus

	integrationtest.AddConfigMap(tests.KubeClient)
	go ctrl.InitController(k8s.K8sinformers{Cs: tests.KubeClient}, registeredInformers, ctrlCh, stopCh, quickSyncCh, waitGroupMap)
	os.Exit(m.Run())
}

/* Positive test cases
 * - Gateway with valid listeners and GatewayClass
 * - Gateway with TLS listeners
 * - Gateway with valid IP address (TODO: cannot be added now since end-to-end code is not present)
 * - Gateway update of listeners (removing 1 listener from say 3)
 */
func TestGatewayWithValidListenersAndGatewayClass(t *testing.T) {

	gatewayName := "gateway-01"
	gatewayClassName := "gateway-class-01"
	ports := []int32{8080, 8081}

	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := tests.GetListenersV1(ports, false, false)
	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionProgrammed)) != nil
	}, 40*time.Second).Should(gomega.Equal(true))

	expectedStatus := &gatewayv1.GatewayStatus{
		Conditions: []metav1.Condition{
			{
				Type:               string(gatewayv1.GatewayConditionAccepted),
				Status:             metav1.ConditionTrue,
				Message:            "Gateway configuration is valid",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonAccepted),
			},
			{
				Type:               string(gatewayv1.GatewayConditionProgrammed),
				Status:             metav1.ConditionTrue,
				Message:            "Virtual service configured/updated",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonProgrammed),
			},
		},
		Listeners: tests.GetListenerStatusV1(ports, []int32{0, 0}, true, true),
	}

	gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	if err != nil || gateway == nil {
		t.Fatalf("Couldn't get the gateway, err: %+v", err)
	}

	tests.ValidateGatewayStatus(t, &gateway.Status, expectedStatus)
	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
}

func TestGatewayWithTLSListeners(t *testing.T) {

	gatewayName := "gateway-02"
	gatewayClassName := "gateway-class-02"
	ports := []int32{8080, 8081}
	secrets := []string{"secret-01"}
	for _, secret := range secrets {
		integrationtest.AddSecret(secret, DEFAULT_NAMESPACE, "cert", "key")
	}
	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := tests.GetListenersV1(ports, false, false, secrets...)
	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	expectedStatus := &gatewayv1.GatewayStatus{
		Conditions: []metav1.Condition{
			{
				Type:               string(gatewayv1.GatewayConditionAccepted),
				Status:             metav1.ConditionTrue,
				Message:            "Gateway configuration is valid",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonAccepted),
			},
		},
		Listeners: tests.GetListenerStatusV1(ports, []int32{0, 0}, true, false),
	}

	gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	if err != nil || gateway == nil {
		t.Fatalf("Couldn't get the gateway, err: %+v", err)
	}

	tests.ValidateGatewayStatus(t, &gateway.Status, expectedStatus)
	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
	for _, secret := range secrets {
		integrationtest.DeleteSecret(secret, DEFAULT_NAMESPACE)
	}
}

func TestGatewayListenerUpdate(t *testing.T) {

	gatewayName := "gateway-03"
	gatewayClassName := "gateway-class-03"
	ports := []int32{8080, 8081, 8082}

	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := tests.GetListenersV1(ports, false, false)
	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	expectedStatus := &gatewayv1.GatewayStatus{
		Conditions: []metav1.Condition{
			{
				Type:               string(gatewayv1.GatewayConditionAccepted),
				Status:             metav1.ConditionTrue,
				Message:            "Gateway configuration is valid",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonAccepted),
			},
		},
		Listeners: tests.GetListenerStatusV1(ports, []int32{0, 0, 0}, true, false),
	}

	gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	if err != nil || gateway == nil {
		t.Fatalf("Couldn't get the gateway, err: %+v", err)
	}

	tests.ValidateGatewayStatus(t, &gateway.Status, expectedStatus)

	// Update the Gateway with new listeners
	ports = []int32{8080, 8082}
	listeners = tests.GetListenersV1(ports, false, false)
	tests.UpdateGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return len(gateway.Status.Listeners) == len(ports)
	}, 30*time.Second).Should(gomega.Equal(true))

	expectedStatus.Listeners = tests.GetListenerStatusV1(ports, []int32{0, 0}, true, false)
	gateway, err = tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	if err != nil || gateway == nil {
		t.Fatalf("Couldn't get the gateway, err: %+v", err)
	}

	tests.ValidateGatewayStatus(t, &gateway.Status, expectedStatus)

	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
}

/*
Transition test cases
* - Valid Gateway configuration to invalid
* - Invalid Gateway configuration to valid
* - Non AKO gateway controller to AKO gateway controller
* - AKO gateway controller to non AKO gateway controller
*/
func TestGatewayTransitionFromValidToPartiallyValid(t *testing.T) {

	t.Skip("This is invalid test case as Hostname in listener can not be *.")
	gatewayName := "gateway-trans-01"
	gatewayClassName := "gateway-class-trans-01"
	ports := []int32{8080, 8081}

	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := tests.GetListenersV1(ports, false, false)
	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	expectedStatus := &gatewayv1.GatewayStatus{
		Conditions: []metav1.Condition{
			{
				Type:               string(gatewayv1.GatewayConditionAccepted),
				Status:             metav1.ConditionTrue,
				Message:            "Gateway configuration is valid",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonAccepted),
			},
		},
		Listeners: tests.GetListenerStatusV1(ports, []int32{0, 0}, true, false),
	}

	gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	if err != nil || gateway == nil {
		t.Fatalf("Couldn't get the gateway, err: %+v", err)
	}

	tests.ValidateGatewayStatus(t, &gateway.Status, expectedStatus)

	// Update the gateway with an invalid configuration
	invalidHostname := "*"
	listeners[0].Hostname = (*gatewayv1.Hostname)(&invalidHostname)
	tests.UpdateGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return len(gateway.Status.Listeners) == len(ports)
	}, 30*time.Second).Should(gomega.Equal(true))

	expectedStatus = &gatewayv1.GatewayStatus{
		Conditions: []metav1.Condition{
			{
				Type:               string(gatewayv1.GatewayConditionAccepted),
				Status:             metav1.ConditionTrue,
				Message:            "Gateway contains atleast one valid listener",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonListenersNotValid),
			},
		},
		Listeners: tests.GetListenerStatusV1(ports, []int32{0, 0}, true, false),
	}
	expectedStatus.Listeners[0].Conditions[0].Reason = string(gatewayv1.ListenerReasonInvalid)
	expectedStatus.Listeners[0].Conditions[0].Status = metav1.ConditionFalse
	expectedStatus.Listeners[0].Conditions[0].Message = "Hostname not found or Hostname has invalid configuration"

	expectedStatus.Listeners[0].Conditions[1].Type = string(gatewayv1.ListenerConditionProgrammed)
	expectedStatus.Listeners[0].Conditions[1].Status = metav1.ConditionFalse
	expectedStatus.Listeners[0].Conditions[1].Reason = string(gatewayv1.ListenerReasonInvalid)
	expectedStatus.Listeners[0].Conditions[1].Message = "Virtual service not configured/updated for this listener"

	gateway, err = tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	if err != nil || gateway == nil {
		t.Fatalf("Couldn't get the gateway, err: %+v", err)
	}
	tests.ValidateGatewayStatus(t, &gateway.Status, expectedStatus)
	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
}

func TestGatewayTransitionFromPartiallyValidToValid(t *testing.T) {

	gatewayName := "gateway-trans-02"
	gatewayClassName := "gateway-class-trans-02"
	ports := []int32{8080, 8081}

	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, nil)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	expectedStatus := &gatewayv1.GatewayStatus{
		Conditions: []metav1.Condition{
			{
				Type:               string(gatewayv1.GatewayConditionAccepted),
				Status:             metav1.ConditionFalse,
				Message:            "No listeners found",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonInvalid),
			},
		},
	}

	gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	if err != nil || gateway == nil {
		t.Fatalf("Couldn't get the gateway, err: %+v", err)
	}

	tests.ValidateGatewayStatus(t, &gateway.Status, expectedStatus)

	// Update the gateway with a valid configuration
	listeners := tests.GetListenersV1(ports, false, false)
	tests.UpdateGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return len(gateway.Status.Listeners) == len(ports)
	}, 30*time.Second).Should(gomega.Equal(true))

	expectedStatus.Conditions[0].Status = metav1.ConditionTrue
	expectedStatus.Conditions[0].Reason = string(gatewayv1.GatewayReasonAccepted)
	expectedStatus.Conditions[0].Message = "Gateway configuration is valid"
	expectedStatus.Listeners = tests.GetListenerStatusV1(ports, []int32{0, 0}, true, false)

	gateway, err = tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	if err != nil || gateway == nil {
		t.Fatalf("Couldn't get the gateway, err: %+v", err)
	}
	tests.ValidateGatewayStatus(t, &gateway.Status, expectedStatus)
	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
}

func TestGatewayTransitionFromNonAKOControllerToAKOController(t *testing.T) {

	gatewayName := "gateway-trans-03"
	gatewayClassNameWithNonAkoController := "gateway-class-trans-03-01"
	gatewayClassName := "gateway-class-trans-03-02"
	ports := []int32{8080, 8081}

	tests.SetupGatewayClass(t, gatewayClassNameWithNonAkoController, "foo.company.com/foo-gateway-controller")
	listeners := tests.GetListenersV1(ports, false, false)
	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassNameWithNonAkoController, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) == nil
	}, 30*time.Second).Should(gomega.Equal(true))

	// Update the gateway with ako as gateway controller
	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	tests.UpdateGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	expectedStatus := &gatewayv1.GatewayStatus{
		Conditions: []metav1.Condition{
			{
				Type:               string(gatewayv1.GatewayConditionAccepted),
				Status:             metav1.ConditionTrue,
				Message:            "Gateway configuration is valid",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonAccepted),
			},
		},
		Listeners: tests.GetListenerStatusV1(ports, []int32{0, 0}, true, false),
	}

	gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	if err != nil || gateway == nil {
		t.Fatalf("Couldn't get the gateway, err: %+v", err)
	}
	tests.ValidateGatewayStatus(t, &gateway.Status, expectedStatus)
	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
}

func TestGatewayTransitionFromAKOControllerToNonAKOController(t *testing.T) {

	gatewayName := "gateway-trans-04"
	gatewayClassNameWithNonAkoController := "gateway-class-trans-04-01"
	gatewayClassName := "gateway-class-trans-04-02"
	ports := []int32{8080, 8081}

	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := tests.GetListenersV1(ports, false, false)
	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	expectedStatus := &gatewayv1.GatewayStatus{
		Conditions: []metav1.Condition{
			{
				Type:               string(gatewayv1.GatewayConditionAccepted),
				Status:             metav1.ConditionTrue,
				Message:            "Gateway configuration is valid",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonAccepted),
			},
		},
		Listeners: tests.GetListenerStatusV1(ports, []int32{0, 0}, true, false),
	}

	gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	if err != nil || gateway == nil {
		t.Fatalf("Couldn't get the gateway, err: %+v", err)
	}
	tests.ValidateGatewayStatus(t, &gateway.Status, expectedStatus)

	// Update the gateway with non-ako gateway controller
	tests.SetupGatewayClass(t, gatewayClassNameWithNonAkoController, "foo.company.com/foo-gateway-controller")
	tests.UpdateGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassNameWithNonAkoController, nil, listeners)

	// TODO: Need a better logic to check this.
	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
}

/* Negative test cases
 * - Gateway with no listeners
 * - Gateway with more than one static address
 * - Gateway with invalid listeners
 *    - Listeners with unsupported protocol
 *    - Listeners with invalid hostname
 *    - Listeners with invalid TLS
 */
func TestGatewayWithNoListeners(t *testing.T) {

	gatewayName := "gateway-neg-01"
	gatewayClassName := "gateway-class-neg-01"

	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, nil)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	expectedStatus := &gatewayv1.GatewayStatus{
		Conditions: []metav1.Condition{
			{
				Type:               string(gatewayv1.GatewayConditionAccepted),
				Status:             metav1.ConditionFalse,
				Message:            "No listeners found",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonInvalid),
			},
			{
				Type:               string(gatewayv1.GatewayConditionProgrammed),
				Status:             metav1.ConditionFalse,
				Message:            "Gateway not programmed",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonInvalid),
			},
		},
	}

	gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	if err != nil || gateway == nil {
		t.Fatalf("Couldn't get the gateway, err: %+v", err)
	}

	tests.ValidateGatewayStatus(t, &gateway.Status, expectedStatus)
	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
}

func TestGatewayWithMoreThanOneAddress(t *testing.T) {

	gatewayName := "gateway-neg-02"
	gatewayClassName := "gateway-class-neg-02"
	ports := []int32{8080, 8081}

	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := tests.GetListenersV1(ports, false, false)
	fakeGateway := tests.Gateway{}
	addresses := []gatewayv1.GatewaySpecAddress{{Value: "10.10.10.1"}, {Value: "10.10.10.2"}}
	fakeGateway.Gateway = fakeGateway.GatewayV1(gatewayName, DEFAULT_NAMESPACE, gatewayClassName, addresses, listeners)
	fakeGateway.Create(t)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	expectedStatus := &gatewayv1.GatewayStatus{
		Conditions: []metav1.Condition{
			{
				Type:               string(gatewayv1.GatewayConditionAccepted),
				Status:             metav1.ConditionFalse,
				Message:            "More than one address is not supported",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonInvalid),
			},
			{
				Type:               string(gatewayv1.GatewayConditionProgrammed),
				Status:             metav1.ConditionFalse,
				Message:            "Gateway not programmed",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonAddressNotUsable),
			},
		},
	}

	gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	if err != nil || gateway == nil {
		t.Fatalf("Couldn't get the gateway, err: %+v", err)
	}

	tests.ValidateGatewayStatus(t, &gateway.Status, expectedStatus)
	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
}

func TestGatewayWithUnsupportedProtocolInListeners(t *testing.T) {

	gatewayName := "gateway-neg-03"
	gatewayClassName := "gateway-class-neg-03"
	ports := []int32{8080, 8081}

	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := tests.GetListenersV1(ports, false, false)
	listeners[0].Protocol = "GRPC"
	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	expectedStatus := &gatewayv1.GatewayStatus{
		Conditions: []metav1.Condition{
			{
				Type:               string(gatewayv1.GatewayConditionAccepted),
				Status:             metav1.ConditionTrue,
				Message:            "Gateway contains atleast one valid listener",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonListenersNotValid),
			},
		},
		Listeners: tests.GetListenerStatusV1(ports, []int32{0, 0}, true, false),
	}
	expectedStatus.Listeners[0].Conditions[0].Reason = string(gatewayv1.ListenerReasonUnsupportedProtocol)
	expectedStatus.Listeners[0].Conditions[0].Status = metav1.ConditionFalse
	expectedStatus.Listeners[0].Conditions[0].Message = "Unsupported protocol"

	expectedStatus.Listeners[1].Conditions[0].Reason = string(gatewayv1.ListenerReasonAccepted)
	expectedStatus.Listeners[1].Conditions[0].Status = metav1.ConditionTrue
	expectedStatus.Listeners[1].Conditions[0].Message = "Listener is valid"

	gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	if err != nil || gateway == nil {
		t.Fatalf("Couldn't get the gateway, err: %+v", err)
	}

	tests.ValidateGatewayStatus(t, &gateway.Status, expectedStatus)
	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
}

func TestGatewayWithInvalidHostnameInListeners(t *testing.T) {
	t.Skip("Skipping this test case as Hostname in Gateway listener can not be *. Can be removed.")
	gatewayName := "gateway-neg-04"
	gatewayClassName := "gateway-class-neg-04"
	ports := []int32{8080, 8081}

	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := tests.GetListenersV1(ports, false, false)
	invalidHostname := "*"
	listeners[0].Hostname = (*gatewayv1.Hostname)(&invalidHostname)
	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	expectedStatus := &gatewayv1.GatewayStatus{
		Conditions: []metav1.Condition{
			{
				Type:               string(gatewayv1.GatewayConditionAccepted),
				Status:             metav1.ConditionTrue,
				Message:            "Gateway contains atleast one valid listener",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonListenersNotValid),
			},
		},
		Listeners: tests.GetListenerStatusV1(ports, []int32{0, 0}, true, false),
	}
	expectedStatus.Listeners[0].Conditions[0].Reason = string(gatewayv1.ListenerReasonInvalid)
	expectedStatus.Listeners[0].Conditions[0].Status = metav1.ConditionFalse
	expectedStatus.Listeners[0].Conditions[0].Message = "Hostname not found or Hostname has invalid configuration"

	expectedStatus.Listeners[0].Conditions[1].Reason = string(gatewayv1.ListenerReasonInvalid)
	expectedStatus.Listeners[0].Conditions[1].Status = metav1.ConditionFalse
	expectedStatus.Listeners[0].Conditions[1].Type = string(gatewayv1.GatewayConditionProgrammed)
	expectedStatus.Listeners[0].Conditions[1].Message = "Virtual service not configured/updated for this listener"

	gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	if err != nil || gateway == nil {
		t.Fatalf("Couldn't get the gateway, err: %+v", err)
	}

	tests.ValidateGatewayStatus(t, &gateway.Status, expectedStatus)
	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
}

func TestGatewayWithInvalidTLSConfigInListeners(t *testing.T) {

	gatewayName := "gateway-neg-05"
	gatewayClassName := "gateway-class-neg-05"
	ports := []int32{8080, 8081}
	secrets := []string{"secret-02"}
	for _, secret := range secrets {
		integrationtest.AddSecret(secret, DEFAULT_NAMESPACE, "cert", "key")
	}
	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := tests.GetListenersV1(ports, false, false, secrets...)
	invalidTLSMode := "invalid-mode"
	listeners[0].TLS.Mode = (*gatewayv1.TLSModeType)(&invalidTLSMode)
	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	expectedStatus := &gatewayv1.GatewayStatus{
		Conditions: []metav1.Condition{
			{
				Type:               string(gatewayv1.GatewayConditionAccepted),
				Status:             metav1.ConditionTrue,
				Message:            "Gateway contains atleast one valid listener",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonListenersNotValid),
			},
		},
		Listeners: tests.GetListenerStatusV1(ports, []int32{0, 0}, true, true),
	}
	expectedStatus.Listeners[0].Conditions[0].Reason = string(gatewayv1.ListenerReasonInvalid)
	expectedStatus.Listeners[0].Conditions[0].Status = metav1.ConditionFalse
	expectedStatus.Listeners[0].Conditions[0].Message = "Listener is Invalid"

	expectedStatus.Listeners[0].Conditions[1].Reason = string(gatewayv1.ListenerReasonInvalidCertificateRef)
	expectedStatus.Listeners[0].Conditions[1].Status = metav1.ConditionFalse
	expectedStatus.Listeners[0].Conditions[1].Message = "TLS mode or reference not valid"

	expectedStatus.Listeners[0].Conditions[2].Reason = string(gatewayv1.ListenerReasonInvalid)
	expectedStatus.Listeners[0].Conditions[2].Status = metav1.ConditionFalse
	expectedStatus.Listeners[0].Conditions[2].Message = "Virtual service not configured/updated for this listener"

	gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	if err != nil || gateway == nil {
		t.Fatalf("Couldn't get the gateway, err: %+v", err)
	}

	tests.ValidateGatewayStatus(t, &gateway.Status, expectedStatus)
	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
	for _, secret := range secrets {
		integrationtest.DeleteSecret(secret, DEFAULT_NAMESPACE)
	}
}

func TestGatewayWithInvalidAllowedRoute(t *testing.T) {

	gatewayName := "gateway-neg-06"
	gatewayClassName := "gateway-class-06"
	ports := []int32{8080}

	// Checking for Invalid RouteKind -> Kind
	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := tests.GetListenersV1(ports, false, false)
	allowedRoutes := gatewayv1.AllowedRoutes{
		Kinds: []gatewayv1.RouteGroupKind{{
			Kind: "Services",
		},
		},
	}
	listeners[0].AllowedRoutes = &allowedRoutes
	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	expectedStatus := &gatewayv1.GatewayStatus{
		Conditions: []metav1.Condition{
			{
				Type:               string(gatewayv1.GatewayConditionAccepted),
				Status:             metav1.ConditionFalse,
				Message:            "Gateway does not contain any valid listener",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonListenersNotValid),
			},
			{
				Type:               string(gatewayv1.GatewayConditionProgrammed),
				Status:             metav1.ConditionFalse,
				Message:            "Gateway not programmed",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonInvalid),
			},
		},
		Listeners: tests.GetListenerStatusV1(ports, []int32{0, 0}, true, true),
	}
	expectedStatus.Listeners[0].Conditions[0].Reason = string(gatewayv1.ListenerReasonInvalid)
	expectedStatus.Listeners[0].Conditions[0].Status = metav1.ConditionFalse
	expectedStatus.Listeners[0].Conditions[0].Message = "Listener is Invalid"
	//expectedStatus.Listeners[0].Conditions[0].Type = string(gatewayv1.ListenerConditionAccepted)

	expectedStatus.Listeners[0].Conditions[1].Reason = string(gatewayv1.ListenerReasonInvalidRouteKinds)
	expectedStatus.Listeners[0].Conditions[1].Status = metav1.ConditionFalse
	expectedStatus.Listeners[0].Conditions[1].Message = "AllowedRoute kind is invalid. Only HTTPRoute is supported currently"
	//expectedStatus.Listeners[0].Conditions[1].Type = string(gatewayv1.ListenerConditionResolvedRefs)

	expectedStatus.Listeners[0].Conditions[2].Message = "Virtual service not configured/updated for this listener"
	expectedStatus.Listeners[0].Conditions[2].Status = metav1.ConditionFalse
	expectedStatus.Listeners[0].Conditions[2].Reason = string(gatewayv1.ListenerReasonInvalid)

	gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	if err != nil || gateway == nil {
		t.Fatalf("Couldn't get the gateway, err: %+v", err)
	}

	tests.ValidateGatewayStatus(t, &gateway.Status, expectedStatus)
	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)

	// Checking for Invalid RouteKind -> Group
	allowedRoutes.Kinds[0].Kind = "HTTPRoute"
	invalidGroup := "InvalidGroup.example.com"
	allowedRoutes.Kinds[0].Group = (*gatewayv1.Group)(&invalidGroup)

	listeners[0].AllowedRoutes = &allowedRoutes
	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	expectedStatus = &gatewayv1.GatewayStatus{
		Conditions: []metav1.Condition{
			{
				Type:               string(gatewayv1.GatewayConditionAccepted),
				Status:             metav1.ConditionFalse,
				Message:            "Gateway does not contain any valid listener",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonListenersNotValid),
			},
			{
				Type:               string(gatewayv1.GatewayConditionProgrammed),
				Status:             metav1.ConditionFalse,
				Message:            "Gateway not programmed",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonInvalid),
			},
		},
		Listeners: tests.GetListenerStatusV1(ports, []int32{0, 0}, true, true),
	}

	expectedStatus.Listeners[0].Conditions[0].Reason = string(gatewayv1.ListenerReasonInvalid)
	expectedStatus.Listeners[0].Conditions[0].Status = metav1.ConditionFalse
	expectedStatus.Listeners[0].Conditions[0].Message = "Listener is Invalid"
	expectedStatus.Listeners[0].Conditions[0].Type = string(gatewayv1.ListenerConditionAccepted)

	expectedStatus.Listeners[0].Conditions[1].Reason = string(gatewayv1.ListenerReasonInvalidRouteKinds)
	expectedStatus.Listeners[0].Conditions[1].Status = metav1.ConditionFalse
	expectedStatus.Listeners[0].Conditions[1].Message = "AllowedRoute Group is invalid."
	expectedStatus.Listeners[0].Conditions[1].Type = string(gatewayv1.ListenerConditionResolvedRefs)

	expectedStatus.Listeners[0].Conditions[2].Message = "Virtual service not configured/updated for this listener"
	expectedStatus.Listeners[0].Conditions[2].Reason = string(gatewayv1.ListenerReasonInvalid)
	expectedStatus.Listeners[0].Conditions[2].Status = metav1.ConditionFalse

	gateway, err = tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	if err != nil || gateway == nil {
		t.Fatalf("Couldn't get the gateway, err: %+v", err)
	}

	tests.ValidateGatewayStatus(t, &gateway.Status, expectedStatus)

	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
}

func TestGatewayWithValidAllowedRoute(t *testing.T) {

	gatewayName := "gateway-07"
	gatewayClassName := "gateway-class-07"
	ports := []int32{8080}

	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	// Checking for Valid RouteKind -> Kind and Group as nil
	listeners := tests.GetListenersV1(ports, false, false)
	allowedRoutes := gatewayv1.AllowedRoutes{
		Kinds: []gatewayv1.RouteGroupKind{{
			Kind: "HTTPRoute",
		},
		},
	}
	listeners[0].AllowedRoutes = &allowedRoutes
	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	expectedStatus := &gatewayv1.GatewayStatus{
		Conditions: []metav1.Condition{
			{
				Type:               string(gatewayv1.GatewayConditionAccepted),
				Status:             metav1.ConditionTrue,
				Message:            "Gateway configuration is valid",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonAccepted),
			},
		},
		Listeners: tests.GetListenerStatusV1(ports, []int32{0, 0}, true, false),
	}
	expectedStatus.Listeners[0].Conditions[0].Reason = string(gatewayv1.ListenerConditionAccepted)
	expectedStatus.Listeners[0].Conditions[0].Status = metav1.ConditionTrue
	expectedStatus.Listeners[0].Conditions[0].Message = "Listener is valid"

	expectedStatus.Listeners[0].Conditions[1].Reason = string(gatewayv1.ListenerConditionResolvedRefs)
	expectedStatus.Listeners[0].Conditions[1].Status = metav1.ConditionTrue
	expectedStatus.Listeners[0].Conditions[1].Message = "All the references are valid"

	gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	if err != nil || gateway == nil {
		t.Fatalf("Couldn't get the gateway, err: %+v", err)
	}

	tests.ValidateGatewayStatus(t, &gateway.Status, expectedStatus)
	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)

	// Checking for Without RouteKinds-> Kind  and Valid RouteKinds-> Group
	allowedRoutes = gatewayv1.AllowedRoutes{
		Kinds: []gatewayv1.RouteGroupKind{{
			Group: (*gatewayv1.Group)(&gatewayv1.GroupVersion.Group),
		},
		},
	}

	listeners[0].AllowedRoutes = &allowedRoutes
	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	expectedStatus = &gatewayv1.GatewayStatus{
		Conditions: []metav1.Condition{
			{
				Type:               string(gatewayv1.GatewayConditionAccepted),
				Status:             metav1.ConditionTrue,
				Message:            "Gateway configuration is valid",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonAccepted),
			},
		},
		Listeners: tests.GetListenerStatusV1(ports, []int32{0, 0}, true, false),
	}
	expectedStatus.Listeners[0].Conditions[0].Reason = string(gatewayv1.ListenerReasonAccepted)
	expectedStatus.Listeners[0].Conditions[0].Status = metav1.ConditionTrue
	expectedStatus.Listeners[0].Conditions[0].Message = "Listener is valid"

	gateway, err = tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	if err != nil || gateway == nil {
		t.Fatalf("Couldn't get the gateway, err: %+v", err)
	}

	tests.ValidateGatewayStatus(t, &gateway.Status, expectedStatus)
	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)

	//Checking for Without RouteKinds-> Kind  and empty RouteKinds-> Group
	emptyGroupKind := ""
	allowedRoutes.Kinds[0].Group = (*gatewayv1.Group)(&emptyGroupKind)
	listeners[0].AllowedRoutes = &allowedRoutes
	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	expectedStatus = &gatewayv1.GatewayStatus{
		Conditions: []metav1.Condition{
			{
				Type:               string(gatewayv1.GatewayConditionAccepted),
				Status:             metav1.ConditionTrue,
				Message:            "Gateway configuration is valid",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonAccepted),
			},
		},
		Listeners: tests.GetListenerStatusV1(ports, []int32{0, 0}, true, false),
	}
	expectedStatus.Listeners[0].Conditions[0].Reason = string(gatewayv1.ListenerReasonAccepted)
	expectedStatus.Listeners[0].Conditions[0].Status = metav1.ConditionTrue
	expectedStatus.Listeners[0].Conditions[0].Message = "Listener is valid"

	gateway, err = tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	if err != nil || gateway == nil {
		t.Fatalf("Couldn't get the gateway, err: %+v", err)
	}

	tests.ValidateGatewayStatus(t, &gateway.Status, expectedStatus)
	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)

	//Checking with populated RouteKinds-> Kind  and RouteKinds-> Group
	allowedRoutes.Kinds[0].Group = (*gatewayv1.Group)(&emptyGroupKind)
	allowedRoutes.Kinds[0].Kind = "HTTPRoute"
	listeners[0].AllowedRoutes = &allowedRoutes
	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	expectedStatus = &gatewayv1.GatewayStatus{
		Conditions: []metav1.Condition{
			{
				Type:               string(gatewayv1.GatewayConditionAccepted),
				Status:             metav1.ConditionTrue,
				Message:            "Gateway configuration is valid",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonAccepted),
			},
			{
				Type:               string(gatewayv1.GatewayConditionProgrammed),
				Status:             metav1.ConditionTrue,
				Message:            "Gateway configuration is valid",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonProgrammed),
			},
		},
		Listeners: tests.GetListenerStatusV1(ports, []int32{0, 0}, true, true),
	}
	expectedStatus.Listeners[0].Conditions[0].Reason = string(gatewayv1.ListenerReasonAccepted)
	expectedStatus.Listeners[0].Conditions[0].Status = metav1.ConditionTrue
	expectedStatus.Listeners[0].Conditions[0].Message = "Listener is valid"

	gateway, err = tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	if err != nil || gateway == nil {
		t.Fatalf("Couldn't get the gateway, err: %+v", err)
	}

	tests.ValidateGatewayStatus(t, &gateway.Status, expectedStatus)
	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
}

func TestMultipleGatewaySameHostname(t *testing.T) {
	gatewayName1 := "gateway-neg-08"
	gatewayClassName := "gateway-class-neg-08"
	ports := []int32{8080}
	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := tests.GetListenersV1(ports, false, false)
	hostname := "products.example.com"
	listeners[0].Hostname = (*gatewayv1.Hostname)(&hostname)
	tests.SetupGateway(t, gatewayName1, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName1, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	expectedStatus := &gatewayv1.GatewayStatus{
		Conditions: []metav1.Condition{
			{
				Type:               string(gatewayv1.GatewayConditionAccepted),
				Status:             metav1.ConditionTrue,
				Message:            "Gateway configuration is valid",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonAccepted),
			},
		},
		Listeners: tests.GetListenerStatusV1(ports, []int32{0, 0}, true, false),
	}
	expectedStatus.Listeners[0].Conditions[0].Reason = string(gatewayv1.ListenerReasonAccepted)
	expectedStatus.Listeners[0].Conditions[0].Status = metav1.ConditionTrue
	expectedStatus.Listeners[0].Conditions[0].Message = "Listener is valid"

	expectedStatus.Listeners[0].Conditions[1].Reason = string(gatewayv1.ListenerReasonResolvedRefs)
	expectedStatus.Listeners[0].Conditions[1].Status = metav1.ConditionTrue
	expectedStatus.Listeners[0].Conditions[1].Message = "All the references are valid"

	gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName1, metav1.GetOptions{})
	if err != nil || gateway == nil {
		t.Fatalf("Couldn't get the gateway, err: %+v", err)
	}

	tests.ValidateGatewayStatus(t, &gateway.Status, expectedStatus)

	gatewayName2 := "gateway-neg-09"
	tests.SetupGateway(t, gatewayName2, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName2, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	expectedStatus = &gatewayv1.GatewayStatus{
		Conditions: []metav1.Condition{
			{
				Type:               string(gatewayv1.GatewayConditionAccepted),
				Status:             metav1.ConditionFalse,
				Message:            "Gateway does not contain any valid listener",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonListenersNotValid),
			},
			{
				Type:               string(gatewayv1.GatewayConditionProgrammed),
				Status:             metav1.ConditionFalse,
				Message:            "Gateway not programmed",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonInvalid),
			},
		},
		Listeners: tests.GetListenerStatusV1(ports, []int32{0, 0}, false, true),
	}
	expectedStatus.Listeners[0].Conditions[0].Reason = string(gatewayv1.ListenerReasonInvalid)
	expectedStatus.Listeners[0].Conditions[0].Status = metav1.ConditionFalse
	expectedStatus.Listeners[0].Conditions[0].Message = "Hostname overlaps or is same as an existing gateway hostname"

	expectedStatus.Listeners[0].Conditions[1].Reason = string(gatewayv1.ListenerReasonInvalid)
	expectedStatus.Listeners[0].Conditions[1].Status = metav1.ConditionFalse
	expectedStatus.Listeners[0].Conditions[1].Message = "Virtual service not configured/updated for this listener"
	expectedStatus.Listeners[0].Conditions[0].Message = "Hostname is same as an existing gateway hostname"

	gateway, err = tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName2, metav1.GetOptions{})
	if err != nil || gateway == nil {
		t.Fatalf("Couldn't get the gateway, err: %+v", err)
	}

	tests.ValidateGatewayStatus(t, &gateway.Status, expectedStatus)

	tests.TeardownGateway(t, gatewayName2, DEFAULT_NAMESPACE)
	tests.TeardownGateway(t, gatewayName1, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
}

func TestMultipleGatewayOverlappingHostname(t *testing.T) {
	gatewayName1 := "gateway-neg-10"
	gatewayClassName := "gateway-class-neg-10"
	ports := []int32{8080}
	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := tests.GetListenersV1(ports, false, false)
	hostname := "products.example.com"
	listeners[0].Hostname = (*gatewayv1.Hostname)(&hostname)
	tests.SetupGateway(t, gatewayName1, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName1, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	expectedStatus := &gatewayv1.GatewayStatus{
		Conditions: []metav1.Condition{
			{
				Type:               string(gatewayv1.GatewayConditionAccepted),
				Status:             metav1.ConditionTrue,
				Message:            "Gateway configuration is valid",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonAccepted),
			},
		},
		Listeners: tests.GetListenerStatusV1(ports, []int32{0, 0}, true, true),
	}
	expectedStatus.Listeners[0].Conditions[0].Reason = string(gatewayv1.GatewayReasonAccepted)
	expectedStatus.Listeners[0].Conditions[0].Status = metav1.ConditionTrue
	expectedStatus.Listeners[0].Conditions[0].Message = "Listener is valid"

	gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName1, metav1.GetOptions{})
	if err != nil || gateway == nil {
		t.Fatalf("Couldn't get the gateway, err: %+v", err)
	}

	tests.ValidateGatewayStatus(t, &gateway.Status, expectedStatus)

	gatewayName2 := "gateway-neg-11"
	wildCardHostname := "*.example.com"
	listeners[0].Hostname = (*gatewayv1.Hostname)(&wildCardHostname)
	tests.SetupGateway(t, gatewayName2, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName2, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	expectedStatus = &gatewayv1.GatewayStatus{
		Conditions: []metav1.Condition{
			{
				Type:               string(gatewayv1.GatewayConditionAccepted),
				Status:             metav1.ConditionTrue,
				Message:            "Gateway configuration is valid",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonAccepted),
			},
			{
				Type:               string(gatewayv1.GatewayConditionProgrammed),
				Status:             metav1.ConditionFalse,
				Message:            "Gateway not programmed",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonInvalid),
			},
		},
		Listeners: tests.GetListenerStatusV1(ports, []int32{0, 0}, false, true),
	}
	expectedStatus.Listeners[0].Conditions[0].Reason = string(gatewayv1.GatewayReasonAccepted)
	expectedStatus.Listeners[0].Conditions[0].Status = metav1.ConditionTrue
	expectedStatus.Listeners[0].Conditions[0].Message = "Listener is valid"

	gateway, err = tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName2, metav1.GetOptions{})
	if err != nil || gateway == nil {
		t.Fatalf("Couldn't get the gateway, err: %+v", err)
	}

	tests.ValidateGatewayStatus(t, &gateway.Status, expectedStatus)

	tests.TeardownGateway(t, gatewayName2, DEFAULT_NAMESPACE)
	tests.TeardownGateway(t, gatewayName1, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
}

func TestGatewayWithUnsupportedProtocolAndHostnameInListeners(t *testing.T) {

	gatewayName := "gateway-neg-11"
	gatewayClassName := "gateway-class-neg-11"
	ports := []int32{8080, 8081}

	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := tests.GetListenersV1(ports, false, false)
	listeners[0].Protocol = "GRPC"
	hostName := "invalid"
	listeners[0].Hostname = (*gatewayv1.Hostname)(&hostName)
	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	expectedStatus := &gatewayv1.GatewayStatus{
		Conditions: []metav1.Condition{
			{
				Type:               string(gatewayv1.GatewayConditionAccepted),
				Status:             metav1.ConditionTrue,
				Message:            "Gateway contains atleast one valid listener",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayReasonListenersNotValid),
			},
		},
		Listeners: tests.GetListenerStatusV1(ports, []int32{0, 0}, true, false),
	}
	expectedStatus.Listeners[0].Conditions[0].Reason = string(gatewayv1.ListenerReasonUnsupportedProtocol)
	expectedStatus.Listeners[0].Conditions[0].Status = metav1.ConditionFalse
	expectedStatus.Listeners[0].Conditions[0].Message = "Unsupported protocol"

	expectedStatus.Listeners[1].Conditions[0].Reason = string(gatewayv1.ListenerReasonAccepted)
	expectedStatus.Listeners[1].Conditions[0].Status = metav1.ConditionTrue
	expectedStatus.Listeners[1].Conditions[0].Message = "Listener is valid"

	gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	if err != nil || gateway == nil {
		t.Fatalf("Couldn't get the gateway, err: %+v", err)
	}

	tests.ValidateGatewayStatus(t, &gateway.Status, expectedStatus)
	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
}
func TestSecretCreateDelete(t *testing.T) {

	gatewayName := "gateway-12"
	gatewayClassName := "gateway-class-12"
	ports := []int32{8080}
	secrets := []string{"secret-12"}

	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := tests.GetListenersV1(ports, false, false, secrets...)
	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	if err != nil || gateway == nil {
		t.Fatalf("Couldn't get the gateway, err: %+v", err)
	}
	expectedStatus := tests.GetNegativeConditions(ports)
	tests.ValidateGatewayStatus(t, &gateway.Status, expectedStatus)
	integrationtest.AddSecret(secrets[0], DEFAULT_NAMESPACE, "cert", "key")
	// add delay
	expectedStatus = tests.GetPositiveConditions(ports)
	time.Sleep(1 * time.Second)

	gateway, err = tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	if err != nil || gateway == nil {
		t.Fatalf("Couldn't get the gateway, err: %+v", err)
	}
	tests.ValidateGatewayStatus(t, &gateway.Status, expectedStatus)
	// delete
	integrationtest.DeleteSecret(secrets[0], DEFAULT_NAMESPACE)

	expectedStatus = tests.GetNegativeConditions(ports)

	// add delay
	time.Sleep(1 * time.Second)
	gateway, err = tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	if err != nil || gateway == nil {
		t.Fatalf("Couldn't get the gateway, err: %+v", err)
	}
	tests.ValidateGatewayStatus(t, &gateway.Status, expectedStatus)
	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
}
