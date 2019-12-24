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

	"github.com/onsi/gomega"
	avinodes "gitlab.eng.vmware.com/orion/akc/pkg/nodes"
	"gitlab.eng.vmware.com/orion/akc/pkg/objects"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	extensionv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func DetectModelChecksumChange(t *testing.T, key string, counter int) interface{} {
	// This method detects a change in the checksum and returns.
	count := 0
	initialcs := uint32(0)
	found, aviModel := objects.SharedAviGraphLister().Get(key)
	if found {
		initialcs = aviModel.(*avinodes.AviObjectGraph).GraphChecksum
	}
	for count < counter {
		found, aviModel = objects.SharedAviGraphLister().Get(key)
		if found {
			if initialcs == aviModel.(*avinodes.AviObjectGraph).GraphChecksum {
				count = count + 1
				time.Sleep(1 * time.Second)
			} else {
				return aviModel
			}
		}
	}
	return nil
}
func TestNoModel(t *testing.T) {
	model_name := "red-ns/testl7"
	objects.SharedAviGraphLister().Delete(model_name)
	svcExample := &corev1.Service{
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{Name: "foo", Port: 8080, Protocol: "TCP", TargetPort: intstr.FromInt(8080)},
			},
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "red-ns",
			Name:      "testl7",
		},
	}
	_, err := kubeClient.CoreV1().Services("red-ns").Create(svcExample)
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	epExample := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "red-ns",
			Name:      "testl7",
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
	pollForCompletion(t, model_name, 5)
	found, _ := objects.SharedAviGraphLister().Get(model_name)
	if found {
		// We shouldn't get an update for this update since it neither belongs to an ingress nor a L4 LB service
		t.Fatalf("Model found for an unrelated update %v", model_name)
	}
	err = kubeClient.CoreV1().Endpoints("red-ns").Delete("testl7", nil)

	if err != nil {
		t.Fatalf("Couldn't DELETE the Endpoint %v", err)
	}
	err = kubeClient.CoreV1().Services("red-ns").Delete("testl7", nil)
}

func TestL7Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	os.Setenv("shard_vs_size", "LARGE")
	model_name := "default/Shard-VS-6"
	objects.SharedAviGraphLister().Delete(model_name)
	svcExample := &corev1.Service{
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{Name: "foo", Port: 8080, Protocol: "TCP", TargetPort: intstr.FromInt(8080)},
			},
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "avisvc",
		},
	}
	_, err := kubeClient.CoreV1().Services("default").Create(svcExample)
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	epExample := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "avisvc",
		},
		Subsets: []corev1.EndpointSubset{{
			Addresses: []corev1.EndpointAddress{{IP: "1.2.3.4"}},
			Ports:     []corev1.EndpointPort{{Name: "foo", Port: 8080, Protocol: "TCP"}},
		}},
	}
	_, err = kubeClient.CoreV1().Endpoints("default").Create(epExample)
	if err != nil {
		t.Fatalf("error in creating Endpoint: %v", err)
	}
	pollForCompletion(t, model_name, 5)
	found, _ := objects.SharedAviGraphLister().Get(model_name)
	if found {
		// We shouldn't get an update for this update since it neither belongs to an ingress nor a L4 LB service
		t.Fatalf("Couldn't find model for DELETE event %v", model_name)
	}
	ingrFake := (fakeIngress{
		name:        "foo-with-targets",
		namespace:   "default",
		dnsnames:    []string{"foo.com"},
		ips:         []string{"8.8.8.8"},
		hostnames:   []string{"v1"},
		serviceName: "avisvc",
	}).Ingress()

	_, err = kubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	pollForCompletion(t, model_name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	} else {
		t.Fatalf("Could not find model: %v", err)
	}
	err = kubeClient.CoreV1().Endpoints("default").Delete("avisvc", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Endpoint %v", err)
	}
	err = kubeClient.CoreV1().Services("default").Delete("avisvc", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Service %v", err)
	}
	err = kubeClient.ExtensionsV1beta1().Ingresses("default").Delete("foo-with-targets", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}

}

func TestMultiIngressToSameSvc(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	os.Setenv("shard_vs_size", "LARGE")
	model_name := "default/Shard-VS-6"
	objects.SharedAviGraphLister().Delete(model_name)
	svcExample := &corev1.Service{
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{Name: "foo", Port: 8080, Protocol: "TCP", TargetPort: intstr.FromInt(8080)},
			},
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "avisvc",
		},
	}
	_, err := kubeClient.CoreV1().Services("default").Create(svcExample)
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	epExample := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "avisvc",
		},
		Subsets: []corev1.EndpointSubset{{
			Addresses: []corev1.EndpointAddress{{IP: "1.2.3.4"}},
			Ports:     []corev1.EndpointPort{{Name: "foo", Port: 8080, Protocol: "TCP"}},
		}},
	}
	_, err = kubeClient.CoreV1().Endpoints("default").Create(epExample)
	if err != nil {
		t.Fatalf("error in creating Endpoint: %v", err)
	}
	ingrFake1 := (fakeIngress{
		name:        "foo-with-targets1",
		namespace:   "default",
		dnsnames:    []string{"foo.com"},
		ips:         []string{"8.8.8.8"},
		hostnames:   []string{"v1"},
		serviceName: "avisvc",
	}).Ingress()

	_, err = kubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake1)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	ingrFake2 := (fakeIngress{
		name:        "foo-with-targets2",
		namespace:   "default",
		dnsnames:    []string{"bar.com"},
		ips:         []string{"8.8.8.8"},
		hostnames:   []string{"v1"},
		serviceName: "avisvc",
	}).Ingress()

	_, err = kubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake2)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	pollForCompletion(t, model_name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		dsNodes := aviModel.(*avinodes.AviObjectGraph).GetAviHTTPDSNode()
		g.Expect(len(dsNodes)).To(gomega.Equal(1))
		poolNodes := aviModel.(*avinodes.AviObjectGraph).GetAviPoolNodes()
		g.Expect(len(poolNodes)).To(gomega.Equal(2))
		for _, pool := range poolNodes {
			// We should get two pools.
			if pool.Name == "pool-foo.com/foo" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				// The servers should be empty
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("bar.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			}
		}
		// Delete the model.
		objects.SharedAviGraphLister().Delete(model_name)
	} else {
		t.Fatalf("Could not find model on ingress delete: %v", err)
	}
	//====== VERIFICATION OF SERVICE DELETE
	// Now we have cleared the layer 2 queue for both the models. Let's delete the service.
	err = kubeClient.CoreV1().Endpoints("default").Delete("avisvc", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Endpoint %v", err)
	}
	err = kubeClient.CoreV1().Services("default").Delete("avisvc", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Service %v", err)
	}
	// We should be able to get one model now in the queue
	pollForCompletion(t, model_name, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(model_name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		dsNodes := aviModel.(*avinodes.AviObjectGraph).GetAviHTTPDSNode()
		g.Expect(len(dsNodes)).To(gomega.Equal(1))
		poolNodes := aviModel.(*avinodes.AviObjectGraph).GetAviPoolNodes()
		g.Expect(len(poolNodes)).To(gomega.Equal(2))
		for _, pool := range poolNodes {
			// We should get two pools.
			if pool.Name == "pool-foo.com/foo" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				// The servers should be empty
				g.Expect(len(pool.Servers)).To(gomega.Equal(0))
			} else {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("bar.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(0))
			}
		}
	} else {
		t.Fatalf("Could not find model on service delete: %v", err)
	}
	_, err = kubeClient.CoreV1().Endpoints("default").Create(epExample)
	if err != nil {
		t.Fatalf("error in creating Endpoint: %v", err)
	}
	//====== VERIFICATION OF ONE INGRESS DELETE
	// Now let's delete one ingress and expect the update for that.
	err = kubeClient.ExtensionsV1beta1().Ingresses("default").Delete("foo-with-targets1", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	DetectModelChecksumChange(t, model_name, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(model_name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		poolNodes := aviModel.(*avinodes.AviObjectGraph).GetAviPoolNodes()
		g.Expect(len(poolNodes)).To(gomega.Equal(1))
		for _, pool := range poolNodes {
			// We should get two pools.
			if pool.Name == "pool-bar.com/foo" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("bar.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			}
		}
		// Delete the model.
		objects.SharedAviGraphLister().Delete(model_name)
	} else {
		t.Fatalf("Could not find model on ingress delete: %v", err)
	}
	//====== VERIFICATION OF SERVICE ADD
	// Let's add the service back now - the ingress's associated with this service should be returned
	_, err = kubeClient.CoreV1().Services("default").Create(svcExample)
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	pollForCompletion(t, model_name, 5)
	// We should be able to get one model now in the queue
	found, aviModel = objects.SharedAviGraphLister().Get(model_name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		// Delete the model.
		poolNodes := aviModel.(*avinodes.AviObjectGraph).GetAviPoolNodes()
		g.Expect(len(poolNodes)).To(gomega.Equal(1))
		for _, pool := range poolNodes {
			// We should get two pools.
			if pool.Name == "pool-bar.com/foo" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("bar.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			}
		}
		objects.SharedAviGraphLister().Delete(model_name)
	} else {
		t.Fatalf("Could not find model on service ADD: %v", err)
	}
	//====== VERIFICATION OF ONE ENDPOINT DELETE
	err = kubeClient.CoreV1().Endpoints("default").Delete("avisvc", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Endpoint %v", err)
	}
	pollForCompletion(t, model_name, 5)
	// Deletion should also give us the affected ingress objects
	found, aviModel = objects.SharedAviGraphLister().Get(model_name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))

		poolNodes := aviModel.(*avinodes.AviObjectGraph).GetAviPoolNodes()
		g.Expect(len(poolNodes)).To(gomega.Equal(1))
		for _, pool := range poolNodes {
			if pool.Name == "pool-bar.com/foo" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("bar.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(0))
			}
		}
		// Delete the model.
		objects.SharedAviGraphLister().Delete(model_name)
	} else {
		t.Fatalf("Could not find model on service ADD: %v", err)
	}
}

// Fake ingress
type fakeIngress struct {
	dnsnames    []string
	tlsdnsnames [][]string
	ips         []string
	hostnames   []string
	namespace   string
	name        string
	annotations map[string]string
	serviceName string
}

func (ing fakeIngress) Ingress() *extensionv1beta1.Ingress {
	ingress := &extensionv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   ing.namespace,
			Name:        ing.name,
			Annotations: ing.annotations,
		},
		Spec: extensionv1beta1.IngressSpec{
			Rules: []extensionv1beta1.IngressRule{},
		},
		Status: extensionv1beta1.IngressStatus{
			LoadBalancer: v1.LoadBalancerStatus{
				Ingress: []v1.LoadBalancerIngress{},
			},
		},
	}
	for _, dnsname := range ing.dnsnames {
		ingress.Spec.Rules = append(ingress.Spec.Rules, extensionv1beta1.IngressRule{
			Host: dnsname,
			IngressRuleValue: extensionv1beta1.IngressRuleValue{
				HTTP: &extensionv1beta1.HTTPIngressRuleValue{
					Paths: []extensionv1beta1.HTTPIngressPath{extensionv1beta1.HTTPIngressPath{
						Path: "foo",
						Backend: extensionv1beta1.IngressBackend{ServiceName: ing.serviceName, ServicePort: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: 8080,
						}},
					},
					},
				},
			},
		})
	}
	for _, hosts := range ing.tlsdnsnames {
		ingress.Spec.TLS = append(ingress.Spec.TLS, extensionv1beta1.IngressTLS{
			Hosts: hosts,
		})
	}
	for _, ip := range ing.ips {
		ingress.Status.LoadBalancer.Ingress = append(ingress.Status.LoadBalancer.Ingress, v1.LoadBalancerIngress{
			IP: ip,
		})
	}
	for _, hostname := range ing.hostnames {
		ingress.Status.LoadBalancer.Ingress = append(ingress.Status.LoadBalancer.Ingress, v1.LoadBalancerIngress{
			Hostname: hostname,
		})
	}
	return ingress
}
