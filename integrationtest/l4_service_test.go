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

package integrationtest

import (
	"os"
	"testing"
	"time"

	k8sfake "k8s.io/client-go/kubernetes/fake"

	"github.com/onsi/gomega"
	"gitlab.eng.vmware.com/orion/akc/pkg/k8s"
	"gitlab.eng.vmware.com/orion/akc/pkg/nodes"
	"gitlab.eng.vmware.com/orion/akc/pkg/objects"
	meshutils "gitlab.eng.vmware.com/orion/container-lib/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var kubeClient *k8sfake.Clientset
var globalKey string

func syncFuncForTest(key string) error {
	globalKey = key
	return nil
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
	informers := k8s.K8sinformers{Cs: kubeClient}
	go k8s.InitController(informers)
}

func TestSvc(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	svcExample := &corev1.Service{
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeLoadBalancer,
			Ports: []corev1.ServicePort{
				{Name: "foo", Port: 8080, Protocol: "TCP", TargetPort: intstr.FromInt(8080)},
			},
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "red-ns",
			Name:      "testsvc",
		},
	}
	time.Sleep(2 * time.Second)
	_, err := kubeClient.CoreV1().Services("red-ns").Create(svcExample)
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	time.Sleep(3 * time.Second)
	model_name := "red-ns/testsvc"
	found, aviModel := objects.SharedAviGraphLister().Get(model_name)
	if !found {
		t.Fatalf("Couldn't find model %v", model_name)
	} else {
		nodes := aviModel.(*nodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.Equal("testsvc"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("red-ns"))
		g.Expect(nodes[0].EastWest).To(gomega.Equal(false))
		g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))
	}
}
