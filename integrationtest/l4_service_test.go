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

	k8sfake "k8s.io/client-go/kubernetes/fake"

	"github.com/onsi/gomega"
	"gitlab.eng.vmware.com/orion/akc/pkg/k8s"
	avinodes "gitlab.eng.vmware.com/orion/akc/pkg/nodes"
	"gitlab.eng.vmware.com/orion/akc/pkg/objects"
	meshutils "gitlab.eng.vmware.com/orion/container-lib/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var kubeClient *k8sfake.Clientset
var ctrl *k8s.AviController

func TestMain(m *testing.M) {
	setUp()
	ret := m.Run()
	os.Exit(ret)
}

func setUp() {
	kubeClient = k8sfake.NewSimpleClientset()
	registeredInformers := []string{meshutils.ServiceInformer, meshutils.EndpointInformer, meshutils.ExtV1IngressInformer, meshutils.SecretInformer, meshutils.NSInformer, meshutils.NodeInformer, meshutils.ConfigMapInformer}
	meshutils.NewInformers(meshutils.KubeClientIntf{kubeClient}, registeredInformers)
	informers := k8s.K8sinformers{Cs: kubeClient}
	os.Setenv("CTRL_USERNAME", "admin")
	os.Setenv("CTRL_PASSWORD", "admin")
	os.Setenv("CTRL_IPADDRESS", "localhost")
	os.Setenv("INGRESS_API", "extensionv1")
	os.Setenv("FULL_SYNC_INTERVAL", "60")
	ctrl = k8s.SharedAviController()
	stopCh := meshutils.SetupSignalHandler()
	k8s.PopulateCache()
	ctrlCh := ctrl.HandleConfigMap(informers, stopCh)
	go ctrl.InitController(informers, ctrlCh, stopCh)
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
	pollForSyncStart(t, ctrl, 10)
}

func TestAviNodeCreationSinglePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	model_name := "admin/testsvc--red-ns"
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
	_, err := kubeClient.CoreV1().Services("red-ns").Create(svcExample)
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	epExample := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "red-ns",
			Name:      "testsvc",
		},
		Subsets: []corev1.EndpointSubset{{
			Addresses: []corev1.EndpointAddress{{IP: "1.2.3.4"}},
			Ports:     []corev1.EndpointPort{{Name: "foo", Port: 8080, Protocol: "TCP"}},
		}},
	}
	_, err = kubeClient.CoreV1().Endpoints("red-ns").Create(epExample)
	if err != nil {
		t.Fatalf("error in creating Endpoint: %v", err)
	}
	pollForCompletion(t, model_name, 15)

	found, aviModel := objects.SharedAviGraphLister().Get(model_name)
	if !found {
		t.Fatalf("Couldn't find model %v", model_name)
	} else {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.Equal("testsvc--red-ns"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(nodes[0].EastWest).To(gomega.Equal(false))
		g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))
		// Check for the pools
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
		address := "1.2.3.4"
		g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&address))
		g.Expect(len(nodes[0].TCPPoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))

	}
	objects.SharedAviGraphLister().Delete(model_name)
	err = kubeClient.CoreV1().Services("red-ns").Delete("testsvc", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the service %v", err)
	}
	pollForCompletion(t, model_name, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(model_name)
	if !found {
		t.Fatalf("Couldn't find model for DELETE event %v", model_name)
	} else {
		if aviModel != nil {
			t.Fatalf("Avi model: %v not nil for DELETE", model_name)
		}
	}
	err = kubeClient.CoreV1().Endpoints("red-ns").Delete("testsvc", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Endpoint %v", err)
	}
}

func TestAviNodeCreationMultiPort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	model_name := "admin/testsvc--red-ns"
	objects.SharedAviGraphLister().Delete(model_name)
	svcExample := &corev1.Service{
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeLoadBalancer,
			Ports: []corev1.ServicePort{
				{Name: "foo", Port: 8080, Protocol: "TCP", TargetPort: intstr.FromInt(8080)},
				{Name: "bar", Port: 9080, Protocol: "TCP", TargetPort: intstr.FromInt(9080)},
				{Name: "baz", Port: 7080, Protocol: "TCP", TargetPort: intstr.FromInt(7080)},
			},
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
	epExample := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "red-ns",
			Name:      "testsvc",
		},
		Subsets: []corev1.EndpointSubset{{
			Addresses: []corev1.EndpointAddress{{IP: "1.2.3.5"}},
			Ports:     []corev1.EndpointPort{{Name: "foo", Port: 8080, Protocol: "TCP"}},
		},
			{
				Addresses: []corev1.EndpointAddress{{IP: "1.2.3.6"}, {IP: "1.2.3.9"}},
				Ports:     []corev1.EndpointPort{{Name: "bar", Port: 9080, Protocol: "TCP"}},
			},
			{
				Addresses: []corev1.EndpointAddress{{IP: "1.2.3.7"}, {IP: "1.2.3.10"}, {IP: "1.2.3.11"}},
				Ports:     []corev1.EndpointPort{{Name: "baz", Port: 7080, Protocol: "UDP"}},
			}},
	}
	_, err = kubeClient.CoreV1().Endpoints("red-ns").Create(epExample)
	if err != nil {
		t.Fatalf("error in creating Endpoint: %v", err)
	}
	pollForCompletion(t, model_name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_name)
	if !found {
		t.Fatalf("Couldn't find model %v", model_name)
	} else {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.Equal("testsvc--red-ns"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(nodes[0].EastWest).To(gomega.Equal(false))
		g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))
		// Check for the pools
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(3))

		for _, node := range nodes[0].PoolRefs {
			if node.Port == 9080 {
				address := "1.2.3.6"
				g.Expect(len(node.Servers)).To(gomega.Equal(2))
				g.Expect(node.Servers[0].Ip.Addr).To(gomega.Equal(&address))
			} else if node.Port == 8080 {
				address := "1.2.3.5"
				g.Expect(len(node.Servers)).To(gomega.Equal(1))
				g.Expect(node.Servers[0].Ip.Addr).To(gomega.Equal(&address))
			} else {
				address := "1.2.3.7"
				g.Expect(len(node.Servers)).To(gomega.Equal(3))
				g.Expect(node.Servers[0].Ip.Addr).To(gomega.Equal(&address))
			}
		}
		g.Expect(len(nodes[0].TCPPoolGroupRefs)).To(gomega.Equal(3))
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(3))
		g.Expect(nodes[0].ApplicationProfile).To(gomega.Equal(meshutils.DEFAULT_L4_APP_PROFILE))
		g.Expect(nodes[0].NetworkProfile).To(gomega.Equal(meshutils.DEFAULT_TCP_NW_PROFILE))

	}
}

func TestAviNodeMultiPortApplicationProf(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	model_name := "admin/testsvc1--red-ns"
	objects.SharedAviGraphLister().Delete(model_name)
	svcExample := &corev1.Service{
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeLoadBalancer,
			Ports: []corev1.ServicePort{
				{Name: "foo", Port: 8080, Protocol: "UDP", TargetPort: intstr.FromInt(8080)},
				{Name: "bar", Port: 9080, Protocol: "UDP", TargetPort: intstr.FromInt(9080)},
				{Name: "baz", Port: 7080, Protocol: "UDP", TargetPort: intstr.FromInt(7080)},
			},
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "red-ns",
			Name:      "testsvc1",
		},
	}
	_, err := kubeClient.CoreV1().Services("red-ns").Create(svcExample)
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	epExample := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "red-ns",
			Name:      "testsvc1",
		},
		Subsets: []corev1.EndpointSubset{{
			Addresses: []corev1.EndpointAddress{{IP: "1.2.3.5"}},
			Ports:     []corev1.EndpointPort{{Name: "foo", Port: 8080, Protocol: "UDP"}},
		},
			{
				Addresses: []corev1.EndpointAddress{{IP: "1.2.3.6"}, {IP: "1.2.3.9"}},
				Ports:     []corev1.EndpointPort{{Name: "bar", Port: 9080, Protocol: "UDP"}},
			},
			{
				Addresses: []corev1.EndpointAddress{{IP: "1.2.3.7"}, {IP: "1.2.3.10"}, {IP: "1.2.3.11"}},
				Ports:     []corev1.EndpointPort{{Name: "baz", Port: 7080, Protocol: "UDP"}},
			}},
	}
	_, err = kubeClient.CoreV1().Endpoints("red-ns").Create(epExample)
	if err != nil {
		t.Fatalf("error in creating Endpoint: %v", err)
	}
	pollForCompletion(t, model_name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_name)
	if !found {
		t.Fatalf("Couldn't find model %v", model_name)
	} else {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.Equal("testsvc1--red-ns"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(nodes[0].EastWest).To(gomega.Equal(false))
		g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))
		// Check for the pools
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(3))

		for _, node := range nodes[0].PoolRefs {
			if node.Port == 9080 {
				address := "1.2.3.6"
				g.Expect(len(node.Servers)).To(gomega.Equal(2))
				g.Expect(node.Servers[0].Ip.Addr).To(gomega.Equal(&address))
			} else if node.Port == 8080 {
				address := "1.2.3.5"
				g.Expect(len(node.Servers)).To(gomega.Equal(1))
				g.Expect(node.Servers[0].Ip.Addr).To(gomega.Equal(&address))
			} else {
				address := "1.2.3.7"
				g.Expect(len(node.Servers)).To(gomega.Equal(3))
				g.Expect(node.Servers[0].Ip.Addr).To(gomega.Equal(&address))
			}
		}
		g.Expect(len(nodes[0].TCPPoolGroupRefs)).To(gomega.Equal(3))
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(3))
		g.Expect(nodes[0].SharedVS).To(gomega.Equal(false))
		g.Expect(nodes[0].ApplicationProfile).To(gomega.Equal(meshutils.DEFAULT_L4_APP_PROFILE))
		g.Expect(nodes[0].NetworkProfile).To(gomega.Equal(meshutils.SYSTEM_UDP_FAST_PATH))

	}
}

func TestAviNodeUpdateEndpoint(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	model_name := "admin/testsvc--red-ns"
	objects.SharedAviGraphLister().Delete(model_name)
	epExample := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "red-ns",
			Name:      "testsvc",
		},
		Subsets: []corev1.EndpointSubset{{
			Addresses: []corev1.EndpointAddress{{IP: "1.2.3.4"}, {IP: "1.2.3.5"}},
			Ports:     []corev1.EndpointPort{{Name: "foo", Port: 8080, Protocol: "TCP"}},
		}},
	}
	_, err := kubeClient.CoreV1().Endpoints("red-ns").Update(epExample)
	if err != nil {
		t.Fatalf("Error in updating the Endpoint: %v", err)
	}

	pollForCompletion(t, model_name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_name)
	if !found {
		t.Fatalf("Couldn't find model %v", model_name)
	} else {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(3))
		for _, node := range nodes[0].PoolRefs {
			if node.Port == 9080 {
				g.Expect(len(node.Servers)).To(gomega.Equal(0))
			} else if node.Port == 8080 {
				g.Expect(len(node.Servers)).To(gomega.Equal(2))
			}
		}
	}
	objects.SharedAviGraphLister().Delete(model_name)
	err = kubeClient.CoreV1().Services("red-ns").Delete("testsvc", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the service %v", err)
	}
	pollForCompletion(t, model_name, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(model_name)
	if !found {
		t.Fatalf("Couldn't find model for DELETE event %v", model_name)
	} else {
		if aviModel != nil {
			t.Fatalf("Avi model: %v not nil for DELETE", model_name)
		}
	}
}
