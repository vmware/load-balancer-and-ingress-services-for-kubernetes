package tests

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	gwk8s "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/k8s"
	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	gwfake "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned/fake"
)

var kubeClient *k8sfake.Clientset
var gwClient *gwfake.Clientset
var keyChan chan string
var ctrl *gwk8s.GatewayController

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
	case _ = <-waitChan:
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
	kubeClient = k8sfake.NewSimpleClientset()
	gwClient = gwfake.NewSimpleClientset()

	os.Setenv("CLUSTER_NAME", "cluster")
	os.Setenv("CLOUD_NAME", "CLOUD_VCENTER")
	os.Setenv("SEG_NAME", "Default-Group")
	os.Setenv("POD_NAMESPACE", utils.AKO_DEFAULT_NS)

	akoControlConfig := akogatewayapilib.AKOControlConfig()
	akoControlConfig.SetEventRecorder(lib.AKOGatewayEventComponent, kubeClient, true)
	registeredInformers := []string{
		utils.ServiceInformer,
		utils.EndpointInformer,
		utils.SecretInformer,
		utils.NSInformer,
		utils.ConfigMapInformer,
	}
	args := make(map[string]interface{})
	utils.NewInformers(utils.KubeClientIntf{ClientSet: kubeClient}, registeredInformers, args)
	integrationtest.InitializeFakeAKOAPIServer()

	defer integrationtest.AviFakeClientInstance.Close()
	ctrl = gwk8s.SharedGatewayController()
	ctrl.InitGatewayAPIInformers(gwClient)

	stopCh := utils.SetupSignalHandler()
	ctrl.Start(stopCh)
	keyChan = make(chan string)

	ctrl.DisableSync = false

	ctrl.SetupEventHandlers(k8s.K8sinformers{Cs: kubeClient})
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
	SetGatewayGatewayClass(&gateway, "gw-class-example")
	AddGatewayListener(&gateway, "listener-example", 80, gatewayv1beta1.HTTPProtocolType, false)
	SetListenerHostname(&gateway.Spec.Listeners[0], "foo.example.com")

	//create
	gw, err := gwClient.GatewayV1beta1().Gateways("default").Create(context.TODO(), &gateway, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Couldn't create, err: %+v", err)
	}
	t.Logf("Created %+v", gw.Name)
	waitAndverify(t, "Gateway/default/gw-example")

	//update
	SetGatewayGatewayClass(&gateway, "gw-class-new")
	gw, err = gwClient.GatewayV1beta1().Gateways("default").Update(context.TODO(), &gateway, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Couldn't update, err: %+v", err)
	}
	t.Logf("Updated %+v", gw.Name)
	waitAndverify(t, "Gateway/default/gw-example")

	//delete
	err = gwClient.GatewayV1beta1().Gateways("default").Delete(context.TODO(), gateway.Name, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't delete, err: %+v", err)
	}
	t.Logf("Deleted %+v", gw.Name)
	waitAndverify(t, "Gateway/default/gw-example")
}

func TestGatewaClassyCUD(t *testing.T) {
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
	gw, err := gwClient.GatewayV1beta1().GatewayClasses().Create(context.TODO(), &gatewayClass, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Couldn't create, err: %+v", err)
	}
	t.Logf("Created %+v", gw.Name)
	waitAndverify(t, "GatewayClass/gw-class-example")

	//update
	testDesc := "test description for update"
	gatewayClass.Spec.Description = &testDesc
	gw, err = gwClient.GatewayV1beta1().GatewayClasses().Update(context.TODO(), &gatewayClass, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Couldn't update gatewayClass, err: %+v", err)
	}
	t.Logf("Updated %+v", gw.Name)
	waitAndverify(t, "GatewayClass/gw-class-example")

	//delete
	err = gwClient.GatewayV1beta1().GatewayClasses().Delete(context.TODO(), gatewayClass.Name, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't delete, err: %+v", err)
	}
	t.Logf("Deleted %+v", gw.Name)
	waitAndverify(t, "GatewayClass/gw-class-example")
}
