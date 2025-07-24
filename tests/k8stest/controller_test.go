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
package k8stest

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha1/clientset/versioned/fake"
	v1beta1crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1beta1/clientset/versioned/fake"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	corev1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	// To Do: add test for openshift route
	//oshiftfake "github.com/openshift/client-go/route/clientset/versioned/fake"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

var kubeClient *k8sfake.Clientset
var crdClient *crdfake.Clientset
var v1beta1crdClient *v1beta1crdfake.Clientset
var dynamicClient *dynamicfake.FakeDynamicClient
var keyChan chan string
var ctrl *k8s.AviController

func syncFuncForTest(key interface{}, wg *sync.WaitGroup) error {
	keyStr, ok := key.(string)
	if !ok {
		return nil
	}
	keyChan <- keyStr
	return nil
}

// empty key ("") means we are not expecting the key
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

func setupQueue(stopCh <-chan struct{}) {
	ingestionQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	wgIngestion := &sync.WaitGroup{}

	ingestionQueue.SyncFunc = syncFuncForTest
	ingestionQueue.Run(stopCh, wgIngestion)
}

func TestMain(m *testing.M) {
	kubeClient = k8sfake.NewSimpleClientset()
	dynamicClient = dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())
	os.Setenv("VIP_NETWORK_LIST", `[{"networkName":"net123"}]`)
	os.Setenv("CLUSTER_NAME", "cluster")
	os.Setenv("CLOUD_NAME", "CLOUD_VCENTER")
	os.Setenv("SEG_NAME", "Default-Group")
	os.Setenv("NODE_NETWORK_LIST", `[{"networkName":"net123","cidrs":["10.79.168.0/22"]}]`)
	os.Setenv("POD_NAMESPACE", utils.AKO_DEFAULT_NS)
	os.Setenv("SHARD_VS_SIZE", "LARGE")
	os.Setenv("MCI_ENABLED", "true")
	os.Setenv("POD_NAME", "ako-0")

	data := map[string][]byte{
		"username": []byte("admin"),
		"password": []byte("admin"),
	}
	object := metav1.ObjectMeta{Name: "avi-secret", Namespace: utils.GetAKONamespace()}
	secret := &corev1.Secret{Data: data, ObjectMeta: object}
	kubeClient.CoreV1().Secrets(utils.GetAKONamespace()).Create(context.TODO(), secret, metav1.CreateOptions{})

	akoControlConfig := lib.AKOControlConfig()
	crdClient = crdfake.NewSimpleClientset()
	v1beta1crdClient = v1beta1crdfake.NewSimpleClientset()
	akoControlConfig.SetCRDClientset(crdClient)
	akoControlConfig.Setv1beta1CRDClientset(v1beta1crdClient)
	akoControlConfig.SetAKOInstanceFlag(true)
	akoControlConfig.SetEventRecorder(lib.AKOEventComponent, kubeClient, true)
	akoControlConfig.SetDefaultLBController(true)

	registeredInformers := []string{
		utils.ServiceInformer,
		utils.IngressInformer,
		utils.IngressClassInformer,
		utils.SecretInformer,
		utils.NSInformer,
		utils.NodeInformer,
		utils.ConfigMapInformer,
		utils.MultiClusterIngressInformer,
		utils.ServiceImportInformer,
	}

	registeredInformers = append(registeredInformers, utils.EndpointSlicesInformer)

	args := make(map[string]interface{})
	args[utils.INFORMERS_AKO_CLIENT] = crdClient
	utils.NewInformers(utils.KubeClientIntf{ClientSet: kubeClient}, registeredInformers, args)
	k8s.NewCRDInformers()
	integrationtest.InitializeFakeAKOAPIServer()

	integrationtest.NewAviFakeClientInstance(kubeClient)
	defer integrationtest.AviFakeClientInstance.Close()

	ctrl = k8s.SharedAviController()
	stopCh := utils.SetupSignalHandler()
	ctrl.Start(stopCh)
	keyChan = make(chan string)
	ctrlCh := make(chan struct{})
	quickSyncCh := make(chan struct{})

	integrationtest.AddConfigMap(kubeClient)
	integrationtest.PollForSyncStart(ctrl, 10)

	ctrl.HandleConfigMap(k8s.K8sinformers{Cs: kubeClient, DynamicClient: dynamicClient}, ctrlCh, stopCh, quickSyncCh)
	ctrl.SetupEventHandlers(k8s.K8sinformers{Cs: kubeClient, DynamicClient: dynamicClient})
	setupQueue(stopCh)
	os.Exit(m.Run())
}

func TestSvc(t *testing.T) {
	waitAndverify(t, fmt.Sprintf("Secret/%s/avi-secret", utils.GetAKONamespace()))
	svcExample := &corev1.Service{
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeLoadBalancer,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "red-ns",
			Name:      "testsvc",
		},
	}
	_, err := kubeClient.CoreV1().Services("red-ns").Create(context.TODO(), svcExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	waitAndverify(t, "L4LBService/red-ns/testsvc")
}

func TestEndpointSlice(t *testing.T) {
	epExample := &discovery.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "red-ns",
			Name:      "eps",
			Labels:    map[string]string{discovery.LabelServiceName: "testep"},
		},
		Endpoints: []discovery.Endpoint{},
	}
	_, err := kubeClient.DiscoveryV1().EndpointSlices("red-ns").Create(context.TODO(), epExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in creating Endpoint: %v", err)
	}
	waitAndverify(t, "Endpointslices/red-ns/testep")
}

func TestIngressClass(t *testing.T) {
	apiGroup := "ako.vmware.com"
	ingrClassExample := &networkingv1.IngressClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "avi-lb",
			Annotations: map[string]string{
				"ingressclass.kubernetes.io/is-default-class": "true",
			},
		},
		Spec: networkingv1.IngressClassSpec{
			Controller: "ako.vmware.com/avi-lb",
			Parameters: &networkingv1.IngressClassParametersReference{
				APIGroup: &apiGroup,
				Kind:     "IngressParameters",
				Name:     "external-lb",
			},
		},
	}
	_, err := kubeClient.NetworkingV1().IngressClasses().Create(context.TODO(), ingrClassExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding IngressClass: %v", err)
	}
	waitAndverify(t, "IngressClass/avi-lb")
}

func TestIngress(t *testing.T) {
	ingrExample := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "red-ns",
			Name:      "testingr",
		},
		Spec: networkingv1.IngressSpec{
			DefaultBackend: &networkingv1.IngressBackend{
				Service: &networkingv1.IngressServiceBackend{
					Name: "testsvc",
				},
			},
		},
	}
	_, err := kubeClient.NetworkingV1().Ingresses("red-ns").Create(context.TODO(), ingrExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	waitAndverify(t, "Ingress/red-ns/testingr")
}

func TestIngressUpdate(t *testing.T) {
	ingrUpdate := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "red-ns",
			Name:      "testingr-update",
		},
		Spec: networkingv1.IngressSpec{
			DefaultBackend: &networkingv1.IngressBackend{
				Service: &networkingv1.IngressServiceBackend{
					Name: "testsvc",
				},
			},
		},
	}
	_, err := kubeClient.NetworkingV1().Ingresses("red-ns").Create(context.TODO(), ingrUpdate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	waitAndverify(t, "Ingress/red-ns/testingr-update")

	ingrUpdate = &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "red-ns",
			Name:      "testingr-update",
		},
		Spec: networkingv1.IngressSpec{
			DefaultBackend: &networkingv1.IngressBackend{
				Service: &networkingv1.IngressServiceBackend{
					Name: "testsvc2",
				},
			},
		},
	}
	ingrUpdate.ResourceVersion = "2"
	_, err = kubeClient.NetworkingV1().Ingresses("red-ns").Update(context.TODO(), ingrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}
	waitAndverify(t, "Ingress/red-ns/testingr-update")
}

// If spec/annotation is not updated, the ingress key should not be added to ingestion queue
func TestIngressNoUpdate(t *testing.T) {
	ingrNoUpdate := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "red-ns",
			Name:      "testingr-noupdate",
		},
		Spec: networkingv1.IngressSpec{
			DefaultBackend: &networkingv1.IngressBackend{
				Service: &networkingv1.IngressServiceBackend{
					Name: "testsvc",
				},
			},
		},
	}
	_, err := kubeClient.NetworkingV1().Ingresses("red-ns").Create(context.TODO(), ingrNoUpdate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	waitAndverify(t, "Ingress/red-ns/testingr-noupdate")

	ingrNoUpdate.Status = networkingv1.IngressStatus{
		LoadBalancer: networkingv1.IngressLoadBalancerStatus{
			Ingress: []networkingv1.IngressLoadBalancerIngress{
				{
					IP:       "1.1.1.1",
					Hostname: "testingr.avi.internal",
				},
			},
		},
	}
	ingrNoUpdate.ResourceVersion = "2"
	_, err = kubeClient.NetworkingV1().Ingresses("red-ns").Update(context.TODO(), ingrNoUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	ingrNoUpdate.Status = networkingv1.IngressStatus{
		LoadBalancer: networkingv1.IngressLoadBalancerStatus{
			Ingress: []networkingv1.IngressLoadBalancerIngress{
				{
					IP:       "1.1.1.1",
					Hostname: "testingr.avi.internal",
				},
				{
					IP:       "2.3.4.5",
					Hostname: "testingr2.avi.internal",
				},
			},
		},
	}
	ingrNoUpdate.ResourceVersion = "3"
	_, err = kubeClient.NetworkingV1().Ingresses("red-ns").Update(context.TODO(), ingrNoUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	waitAndverify(t, "")
}

func TestNode(t *testing.T) {
	nodeExample := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "testnode",
			ResourceVersion: "1",
		},
		Spec: corev1.NodeSpec{
			PodCIDR:  "10.244.0.0/24",
			PodCIDRs: []string{"10.244.0.0/24"},
		},
		Status: corev1.NodeStatus{
			Addresses: []corev1.NodeAddress{
				{
					Type:    "InternalIP",
					Address: "10.1.1.2",
				},
			},
		},
	}
	_, err := kubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}
	waitAndverify(t, utils.NodeObj+"/testnode")

	nodeExample.ObjectMeta.ResourceVersion = "2"
	nodeExample.Spec.PodCIDR = "10.230.0.0/24"
	_, err = kubeClient.CoreV1().Nodes().Update(context.TODO(), nodeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Node: %v", err)
	}
	waitAndverify(t, utils.NodeObj+"/testnode")

	err = kubeClient.CoreV1().Nodes().Delete(context.TODO(), "testnode", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error in Deleting Node: %v", err)
	}
	waitAndverify(t, utils.NodeObj+"/testnode")
}

func TestMultiClusterIngress(t *testing.T) {

	os.Setenv("ENABLE_EVH", "true")
	os.Setenv("SERVICE_TYPE", "NodePort")
	defer func() {
		os.Setenv("ENABLE_EVH", "false")
		os.Setenv("SERVICE_TYPE", "ClusterIP")
	}()
	ingressObject := integrationtest.FakeMultiClusterIngress{
		Name:       "MCI-01",
		HostName:   "foo.com",
		SecretName: "my-secret",
	}
	cluster := "cluster-01"
	weight := 50
	serviceName := "service-01"
	ingressObject.Namespaces = append(ingressObject.Namespaces, "default")
	ingressObject.Ports = append(ingressObject.Ports, 8080)
	ingressObject.Clusters = append(ingressObject.Clusters, cluster)
	ingressObject.Weights = append(ingressObject.Weights, weight)
	ingressObject.Paths = append(ingressObject.Paths, "bar")
	ingressObject.ServiceNames = append(ingressObject.ServiceNames, serviceName)

	fakeMCI := ingressObject.Create()
	_, err := crdClient.AkoV1alpha1().MultiClusterIngresses("avi-system").Create(context.TODO(), fakeMCI, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Multi-cluster ingress: %v", err)
	}
	waitAndverify(t, "MultiClusterIngress/avi-system/MCI-01")
}

func TestServiceImport(t *testing.T) {

	os.Setenv("ENABLE_EVH", "true")
	os.Setenv("SERVICE_TYPE", "NodePort")
	defer func() {
		os.Setenv("ENABLE_EVH", "false")
		os.Setenv("SERVICE_TYPE", "ClusterIP")
	}()
	siObj := integrationtest.FakeServiceImport{
		Name:          "SI-01",
		Cluster:       "cluster-01",
		Namespace:     "default",
		ServiceName:   "service-01",
		EndPointIPs:   []string{"100.1.1.1", "100.1.1.2"},
		EndPointPorts: []int32{31030, 31030},
	}
	fakeSI := siObj.Create()
	_, err := crdClient.AkoV1alpha1().ServiceImports("avi-system").Create(context.TODO(), fakeSI, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service Import: %v", err)
	}
	waitAndverify(t, "ServiceImport/avi-system/SI-01")
}
