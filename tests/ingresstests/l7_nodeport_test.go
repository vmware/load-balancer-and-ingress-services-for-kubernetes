/*
 * Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
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

package ingresstests

import (
	"context"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func SetUpTestForIngressInNodePortMode(t *testing.T, svcName, model_Name, externalTrafficPolicy string) {
	objects.SharedAviGraphLister().Delete(model_Name)
	if externalTrafficPolicy == "" {
		integrationtest.CreateSVC(t, "default", svcName, corev1.ProtocolTCP, corev1.ServiceTypeNodePort, false)
	} else {
		integrationtest.CreateSvcWithExternalTrafficPolicy(t, "default", svcName, corev1.ProtocolTCP, corev1.ServiceTypeNodePort, false, externalTrafficPolicy)
	}
}

func TearDownTestForIngressInNodePortMode(t *testing.T, svcName, model_Name string) {
	objects.SharedAviGraphLister().Delete(model_Name)
	integrationtest.DelSVC(t, "default", svcName)
}

// TestL7ModelInNodePort checks if models are not updated for svc if its not referred in ingress.
func TestL7ModelInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	nodeName := "testNodeNP"
	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, nodeName, nodeIP)
	defer integrationtest.DeleteNode(t, nodeName)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	ingName := objNameMap.GenerateName("foo-with-targets")
	SetUpTestForIngressInNodePortMode(t, svcName, modelName, "")

	integrationtest.PollForCompletion(t, modelName, 5)
	found, _ := objects.SharedAviGraphLister().Get(modelName)
	if found {
		// We shouldn't get an update for this update since it neither belongs to an ingress nor a L4 LB service
		t.Fatalf("Couldn't find Model for DELETE event %v", modelName)
	}
	ingrFake := (integrationtest.FakeIngress{
		Name:        ingName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
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
	g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
	// pool server is added for testNodeNP node even though endpointslice/endpoint does not exist
	g.Eventually(func() int {
		return len(nodes[0].PoolRefs[0].Servers)
	}, 30*time.Second).Should(gomega.Equal(1))
	g.Expect(*nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(nodeIP))
	g.Expect(len(nodes[0].PoolRefs[0].NetworkPlacementSettings)).To(gomega.Equal(1))
	_, ok := nodes[0].PoolRefs[0].NetworkPlacementSettings["net123"]
	g.Expect(ok).To(gomega.Equal(true))
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngressInNodePortMode(t, svcName, modelName)
}

// TestL7ModelInNodePortExternalTrafficPolicyLocal checks if pool servers are populated in model only for nodes that are running the app pod.
func TestL7ModelInNodePortExternalTrafficPolicyLocal(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	nodeName := "testNodeNP"
	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, nodeName, nodeIP)
	defer integrationtest.DeleteNode(t, nodeName)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	ingName := objNameMap.GenerateName("foo-with-targets")
	SetUpTestForIngressInNodePortMode(t, svcName, modelName, "Local")

	integrationtest.PollForCompletion(t, modelName, 5)
	found, _ := objects.SharedAviGraphLister().Get(modelName)
	if found {
		// We shouldn't get an update for this update since it neither belongs to an ingress nor a L4 LB service
		t.Fatalf("Couldn't find Model for DELETE event %v", modelName)
	}
	ingrFake := (integrationtest.FakeIngress{
		Name:        ingName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
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
	g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
	// No pool server is added as endpointslice/endpoint does not exist
	g.Expect(nodes[0].PoolRefs[0].Servers).To(gomega.HaveLen(0))
	g.Expect(len(nodes[0].PoolRefs[0].NetworkPlacementSettings)).To(gomega.Equal(1))
	_, ok := nodes[0].PoolRefs[0].NetworkPlacementSettings["net123"]
	g.Expect(ok).To(gomega.Equal(true))

	integrationtest.CreateEPSNodeName(t, "default", svcName, false, false, "1.1.1", nodeName)
	// After creating the endpointslice/endpoint, pool server should be added for testNodeNP node
	g.Eventually(func() int {
		return len(nodes[0].PoolRefs[0].Servers)
	}, 30*time.Second).Should(gomega.Equal(1))
	g.Expect(*nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(nodeIP))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DelEPS(t, "default", svcName)
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngressInNodePortMode(t, svcName, modelName)
}

// TestMultiIngressToSameSvcInNodePort tests if clusterIP is ignored in nodeport mode.
// there will be no backend servers for all the pools created for this ingress
func TestMultiIngressToSameClusterIPSvcInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	ingName := objNameMap.GenerateName("foo-with-targets")
	ingName2 := objNameMap.GenerateName("foo-with-targets")
	objects.SharedAviGraphLister().Delete(modelName)
	svcExample := (integrationtest.FakeService{
		Name:         svcName,
		Namespace:    "default",
		Type:         corev1.ServiceTypeClusterIP,
		ServicePorts: []integrationtest.Serviceport{{PortName: "foo", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()

	_, err := KubeClient.CoreV1().Services("default").Create(context.TODO(), svcExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	integrationtest.CreateEPS(t, "default", svcName, false, false, "1.2.3.4")
	ingrFake1 := (integrationtest.FakeIngress{
		Name:        ingName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
	}).Ingress()

	_, err = KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	ingrFake2 := (integrationtest.FakeIngress{
		Name:        ingName2,
		Namespace:   "default",
		DnsNames:    []string{"bar.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
	}).Ingress()

	_, err = KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake2, metav1.CreateOptions{})
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
			if pool.Name == "cluster--foo.com_foo-default-"+ingName {
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
	integrationtest.DelEPS(t, "default", svcName)
	err = KubeClient.CoreV1().Services("default").Delete(context.TODO(), svcName, metav1.DeleteOptions{})
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
	integrationtest.CreateEPS(t, "default", svcName, false, false, "1.1.1")
	//====== VERIFICATION OF ONE INGRESS DELETE
	// Now let's delete one ingress and expect the update for that.
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingName, metav1.DeleteOptions{})
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

	integrationtest.DelEPS(t, "default", svcName)

	err = KubeClient.CoreV1().Services("default").Delete(context.TODO(), svcName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Service %v", err)
	}
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingName2, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
}

// TestMultiIngressToSameNodePortSvcInNodePort tests if multiple ingresses referring to same nodeport service
// nodeIP should be set in backend server, and pool's port is set to nodePort.
func TestMultiIngressToSameNodePortSvcInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	nodePort := int32(31030)
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	ingName := objNameMap.GenerateName("foo-with-targets")
	ingName2 := objNameMap.GenerateName("foo-with-targets")
	objects.SharedAviGraphLister().Delete(modelName)
	svcExample := (integrationtest.FakeService{
		Name:         svcName,
		Namespace:    "default",
		Type:         corev1.ServiceTypeNodePort,
		ServicePorts: []integrationtest.Serviceport{{PortName: "foo", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080), NodePort: nodePort}},
	}).Service()

	_, err := KubeClient.CoreV1().Services("default").Create(context.TODO(), svcExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	// Endpoint is not created in NodePort Mode.
	ingrFake1 := (integrationtest.FakeIngress{
		Name:        ingName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
	}).Ingress()

	_, err = KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	ingrFake2 := (integrationtest.FakeIngress{
		Name:        ingName2,
		Namespace:   "default",
		DnsNames:    []string{"bar.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
	}).Ingress()

	_, err = KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake2, metav1.CreateOptions{})
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
			if pool.Name == "cluster--foo.com_foo-default-"+ingName {
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
	err = KubeClient.CoreV1().Services("default").Delete(context.TODO(), svcName, metav1.DeleteOptions{})
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
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingName, metav1.DeleteOptions{})
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

	err = KubeClient.CoreV1().Services("default").Delete(context.TODO(), svcName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Service %v", err)
	}
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingName2, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
}

// TestMultiVSIngressInNodePort tests multiple ingresses creation
// nodeIP should be set in backend server, and pool's port is set to nodePort.
func TestMultiVSIngressInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	nodePort := int32(31030)
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	ingName := objNameMap.GenerateName("foo-with-targets")
	SetUpTestForIngressInNodePortMode(t, svcName, modelName, "")

	ingrFake := (integrationtest.FakeIngress{
		Name:        ingName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
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
	randomNamespace := "randomNamespacethatyeildsdiff"
	integrationtest.AddDefaultNamespace(randomNamespace)
	_, err = KubeClient.CoreV1().Namespaces().Create(context.TODO(), &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: randomNamespace}}, metav1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		t.Fatalf("error creating namespace: %s", randomNamespace)
	}
	randoming := (integrationtest.FakeIngress{
		Name:        randomNamespace,
		Namespace:   randomNamespace,
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
	}).Ingress()
	_, err = KubeClient.NetworkingV1().Ingresses(randomNamespace).Create(context.TODO(), randoming, metav1.CreateOptions{})

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
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				return len(nodes[0].PoolRefs)
			}
			return 0
		}, 10*time.Second).Should(gomega.Equal(2))

	} else {
		t.Fatalf("Could not find model: %v", err)
	}
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	err = KubeClient.NetworkingV1().Ingresses(randomNamespace).Delete(context.TODO(), randomNamespace, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngressInNodePortMode(t, svcName, modelName)
}

// TestMultipleNodeCreationAndDeletionInNodePort tests addition of node and its effect of pool server and deletion of node.
func TestMultipleNodeCreationAndDeletionInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP1 := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP1", nodeIP1)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngressInNodePortMode(t, svcName, modelName, "")

	ingrFake1 := (integrationtest.FakeIngress{
		Name:        "ingress-multi1",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
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
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-multi1", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)
	TearDownTestForIngressInNodePortMode(t, svcName, modelName)

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

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngressInNodePortMode(t, svcName, modelName, "")

	ingrFake := (integrationtest.FakeIngress{
		Name:        "ingress-multipath",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: svcName,
	}).IngressMultiPath()

	_, err = KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
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
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-multipath", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngressInNodePortMode(t, svcName, modelName)
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

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	objects.SharedAviGraphLister().Delete(modelName)
	integrationtest.CreateSVC(t, "default", svcName, corev1.ProtocolTCP, corev1.ServiceTypeNodePort, true)
	ingrFake := (integrationtest.FakeIngress{
		Name:        "ingress-multipath",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}).IngressMultiPort()

	_, err = KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
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
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-multipath", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngressInNodePortMode(t, svcName, modelName)
}

func TestDeleteServiceInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngressInNodePortMode(t, svcName, modelName, "")

	ingrFake1 := (integrationtest.FakeIngress{
		Name:        "ingress-multi1",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake2 := (integrationtest.FakeIngress{
		Name:        "ingress-multi2",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/bar"},
		ServiceName: svcName,
	}).Ingress()

	_, err = KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake2, metav1.CreateOptions{})
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
	integrationtest.DelSVC(t, "default", svcName)
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
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-multi1", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 1)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--foo.com_bar-default-ingress-multi2"))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-multi2", metav1.DeleteOptions{})
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

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngressInNodePortMode(t, svcName, modelName, "")

	ingrFake1 := (integrationtest.FakeIngress{
		Name:        "ingress-multi1",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
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
		Name:     nodeName,
		PodCIDR:  "10.244.0.0/24",
		PodCIDRs: []string{"10.244.0.0/24"},
		Version:  "1",
		NodeIP:   nodeIP,
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
		Name:     nodeName,
		PodCIDR:  "10.244.0.0/24",
		PodCIDRs: []string{"10.244.0.0/24"},
		Version:  "1",
		NodeIP:   "10.1.1.3",
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
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-multi1", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 1)
	TearDownTestForIngressInNodePortMode(t, svcName, modelName)

}

func TestDeleteNodeInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngressInNodePortMode(t, svcName, modelName, "")

	ingrFake1 := (integrationtest.FakeIngress{
		Name:        "ingress-multi1",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake2 := (integrationtest.FakeIngress{
		Name:        "ingress-multi2",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/bar"},
		ServiceName: svcName,
	}).Ingress()

	_, err = KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake2, metav1.CreateOptions{})
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
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-multi1", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 1)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--foo.com_bar-default-ingress-multi2"))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-multi2", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)
	TearDownTestForIngressInNodePortMode(t, svcName, modelName)

}

func TestFullSyncCacheNoOpInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")
	secretName := objNameMap.GenerateName("my-secret")

	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngressInNodePortMode(t, svcName, modelName, "")
	//create multipath ingress with tls secret
	ingrFake1 := (integrationtest.FakeIngress{
		Name:      "ingress-fsno",
		Namespace: "default",
		DnsNames:  []string{"foo.com"},
		Ips:       []string{"8.8.8.8"},
		Paths:     []string{"/foo", "/bar"},
		HostNames: []string{"v1"},
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
		ServiceName: svcName,
	}).IngressMultiPath()
	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
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
	ctrl.FullSyncK8s(true)

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

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-fsno", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), secretName, metav1.DeleteOptions{})
	VerifySNIIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngressInNodePortMode(t, svcName, modelName)
}

func TestL7ModelMultiSNIInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")
	secretName := objNameMap.GenerateName("my-secret")

	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	ingName := objNameMap.GenerateName("foo-with-targets")
	SetUpTestForIngressInNodePortMode(t, svcName, modelName, "")

	ingrFake := (integrationtest.FakeIngress{
		Name:        ingName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "bar.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com", "bar.com"},
		},
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
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

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), secretName, metav1.DeleteOptions{})
	VerifySNIIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngressInNodePortMode(t, svcName, modelName)
}
