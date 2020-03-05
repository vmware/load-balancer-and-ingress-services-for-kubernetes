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

package hostnameshardtests

import (
	"os"
	"testing"
	"time"

	"github.com/onsi/gomega"
	avinodes "gitlab.eng.vmware.com/orion/akc/pkg/nodes"
	"gitlab.eng.vmware.com/orion/akc/pkg/objects"
	"gitlab.eng.vmware.com/orion/akc/tests/integrationtest"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMain(m *testing.M) {
	integrationtest.SetUp()
	ret := m.Run()
	os.Exit(ret)
}

func SetUpTestForIngress(t *testing.T, Model_Name string) {
	os.Setenv("SHARD_VS_SIZE", "LARGE")
	os.Setenv("CLOUD_NAME", "Shard-VS-")
	os.Setenv("VRF_CONTEXT", "global")
	os.Setenv("L7_SHARD_SCHEME", "hostname")

	objects.SharedAviGraphLister().Delete(Model_Name)
	integrationtest.CreateSVC(t, "default", "avisvc")
	integrationtest.CreateEP(t, "default", "avisvc")
}

func TearDownTestForIngress(t *testing.T, Model_Name string) {
	os.Setenv("SHARD_VS_SIZE", "")
	os.Setenv("CLOUD_NAME", "")
	os.Setenv("VRF_CONTEXT", "")

	objects.SharedAviGraphLister().Delete(Model_Name)
	integrationtest.DelSVC(t, "default", "avisvc")
	integrationtest.DelEP(t, "default", "avisvc")
}

func VerifyIngressDeletion(t *testing.T, g *gomega.WithT, aviModel interface{}, poolCount int) {
	var nodes []*avinodes.AviVsNode
	g.Eventually(func() []*avinodes.AviPoolNode {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].PoolRefs
	}, 5*time.Second).Should(gomega.HaveLen(poolCount))

	g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(poolCount))
}

func TestL7Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	Model_Name := "admin/Shard-VS---global-0"
	SetUpTestForIngress(t, Model_Name)

	integrationtest.PollForCompletion(t, Model_Name, 5)
	found, _ := objects.SharedAviGraphLister().Get(Model_Name)
	if found {
		// We shouldn't get an update for this update since it neither belongs to an ingress nor a L4 LB service
		t.Fatalf("Couldn't find Model for DELETE event %v", Model_Name)
	}
	ingrFake := (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := integrationtest.KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, Model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(Model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	} else {
		t.Fatalf("Could not find Model: %v", err)
	}
	err = integrationtest.KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("foo-with-targets", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, Model_Name)
}

func TestMultiIngressToSameSvc(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	os.Setenv("SHARD_VS_SIZE", "LARGE")
	os.Setenv("CLOUD_NAME", "Shard-VS-")
	os.Setenv("VRF_CONTEXT", "global")

	model_Name := "admin/Shard-VS---global-0"
	objects.SharedAviGraphLister().Delete(model_Name)
	svcExample := (integrationtest.FakeService{
		Name:         "avisvc",
		Namespace:    "default",
		ServicePorts: []integrationtest.Serviceport{{PortName: "foo", Protocol: "TCP", PortNumber: 8080, TargetPort: 8080}},
	}).Service()

	_, err := integrationtest.KubeClient.CoreV1().Services("default").Create(svcExample)
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
	_, err = integrationtest.KubeClient.CoreV1().Endpoints("default").Create(epExample)
	if err != nil {
		t.Fatalf("error in creating Endpoint: %v", err)
	}
	ingrFake1 := (integrationtest.FakeIngress{
		Name:        "foo-with-targets1",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err = integrationtest.KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake1)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	ingrFake2 := (integrationtest.FakeIngress{
		Name:        "foo-with-targets2",
		Namespace:   "default",
		DnsNames:    []string{"bar.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err = integrationtest.KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake2)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, model_Name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(nodes[0].SharedVS).To(gomega.Equal(true))
		dsNodes := aviModel.(*avinodes.AviObjectGraph).GetAviHTTPDSNode()
		g.Expect(len(dsNodes)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
		for _, pool := range nodes[0].PoolRefs {
			if pool.Name == "foo.com/foo--default--foo-with-targets1" {
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
	err = integrationtest.KubeClient.CoreV1().Endpoints("default").Delete("avisvc", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Endpoint %v", err)
	}
	err = integrationtest.KubeClient.CoreV1().Services("default").Delete("avisvc", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Service %v", err)
	}
	// We should be able to get one model now in the queue
	integrationtest.PollForCompletion(t, model_Name, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		dsNodes := aviModel.(*avinodes.AviObjectGraph).GetAviHTTPDSNode()
		g.Expect(len(dsNodes)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))

	} else {
		t.Fatalf("Could not find model on service delete: %v", err)
	}
	_, err = integrationtest.KubeClient.CoreV1().Endpoints("default").Create(epExample)
	if err != nil {
		t.Fatalf("error in creating Endpoint: %v", err)
	}
	//====== VERIFICATION OF ONE INGRESS DELETE
	// Now let's delete one ingress and expect the update for that.
	err = integrationtest.KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("foo-with-targets1", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DetectModelChecksumChange(t, model_Name, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(model_Name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shard-VS"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(0))

		// Delete the model.
		objects.SharedAviGraphLister().Delete(model_Name)
	} else {
		t.Fatalf("Could not find model on ingress delete: %v", err)
	}
	//====== VERIFICATION OF SERVICE ADD
	// Let's add the service back now - the ingress's associated with this service should be returned
	_, err = integrationtest.KubeClient.CoreV1().Services("default").Create(svcExample)
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	model_Name = "admin/Shard-VS---global-1"
	integrationtest.PollForCompletion(t, model_Name, 5)
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
	err = integrationtest.KubeClient.CoreV1().Endpoints("default").Delete("avisvc", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Endpoint %v", err)
	}
	integrationtest.PollForCompletion(t, model_Name, 5)
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
	err = integrationtest.KubeClient.CoreV1().Services("default").Delete("avisvc", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Service %v", err)
	}
	err = integrationtest.KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("foo-with-targets2", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
}
