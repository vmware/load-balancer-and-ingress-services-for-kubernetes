/*
 * [2013] - [2019] Avi Networks Incorporated
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
package k8s

import (
	"os"
	"testing"
	"time"

	k8sfake "k8s.io/client-go/kubernetes/fake"

	// To Do: add test for openshift route
	//oshiftfake "github.com/openshift/client-go/route/clientset/versioned/fake"

	"gitlab.eng.vmware.com/orion/container-lib/utils"
	meshutils "gitlab.eng.vmware.com/orion/container-lib/utils"
	corev1 "k8s.io/api/core/v1"
	extensionv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var kubeClient *k8sfake.Clientset
var keyChan chan string
var ctrl AviController

func syncFuncForTest(key string) error {
	keyChan <- key
	return nil
}

func waitAndverify(t *testing.T, key string) {
	waitChan := make(chan int)
	go func() {
		time.Sleep(5 * time.Second)
		waitChan <- 1
	}()

	select {
	case data := <-keyChan:
		if data != key {
			t.Fatalf("error in match expected: %v, got: %v", key, data)
		}
	case _ = <-waitChan:
		t.Fatalf("timed out waitig for %v", key)
	}
}

func setupQueue(stopCh <-chan struct{}) {
	ingestionQueue := meshutils.SharedWorkQueue().GetQueueByName(meshutils.ObjectIngestionLayer)
	ingestionQueue.SyncFunc = syncFuncForTest
	ingestionQueue.Run(stopCh)
}

func addConfigMap(t *testing.T) {
	aviCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "avi-system",
			Name:      "avi-k8s-config",
		},
	}
	_, err := kubeClient.CoreV1().ConfigMaps("avi-system").Create(aviCM)
	if err != nil {
		t.Fatalf("error in adding configmap: %v", err)
	}
}

func TestMain(m *testing.M) {
	setUp()
	ret := m.Run()
	os.Exit(ret)
}

func setUp() {
	kubeClient = k8sfake.NewSimpleClientset()
	os.Setenv("INGRESS_API", "extensionv1")
	registeredInformers := []string{meshutils.ServiceInformer, meshutils.EndpointInformer, meshutils.ExtV1IngressInformer, meshutils.SecretInformer, meshutils.NSInformer, meshutils.NodeInformer, utils.ConfigMapInformer}
	meshutils.NewInformers(meshutils.KubeClientIntf{kubeClient}, registeredInformers)
	ctrl := SharedAviController()
	stopCh := meshutils.SetupSignalHandler()
	ctrl.Start(stopCh)
	keyChan = make(chan string)
	ctrl.HandleConfigMap(K8sinformers{kubeClient}, stopCh)
	ctrl.SetupEventHandlers(K8sinformers{kubeClient})
	setupQueue(stopCh)
}

func TestAviConfigMap(t *testing.T) {
	aviCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "avi-system",
			Name:      "avi-k8s-config",
		},
	}
	_, err := kubeClient.CoreV1().ConfigMaps("avi-system").Create(aviCM)
	if err != nil {
		t.Fatalf("error in adding configmap: %v", err)
	}
	time.Sleep(10 * time.Second)
	if ctrl.DisableSync {
		t.Fatalf("sync not enabled after adding configmap")
	}
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
	_, err := kubeClient.CoreV1().Services("red-ns").Create(svcExample)
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
	_, err := kubeClient.CoreV1().Endpoints("red-ns").Create(epExample)
	if err != nil {
		t.Fatalf("error in creating Endpoint: %v", err)
	}
	waitAndverify(t, "Endpoints/red-ns/testep")
}

func TestIngress(t *testing.T) {
	ingrExample := &extensionv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "red-ns",
			Name:      "testingr",
		},
		Spec: extensionv1beta1.IngressSpec{
			Backend: &extensionv1beta1.IngressBackend{
				ServiceName: "testsvc",
			},
		},
	}
	_, err := kubeClient.ExtensionsV1beta1().Ingresses("red-ns").Create(ingrExample)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	waitAndverify(t, "Ingress/red-ns/testingr")
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
	_, err := kubeClient.CoreV1().Nodes().Create(nodeExample)
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}
	waitAndverify(t, utils.NodeObj+"/testnode")

	nodeExample.ObjectMeta.ResourceVersion = "2"
	nodeExample.Spec.PodCIDR = "10.230.0.0/24"
	_, err = kubeClient.CoreV1().Nodes().Update(nodeExample)
	if err != nil {
		t.Fatalf("error in updating Node: %v", err)
	}
	waitAndverify(t, utils.NodeObj+"/testnode")

	err = kubeClient.CoreV1().Nodes().Delete("testnode", nil)
	if err != nil {
		t.Fatalf("error in Deleting Node: %v", err)
	}
	waitAndverify(t, utils.NodeObj+"/testnode")
}
