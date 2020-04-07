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
	"strings"
	"testing"
	"time"

	avinodes "ako/pkg/nodes"
	"ako/pkg/objects"

	"github.com/avinetworks/sdk/go/models"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func SetUpTestForIngress(t *testing.T, model_Name string) {
	os.Setenv("SHARD_VS_SIZE", "LARGE")
	os.Setenv("CLOUD_NAME", "Shard-VS-")
	os.Setenv("VRF_CONTEXT", "global")
	os.Setenv("L7_SHARD_SCHEME", "namespace")
	objects.SharedAviGraphLister().Delete(model_Name)
	CreateSVC(t, "default", "avisvc", corev1.ServiceTypeClusterIP, false)
	CreateEP(t, "default", "avisvc", false, false)
}

func TearDownTestForIngress(t *testing.T, model_Name string) {
	os.Setenv("SHARD_VS_SIZE", "")
	os.Setenv("CLOUD_NAME", "")
	os.Setenv("VRF_CONTEXT", "")

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
	_, err := KubeClient.CoreV1().Services("red-ns").Create(svcExample)
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
	_, err = KubeClient.CoreV1().Endpoints("red-ns").Create(epExample)
	if err != nil {
		t.Fatalf("error in creating Endpoint: %v", err)
	}
	PollForCompletion(t, model_Name, 5)
	found, _ := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		// We shouldn't get an update for this update since it neither belongs to an ingress nor a L4 LB service
		t.Fatalf("Model found for an unrelated update %v", model_Name)
	}
	err = KubeClient.CoreV1().Endpoints("red-ns").Delete("testl7", nil)

	if err != nil {
		t.Fatalf("Couldn't DELETE the Endpoint %v", err)
	}
	err = KubeClient.CoreV1().Services("red-ns").Delete("testl7", nil)
}

func TestL7Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	model_Name := "admin/Shard-VS---global-6"
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
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	PollForCompletion(t, model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	} else {
		t.Fatalf("Could not find model: %v", err)
	}
	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("foo-with-targets", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_Name)
}

func TestMultiIngressToSameSvc(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	os.Setenv("SHARD_VS_SIZE", "LARGE")
	os.Setenv("CLOUD_NAME", "Shard-VS-")
	os.Setenv("VRF_CONTEXT", "global")

	model_Name := "admin/Shard-VS---global-6"
	objects.SharedAviGraphLister().Delete(model_Name)
	svcExample := (FakeService{
		Name:         "avisvc",
		Namespace:    "default",
		Type:         corev1.ServiceTypeClusterIP,
		ServicePorts: []Serviceport{{PortName: "foo", Protocol: "TCP", PortNumber: 8080, TargetPort: 8080}},
	}).Service()

	_, err := KubeClient.CoreV1().Services("default").Create(svcExample)
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
	_, err = KubeClient.CoreV1().Endpoints("default").Create(epExample)
	if err != nil {
		t.Fatalf("error in creating Endpoint: %v", err)
	}
	ingrFake1 := (FakeIngress{
		Name:        "foo-with-targets1",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake1)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	ingrFake2 := (FakeIngress{
		Name:        "foo-with-targets2",
		Namespace:   "default",
		DnsNames:    []string{"bar.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake2)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	PollForCompletion(t, model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(nodes[0].SharedVS).To(gomega.Equal(true))
		dsNodes := aviModel.(*avinodes.AviObjectGraph).GetAviHTTPDSNode()
		g.Expect(len(dsNodes)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolRefs {
			if strings.Contains(pool.Name, "foo.com/foo--default--foo-with-targets1") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("bar.com/foo"))
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
	err = KubeClient.CoreV1().Endpoints("default").Delete("avisvc", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Endpoint %v", err)
	}
	err = KubeClient.CoreV1().Services("default").Delete("avisvc", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Service %v", err)
	}
	// We should be able to get one model now in the queue
	PollForCompletion(t, model_Name, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		dsNodes := aviModel.(*avinodes.AviObjectGraph).GetAviHTTPDSNode()
		g.Expect(len(dsNodes)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))

	} else {
		t.Fatalf("Could not find model on service delete: %v", err)
	}
	_, err = KubeClient.CoreV1().Endpoints("default").Create(epExample)
	if err != nil {
		t.Fatalf("error in creating Endpoint: %v", err)
	}
	//====== VERIFICATION OF ONE INGRESS DELETE
	// Now let's delete one ingress and expect the update for that.
	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("foo-with-targets1", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	DetectModelChecksumChange(t, model_Name, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))

		// Delete the model.
		objects.SharedAviGraphLister().Delete(model_Name)
	} else {
		t.Fatalf("Could not find model on ingress delete: %v", err)
	}
	//====== VERIFICATION OF SERVICE ADD
	// Let's add the service back now - the ingress's associated with this service should be returned
	_, err = KubeClient.CoreV1().Services("default").Create(svcExample)
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	PollForCompletion(t, model_Name, 5)
	// We should be able to get one model now in the queue
	found, aviModel = objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))

		objects.SharedAviGraphLister().Delete(model_Name)
	} else {
		t.Fatalf("Could not find model on service ADD: %v", err)
	}
	//====== VERIFICATION OF ONE ENDPOINT DELETE
	err = KubeClient.CoreV1().Endpoints("default").Delete("avisvc", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Endpoint %v", err)
	}
	PollForCompletion(t, model_Name, 5)
	// Deletion should also give us the affected ingress objects
	found, aviModel = objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		// Delete the model.
		objects.SharedAviGraphLister().Delete(model_Name)
	} else {
		t.Fatalf("Could not find model on service ADD: %v", err)
	}
	err = KubeClient.CoreV1().Services("default").Delete("avisvc", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Service %v", err)
	}
	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("foo-with-targets2", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
}

func TestMultiVSIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	model_Name := "admin/Shard-VS---global-6"
	SetUpTestForIngress(t, model_Name)

	ingrFake := (FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	PollForCompletion(t, model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		dsNodes := aviModel.(*avinodes.AviObjectGraph).GetAviHTTPDSNode()
		g.Expect(len(dsNodes)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
		for _, pool := range nodes[0].PoolRefs {
			// We should get two pools.
			if strings.Contains(pool.Name, "foo.com/foo") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			}
		}
	} else {
		t.Fatalf("Could not find model: %v", err)
	}
	randoming := (FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "randomNamespacethatyeildsdiff",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()
	_, err = KubeClient.ExtensionsV1beta1().Ingresses("randomNamespacethatyeildsdiff").Create(randoming)
	model_Name = "admin/Shard-VS---global-5"
	PollForCompletion(t, model_Name, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
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
	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("foo-with-targets", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	err = KubeClient.ExtensionsV1beta1().Ingresses("randomNamespacethatyeildsdiff").Delete("foo-with-targets", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_Name)
}

func TestMultiPathIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var err error

	model_Name := "admin/Shard-VS---global-6"
	SetUpTestForIngress(t, model_Name)

	ingrFake := (FakeIngress{
		Name:        "ingress-multipath",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
	}).IngressMultiPath()

	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	PollForCompletion(t, model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolRefs {
			if strings.Contains(pool.Name, "foo.com/foo--default--ingress-multipath") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if strings.Contains(pool.Name, "foo.com/bar--default--ingress-multipath") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if strings.Contains(*pool.PoolRef, "/api/pool?name=global--foo.com/foo--default--ingress-multipath") {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			} else if strings.Contains(*pool.PoolRef, "/api/pool?name=global--foo.com/bar--default--ingress-multipath") {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", model_Name)
	}
	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-multipath", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_Name)
}

func TestMultiIngressSameHost(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	model_Name := "admin/Shard-VS---global-6"
	SetUpTestForIngress(t, model_Name)

	ingrFake1 := (FakeIngress{
		Name:        "ingress-multi1",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake1)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake2 := (FakeIngress{
		Name:        "ingress-multi2",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/bar"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake2)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	PollForCompletion(t, model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolRefs {
			if strings.Contains(pool.Name, "foo.com/foo--default--ingress-multi1") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if strings.Contains(pool.Name, "foo.com/bar--default--ingress-multi2") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if strings.Contains(*pool.PoolRef, "/api/pool?name=global--foo.com/foo--default--ingress-multi1") {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			} else if strings.Contains(*pool.PoolRef, "/api/pool?name=global--foo.com/bar--default--ingress-multi2") {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", model_Name)
	}
	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-multi1", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 1)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("global--foo.com/bar--default--ingress-multi2"))

	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-multi2", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_Name)
}

func TestMultiHostIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	model_Name := "admin/Shard-VS---global-6"
	SetUpTestForIngress(t, model_Name)

	ingrFake := (FakeIngress{
		Name:        "ingress-multihost",
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "bar.com"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	PollForCompletion(t, model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolRefs {
			if strings.Contains(pool.Name, "foo.com/foo--default--ingress-multihost") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if strings.Contains(pool.Name, "bar.com/bar--default--ingress-multihost") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("bar.com/bar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if strings.Contains(*pool.PoolRef, "/api/pool?name=global--foo.com/foo--default--ingress-multihost") {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			} else if strings.Contains(*pool.PoolRef, "/api/pool?name=global--bar.com/bar--default--ingress-multihost") {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("bar.com/bar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", model_Name)
	}

	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-multihost", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_Name)
}

func TestMultiHostSameHostNameIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	model_Name := "admin/Shard-VS---global-6"
	SetUpTestForIngress(t, model_Name)

	ingrFake := (FakeIngress{
		Name:        "ingress-multihost",
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "foo.com"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	PollForCompletion(t, model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolRefs {
			if strings.Contains(pool.Name, "foo.com/foo--default--ingress-multihost") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if strings.Contains(pool.Name, "foo.com/bar--default--ingress-multihost") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if strings.Contains(*pool.PoolRef, "/api/pool?name=global--foo.com/foo--default--ingress-multihost") {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			} else if strings.Contains(*pool.PoolRef, "/api/pool?name=global--foo.com/bar--default--ingress-multihost") {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", model_Name)
	}

	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-multihost", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_Name)
}

func TestEditPathIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	model_Name := "admin/Shard-VS---global-6"
	SetUpTestForIngress(t, model_Name)

	ingrFake := (FakeIngress{
		Name:        "ingress-edit",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).Ingress()
	ingrFake.ResourceVersion = "1"
	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	PollForCompletion(t, model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Eventually(len(nodes), 5*time.Second).Should(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Eventually(func() []*avinodes.AviPoolNode {
			return nodes[0].PoolRefs
		}, 5*time.Second).Should(gomega.HaveLen(1))

		g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.ContainSubstring("global--foo.com/foo--default--ingress-edit"))
		g.Expect(nodes[0].PoolRefs[0].PriorityLabel).To(gomega.Equal("foo.com/foo"))
		g.Expect(len(nodes[0].PoolRefs[0].Servers)).To(gomega.Equal(1))

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(1))

		pool := nodes[0].PoolGroupRefs[0].Members[0]
		g.Expect(*pool.PoolRef).To(gomega.ContainSubstring("/api/pool?name=global--foo.com/foo--default--ingress-edit"))
		g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))

	} else {
		t.Fatalf("Could not find model: %s", model_Name)
	}

	ingrFake = (FakeIngress{
		Name:        "ingress-edit",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/bar"},
		ServiceName: "avisvc",
	}).Ingress()
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Update(ingrFake)
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	found, aviModel = objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Eventually(len(nodes), 5*time.Second).Should(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Eventually(func() []*avinodes.AviPoolNode {
			return nodes[0].PoolRefs
		}, 5*time.Second).Should(gomega.HaveLen(1))
		g.Eventually(func() string {
			return nodes[0].PoolRefs[0].Name
		}, 5*time.Second).Should(gomega.Equal("global--foo.com/bar--default--ingress-edit"))
		g.Expect(nodes[0].PoolRefs[0].PriorityLabel).To(gomega.Equal("foo.com/bar"))
		g.Expect(len(nodes[0].PoolRefs[0].Servers)).To(gomega.Equal(1))

		pool := nodes[0].PoolGroupRefs[0].Members[0]
		g.Expect(*pool.PoolRef).To(gomega.Equal("/api/pool?name=global--foo.com/bar--default--ingress-edit"))
		g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
	} else {
		t.Fatalf("Could not find model: %s", model_Name)
	}

	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-edit", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_Name)
}

func TestEditMultiPathIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	model_Name := "admin/Shard-VS---global-6"
	SetUpTestForIngress(t, model_Name)

	ingrFake := (FakeIngress{
		Name:        "ingress-multipath-edit",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).Ingress()
	ingrFake.ResourceVersion = "1"

	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	PollForCompletion(t, model_Name, 5)
	ingrFake = (FakeIngress{
		Name:        "ingress-multipath-edit",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
	}).IngressMultiPath()
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Update(ingrFake)
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Eventually(len(nodes), 5*time.Second).Should(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		//g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		g.Eventually(func() []*avinodes.AviPoolNode {
			return nodes[0].PoolRefs
		}, 5*time.Second).Should(gomega.HaveLen(2))
		for _, pool := range nodes[0].PoolRefs {
			if strings.Contains(pool.Name, "foo.com/foo--default--ingress-multipath-edit") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if strings.Contains(pool.Name, "foo.com/bar--default--ingress-multipath-edit") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=global--foo.com/foo--default--ingress-multipath-edit" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			} else if *pool.PoolRef == "/api/pool?name=global--foo.com/bar--default--ingress-multipath-edit" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
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
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo", "/foobar"},
		ServiceName: "avisvc",
	}).IngressMultiPath()
	ingrFake.ResourceVersion = "3"
	objects.SharedAviGraphLister().Delete(model_Name)
	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Update(ingrFake)
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	PollForCompletion(t, model_Name, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Eventually(len(nodes), 5*time.Second).Should(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Eventually(func() []*avinodes.AviPoolNode {
			return nodes[0].PoolRefs
		}, 5*time.Second).Should(gomega.HaveLen(2))
		for _, pool := range nodes[0].PoolRefs {
			if strings.Contains(pool.Name, "foo.com/foo--default--ingress-multipath-edit") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if strings.Contains(pool.Name, "foo.com/foobar--default--ingress-multipath-edit") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foobar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=global--foo.com/foo--default--ingress-multipath-edit" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			} else if *pool.PoolRef == "/api/pool?name=global--foo.com/foobar--default--ingress-multipath-edit" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foobar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", model_Name)
	}

	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-multipath-edit", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_Name)
}

func TestEditMultiIngressSameHost(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	model_name := "admin/Shard-VS---global-6"
	SetUpTestForIngress(t, model_name)

	ingrFake1 := (FakeIngress{
		Name:        "ingress-multi1",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake1)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake2 := (FakeIngress{
		Name:        "ingress-multi2",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/bar"},
		ServiceName: "avisvc",
	}).Ingress()
	ingrFake2.ResourceVersion = "2"
	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake2)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake2 = (FakeIngress{
		Name:        "ingress-multi2",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foobar"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Update(ingrFake2)
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	PollForCompletion(t, model_name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolRefs {
			if strings.Contains(pool.Name, "foo.com/foo--default--ingress-multi1") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if strings.Contains(pool.Name, "foo.com/foobar--default--ingress-multi2") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foobar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=global--foo.com/foo--default--ingress-multi1" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			} else if *pool.PoolRef == "/api/pool?name=global--foo.com/foobar--default--ingress-multi2" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foobar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", model_name)
	}
	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-multi1", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 1)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("global--foo.com/foobar--default--ingress-multi2"))

	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-multi2", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_name)
}

func TestEditMultiHostIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	model_Name := "admin/Shard-VS---global-6"
	SetUpTestForIngress(t, model_Name)

	ingrFake := (FakeIngress{
		Name:        "ingress-multihost",
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "bar.com"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake = (FakeIngress{
		Name:        "ingress-multihost",
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "foobar.com"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
	}).Ingress()
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Update(ingrFake)

	PollForCompletion(t, model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolRefs {
			if strings.Contains(pool.Name, "foo.com/foo--default--ingress-multihost") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if strings.Contains(pool.Name, "foobar.com/bar--default--ingress-multihost") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foobar.com/bar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=global--foo.com/foo--default--ingress-multihost" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			} else if *pool.PoolRef == "/api/pool?name=global--foobar.com/bar--default--ingress-multihost" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foobar.com/bar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", model_Name)
	}

	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-multihost", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_Name)
}

func TestNoHostIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	model_Name := "admin/Shard-VS---global-6"
	SetUpTestForIngress(t, model_Name)

	ingrFake := (FakeIngress{
		Name:        "ingress-nohost",
		Namespace:   "default",
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).IngressNoHost()

	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	PollForCompletion(t, model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
		g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("global--ingress-nohost.default.avi.internal/foo--default--ingress-nohost"))
		g.Expect(nodes[0].PoolRefs[0].PriorityLabel).To(gomega.Equal("ingress-nohost.default.avi.internal/foo"))

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(1))

		pool := nodes[0].PoolGroupRefs[0].Members[0]
		g.Expect(*pool.PoolRef).To(gomega.Equal("/api/pool?name=global--ingress-nohost.default.avi.internal/foo--default--ingress-nohost"))
		g.Expect(*pool.PriorityLabel).To(gomega.Equal("ingress-nohost.default.avi.internal/foo"))
	} else {
		t.Fatalf("Could not find model: %s", model_Name)
	}

	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-nohost", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_Name)
}

func TestEditNoHostIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	model_Name := "admin/Shard-VS---global-6"
	SetUpTestForIngress(t, model_Name)

	ingrFake := (FakeIngress{
		Name:        "ingress-nohost",
		Namespace:   "default",
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).IngressNoHost()

	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
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
	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Update(ingrFake)
	if err != nil {
		t.Fatalf("error in Updating Ingress: %v", err)
	}

	PollForCompletion(t, model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
		g.Eventually(func() string {
			return nodes[0].PoolRefs[0].Name
		}, 5*time.Second).Should(gomega.Equal("global--ingress-nohost.default.avi.internal/bar--default--ingress-nohost"))
		g.Expect(nodes[0].PoolRefs[0].PriorityLabel).To(gomega.Equal("ingress-nohost.default.avi.internal/bar"))

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(1))

		pool := nodes[0].PoolGroupRefs[0].Members[0]
		g.Expect(*pool.PoolRef).To(gomega.Equal("/api/pool?name=global--ingress-nohost.default.avi.internal/bar--default--ingress-nohost"))
		g.Expect(*pool.PriorityLabel).To(gomega.Equal("ingress-nohost.default.avi.internal/bar"))
	} else {
		t.Fatalf("Could not find model: %s", model_Name)
	}

	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-nohost", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_Name)
}

func TestEditNoHostToHostIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	model_Name := "admin/Shard-VS---global-6"
	SetUpTestForIngress(t, model_Name)

	ingrFake := (FakeIngress{
		Name:        "ingress-nohost",
		Namespace:   "default",
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).IngressNoHost()

	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	PollForCompletion(t, model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
		g.Eventually(func() string {
			return nodes[0].PoolRefs[0].Name
		}, 5*time.Second).Should(gomega.Equal("global--ingress-nohost.default.avi.internal/foo--default--ingress-nohost"))
		g.Expect(nodes[0].PoolRefs[0].PriorityLabel).To(gomega.Equal("ingress-nohost.default.avi.internal/foo"))

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(1))

		pool := nodes[0].PoolGroupRefs[0].Members[0]
		g.Expect(*pool.PoolRef).To(gomega.Equal("/api/pool?name=global--ingress-nohost.default.avi.internal/foo--default--ingress-nohost"))
		g.Expect(*pool.PriorityLabel).To(gomega.Equal("ingress-nohost.default.avi.internal/foo"))
	} else {
		t.Fatalf("Could not find model: %s", model_Name)
	}

	ingrFake = (FakeIngress{
		Name:        "ingress-nohost",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/bar"},
		ServiceName: "avisvc",
	}).Ingress()
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Update(ingrFake)
	if err != nil {
		t.Fatalf("error in Updating Ingress: %v", err)
	}

	PollForCompletion(t, model_Name, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
		g.Eventually(func() string {
			return nodes[0].PoolRefs[0].Name
		}, 5*time.Second).Should(gomega.Equal("global--foo.com/bar--default--ingress-nohost"))
		g.Expect(nodes[0].PoolRefs[0].PriorityLabel).To(gomega.Equal("foo.com/bar"))

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(1))

		pool := nodes[0].PoolGroupRefs[0].Members[0]
		g.Expect(*pool.PoolRef).To(gomega.Equal("/api/pool?name=global--foo.com/bar--default--ingress-nohost"))
		g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
	} else {
		t.Fatalf("Could not find model: %s", model_Name)
	}

	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-nohost", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_Name)
}

func TestNoHostMultiPathIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	model_Name := "admin/Shard-VS---global-6"
	SetUpTestForIngress(t, model_Name)

	ingrFake := (FakeIngress{
		Name:        "nohost-multipath",
		Namespace:   "default",
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
	}).IngressNoHost()

	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	PollForCompletion(t, model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))

		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolRefs {
			if strings.Contains(pool.Name, "nohost-multipath.default.avi.internal/foo--default--nohost-multipath") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("nohost-multipath.default.avi.internal/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if strings.Contains(pool.Name, "nohost-multipath.default.avi.internal/bar--default--nohost-multipath") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("nohost-multipath.default.avi.internal/bar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=global--nohost-multipath.default.avi.internal/foo--default--nohost-multipath" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("nohost-multipath.default.avi.internal/foo"))
			} else if *pool.PoolRef == "/api/pool?name=global--nohost-multipath.default.avi.internal/bar--default--nohost-multipath" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("nohost-multipath.default.avi.internal/bar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}

	} else {
		t.Fatalf("Could not find model: %s", model_Name)
	}

	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("nohost-multipath", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_Name)
}

func TestEditNoHostMultiPathIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	model_Name := "admin/Shard-VS---global-6"
	SetUpTestForIngress(t, model_Name)

	ingrFake := (FakeIngress{
		Name:        "nohost-multipath",
		Namespace:   "default",
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
	}).IngressNoHost()

	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
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
	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Update(ingrFake)
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	PollForCompletion(t, model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))

		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolRefs {
			if strings.Contains(pool.Name, "nohost-multipath.default.avi.internal/foo--default--nohost-multipath") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("nohost-multipath.default.avi.internal/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if strings.Contains(pool.Name, "nohost-multipath.default.avi.internal/foobar--default--nohost-multipath") {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("nohost-multipath.default.avi.internal/foobar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=global--nohost-multipath.default.avi.internal/foo--default--nohost-multipath" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("nohost-multipath.default.avi.internal/foo"))
			} else if *pool.PoolRef == "/api/pool?name=global--nohost-multipath.default.avi.internal/foobar--default--nohost-multipath" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("nohost-multipath.default.avi.internal/foobar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}

	} else {
		t.Fatalf("Could not find model: %s", model_Name)
	}

	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("nohost-multipath", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_Name)
}
