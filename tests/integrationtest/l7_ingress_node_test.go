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

package integrationtest

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"

	"github.com/avinetworks/sdk/go/models"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func SetUpTestForIngress(t *testing.T, model_Name string) {
	os.Setenv("SHARD_VS_SIZE", "LARGE")
	os.Setenv("L7_SHARD_SCHEME", "namespace")
	objects.SharedAviGraphLister().Delete(model_Name)
	CreateSVC(t, "default", "avisvc", corev1.ServiceTypeClusterIP, false)
	CreateEP(t, "default", "avisvc", false, false, "1.1.1")
}

func TearDownTestForIngress(t *testing.T, model_Name string) {
	os.Setenv("SHARD_VS_SIZE", "")

	objects.SharedAviGraphLister().Delete(model_Name)
	DelSVC(t, "default", "avisvc")
	DelEP(t, "default", "avisvc")
}

func VerifyIngressDeletion(t *testing.T, g *gomega.WithT, aviModel interface{}, poolCount int) {
	var nodes []*avinodes.AviVsNode
	g.Eventually(func() []*avinodes.AviPoolNode {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].PoolRefs
	}, 5*time.Second).Should(gomega.HaveLen(poolCount))

	g.Eventually(func() []*models.PoolGroupMember {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].PoolGroupRefs[0].Members
	}, 5*time.Second).Should(gomega.HaveLen(poolCount))
}

func TestNoModel(t *testing.T) {
	model_Name := "admin/testl7"
	objects.SharedAviGraphLister().Delete(model_Name)

	svcExample := (FakeService{
		Name:         "testl7",
		Namespace:    "red-ns",
		Type:         corev1.ServiceTypeClusterIP,
		ServicePorts: []Serviceport{{PortName: "foo", Protocol: "TCP", PortNumber: 8080, TargetPort: 8080}},
	}).Service()
	_, err := KubeClient.CoreV1().Services("red-ns").Create(context.TODO(), svcExample, metav1.CreateOptions{})
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
	_, err = KubeClient.CoreV1().Endpoints("red-ns").Create(context.TODO(), epExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in creating Endpoint: %v", err)
	}
	PollForCompletion(t, model_Name, 5)
	found, _ := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		// We shouldn't get an update for this update since it neither belongs to an ingress nor a L4 LB service
		t.Fatalf("Model found for an unrelated update %v", model_Name)
	}

	err = KubeClient.CoreV1().Endpoints("red-ns").Delete(context.TODO(), "testl7", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Endpoint %v", err)
	}

	err = KubeClient.CoreV1().Services("red-ns").Delete(context.TODO(), "testl7", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Service %v", err)
	}
}

func TestL7Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	model_Name := "admin/cluster--Shared-L7-6"
	SetUpTestForIngress(t, model_Name)

	PollForCompletion(t, model_Name, 5)
	found, _ := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		// We shouldn't get an update for this update since it neither belongs to an ingress nor a L4 LB service
		t.Fatalf("Couldn't find model for DELETE event %v", model_Name)
	}
	ingrFake := (FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com.avi.internal"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	PollForCompletion(t, model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	} else {
		t.Fatalf("Could not find model: %v", err)
	}
	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "foo-with-targets", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_Name)
}

func TestNamespaceShardNamingConvention(t *testing.T) {
	// checks naming convention of all generated nodes
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-6"
	SetUpTestForIngress(t, modelName)
	AddSecret("my-secret", "default", "tlsCert", "tlsKey")

	ingrFake := (FakeIngress{
		Name:      "foo-with-targets",
		Namespace: "default",
		DnsNames:  []string{"foo.com", "noo.com"},
		Ips:       []string{"8.8.8.8"},
		Paths:     []string{"/foo/bar"},
		HostNames: []string{"v1"},
		TlsSecretDNS: map[string][]string{
			"my-secret": {"foo.com"},
		},
		ServiceName: "avisvc",
	}).IngressMultiPath()

	_, err := KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	PollForCompletion(t, modelName, 5)

	verifyIng, _ := KubeClient.NetworkingV1beta1().Ingresses("default").Get(context.TODO(), "foo-with-targets", metav1.GetOptions{})
	for i, host := range []string{"foo.com", "noo.com"} {
		if verifyIng.Spec.Rules[i].Host == host {
			g.Expect(verifyIng.Spec.Rules[i].Host).To(gomega.Equal(host))
			g.Expect(verifyIng.Spec.Rules[i].HTTP.Paths[0].Path).To(gomega.Equal("/foo/bar"))
		}
	}

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 15*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].Name).To(gomega.Equal("cluster--Shared-L7-6"))
	g.Expect(nodes[0].PoolGroupRefs[0].Name).To(gomega.Equal("cluster--Shared-L7-6"))
	g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--noo.com_foo_bar-default-foo-with-targets"))
	g.Expect(nodes[0].HTTPDSrefs[0].Name).To(gomega.Equal("cluster--Shared-L7-6"))
	g.Expect(nodes[0].VSVIPRefs[0].Name).To(gomega.Equal("cluster--Shared-L7-6"))
	g.Expect(nodes[0].SniNodes[0].Name).To(gomega.Equal("cluster--foo-with-targets-default-my-secret"))
	g.Expect(nodes[0].SniNodes[0].PoolGroupRefs[0].Name).To(gomega.Equal("cluster--default-foo.com_foo_bar-foo-with-targets"))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--default-foo.com_foo_bar-foo-with-targets"))
	g.Expect(nodes[0].SniNodes[0].SSLKeyCertRefs[0].Name).To(gomega.Equal("cluster--default-my-secret"))
	g.Expect(nodes[0].SniNodes[0].HttpPolicyRefs[0].Name).To(gomega.Equal("cluster--default-foo.com_foo_bar-foo-with-targets"))

	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "foo-with-targets", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)
	TearDownTestForIngress(t, modelName)
}

func TestL7ModelWithMultiTenant(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetAkoTenant()
	defer ResetAkoTenant()
	model_Name := fmt.Sprintf("%s/cluster--Shared-L7-6", AKOTENANT)
	SetUpTestForIngress(t, model_Name)
	PollForCompletion(t, model_Name, 5)
	found, _ := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		// We shouldn't get an update for this update since it neither belongs to an ingress nor a L4 LB service
		t.Fatalf("Couldn't find model for DELETE event %v", model_Name)
	}
	ingrFake := (FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com.avi.internal"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	PollForCompletion(t, model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		// Tenant should be akotenant instead of admin
		g.Expect(nodes[0].Tenant).To(gomega.Equal(AKOTENANT))
	} else {
		t.Fatalf("Could not find model: %v", err)
	}
	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "foo-with-targets", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_Name)
}

func TestMultiIngressToSameSvc(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	os.Setenv("SHARD_VS_SIZE", "LARGE")

	model_Name := "admin/cluster--Shared-L7-6"
	objects.SharedAviGraphLister().Delete(model_Name)
	svcExample := (FakeService{
		Name:         "avisvc",
		Namespace:    "default",
		Type:         corev1.ServiceTypeClusterIP,
		ServicePorts: []Serviceport{{PortName: "foo", Protocol: "TCP", PortNumber: 8080, TargetPort: 8080}},
	}).Service()

	_, err := KubeClient.CoreV1().Services("default").Create(context.TODO(), svcExample, metav1.CreateOptions{})
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
	_, err = KubeClient.CoreV1().Endpoints("default").Create(context.TODO(), epExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in creating Endpoint: %v", err)
	}
	ingrFake1 := (FakeIngress{
		Name:        "foo-with-targets1",
		Namespace:   "default",
		DnsNames:    []string{"foo.com.avi.internal"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err = KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	ingrFake2 := (FakeIngress{
		Name:        "foo-with-targets2",
		Namespace:   "default",
		DnsNames:    []string{"bar.com.avi.internal"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err = KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	PollForCompletion(t, model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(nodes[0].SharedVS).To(gomega.Equal(true))
		dsNodes := aviModel.(*avinodes.AviObjectGraph).GetAviHTTPDSNode()
		g.Expect(len(dsNodes)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolRefs {
			if strings.Contains(pool.Name, "foo.com.avi.internal_foo-default-foo-with-targets1") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("bar.com.avi.internal/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			}
		}
		// Delete the model.
		objects.SharedAviGraphLister().Delete(model_Name)
	} else {
		t.Fatalf("Could not find model on ingress delete: %v", err)
	}
	//====== VERIFICATION OF SERVICE DELETE
	// Now we have cleared the layer 2 queue for both the models. Let's delete the service.
	err = KubeClient.CoreV1().Endpoints("default").Delete(context.TODO(), "avisvc", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Endpoint %v", err)
	}
	err = KubeClient.CoreV1().Services("default").Delete(context.TODO(), "avisvc", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Service %v", err)
	}
	// We should be able to get one model now in the queue
	PollForCompletion(t, model_Name, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		dsNodes := aviModel.(*avinodes.AviObjectGraph).GetAviHTTPDSNode()
		g.Expect(len(dsNodes)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))

	} else {
		t.Fatalf("Could not find model on service delete: %v", err)
	}
	_, err = KubeClient.CoreV1().Endpoints("default").Create(context.TODO(), epExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in creating Endpoint: %v", err)
	}
	//====== VERIFICATION OF ONE INGRESS DELETE
	// Now let's delete one ingress and expect the update for that.
	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "foo-with-targets1", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	DetectModelChecksumChange(t, model_Name, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))

		// Delete the model.
		objects.SharedAviGraphLister().Delete(model_Name)
	} else {
		t.Fatalf("Could not find model on ingress delete: %v", err)
	}
	//====== VERIFICATION OF SERVICE ADD
	// Let's add the service back now - the ingress's associated with this service should be returned
	_, err = KubeClient.CoreV1().Services("default").Create(context.TODO(), svcExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	PollForCompletion(t, model_Name, 5)
	// We should be able to get one model now in the queue
	found, aviModel = objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))

		objects.SharedAviGraphLister().Delete(model_Name)
	} else {
		t.Fatalf("Could not find model on service ADD: %v", err)
	}
	//====== VERIFICATION OF ONE ENDPOINT DELETE
	err = KubeClient.CoreV1().Endpoints("default").Delete(context.TODO(), "avisvc", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Endpoint %v", err)
	}
	PollForCompletion(t, model_Name, 5)
	// Deletion should also give us the affected ingress objects
	found, aviModel = objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		// Delete the model.
		objects.SharedAviGraphLister().Delete(model_Name)
	} else {
		t.Fatalf("Could not find model on service ADD: %v", err)
	}
	err = KubeClient.CoreV1().Services("default").Delete(context.TODO(), "avisvc", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Service %v", err)
	}
	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "foo-with-targets2", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
}

func TestMultiVSIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	model_Name := "admin/cluster--Shared-L7-6"
	SetUpTestForIngress(t, model_Name)

	ingrFake := (FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com.avi.internal"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	PollForCompletion(t, model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		dsNodes := aviModel.(*avinodes.AviObjectGraph).GetAviHTTPDSNode()
		g.Expect(len(dsNodes)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
		for _, pool := range nodes[0].PoolRefs {
			// We should get two pools.
			if strings.Contains(pool.Name, "foo.com.avi.internal_foo") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			}
		}
	} else {
		t.Fatalf("Could not find model: %v", err)
	}
	randoming := (FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "randomNamespacethatyeildsdiff",
		DnsNames:    []string{"foo.com.avi.internal"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()
	_, err = KubeClient.NetworkingV1beta1().Ingresses("randomNamespacethatyeildsdiff").Create(context.TODO(), randoming, metav1.CreateOptions{})
	model_Name = "admin/cluster--Shared-L7-5"
	PollForCompletion(t, model_Name, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		dsNodes := aviModel.(*avinodes.AviObjectGraph).GetAviHTTPDSNode()
		g.Expect(len(dsNodes)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolRefs[0].Servers)).To(gomega.Equal(0))
	} else {
		t.Fatalf("Could not find model: %v", err)
	}
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "foo-with-targets", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	err = KubeClient.NetworkingV1beta1().Ingresses("randomNamespacethatyeildsdiff").Delete(context.TODO(), "foo-with-targets", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_Name)
}

func TestMultiPathIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var err error

	model_Name := "admin/cluster--Shared-L7-6"
	SetUpTestForIngress(t, model_Name)

	ingrFake := (FakeIngress{
		Name:        "ingress-multipath",
		Namespace:   "default",
		DnsNames:    []string{"foo.com.avi.internal"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
	}).IngressMultiPath()

	_, err = KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	PollForCompletion(t, model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolRefs {
			if strings.Contains(pool.Name, "foo.com.avi.internal_foo-default-ingress-multipath") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if strings.Contains(pool.Name, "foo.com.avi.internal_bar-default-ingress-multipath") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/bar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if strings.Contains(*pool.PoolRef, "/api/pool?name=cluster--foo.com.avi.internal_foo-default-ingress-multipath") {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/foo"))
			} else if strings.Contains(*pool.PoolRef, "/api/pool?name=cluster--foo.com.avi.internal_bar-default-ingress-multipath") {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/bar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", model_Name)
	}
	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "ingress-multipath", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_Name)
}

func TestMultiIngressSameHost(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	model_Name := "admin/cluster--Shared-L7-6"
	SetUpTestForIngress(t, model_Name)

	ingrFake1 := (FakeIngress{
		Name:        "ingress-multi1",
		Namespace:   "default",
		DnsNames:    []string{"foo.com.avi.internal"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake2 := (FakeIngress{
		Name:        "ingress-multi2",
		Namespace:   "default",
		DnsNames:    []string{"foo.com.avi.internal"},
		Paths:       []string{"/bar"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err = KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	PollForCompletion(t, model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolRefs {
			if strings.Contains(pool.Name, "foo.com.avi.internal_foo-default-ingress-multi1") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if strings.Contains(pool.Name, "foo.com.avi.internal_bar-default-ingress-multi2") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/bar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if strings.Contains(*pool.PoolRef, "/api/pool?name=cluster--foo.com.avi.internal_foo-default-ingress-multi1") {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/foo"))
			} else if strings.Contains(*pool.PoolRef, "/api/pool?name=cluster--foo.com.avi.internal_bar-default-ingress-multi2") {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/bar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", model_Name)
	}
	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "ingress-multi1", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 1)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--foo.com.avi.internal_bar-default-ingress-multi2"))

	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "ingress-multi2", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_Name)
}

func TestMultiHostIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	model_Name := "admin/cluster--Shared-L7-6"
	SetUpTestForIngress(t, model_Name)

	ingrFake := (FakeIngress{
		Name:        "ingress-multihost",
		Namespace:   "default",
		DnsNames:    []string{"foo.com.avi.internal", "bar.com.avi.internal"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	PollForCompletion(t, model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolRefs {
			if strings.Contains(pool.Name, "foo.com.avi.internal_foo-default-ingress-multihost") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if strings.Contains(pool.Name, "bar.com.avi.internal_bar-default-ingress-multihost") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("bar.com.avi.internal/bar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if strings.Contains(*pool.PoolRef, "/api/pool?name=cluster--foo.com.avi.internal_foo-default-ingress-multihost") {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/foo"))
			} else if strings.Contains(*pool.PoolRef, "/api/pool?name=cluster--bar.com.avi.internal_bar-default-ingress-multihost") {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("bar.com.avi.internal/bar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", model_Name)
	}

	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "ingress-multihost", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_Name)
}

func TestMultiHostSameHostNameIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	model_Name := "admin/cluster--Shared-L7-6"
	SetUpTestForIngress(t, model_Name)

	ingrFake := (FakeIngress{
		Name:        "ingress-multihost",
		Namespace:   "default",
		DnsNames:    []string{"foo.com.avi.internal", "foo.com.avi.internal"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	PollForCompletion(t, model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolRefs {
			if strings.Contains(pool.Name, "foo.com.avi.internal_foo-default-ingress-multihost") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if strings.Contains(pool.Name, "foo.com.avi.internal_bar-default-ingress-multihost") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/bar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if strings.Contains(*pool.PoolRef, "/api/pool?name=cluster--foo.com.avi.internal_foo-default-ingress-multihost") {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/foo"))
			} else if strings.Contains(*pool.PoolRef, "/api/pool?name=cluster--foo.com.avi.internal_bar-default-ingress-multihost") {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/bar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", model_Name)
	}

	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "ingress-multihost", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_Name)
}

func TestEditPathIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	model_Name := "admin/cluster--Shared-L7-6"
	SetUpTestForIngress(t, model_Name)

	ingrFake := (FakeIngress{
		Name:        "ingress-edit",
		Namespace:   "default",
		DnsNames:    []string{"foo.com.avi.internal"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).Ingress()
	ingrFake.ResourceVersion = "1"
	_, err := KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	fmt.Println("SHIT1")

	PollForCompletion(t, model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Eventually(len(nodes), 5*time.Second).Should(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Eventually(func() []*avinodes.AviPoolNode {
			return nodes[0].PoolRefs
		}, 5*time.Second).Should(gomega.HaveLen(1))

		g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.ContainSubstring("cluster--foo.com.avi.internal_foo-default-ingress-edit"))
		g.Expect(nodes[0].PoolRefs[0].PriorityLabel).To(gomega.Equal("foo.com.avi.internal/foo"))
		g.Expect(len(nodes[0].PoolRefs[0].Servers)).To(gomega.Equal(1))

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(1))

		pool := nodes[0].PoolGroupRefs[0].Members[0]
		g.Expect(*pool.PoolRef).To(gomega.ContainSubstring("/api/pool?name=cluster--foo.com.avi.internal_foo-default-ingress-edit"))
		g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/foo"))

	} else {
		t.Fatalf("Could not find model: %s", model_Name)
	}

	ingrFake = (FakeIngress{
		Name:        "ingress-edit",
		Namespace:   "default",
		DnsNames:    []string{"foo.com.avi.internal"},
		Paths:       []string{"/bar"},
		ServiceName: "avisvc",
	}).Ingress()
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1beta1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}
	fmt.Println("SHIT2")

	found, aviModel = objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Eventually(len(nodes), 5*time.Second).Should(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Eventually(func() []*avinodes.AviPoolNode {
			return nodes[0].PoolRefs
		}, 20*time.Second).Should(gomega.HaveLen(1))
		g.Eventually(func() string {
			return nodes[0].PoolRefs[0].Name
		}, 20*time.Second).Should(gomega.Equal("cluster--foo.com.avi.internal_bar-default-ingress-edit"))
		g.Expect(nodes[0].PoolRefs[0].PriorityLabel).To(gomega.Equal("foo.com.avi.internal/bar"))
		g.Expect(len(nodes[0].PoolRefs[0].Servers)).To(gomega.Equal(1))

		pool := nodes[0].PoolGroupRefs[0].Members[0]
		g.Expect(*pool.PoolRef).To(gomega.Equal("/api/pool?name=cluster--foo.com.avi.internal_bar-default-ingress-edit"))
		g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/bar"))
	} else {
		t.Fatalf("Could not find model: %s", model_Name)
	}

	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "ingress-edit", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	fmt.Println("SHIT3")
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_Name)
}

func TestEditMultiPathIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	model_Name := "admin/cluster--Shared-L7-6"
	SetUpTestForIngress(t, model_Name)

	ingrFake := (FakeIngress{
		Name:        "ingress-multipath-edit",
		Namespace:   "default",
		DnsNames:    []string{"foo.com.avi.internal"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).Ingress()
	ingrFake.ResourceVersion = "1"

	_, err := KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	PollForCompletion(t, model_Name, 5)
	ingrFake = (FakeIngress{
		Name:        "ingress-multipath-edit",
		Namespace:   "default",
		DnsNames:    []string{"foo.com.avi.internal"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
	}).IngressMultiPath()
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1beta1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Eventually(len(nodes), 5*time.Second).Should(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		//g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		g.Eventually(func() []*avinodes.AviPoolNode {
			return nodes[0].PoolRefs
		}, 5*time.Second).Should(gomega.HaveLen(2))
		for _, pool := range nodes[0].PoolRefs {
			if strings.Contains(pool.Name, "foo.com.avi.internal_foo-default-ingress-multipath-edit") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if strings.Contains(pool.Name, "foo.com.avi.internal_bar-default-ingress-multipath-edit") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/bar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=cluster--foo.com.avi.internal_foo-default-ingress-multipath-edit" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/foo"))
			} else if *pool.PoolRef == "/api/pool?name=cluster--foo.com.avi.internal_bar-default-ingress-multipath-edit" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/bar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", model_Name)
	}
	ingrFake = (FakeIngress{
		Name:        "ingress-multipath-edit",
		Namespace:   "default",
		DnsNames:    []string{"foo.com.avi.internal"},
		Paths:       []string{"/foo", "/foobar"},
		ServiceName: "avisvc",
	}).IngressMultiPath()
	ingrFake.ResourceVersion = "3"
	objects.SharedAviGraphLister().Delete(model_Name)
	_, err = KubeClient.NetworkingV1beta1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	PollForCompletion(t, model_Name, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Eventually(len(nodes), 5*time.Second).Should(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Eventually(func() []*avinodes.AviPoolNode {
			return nodes[0].PoolRefs
		}, 5*time.Second).Should(gomega.HaveLen(2))
		for _, pool := range nodes[0].PoolRefs {
			if strings.Contains(pool.Name, "foo.com.avi.internal_foo-default-ingress-multipath-edit") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if strings.Contains(pool.Name, "foo.com.avi.internal_foobar-default-ingress-multipath-edit") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/foobar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=cluster--foo.com.avi.internal_foo-default-ingress-multipath-edit" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/foo"))
			} else if *pool.PoolRef == "/api/pool?name=cluster--foo.com.avi.internal_foobar-default-ingress-multipath-edit" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/foobar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", model_Name)
	}

	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "ingress-multipath-edit", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_Name)
}

func TestEditMultiIngressSameHost(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	model_name := "admin/cluster--Shared-L7-6"
	SetUpTestForIngress(t, model_name)

	ingrFake1 := (FakeIngress{
		Name:        "ingress-multi1",
		Namespace:   "default",
		DnsNames:    []string{"foo.com.avi.internal"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake2 := (FakeIngress{
		Name:        "ingress-multi2",
		Namespace:   "default",
		DnsNames:    []string{"foo.com.avi.internal"},
		Paths:       []string{"/bar"},
		ServiceName: "avisvc",
	}).Ingress()
	ingrFake2.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake2 = (FakeIngress{
		Name:        "ingress-multi2",
		Namespace:   "default",
		DnsNames:    []string{"foo.com.avi.internal"},
		Paths:       []string{"/foobar"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err = KubeClient.NetworkingV1beta1().Ingresses("default").Update(context.TODO(), ingrFake2, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	PollForCompletion(t, model_name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolRefs {
			if strings.Contains(pool.Name, "foo.com.avi.internal_foo-default-ingress-multi1") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if strings.Contains(pool.Name, "foo.com.avi.internal_foobar-default-ingress-multi2") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/foobar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=cluster--foo.com.avi.internal_foo-default-ingress-multi1" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/foo"))
			} else if *pool.PoolRef == "/api/pool?name=cluster--foo.com.avi.internal_foobar-default-ingress-multi2" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/foobar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", model_name)
	}
	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "ingress-multi1", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 1)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--foo.com.avi.internal_foobar-default-ingress-multi2"))

	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "ingress-multi2", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_name)
}

func TestEditMultiHostIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	model_Name := "admin/cluster--Shared-L7-6"
	SetUpTestForIngress(t, model_Name)

	ingrFake := (FakeIngress{
		Name:        "ingress-multihost",
		Namespace:   "default",
		DnsNames:    []string{"foo.com.avi.internal", "bar.com.avi.internal"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake = (FakeIngress{
		Name:        "ingress-multihost",
		Namespace:   "default",
		DnsNames:    []string{"foo.com.avi.internal", "foobar.com.avi.internal"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
	}).Ingress()
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1beta1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	PollForCompletion(t, model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolRefs {
			if strings.Contains(pool.Name, "foo.com.avi.internal_foo-default-ingress-multihost") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if strings.Contains(pool.Name, "foobar.com.avi.internal_bar-default-ingress-multihost") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foobar.com.avi.internal/bar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=cluster--foo.com.avi.internal_foo-default-ingress-multihost" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/foo"))
			} else if *pool.PoolRef == "/api/pool?name=cluster--foobar.com.avi.internal_bar-default-ingress-multihost" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foobar.com.avi.internal/bar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", model_Name)
	}

	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "ingress-multihost", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_Name)
}

func TestNoHostIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	model_Name := "admin/cluster--Shared-L7-6"
	SetupDomain()
	SetUpTestForIngress(t, model_Name)

	ingrFake := (FakeIngress{
		Name:        "ingress-nohost",
		Namespace:   "default",
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).IngressNoHost()

	_, err := KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	PollForCompletion(t, model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
		fmt.Println(nodes[0].PoolRefs[0].Name)
		g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--ingress-nohost.default.com_foo-default-ingress-nohost"))
		g.Expect(nodes[0].PoolRefs[0].PriorityLabel).To(gomega.Equal("ingress-nohost.default.com/foo"))

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(1))

		pool := nodes[0].PoolGroupRefs[0].Members[0]
		g.Expect(*pool.PoolRef).To(gomega.Equal("/api/pool?name=cluster--ingress-nohost.default.com_foo-default-ingress-nohost"))
		g.Expect(*pool.PriorityLabel).To(gomega.Equal("ingress-nohost.default.com/foo"))
	} else {
		t.Fatalf("Could not find model: %s", model_Name)
	}

	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "ingress-nohost", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_Name)
}

func TestEditNoHostIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	model_Name := "admin/cluster--Shared-L7-6"
	SetUpTestForIngress(t, model_Name)

	ingrFake := (FakeIngress{
		Name:        "ingress-nohost",
		Namespace:   "default",
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).IngressNoHost()

	_, err := KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake = (FakeIngress{
		Name:        "ingress-nohost",
		Namespace:   "default",
		Paths:       []string{"/bar"},
		ServiceName: "avisvc",
	}).IngressNoHost()
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1beta1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in Updating Ingress: %v", err)
	}

	PollForCompletion(t, model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
		g.Eventually(func() string {
			return nodes[0].PoolRefs[0].Name
		}, 5*time.Second).Should(gomega.Equal("cluster--ingress-nohost.default.com_bar-default-ingress-nohost"))
		g.Expect(nodes[0].PoolRefs[0].PriorityLabel).To(gomega.Equal("ingress-nohost.default.com/bar"))

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(1))

		pool := nodes[0].PoolGroupRefs[0].Members[0]
		g.Expect(*pool.PoolRef).To(gomega.Equal("/api/pool?name=cluster--ingress-nohost.default.com_bar-default-ingress-nohost"))
		g.Expect(*pool.PriorityLabel).To(gomega.Equal("ingress-nohost.default.com/bar"))
	} else {
		t.Fatalf("Could not find model: %s", model_Name)
	}

	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "ingress-nohost", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_Name)
}

func TestEditNoHostToHostIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	model_Name := "admin/cluster--Shared-L7-6"
	SetUpTestForIngress(t, model_Name)

	ingrFake := (FakeIngress{
		Name:        "ingress-nohost",
		Namespace:   "default",
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).IngressNoHost()

	_, err := KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	PollForCompletion(t, model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
		g.Eventually(func() string {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes[0].PoolRefs) == 1 {
				return nodes[0].PoolRefs[0].Name
			}
			return ""
		}, 5*time.Second).Should(gomega.Equal("cluster--ingress-nohost.default.com_foo-default-ingress-nohost"))
		g.Expect(nodes[0].PoolRefs[0].PriorityLabel).To(gomega.Equal("ingress-nohost.default.com/foo"))

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(1))

		pool := nodes[0].PoolGroupRefs[0].Members[0]
		g.Expect(*pool.PoolRef).To(gomega.Equal("/api/pool?name=cluster--ingress-nohost.default.com_foo-default-ingress-nohost"))
		g.Expect(*pool.PriorityLabel).To(gomega.Equal("ingress-nohost.default.com/foo"))
	} else {
		t.Fatalf("Could not find model: %s", model_Name)
	}

	ingrFake = (FakeIngress{
		Name:        "ingress-nohost",
		Namespace:   "default",
		DnsNames:    []string{"foo.com.avi.internal"},
		Paths:       []string{"/bar"},
		ServiceName: "avisvc",
	}).Ingress()
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1beta1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in Updating Ingress: %v", err)
	}

	PollForCompletion(t, model_Name, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
		g.Eventually(func() string {
			if len(nodes[0].PoolRefs) == 0 {
				return ""
			}
			return nodes[0].PoolRefs[0].Name
		}, 20*time.Second).Should(gomega.Equal("cluster--foo.com.avi.internal_bar-default-ingress-nohost"))
		g.Expect(nodes[0].PoolRefs[0].PriorityLabel).To(gomega.Equal("foo.com.avi.internal/bar"))

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(1))

		pool := nodes[0].PoolGroupRefs[0].Members[0]
		g.Expect(*pool.PoolRef).To(gomega.Equal("/api/pool?name=cluster--foo.com.avi.internal_bar-default-ingress-nohost"))
		g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com.avi.internal/bar"))
	} else {
		t.Fatalf("Could not find model: %s", model_Name)
	}

	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "ingress-nohost", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_Name)
}

func TestNoHostMultiPathIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	model_Name := "admin/cluster--Shared-L7-6"
	SetUpTestForIngress(t, model_Name)

	ingrFake := (FakeIngress{
		Name:        "nohost-multipath",
		Namespace:   "default",
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
	}).IngressNoHost()

	_, err := KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	PollForCompletion(t, model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))

		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolRefs {
			if strings.Contains(pool.Name, "nohost-multipath.default.com_foo-default-nohost-multipath") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("nohost-multipath.default.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if strings.Contains(pool.Name, "nohost-multipath.default.com_bar-default-nohost-multipath") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("nohost-multipath.default.com/bar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=cluster--nohost-multipath.default.com_foo-default-nohost-multipath" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("nohost-multipath.default.com/foo"))
			} else if *pool.PoolRef == "/api/pool?name=cluster--nohost-multipath.default.com_bar-default-nohost-multipath" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("nohost-multipath.default.com/bar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}

	} else {
		t.Fatalf("Could not find model: %s", model_Name)
	}

	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "nohost-multipath", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_Name)
}

func TestEditNoHostMultiPathIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	model_Name := "admin/cluster--Shared-L7-6"
	SetUpTestForIngress(t, model_Name)

	ingrFake := (FakeIngress{
		Name:        "nohost-multipath",
		Namespace:   "default",
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
	}).IngressNoHost()

	_, err := KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake = (FakeIngress{
		Name:        "nohost-multipath",
		Namespace:   "default",
		Paths:       []string{"/foo", "/foobar"},
		ServiceName: "avisvc",
	}).IngressNoHost()
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1beta1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	PollForCompletion(t, model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))

		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolRefs {
			if strings.Contains(pool.Name, "nohost-multipath.default.com_foo-default-nohost-multipath") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("nohost-multipath.default.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if strings.Contains(pool.Name, "nohost-multipath.default.com_foobar-default-nohost-multipath") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("nohost-multipath.default.com/foobar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=cluster--nohost-multipath.default.com_foo-default-nohost-multipath" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("nohost-multipath.default.com/foo"))
			} else if *pool.PoolRef == "/api/pool?name=cluster--nohost-multipath.default.com_foobar-default-nohost-multipath" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("nohost-multipath.default.com/foobar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}

	} else {
		t.Fatalf("Could not find model: %s", model_Name)
	}

	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "nohost-multipath", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_Name)
}
