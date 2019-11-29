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

	meshutils "gitlab.eng.vmware.com/orion/container-lib/utils"
	corev1 "k8s.io/api/core/v1"
	extensionv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var kubeClient *k8sfake.Clientset
var globalKey string

func syncFuncForTest(key string) error {
	globalKey = key
	return nil
}

func setupQueue(stopCh <-chan struct{}) {
	ingestionQueue := meshutils.SharedWorkQueue().GetQueueByName(meshutils.ObjectIngestionLayer)
	ingestionQueue.SyncFunc = syncFuncForTest
	ingestionQueue.Run(stopCh)
}

func TestMain(m *testing.M) {
	setUp()
	ret := m.Run()
	os.Exit(ret)
}

func setUp() {
	kubeClient = k8sfake.NewSimpleClientset()
	registeredInformers := []string{meshutils.ServiceInformer, meshutils.EndpointInformer, meshutils.IngressInformer}
	meshutils.NewInformers(meshutils.KubeClientIntf{kubeClient}, registeredInformers)
	ctrl := SharedAviController()
	stopCh := meshutils.SetupSignalHandler()
	ctrl.Start(stopCh)
	ctrl.SetupEventHandlers(K8sinformers{kubeClient})
	setupQueue(stopCh)
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
	time.Sleep(2 * time.Second)
	if globalKey != "Service/red-ns/testsvc" {
		t.Fatalf("error in adding Service: %v", globalKey)
	}
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
		t.Fatalf("error in adding Service: %v", err)
	}
	time.Sleep(2 * time.Second)
	if globalKey != "Endpoints/red-ns/testep" {
		t.Fatalf("error in adding Service: %v", globalKey)
	}
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
	time.Sleep(2 * time.Second)
	if globalKey != "Ingress/red-ns/testingr" {
		t.Fatalf("error in adding Ingress: %v", globalKey)
	}
}
