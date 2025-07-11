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
	"os"
	"sync"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayfake "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned/fake"

	akogatewayapik8s "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/k8s"
	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	akogatewayapitests "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/gatewayapitests"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"
)

var keyChan chan string

const (
	DEFAULT_NAMESPACE = "default"
)

func waitAndverify(t *testing.T, key string) {
	waitChan := make(chan int)
	go func() {
		time.Sleep(20 * time.Second)
		waitChan <- 1
	}()

	select {
	case data := <-keyChan:
		if key == "" {
			t.Fatalf("unpexpected key: %v", data)
		} else if data != key {
			t.Fatalf("error in match expected: %v, got: %v", key, data)
		}
	case <-waitChan:
		if key != "" {
			t.Fatalf("timed out waiting for %v", key)
		}
	}

}

func syncFuncForTest(key interface{}, wg *sync.WaitGroup) error {
	keyStr, ok := key.(string)
	if !ok {
		return nil
	}
	keyChan <- keyStr
	return nil
}

func setupQueue(stopCh <-chan struct{}) {
	statusQueueParams := utils.WorkerQueue{NumWorkers: 1, WorkqueueName: utils.StatusQueue}
	ingestionQueueParams := utils.WorkerQueue{NumWorkers: 1, WorkqueueName: utils.ObjectIngestionLayer}
	statusQueue := utils.SharedWorkQueue(&ingestionQueueParams, &statusQueueParams).GetQueueByName(utils.StatusQueue)
	wgStatus := &sync.WaitGroup{}
	statusQueue.SyncFunc = akogatewayapik8s.SyncFromStatusQueue
	statusQueue.Run(stopCh, wgStatus)
	ingestionQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	wgIngestion := &sync.WaitGroup{}

	ingestionQueue.SyncFunc = syncFuncForTest
	ingestionQueue.Run(stopCh, wgIngestion)
}

func TestMain(m *testing.M) {
	testData := akogatewayapitests.GetL7RuleFakeData()
	akogatewayapitests.KubeClient = k8sfake.NewSimpleClientset()
	akogatewayapitests.GatewayClient = gatewayfake.NewSimpleClientset()
	//akogatewayapitests.DynamicClient = dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())
	akogatewayapitests.DynamicClient = dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), akogatewayapitests.GvrToKind, &testData)

	integrationtest.KubeClient = akogatewayapitests.KubeClient

	os.Setenv("CLUSTER_NAME", "cluster")
	os.Setenv("CLOUD_NAME", "CLOUD_VCENTER")
	os.Setenv("SEG_NAME", "Default-Group")
	os.Setenv("POD_NAMESPACE", utils.AKO_DEFAULT_NS)
	os.Setenv("POD_NAME", "ako-0")

	// Set the user with prefix
	_ = lib.AKOControlConfig()
	lib.SetAKOUser(akogatewayapilib.Prefix)
	lib.SetNamePrefix(akogatewayapilib.Prefix)
	akoControlConfig := akogatewayapilib.AKOControlConfig()
	akoControlConfig.SetEventRecorder(lib.AKOGatewayEventComponent, akogatewayapitests.KubeClient, true)
	akogatewayapilib.SetDynamicClientSet(akogatewayapitests.DynamicClient)
	akogatewayapilib.NewDynamicInformers(akogatewayapitests.DynamicClient, false)
	registeredInformers := []string{
		utils.ServiceInformer,
		utils.SecretInformer,
		utils.NSInformer,
	}

	registeredInformers = append(registeredInformers, utils.EndpointSlicesInformer)

	args := make(map[string]interface{})
	utils.NewInformers(utils.KubeClientIntf{ClientSet: akogatewayapitests.KubeClient}, registeredInformers, args)
	akoApi := integrationtest.InitializeFakeAKOAPIServer()
	defer akoApi.ShutDown()

	defer integrationtest.AviFakeClientInstance.Close()
	ctrl := akogatewayapik8s.SharedGatewayController()
	ctrl.InitGatewayAPIInformers(akogatewayapitests.GatewayClient)
	akoControlConfig.SetGatewayAPIClientset(akogatewayapitests.GatewayClient)
	stopCh := utils.SetupSignalHandler()
	ctrl.Start(stopCh)
	keyChan = make(chan string)

	ctrl.DisableSync = false
	setupQueue(stopCh)
	ctrl.SetupEventHandlers(k8s.K8sinformers{Cs: akogatewayapitests.KubeClient, DynamicClient: akogatewayapitests.DynamicClient})
	numWorkers := uint32(1)
	ctrl.SetupGatewayApiEventHandlers(numWorkers)
	os.Exit(m.Run())
}

func TestGatewayCUD(t *testing.T) {
	gwName := "gw-example-00"
	gwClassName := "gw-class-example-00"
	gwKey := "Gateway/" + DEFAULT_NAMESPACE + "/" + gwName
	gateway := gatewayv1.Gateway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "gateway.networking.k8s.io/v1beta1",
			Kind:       "Gateway",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      gwName,
			Namespace: "default",
		},
		Spec:   gatewayv1.GatewaySpec{},
		Status: gatewayv1.GatewayStatus{},
	}
	akogatewayapitests.SetGatewayGatewayClass(&gateway, gwClassName)
	akogatewayapitests.AddGatewayListener(&gateway, "listener-example", 80, gatewayv1.HTTPProtocolType, false)
	akogatewayapitests.SetListenerHostname(&gateway.Spec.Listeners[0], "foo.example.com")

	//create
	gw, err := akogatewayapitests.GatewayClient.GatewayV1().Gateways("default").Create(context.TODO(), &gateway, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Couldn't create, err: %+v", err)
	}
	t.Logf("Created %+v", gw.Name)
	waitAndverify(t, gwKey)

	//update
	akogatewayapitests.SetGatewayGatewayClass(&gateway, "gw-class-new")
	gw, err = akogatewayapitests.GatewayClient.GatewayV1().Gateways("default").Update(context.TODO(), &gateway, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Couldn't update, err: %+v", err)
	}
	t.Logf("Updated %+v", gw.Name)
	waitAndverify(t, gwKey)

	//delete
	err = akogatewayapitests.GatewayClient.GatewayV1().Gateways("default").Delete(context.TODO(), gateway.Name, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't delete, err: %+v", err)
	}
	t.Logf("Deleted %+v", gw.Name)
	waitAndverify(t, gwKey)
}

func TestGatewayInvalidListenerCount(t *testing.T) {
	gwName := "gw-example-01"
	gwClassName := "gw-class-example-01"
	gwKey := "Gateway/" + DEFAULT_NAMESPACE + "/" + gwName
	gateway := gatewayv1.Gateway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "gateway.networking.k8s.io/v1beta1",
			Kind:       "Gateway",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      gwName,
			Namespace: "default",
		},
		Spec:   gatewayv1.GatewaySpec{},
		Status: gatewayv1.GatewayStatus{},
	}
	akogatewayapitests.SetGatewayGatewayClass(&gateway, gwClassName)

	//create
	gw, err := akogatewayapitests.GatewayClient.GatewayV1().Gateways("default").Create(context.TODO(), &gateway, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Couldn't create, err: %+v", err)
	}
	t.Logf("Created %+v", gw.Name)
	waitAndverify(t, "")

	//update
	akogatewayapitests.AddGatewayListener(&gateway, "listener-example", 80, gatewayv1.HTTPProtocolType, false)
	akogatewayapitests.SetListenerHostname(&gateway.Spec.Listeners[0], "*.example.com")
	gw, err = akogatewayapitests.GatewayClient.GatewayV1().Gateways("default").Update(context.TODO(), &gateway, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Couldn't update, err: %+v", err)
	}
	t.Logf("Updated %+v", gw.Name)
	waitAndverify(t, gwKey)

	akogatewayapitests.TeardownGateway(t, gwName, DEFAULT_NAMESPACE)
	waitAndverify(t, gwKey)

}

func TestGatewayInvalidAddress(t *testing.T) {
	gwName := "gw-example-02"
	gwClassName := "gw-class-example-02"
	gwKey := "Gateway/" + DEFAULT_NAMESPACE + "/" + gwName
	gateway := gatewayv1.Gateway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "gateway.networking.k8s.io/v1beta1",
			Kind:       "Gateway",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      gwName,
			Namespace: "default",
		},
		Spec:   gatewayv1.GatewaySpec{},
		Status: gatewayv1.GatewayStatus{},
	}
	akogatewayapitests.SetGatewayGatewayClass(&gateway, gwClassName)
	akogatewayapitests.AddGatewayListener(&gateway, "listener-example", 80, gatewayv1.HTTPProtocolType, false)
	akogatewayapitests.SetListenerHostname(&gateway.Spec.Listeners[0], "foo.example.com")
	hostnameType := gatewayv1.AddressType("Hostname")
	gateway.Spec.Addresses = []gatewayv1.GatewaySpecAddress{
		{
			Type:  &hostnameType,
			Value: "some.fqdn.address",
		},
	}

	//create
	gw, err := akogatewayapitests.GatewayClient.GatewayV1().Gateways("default").Create(context.TODO(), &gateway, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Couldn't create, err: %+v", err)
	}
	t.Logf("Created %+v", gw.Name)
	waitAndverify(t, "")

	//update with IPv6
	ipAddressType := gatewayv1.AddressType("IPAddress")
	gateway.Spec.Addresses = []gatewayv1.GatewaySpecAddress{
		{
			Type: &ipAddressType,
			//TODO replace with constant from utils
			Value: "2001:db8:3333:4444:5555:6666:7777:8888",
		},
	}
	gw, err = akogatewayapitests.GatewayClient.GatewayV1().Gateways("default").Update(context.TODO(), &gateway, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Couldn't update, err: %+v", err)
	}
	t.Logf("Updated %+v", gw.Name)
	waitAndverify(t, "")

	//update with IPv4
	gateway.Spec.Addresses = []gatewayv1.GatewaySpecAddress{
		{
			Type: &ipAddressType,
			//TODO replace with constant from utils
			Value: "1.2.3.4",
		},
	}
	gw, err = akogatewayapitests.GatewayClient.GatewayV1().Gateways("default").Update(context.TODO(), &gateway, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Couldn't update, err: %+v", err)
	}
	t.Logf("Updated %+v", gw.Name)
	waitAndverify(t, gwKey)

	//delete
	akogatewayapitests.TeardownGateway(t, gwName, DEFAULT_NAMESPACE)
	waitAndverify(t, gwKey)
}

func TestGatewayWildcardHostname(t *testing.T) {
	gwName := "gw-example-03"
	gwClassName := "gw-class-example-03"
	gwKey := "Gateway/" + DEFAULT_NAMESPACE + "/" + gwName
	gateway := gatewayv1.Gateway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "gateway.networking.k8s.io/v1beta1",
			Kind:       "Gateway",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      gwName,
			Namespace: "default",
		},
		Spec:   gatewayv1.GatewaySpec{},
		Status: gatewayv1.GatewayStatus{},
	}
	akogatewayapitests.SetGatewayGatewayClass(&gateway, gwClassName)
	akogatewayapitests.AddGatewayListener(&gateway, "listener-example", 80, gatewayv1.HTTPProtocolType, false)
	akogatewayapitests.SetListenerHostname(&gateway.Spec.Listeners[0], "*")

	//create
	gw, err := akogatewayapitests.GatewayClient.GatewayV1().Gateways("default").Create(context.TODO(), &gateway, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Couldn't create, err: %+v", err)
	}
	t.Logf("Created %+v", gw.Name)
	waitAndverify(t, gwKey)

	//update with empty
	akogatewayapitests.SetListenerHostname(&gateway.Spec.Listeners[0], "")
	gw, err = akogatewayapitests.GatewayClient.GatewayV1().Gateways("default").Update(context.TODO(), &gateway, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Couldn't update, err: %+v", err)
	}
	t.Logf("Updated %+v", gw.Name)
	waitAndverify(t, gwKey)

	//update with wildcard fqdn
	akogatewayapitests.SetListenerHostname(&gateway.Spec.Listeners[0], "*.example.com")
	gw, err = akogatewayapitests.GatewayClient.GatewayV1().Gateways("default").Update(context.TODO(), &gateway, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Couldn't update, err: %+v", err)
	}
	t.Logf("Updated %+v", gw.Name)
	waitAndverify(t, gwKey)

	//delete
	akogatewayapitests.TeardownGateway(t, gwName, DEFAULT_NAMESPACE)
	waitAndverify(t, gwKey)
}

func TestGatewayInvalidListenerProtocol(t *testing.T) {
	gwName := "gw-example-04"
	gwClassName := "gw-class-example-04"
	gwKey := "Gateway/" + DEFAULT_NAMESPACE + "/" + gwName
	gateway := gatewayv1.Gateway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "gateway.networking.k8s.io/v1beta1",
			Kind:       "Gateway",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      gwName,
			Namespace: "default",
		},
		Spec:   gatewayv1.GatewaySpec{},
		Status: gatewayv1.GatewayStatus{},
	}
	akogatewayapitests.SetGatewayGatewayClass(&gateway, gwClassName)
	akogatewayapitests.AddGatewayListener(&gateway, "listener-example", 80, gatewayv1.TCPProtocolType, false)
	akogatewayapitests.SetListenerHostname(&gateway.Spec.Listeners[0], "*.example.com")

	//create
	gw, err := akogatewayapitests.GatewayClient.GatewayV1().Gateways("default").Create(context.TODO(), &gateway, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Couldn't create, err: %+v", err)
	}
	t.Logf("Created %+v", gw.Name)
	waitAndverify(t, "")

	//update
	gateway.Spec.Listeners[0].Protocol = gatewayv1.HTTPProtocolType
	gw, err = akogatewayapitests.GatewayClient.GatewayV1().Gateways("default").Update(context.TODO(), &gateway, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Couldn't update, err: %+v", err)
	}
	t.Logf("Updated %+v", gw.Name)
	waitAndverify(t, gwKey)

	//delete
	akogatewayapitests.TeardownGateway(t, gwName, DEFAULT_NAMESPACE)
	waitAndverify(t, gwKey)
}

func TestGatewayInvalidListenerTLS(t *testing.T) {
	gwName := "gw-example-04"
	gwClassName := "gw-class-example-04"
	gwKey := "Gateway/" + DEFAULT_NAMESPACE + "/" + gwName

	ports := []int32{8080}
	secrets := []string{"secret-01"}
	for _, secret := range secrets {
		integrationtest.AddSecret(secret, DEFAULT_NAMESPACE, "cert", "key")
		waitAndverify(t, "Secret/"+DEFAULT_NAMESPACE+"/"+secret)
	}

	listeners := akogatewayapitests.GetListenersV1(ports, false, false, secrets...)
	tlsModePassthrough := gatewayv1.TLSModePassthrough
	listeners[0].TLS.Mode = &tlsModePassthrough
	//create
	akogatewayapitests.SetupGateway(t, gwName, DEFAULT_NAMESPACE, gwClassName, nil, listeners)
	waitAndverify(t, "")
	//update
	gateway, _ := akogatewayapitests.GatewayClient.GatewayV1().Gateways("default").Get(context.TODO(), gwName, metav1.GetOptions{})
	tlsModeTerminate := gatewayv1.TLSModeTerminate
	gateway.Spec.Listeners[0].TLS.Mode = &tlsModeTerminate
	gw, err := akogatewayapitests.GatewayClient.GatewayV1().Gateways("default").Update(context.TODO(), gateway, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Couldn't update, err: %+v", err)
	}
	t.Logf("Updated %+v", gw.Name)
	waitAndverify(t, gwKey)

	//delete
	akogatewayapitests.TeardownGateway(t, gwName, DEFAULT_NAMESPACE)
	waitAndverify(t, gwKey)
}

func TestMultipleGatewaySameHostname(t *testing.T) {
	//create first gateway
	gwName1 := "gw-example-05"
	gwClassName := "gw-class-example-05"
	gwKey1 := "Gateway/" + DEFAULT_NAMESPACE + "/" + gwName1
	gateway1 := gatewayv1.Gateway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "gateway.networking.k8s.io/v1beta1",
			Kind:       "Gateway",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      gwName1,
			Namespace: "default",
		},
		Spec:   gatewayv1.GatewaySpec{},
		Status: gatewayv1.GatewayStatus{},
	}
	akogatewayapitests.SetGatewayGatewayClass(&gateway1, gwClassName)
	akogatewayapitests.AddGatewayListener(&gateway1, "listener-example", 80, gatewayv1.HTTPProtocolType, false)
	akogatewayapitests.SetListenerHostname(&gateway1.Spec.Listeners[0], "*.example.com")

	gw, err := akogatewayapitests.GatewayClient.GatewayV1().Gateways("default").Create(context.TODO(), &gateway1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Couldn't create, err: %+v", err)
	}
	t.Logf("Created %+v", gw.Name)
	waitAndverify(t, gwKey1)

	gwName2 := "gw-example-06"
	gwKey2 := "Gateway/" + DEFAULT_NAMESPACE + "/" + gwName2
	gateway2 := gatewayv1.Gateway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "gateway.networking.k8s.io/v1beta1",
			Kind:       "Gateway",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      gwName2,
			Namespace: "default",
		},
		Spec:   gatewayv1.GatewaySpec{},
		Status: gatewayv1.GatewayStatus{},
	}
	akogatewayapitests.SetGatewayGatewayClass(&gateway2, gwClassName)
	akogatewayapitests.AddGatewayListener(&gateway2, "listener-example", 80, gatewayv1.HTTPProtocolType, false)
	akogatewayapitests.SetListenerHostname(&gateway2.Spec.Listeners[0], "*.example.com")

	//create second gateway
	gw, err = akogatewayapitests.GatewayClient.GatewayV1().Gateways("default").Create(context.TODO(), &gateway2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Couldn't create, err: %+v", err)
	}
	t.Logf("Created %+v", gw.Name)
	waitAndverify(t, "")

	//delete
	akogatewayapitests.TeardownGateway(t, gwName1, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gwName2, DEFAULT_NAMESPACE)
	waitAndverify(t, gwKey1)
	waitAndverify(t, gwKey2)
}

func TestMultipleGatewayOverlappingHostname(t *testing.T) {
	//create first gateway
	gwName1 := "gw-example-07"
	gwClassName := "gw-class-example-07"
	gwKey1 := "Gateway/" + DEFAULT_NAMESPACE + "/" + gwName1
	gateway1 := gatewayv1.Gateway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "gateway.networking.k8s.io/v1beta1",
			Kind:       "Gateway",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      gwName1,
			Namespace: "default",
		},
		Spec:   gatewayv1.GatewaySpec{},
		Status: gatewayv1.GatewayStatus{},
	}
	akogatewayapitests.SetGatewayGatewayClass(&gateway1, gwClassName)
	akogatewayapitests.AddGatewayListener(&gateway1, "listener-example", 80, gatewayv1.HTTPProtocolType, false)
	akogatewayapitests.SetListenerHostname(&gateway1.Spec.Listeners[0], "*.example.com")

	gw, err := akogatewayapitests.GatewayClient.GatewayV1().Gateways("default").Create(context.TODO(), &gateway1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Couldn't create, err: %+v", err)
	}
	t.Logf("Created %+v", gw.Name)
	waitAndverify(t, gwKey1)

	gwName2 := "gw-example-08"
	gwKey2 := "Gateway/" + DEFAULT_NAMESPACE + "/" + gwName2
	gateway2 := gatewayv1.Gateway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "gateway.networking.k8s.io/v1beta1",
			Kind:       "Gateway",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      gwName2,
			Namespace: "default",
		},
		Spec:   gatewayv1.GatewaySpec{},
		Status: gatewayv1.GatewayStatus{},
	}
	akogatewayapitests.SetGatewayGatewayClass(&gateway2, gwClassName)
	akogatewayapitests.AddGatewayListener(&gateway2, "listener-example", 80, gatewayv1.HTTPProtocolType, false)
	akogatewayapitests.SetListenerHostname(&gateway2.Spec.Listeners[0], "products.example.com")

	//create second gateway
	gw, err = akogatewayapitests.GatewayClient.GatewayV1().Gateways("default").Create(context.TODO(), &gateway2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Couldn't create, err: %+v", err)
	}
	t.Logf("Created %+v", gw.Name)
	waitAndverify(t, gwKey2)

	//delete
	akogatewayapitests.TeardownGateway(t, gwName1, DEFAULT_NAMESPACE)
	akogatewayapitests.TeardownGateway(t, gwName2, DEFAULT_NAMESPACE)
	waitAndverify(t, gwKey1)
	waitAndverify(t, gwKey2)
}

func TestGatewayClassCUD(t *testing.T) {
	gatewayClass := gatewayv1.GatewayClass{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "gateway.networking.k8s.io/v1beta1",
			Kind:       "GatewayClass",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "gw-class-example",
		},
		Spec: gatewayv1.GatewayClassSpec{
			ControllerName: "ako.vmware.com/avi-lb",
		},
		Status: gatewayv1.GatewayClassStatus{},
	}

	//create
	gw, err := akogatewayapitests.GatewayClient.GatewayV1().GatewayClasses().Create(context.TODO(), &gatewayClass, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Couldn't create, err: %+v", err)
	}
	t.Logf("Created %+v", gw.Name)
	waitAndverify(t, "GatewayClass/gw-class-example")

	//update
	testDesc := "test description for update"
	gatewayClass.Spec.Description = &testDesc
	gw, err = akogatewayapitests.GatewayClient.GatewayV1().GatewayClasses().Update(context.TODO(), &gatewayClass, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Couldn't update gatewayClass, err: %+v", err)
	}
	t.Logf("Updated %+v", gw.Name)
	waitAndverify(t, "GatewayClass/gw-class-example")

	//delete
	err = akogatewayapitests.GatewayClient.GatewayV1().GatewayClasses().Delete(context.TODO(), gatewayClass.Name, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't delete, err: %+v", err)
	}
	t.Logf("Deleted %+v", gw.Name)
	waitAndverify(t, "GatewayClass/gw-class-example")
}

func TestGatewayWithInvalidAllowedRoute(t *testing.T) {
	gwName := "gw-example-03"
	gwClassName := "gw-class-example-03"
	gwKey := "Gateway/" + DEFAULT_NAMESPACE + "/" + gwName
	gateway := gatewayv1.Gateway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "gateway.networking.k8s.io/v1",
			Kind:       "Gateway",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      gwName,
			Namespace: "default",
		},
		Spec:   gatewayv1.GatewaySpec{},
		Status: gatewayv1.GatewayStatus{},
	}
	akogatewayapitests.AddGatewayListener(&gateway, "listener-example", 80, gatewayv1.HTTPProtocolType, false)
	akogatewayapitests.SetListenerHostname(&gateway.Spec.Listeners[0], "foo.example.com")

	// Checking for Invalid RouteKind -> Kind
	allowedRoutes := gatewayv1.AllowedRoutes{
		Kinds: []gatewayv1.RouteGroupKind{{
			Kind: "Services",
		},
		},
	}
	gateway.Spec.Listeners[0].AllowedRoutes = &allowedRoutes
	akogatewayapitests.SetGatewayGatewayClass(&gateway, gwClassName)

	//create
	gw, err := akogatewayapitests.GatewayClient.GatewayV1().Gateways("default").Create(context.TODO(), &gateway, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Couldn't create, err: %+v", err)
	}
	t.Logf("Created %+v", gw.Name)
	waitAndverify(t, "")

	akogatewayapitests.TeardownGateway(t, gwName, DEFAULT_NAMESPACE)
	waitAndverify(t, gwKey)

	// Checking for Invalid RouteKind -> Group
	allowedRoutes.Kinds[0].Kind = "HTTPRoute"
	invalidGroup := "InvalidGroup.example.com"
	allowedRoutes.Kinds[0].Group = (*gatewayv1.Group)(&invalidGroup)
	gateway.Spec.Listeners[0].AllowedRoutes = &allowedRoutes
	akogatewayapitests.SetGatewayGatewayClass(&gateway, gwClassName)

	//create
	gw, err = akogatewayapitests.GatewayClient.GatewayV1().Gateways("default").Create(context.TODO(), &gateway, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Couldn't create, err: %+v", err)
	}
	t.Logf("Created %+v", gw.Name)
	waitAndverify(t, "")

	akogatewayapitests.TeardownGateway(t, gwName, DEFAULT_NAMESPACE)
	waitAndverify(t, gwKey)
}

func TestGatewayWithValidAllowedRoute(t *testing.T) {
	gwName := "gw-example-04"
	gwClassName := "gw-class-example-04"
	gwKey := "Gateway/" + DEFAULT_NAMESPACE + "/" + gwName
	gateway := gatewayv1.Gateway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "gateway.networking.k8s.io/v1",
			Kind:       "Gateway",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      gwName,
			Namespace: "default",
		},
		Spec:   gatewayv1.GatewaySpec{},
		Status: gatewayv1.GatewayStatus{},
	}
	akogatewayapitests.AddGatewayListener(&gateway, "listener-example", 80, gatewayv1.HTTPProtocolType, false)
	akogatewayapitests.SetListenerHostname(&gateway.Spec.Listeners[0], "foo.example.com")

	//Checking with populated RouteKinds-> Kind  and RouteKinds-> Group
	allowedRoutes := gatewayv1.AllowedRoutes{
		Kinds: []gatewayv1.RouteGroupKind{{
			Kind:  "HTTPRoute",
			Group: (*gatewayv1.Group)(&gatewayv1.GroupVersion.Group),
		},
		},
	}
	gateway.Spec.Listeners[0].AllowedRoutes = &allowedRoutes
	akogatewayapitests.SetGatewayGatewayClass(&gateway, gwClassName)

	//create
	gw, err := akogatewayapitests.GatewayClient.GatewayV1().Gateways("default").Create(context.TODO(), &gateway, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Couldn't create, err: %+v", err)
	}
	t.Logf("Created %+v", gw.Name)
	waitAndverify(t, gwKey)

	akogatewayapitests.TeardownGateway(t, gwName, DEFAULT_NAMESPACE)
	waitAndverify(t, gwKey)

	// Checking for Valid RouteKind -> Kind and Group as nil
	allowedRoutes = gatewayv1.AllowedRoutes{
		Kinds: []gatewayv1.RouteGroupKind{{
			Kind: "HTTPRoute",
		},
		},
	}
	gateway.Spec.Listeners[0].AllowedRoutes = &allowedRoutes
	akogatewayapitests.SetGatewayGatewayClass(&gateway, gwClassName)

	//create
	gw, err = akogatewayapitests.GatewayClient.GatewayV1().Gateways("default").Create(context.TODO(), &gateway, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Couldn't create, err: %+v", err)
	}
	t.Logf("Created %+v", gw.Name)
	waitAndverify(t, gwKey)

	akogatewayapitests.TeardownGateway(t, gwName, DEFAULT_NAMESPACE)
	waitAndverify(t, gwKey)

	// Checking for Without RouteKinds-> Kind  and Valid RouteKinds-> Group
	allowedRoutes = gatewayv1.AllowedRoutes{
		Kinds: []gatewayv1.RouteGroupKind{{
			Group: (*gatewayv1.Group)(&gatewayv1.GroupVersion.Group),
		},
		},
	}
	gateway.Spec.Listeners[0].AllowedRoutes = &allowedRoutes
	akogatewayapitests.SetGatewayGatewayClass(&gateway, gwClassName)

	//create
	gw, err = akogatewayapitests.GatewayClient.GatewayV1().Gateways("default").Create(context.TODO(), &gateway, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Couldn't create, err: %+v", err)
	}
	t.Logf("Created %+v", gw.Name)
	waitAndverify(t, gwKey)

	akogatewayapitests.TeardownGateway(t, gwName, DEFAULT_NAMESPACE)
	waitAndverify(t, gwKey)

	//Checking for Without RouteKinds-> Kind  and empty RouteKinds-> Group
	emptyGroupKind := ""
	allowedRoutes.Kinds[0].Group = (*gatewayv1.Group)(&emptyGroupKind)
	akogatewayapitests.SetGatewayGatewayClass(&gateway, gwClassName)

	//create
	gw, err = akogatewayapitests.GatewayClient.GatewayV1().Gateways("default").Create(context.TODO(), &gateway, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Couldn't create, err: %+v", err)
	}
	t.Logf("Created %+v", gw.Name)
	waitAndverify(t, gwKey)

	akogatewayapitests.TeardownGateway(t, gwName, DEFAULT_NAMESPACE)
	waitAndverify(t, gwKey)
}
