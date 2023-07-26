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
	"context"
	"os"
	"sync"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	gatewayfake "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned/fake"

	akogatewayapik8s "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/k8s"
	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	akogatewayapiobjects "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/objects"
	akogatewayapitests "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/tests"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"
)

var keyChan chan string

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
	ingestionQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	wgIngestion := &sync.WaitGroup{}

	ingestionQueue.SyncFunc = syncFuncForTest
	ingestionQueue.Run(stopCh, wgIngestion)
}

func TestMain(m *testing.M) {
	akogatewayapitests.KubeClient = k8sfake.NewSimpleClientset()
	akogatewayapitests.GatewayClient = gatewayfake.NewSimpleClientset()

	os.Setenv("CLUSTER_NAME", "cluster")
	os.Setenv("CLOUD_NAME", "CLOUD_VCENTER")
	os.Setenv("SEG_NAME", "Default-Group")
	os.Setenv("POD_NAMESPACE", utils.AKO_DEFAULT_NS)

	// Set the user with prefix
	_ = lib.AKOControlConfig()
	lib.SetAKOUser(akogatewayapilib.Prefix)
	lib.SetNamePrefix(akogatewayapilib.Prefix)
	akoControlConfig := akogatewayapilib.AKOControlConfig()
	akoControlConfig.SetEventRecorder(lib.AKOGatewayEventComponent, akogatewayapitests.KubeClient, true)
	registeredInformers := []string{
		utils.ServiceInformer,
		utils.EndpointInformer,
		utils.SecretInformer,
		utils.NSInformer,
	}
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

	ctrl.SetupEventHandlers(k8s.K8sinformers{Cs: akogatewayapitests.KubeClient})
	numWorkers := uint32(1)
	ctrl.SetupGatewayApiEventHandlers(numWorkers)
	setupQueue(stopCh)
	os.Exit(m.Run())
}

func TestGatewayCUD(t *testing.T) {
	gateway := gatewayv1beta1.Gateway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "gateway.networking.k8s.io/v1beta1",
			Kind:       "Gateway",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gw-example",
			Namespace: "default",
		},
		Spec:   gatewayv1beta1.GatewaySpec{},
		Status: gatewayv1beta1.GatewayStatus{},
	}
	akogatewayapitests.SetGatewayGatewayClass(&gateway, "gw-class-example")
	akogatewayapitests.AddGatewayListener(&gateway, "listener-example", 80, gatewayv1beta1.HTTPProtocolType, false)
	akogatewayapitests.SetListenerHostname(&gateway.Spec.Listeners[0], "foo.example.com")

	//create
	gw, err := akogatewayapitests.GatewayClient.GatewayV1beta1().Gateways("default").Create(context.TODO(), &gateway, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Couldn't create, err: %+v", err)
	}
	t.Logf("Created %+v", gw.Name)
	waitAndverify(t, "Gateway/default/gw-example")

	//update
	akogatewayapitests.SetGatewayGatewayClass(&gateway, "gw-class-new")
	gw, err = akogatewayapitests.GatewayClient.GatewayV1beta1().Gateways("default").Update(context.TODO(), &gateway, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Couldn't update, err: %+v", err)
	}
	t.Logf("Updated %+v", gw.Name)
	waitAndverify(t, "Gateway/default/gw-example")

	//delete
	err = akogatewayapitests.GatewayClient.GatewayV1beta1().Gateways("default").Delete(context.TODO(), gateway.Name, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't delete, err: %+v", err)
	}
	t.Logf("Deleted %+v", gw.Name)
	waitAndverify(t, "Gateway/default/gw-example")
}

func TestGatewayClassCUD(t *testing.T) {
	gatewayClass := gatewayv1beta1.GatewayClass{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "gateway.networking.k8s.io/v1beta1",
			Kind:       "GatewayClass",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "gw-class-example",
		},
		Spec: gatewayv1beta1.GatewayClassSpec{
			ControllerName: "ako.vmware.com/avi-lb",
		},
		Status: gatewayv1beta1.GatewayClassStatus{},
	}

	//create
	gw, err := akogatewayapitests.GatewayClient.GatewayV1beta1().GatewayClasses().Create(context.TODO(), &gatewayClass, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Couldn't create, err: %+v", err)
	}
	t.Logf("Created %+v", gw.Name)
	waitAndverify(t, "GatewayClass/gw-class-example")

	//update
	testDesc := "test description for update"
	gatewayClass.Spec.Description = &testDesc
	gw, err = akogatewayapitests.GatewayClient.GatewayV1beta1().GatewayClasses().Update(context.TODO(), &gatewayClass, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Couldn't update gatewayClass, err: %+v", err)
	}
	t.Logf("Updated %+v", gw.Name)
	waitAndverify(t, "GatewayClass/gw-class-example")

	//delete
	err = akogatewayapitests.GatewayClient.GatewayV1beta1().GatewayClasses().Delete(context.TODO(), gatewayClass.Name, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't delete, err: %+v", err)
	}
	t.Logf("Deleted %+v", gw.Name)
	waitAndverify(t, "GatewayClass/gw-class-example")
}

func TestHTTPRouteCUD(t *testing.T) {
	gatewayClassName := "gateway-class-01"
	gatewayName := "gateway-01"
	httpRouteName := "httproute-01"
	namespace := "default"
	ports := []int32{8080, 8081}
	key := "HTTPRoute" + "/" + namespace + "/" + httpRouteName
	akogatewayapiobjects.GatewayApiLister().UpdateGatewayClass(gatewayClassName, true)

	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	t.Logf("Created GatewayClass %s", gatewayClassName)
	waitAndverify(t, "GatewayClass/gateway-class-01")

	listeners := akogatewayapitests.GetListenersV1Beta1(ports)
	akogatewayapitests.SetupGateway(t, gatewayName, namespace, gatewayClassName, nil, listeners)
	t.Logf("Created GatewayClass %s", gatewayClassName)
	waitAndverify(t, "Gateway/default/gateway-01")

	parentRefs := akogatewayapitests.GetParentReferencesV1Beta1([]string{gatewayName}, namespace, ports)
	hostnames := []gatewayv1beta1.Hostname{"foo-8080.com", "foo-8081.com"}
	rules := akogatewayapitests.GetHTTPRouteRulesV1Beta1()
	akogatewayapitests.SetupHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, rules)
	waitAndverify(t, key)

	// update
	hostnames = []gatewayv1beta1.Hostname{"foo-8080.com"}
	akogatewayapitests.UpdateHTTPRoute(t, httpRouteName, namespace, parentRefs, hostnames, rules)
	waitAndverify(t, key)

	// delete
	akogatewayapitests.TeardownHTTPRoute(t, httpRouteName, namespace)
	waitAndverify(t, key)
}
