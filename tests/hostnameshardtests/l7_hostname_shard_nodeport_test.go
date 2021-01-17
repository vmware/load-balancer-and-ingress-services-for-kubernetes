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

package hostnameshardtests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func SetUpTestForIngressInNodePortMode(t *testing.T, model_Name string) {
	os.Setenv("SHARD_VS_SIZE", "LARGE")
	os.Setenv("L7_SHARD_SCHEME", "hostname")
	objects.SharedAviGraphLister().Delete(model_Name)
	integrationtest.CreateSVC(t, "default", "avisvc", corev1.ServiceTypeNodePort, false)
}

func TearDownTestForIngressInNodePortMode(t *testing.T, model_Name string) {
	os.Setenv("SHARD_VS_SIZE", "")
	objects.SharedAviGraphLister().Delete(model_Name)
	integrationtest.DelSVC(t, "default", "avisvc")
}

// TestL7ModelInNodePort checks if models are not updated for svc if its not referred in ingress.
func TestL7ModelInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	modelName := "admin/cluster--Shared-L7-0"
	SetUpTestForIngressInNodePortMode(t, modelName)

	integrationtest.PollForCompletion(t, modelName, 5)
	found, _ := objects.SharedAviGraphLister().Get(modelName)
	if found {
		// We shouldn't get an update for this update since it neither belongs to an ingress nor a L4 LB service
		t.Fatalf("Couldn't find Model for DELETE event %v", modelName)
	}
	ingrFake := (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 5*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))

	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "foo-with-targets", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngressInNodePortMode(t, modelName)
}

// TestMultiIngressToSameSvcInNodePort tests if clusterIP is ignored in nodeport mode.
// there will be no backend servers for all the pools created for this ingress
func TestMultiIngressToSameClusterIPSvcInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	os.Setenv("SHARD_VS_SIZE", "LARGE")

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	modelName := "admin/cluster--Shared-L7-0"
	objects.SharedAviGraphLister().Delete(modelName)
	svcExample := (integrationtest.FakeService{
		Name:         "avisvc",
		Namespace:    "default",
		Type:         corev1.ServiceTypeClusterIP,
		ServicePorts: []integrationtest.Serviceport{{PortName: "foo", Protocol: "TCP", PortNumber: 8080, TargetPort: 8080}},
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
	ingrFake1 := (integrationtest.FakeIngress{
		Name:        "foo-with-targets1",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err = KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
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

	_, err = KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(nodes[0].SharedVS).To(gomega.Equal(true))
		dsNodes := aviModel.(*avinodes.AviObjectGraph).GetAviHTTPDSNode()
		g.Expect(len(dsNodes)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
		for _, pool := range nodes[0].PoolRefs {
			if pool.Name == "cluster--foo.com_foo-default-foo-with-targets1" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				// since the service is cluster IP the backend servers are not added
				g.Expect(len(pool.Servers)).To(gomega.Equal(0))
			} else {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("bar.com/foo"))
				// since the service is cluster IP the backend servers are not added
				g.Expect(len(pool.Servers)).To(gomega.Equal(0))
			}
		}
		// Delete the model.
		objects.SharedAviGraphLister().Delete(modelName)
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
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		dsNodes := aviModel.(*avinodes.AviObjectGraph).GetAviHTTPDSNode()
		g.Expect(len(dsNodes)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))

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
	integrationtest.DetectModelChecksumChange(t, modelName, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(0))

		// Delete the model.
		objects.SharedAviGraphLister().Delete(modelName)
	} else {
		t.Fatalf("Could not find model on ingress delete: %v", err)
	}
	//====== VERIFICATION OF SERVICE ADD
	// Let's add the service back now - the ingress's associated with this service should be returned
	_, err = KubeClient.CoreV1().Services("default").Create(context.TODO(), svcExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	modelName = "admin/cluster--Shared-L7-1"
	integrationtest.PollForCompletion(t, modelName, 5)
	// We should be able to get one model now in the queue
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))

		objects.SharedAviGraphLister().Delete(modelName)
	} else {
		t.Fatalf("Could not find model on service ADD: %v", err)
	}

	err = KubeClient.CoreV1().Endpoints("default").Delete(context.TODO(), "avisvc", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Endpoint %v", err)
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

// TestMultiIngressToSameNodePortSvcInNodePort tests if multiple ingresses referring to same nodeport service
// nodeIP should be set in backed server, and pool's port is set to nodePort.
func TestMultiIngressToSameNodePortSvcInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	os.Setenv("SHARD_VS_SIZE", "LARGE")

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	nodePort := int32(31030)
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	modelName := "admin/cluster--Shared-L7-0"
	objects.SharedAviGraphLister().Delete(modelName)
	svcExample := (integrationtest.FakeService{
		Name:         "avisvc",
		Namespace:    "default",
		Type:         corev1.ServiceTypeNodePort,
		ServicePorts: []integrationtest.Serviceport{{PortName: "foo", Protocol: "TCP", PortNumber: 8080, TargetPort: 8080, NodePort: nodePort}},
	}).Service()

	_, err := KubeClient.CoreV1().Services("default").Create(context.TODO(), svcExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	// Endpoint is not created in NodePort Mode.
	ingrFake1 := (integrationtest.FakeIngress{
		Name:        "foo-with-targets1",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err = KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
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

	_, err = KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(nodes[0].SharedVS).To(gomega.Equal(true))
		dsNodes := aviModel.(*avinodes.AviObjectGraph).GetAviHTTPDSNode()
		g.Expect(len(dsNodes)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
		for _, pool := range nodes[0].PoolRefs {
			// validate if the pool port is nodeport
			g.Expect(pool.Port).To(gomega.Equal(nodePort))
			if pool.Name == "cluster--foo.com_foo-default-foo-with-targets1" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				// since the service is NodePort type the backend servers  added and its nodeIP
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
				g.Expect(pool.Servers[0].Ip.Addr).To(gomega.Equal(&nodeIP))
			} else {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("bar.com/foo"))
				// since the service is NodePort type the backend servers  added and its nodeIP
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
				g.Expect(pool.Servers[0].Ip.Addr).To(gomega.Equal(&nodeIP))
			}
		}
		// Delete the model.
		objects.SharedAviGraphLister().Delete(modelName)
	} else {
		t.Fatalf("Could not find model on ingress delete: %v", err)
	}
	//====== VERIFICATION OF SERVICE DELETE
	// Now we have cleared the layer 2 queue for both the models. Let's delete the service.
	err = KubeClient.CoreV1().Services("default").Delete(context.TODO(), "avisvc", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Service %v", err)
	}
	// We should be able to get one model now in the queue
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		dsNodes := aviModel.(*avinodes.AviObjectGraph).GetAviHTTPDSNode()
		g.Expect(len(dsNodes)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))

	} else {
		t.Fatalf("Could not find model on service delete: %v", err)
	}

	//====== VERIFICATION OF ONE INGRESS DELETE
	// Now let's delete one ingress and expect the update for that.
	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "foo-with-targets1", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DetectModelChecksumChange(t, modelName, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(0))

		// Delete the model.
		objects.SharedAviGraphLister().Delete(modelName)
	} else {
		t.Fatalf("Could not find model on ingress delete: %v", err)
	}
	//====== VERIFICATION OF SERVICE ADD
	// Let's add the service back now - the ingress's associated with this service should be returned
	_, err = KubeClient.CoreV1().Services("default").Create(context.TODO(), svcExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	modelName = "admin/cluster--Shared-L7-1"
	integrationtest.PollForCompletion(t, modelName, 5)
	// We should be able to get one model now in the queue
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))

		objects.SharedAviGraphLister().Delete(modelName)
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

// TestMultiVSIngressInNodePort tests multiple ingresses creation
// nodeIP should be set in backed server, and pool's port is set to nodePort.
func TestMultiVSIngressInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	nodePort := int32(31030)
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	modelName := "admin/cluster--Shared-L7-0"
	SetUpTestForIngressInNodePortMode(t, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		dsNodes := aviModel.(*avinodes.AviObjectGraph).GetAviHTTPDSNode()
		g.Expect(len(dsNodes)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
		for _, pool := range nodes[0].PoolRefs {
			g.Expect(pool.Port).To(gomega.Equal(nodePort))
			if pool.Name == "cluster--foo.com_foo" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
				g.Expect(pool.Servers[0].Ip.Addr).To(gomega.Equal(&nodeIP))
			}
		}
	} else {
		t.Fatalf("Could not find model: %v", err)
	}
	randoming := (integrationtest.FakeIngress{
		Name:        "randomNamespacethatyeildsdiff",
		Namespace:   "randomNamespacethatyeildsdiff",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()
	_, err = KubeClient.NetworkingV1beta1().Ingresses("randomNamespacethatyeildsdiff").Create(context.TODO(), randoming, metav1.CreateOptions{})
	integrationtest.PollForCompletion(t, modelName, 10)
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		dsNodes := aviModel.(*avinodes.AviObjectGraph).GetAviHTTPDSNode()
		g.Expect(len(dsNodes)).To(gomega.Equal(1))
		g.Eventually(func() int {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return len(nodes[0].PoolRefs)
		}, 10*time.Second).Should(gomega.Equal(2))

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
	err = KubeClient.NetworkingV1beta1().Ingresses("randomNamespacethatyeildsdiff").Delete(context.TODO(), "randomNamespacethatyeildsdiff", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngressInNodePortMode(t, modelName)
}

// TestMultipleNodeCreationAndDeletionInNodePort tests addition of node and its effect of pool server and deletion of node.
func TestMultipleNodeCreationAndDeletionInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP1 := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP1", nodeIP1)

	modelName := "admin/cluster--Shared-L7-0"
	SetUpTestForIngressInNodePortMode(t, modelName)

	ingrFake1 := (integrationtest.FakeIngress{
		Name:        "ingress-multi1",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
		for _, pool := range nodes[0].PoolRefs {
			if pool.Name == "cluster--foo.com_foo-default-ingress-multi1" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
				g.Expect(pool.Servers[0].Ip.Addr).To(gomega.Equal(&nodeIP1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(1))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=cluster--foo.com_foo-default-ingress-multi1" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	// Add Another Node and check if pool server got added.
	nodeIP2 := "10.1.1.20"
	integrationtest.CreateNode(t, "testNodeNP2", nodeIP2)
	defer integrationtest.DeleteNode(t, "testNodeNP2")
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
		for _, pool := range nodes[0].PoolRefs {
			if pool.Name == "cluster--foo.com_foo-default-ingress-multi1" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				// Check if the pool server got added
				g.Expect(len(pool.Servers)).To(gomega.Equal(2))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}

	// Delete the Node1 and check if the pool server gets deleted, also check if the only pool remaining should have server ip address as NodeIP2
	integrationtest.DeleteNode(t, "testNodeNP1")
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
		for _, pool := range nodes[0].PoolRefs {
			if pool.Name == "cluster--foo.com_foo-default-ingress-multi1" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				// Check if the pool server got added
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
				g.Expect(pool.Servers[0].Ip.Addr).To(gomega.Equal(&nodeIP2))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "ingress-multi1", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)
	TearDownTestForIngressInNodePortMode(t, modelName)

}

func TestMultiPathIngressInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var err error

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	nodePort := int32(31030)
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	modelName := "admin/cluster--Shared-L7-0"
	SetUpTestForIngressInNodePortMode(t, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        "ingress-multipath",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
	}).IngressMultiPath()

	_, err = KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolRefs {
			g.Expect(pool.Port).To(gomega.Equal(nodePort))
			if pool.Name == "cluster--foo.com_foo-default-ingress-multipath" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
				g.Expect(pool.Servers[0].Ip.Addr).To(gomega.Equal(&nodeIP))

			} else if pool.Name == "cluster--foo.com_bar-default-ingress-multipath" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
				g.Expect(pool.Servers[0].Ip.Addr).To(gomega.Equal(&nodeIP))

			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=cluster--foo.com_foo-default-ingress-multipath" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			} else if *pool.PoolRef == "/api/pool?name=cluster--foo.com_bar-default-ingress-multipath" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "ingress-multipath", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngressInNodePortMode(t, modelName)
}

func TestMultiPortServiceIngressInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var err error

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	nodePort := int32(31030)
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	modelName := "admin/cluster--Shared-L7-0"
	objects.SharedAviGraphLister().Delete(modelName)
	integrationtest.CreateSVC(t, "default", "avisvc", corev1.ServiceTypeNodePort, true)
	ingrFake := (integrationtest.FakeIngress{
		Name:        "ingress-multipath",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).Ingress(true)

	_, err = KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		for k, pool := range nodes[0].PoolRefs {
			// irrespective of which pool the port is always nodeport
			g.Expect(pool.Port).To(gomega.Equal(nodePort + int32(k)))
			if pool.Name == "cluster--foo.com_foo-default-ingress-multipath" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				// In clusterIP case this would have been 3 but in case of NodePort this is 1 as single node is present in system
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if pool.Name == "cluster--foo.com_bar-default-ingress-multipath" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=cluster--foo.com_foo-default-ingress-multipath" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			} else if *pool.PoolRef == "/api/pool?name=cluster--foo.com_bar-default-ingress-multipath" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "ingress-multipath", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngressInNodePortMode(t, modelName)
}

func TestDeleteServiceInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	modelName := "admin/cluster--Shared-L7-0"
	SetUpTestForIngressInNodePortMode(t, modelName)

	ingrFake1 := (integrationtest.FakeIngress{
		Name:        "ingress-multi1",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake2 := (integrationtest.FakeIngress{
		Name:        "ingress-multi2",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/bar"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err = KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolRefs {
			if pool.Name == "cluster--foo.com_foo-default-ingress-multi1" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if pool.Name == "cluster--foo.com_bar-default-ingress-multi2" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=cluster--foo.com_foo-default-ingress-multi1" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			} else if *pool.PoolRef == "/api/pool?name=cluster--foo.com_bar-default-ingress-multi2" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	// Delete the service
	integrationtest.DelSVC(t, "default", "avisvc")
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))

		g.Eventually(func() int {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return len(nodes[0].PoolRefs[0].Servers)
		}, 10*time.Second).Should(gomega.Equal(0))

		g.Eventually(func() int {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes[0].PoolRefs) > 1 {
				return len(nodes[0].PoolRefs[1].Servers)
			}
			return 1
		}, 10*time.Second).Should(gomega.Equal(0))

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "ingress-multi1", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 1)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--foo.com_bar-default-ingress-multi2"))

	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "ingress-multi2", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)
	objects.SharedAviGraphLister().Delete(modelName)
}

func TestUpdateNodeInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	nodeName := "testNodeNP"
	integrationtest.CreateNode(t, nodeName, nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	modelName := "admin/cluster--Shared-L7-0"
	SetUpTestForIngressInNodePortMode(t, modelName)

	ingrFake1 := (integrationtest.FakeIngress{
		Name:        "ingress-multi1",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		for _, pool := range nodes[0].PoolRefs {
			if pool.Name == "cluster--foo.com_foo-default-ingress-multi1" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	// Update the Node's resource version
	objects.SharedAviGraphLister().Delete(modelName)
	nodeExample := (integrationtest.FakeNode{
		Name:    nodeName,
		PodCIDR: "10.244.0.0/24",
		Version: "1",
		NodeIP:  nodeIP,
	}).Node()
	nodeExample.ResourceVersion = "2"

	_, err = KubeClient.CoreV1().Nodes().Update(context.TODO(), nodeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	found, _ = objects.SharedAviGraphLister().Get(modelName)
	if found {
		// model should not be sent for this update as only the resource ver is changed.
		t.Fatalf("Model found for node add %v", modelName)
	}

	// Update the Node's IP
	objects.SharedAviGraphLister().Delete(modelName)
	nodeExample = (integrationtest.FakeNode{
		Name:    nodeName,
		PodCIDR: "10.244.0.0/24",
		Version: "1",
		NodeIP:  "10.1.1.3",
	}).Node()
	nodeExample.ResourceVersion = "3"

	_, err = KubeClient.CoreV1().Nodes().Update(context.TODO(), nodeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		for _, pool := range nodes[0].PoolRefs {
			if pool.Name == "cluster--foo.com_foo-default-ingress-multi1" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "ingress-multi1", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 1)
	TearDownTestForIngressInNodePortMode(t, modelName)

}

func TestDeleteNodeInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)

	modelName := "admin/cluster--Shared-L7-0"
	SetUpTestForIngressInNodePortMode(t, modelName)

	ingrFake1 := (integrationtest.FakeIngress{
		Name:        "ingress-multi1",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake2 := (integrationtest.FakeIngress{
		Name:        "ingress-multi2",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/bar"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err = KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolRefs {
			if pool.Name == "cluster--foo.com_foo-default-ingress-multi1" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if pool.Name == "cluster--foo.com_bar-default-ingress-multi2" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=cluster--foo.com_foo-default-ingress-multi1" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			} else if *pool.PoolRef == "/api/pool?name=cluster--foo.com_bar-default-ingress-multi2" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	// Delete the Node
	integrationtest.DeleteNode(t, "testNodeNP")

	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))

		g.Eventually(func() int {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return len(nodes[0].PoolRefs[0].Servers)
		}, 10*time.Second).Should(gomega.Equal(0))

		g.Eventually(func() int {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes[0].PoolRefs) > 1 {
				return len(nodes[0].PoolRefs[1].Servers)
			}
			return 1
		}, 10*time.Second).Should(gomega.Equal(0))

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "ingress-multi1", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 1)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--foo.com_bar-default-ingress-multi2"))

	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "ingress-multi2", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)
	TearDownTestForIngressInNodePortMode(t, modelName)

}

func TestFullSyncCacheNoOpInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	integrationtest.AddSecret("my-secret", "default", "tlsCert", "tlsKey")
	modelName := "admin/cluster--Shared-L7-0"
	SetUpTestForIngressInNodePortMode(t, modelName)
	//create multipath ingress with tls secret
	ingrFake1 := (integrationtest.FakeIngress{
		Name:      "ingress-fsno",
		Namespace: "default",
		DnsNames:  []string{"foo.com"},
		Ips:       []string{"8.8.8.8"},
		Paths:     []string{"/foo", "/bar"},
		HostNames: []string{"v1"},
		TlsSecretDNS: map[string][]string{
			"my-secret": {"foo.com"},
		},
		ServiceName: "avisvc",
	}).IngressMultiPath()
	_, err := KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}

	//store old chksum
	mcache := cache.SharedAviObjCache()
	oldSniCache, _ := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
	oldSniCacheObj, _ := oldSniCache.(*cache.AviVsCache)
	oldChksum := oldSniCacheObj.CloudConfigCksum

	//call fullsync
	ctrl.FullSync()
	ctrl.FullSyncK8s()

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	integrationtest.PollForCompletion(t, modelName, 5)

	//compare with new chksum
	g.Eventually(func() string {
		mcache := cache.SharedAviObjCache()
		newSniCache, _ := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
		newSniCacheObj, _ := newSniCache.(*cache.AviVsCache)
		if newSniCacheObj != nil {
			return newSniCacheObj.CloudConfigCksum
		}
		return ""
	}, 30*time.Second).Should(gomega.Equal(oldChksum))

	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "ingress-fsno", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), "my-secret", metav1.DeleteOptions{})
	VerifySNIIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngressInNodePortMode(t, modelName)
}

func TestL7ModelMultiSNIInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	integrationtest.AddSecret("my-secret", "default", "tlsCert", "tlsKey")
	modelName := "admin/cluster--Shared-L7-0"
	SetUpTestForIngressInNodePortMode(t, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "bar.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			"my-secret": {"foo.com", "bar.com"},
		},
	}).Ingress()
	_, err := KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(nodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
		g.Expect(len(nodes[0].SniNodes)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].HttpPolicyRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].PoolRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].PoolRefs[0].Servers)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].SSLKeyCertRefs)).To(gomega.Equal(1))
		g.Expect(nodes[0].SniNodes[0].VHDomainNames).To(gomega.HaveLen(1))
	} else {
		t.Fatalf("Could not find Model: %v", err)
	}

	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "foo-with-targets", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), "my-secret", metav1.DeleteOptions{})
	VerifySNIIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngressInNodePortMode(t, modelName)
}
