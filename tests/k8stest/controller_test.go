/*
 * Copyright 2019-2020 VMware, Inc.
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
	"os"
	"sync"
	"testing"
	"time"

	crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/client/v1alpha1/clientset/versioned/fake"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	corev1 "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
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
var dynamicClient *dynamicfake.FakeDynamicClient
var keyChan chan string
var ctrl *k8s.AviController

func syncFuncForTest(key string, wg *sync.WaitGroup) error {
	keyChan <- key
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

func addConfigMap() {
	aviCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "avi-system",
			Name:      "avi-k8s-config",
		},
	}
	kubeClient.CoreV1().ConfigMaps("avi-system").Create(context.TODO(), aviCM, metav1.CreateOptions{})
}

func TestMain(m *testing.M) {
	kubeClient = k8sfake.NewSimpleClientset()
	dynamicClient = dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())
	os.Setenv("NETWORK_NAME", "net123")
	os.Setenv("CLUSTER_NAME", "cluster")
	os.Setenv("CLOUD_NAME", "CLOUD_VCENTER")
	os.Setenv("SEG_NAME", "Default-Group")
	os.Setenv("NODE_NETWORK_LIST", `[{"networkName":"net123","cidrs":["10.79.168.0/22"]}]`)
	crdClient = crdfake.NewSimpleClientset()
	lib.SetCRDClientset(crdClient)

	registeredInformers := []string{
		utils.ServiceInformer,
		utils.EndpointInformer,
		utils.IngressInformer,
		utils.IngressClassInformer,
		utils.SecretInformer,
		utils.NSInformer,
		utils.NodeInformer,
		utils.ConfigMapInformer,
	}
	utils.NewInformers(utils.KubeClientIntf{ClientSet: kubeClient}, registeredInformers)
	k8s.NewCRDInformers(crdClient)
	integrationtest.InitializeFakeAKOAPIServer()

	integrationtest.NewAviFakeClientInstance()
	defer integrationtest.AviFakeClientInstance.Close()

	ctrl = k8s.SharedAviController()
	stopCh := utils.SetupSignalHandler()
	ctrl.Start(stopCh)
	keyChan = make(chan string)
	ctrlCh := make(chan struct{})
	quickSyncCh := make(chan struct{})
	addConfigMap()
	ctrl.HandleConfigMap(k8s.K8sinformers{Cs: kubeClient, DynamicClient: dynamicClient}, ctrlCh, stopCh, quickSyncCh)
	ctrl.SetupEventHandlers(k8s.K8sinformers{Cs: kubeClient, DynamicClient: dynamicClient})
	setupQueue(stopCh)
	os.Exit(m.Run())
}

func TestSvc(t *testing.T) {
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

func TestEndpoint(t *testing.T) {
	epExample := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "red-ns",
			Name:      "testep",
		},
		Subsets: []corev1.EndpointSubset{},
	}
	_, err := kubeClient.CoreV1().Endpoints("red-ns").Create(context.TODO(), epExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in creating Endpoint: %v", err)
	}
	waitAndverify(t, "Endpoints/red-ns/testep")
}

func TestIngress(t *testing.T) {
	ingrExample := &networkingv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "red-ns",
			Name:      "testingr",
		},
		Spec: networkingv1beta1.IngressSpec{
			Backend: &networkingv1beta1.IngressBackend{
				ServiceName: "testsvc",
			},
		},
	}
	_, err := kubeClient.NetworkingV1beta1().Ingresses("red-ns").Create(context.TODO(), ingrExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	waitAndverify(t, "Ingress/red-ns/testingr")
}

func TestIngressUpdate(t *testing.T) {
	ingrUpdate := &networkingv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "red-ns",
			Name:      "testingr-update",
		},
		Spec: networkingv1beta1.IngressSpec{
			Backend: &networkingv1beta1.IngressBackend{
				ServiceName: "testsvc",
			},
		},
	}
	_, err := kubeClient.NetworkingV1beta1().Ingresses("red-ns").Create(context.TODO(), ingrUpdate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	waitAndverify(t, "Ingress/red-ns/testingr-update")

	ingrUpdate = &networkingv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "red-ns",
			Name:      "testingr-update",
		},
		Spec: networkingv1beta1.IngressSpec{
			Backend: &networkingv1beta1.IngressBackend{
				ServiceName: "testsvc2",
			},
		},
	}
	ingrUpdate.ResourceVersion = "2"
	_, err = kubeClient.NetworkingV1beta1().Ingresses("red-ns").Update(context.TODO(), ingrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}
	waitAndverify(t, "Ingress/red-ns/testingr-update")
}

// If spec/annotation is not updated, the ingress key should not be added to ingestion queue
func TestIngressNoUpdate(t *testing.T) {
	ingrNoUpdate := &networkingv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "red-ns",
			Name:      "testingr-noupdate",
		},
		Spec: networkingv1beta1.IngressSpec{
			Backend: &networkingv1beta1.IngressBackend{
				ServiceName: "testsvc",
			},
		},
	}
	_, err := kubeClient.NetworkingV1beta1().Ingresses("red-ns").Create(context.TODO(), ingrNoUpdate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	waitAndverify(t, "Ingress/red-ns/testingr-noupdate")

	ingrNoUpdate.Status = networkingv1beta1.IngressStatus{
		LoadBalancer: corev1.LoadBalancerStatus{
			Ingress: []corev1.LoadBalancerIngress{
				{
					IP:       "1.1.1.1",
					Hostname: "testingr.avi.internal",
				},
			},
		},
	}
	ingrNoUpdate.ResourceVersion = "2"
	_, err = kubeClient.NetworkingV1beta1().Ingresses("red-ns").Update(context.TODO(), ingrNoUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	ingrNoUpdate.Status = networkingv1beta1.IngressStatus{
		LoadBalancer: corev1.LoadBalancerStatus{
			Ingress: []corev1.LoadBalancerIngress{
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
	_, err = kubeClient.NetworkingV1beta1().Ingresses("red-ns").Update(context.TODO(), ingrNoUpdate, metav1.UpdateOptions{})
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
			PodCIDR: "10.244.0.0/24",
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
