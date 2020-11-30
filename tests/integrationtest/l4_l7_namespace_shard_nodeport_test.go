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

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func SetUpTestForIngressInNodePortMode(t *testing.T, model_Name string) {
	os.Setenv("SHARD_VS_SIZE", "LARGE")
	os.Setenv("L7_SHARD_SCHEME", "namespace")
	objects.SharedAviGraphLister().Delete(model_Name)
	AddConfigMap()
	CreateSVC(t, "default", "avisvc", corev1.ServiceTypeNodePort, false)
}

func TearDownTestForIngressInNodePortMode(t *testing.T, model_Name string) {
	os.Setenv("SHARD_VS_SIZE", "")
	objects.SharedAviGraphLister().Delete(model_Name)
	DelSVC(t, "default", "avisvc")
}

// TestNodeAddInNodePortMode tests if VRF creation is skipped in NodePort mode for node addition
func TestNodeAddInNodePortMode(t *testing.T) {
	SetNodePortMode()
	defer SetClusterIPMode()
	nodeIP := "10.1.1.2"
	CreateNode(t, "testNodeNP", nodeIP)
	defer DeleteNode(t, "testNodeNP")
	modelName := "admin/global"
	found, _ := objects.SharedAviGraphLister().Get(modelName)
	if found {
		t.Fatalf("Model found for node add %v in NodePort mode.", modelName)
	}
}

// TestSinglePortL4SvcNodePort tests L4 service with single port
func TestSinglePortL4SvcNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	SetNodePortMode()
	defer SetClusterIPMode()
	nodeIP := "10.1.1.2"
	nodePort := int32(31030)
	CreateNode(t, "testNode1", nodeIP)
	defer DeleteNode(t, "testNode1")

	SetUpTestForSvcLB(t)
	found, aviModel := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
	if !found {
		t.Fatalf("Couldn't find model %v", SINGLEPORTMODEL)
	} else {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(nodes).To(gomega.HaveLen(1))
		g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, SINGLEPORTSVC)))
		g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
		g.Expect(nodes[0].EastWest).To(gomega.Equal(false))
		g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

		// Check for the pools
		g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
		g.Expect(nodes[0].PoolRefs[0].Port).To(gomega.Equal(nodePort))
		g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&nodeIP))
		g.Expect(nodes[0].L4PolicyRefs).To(gomega.HaveLen(1))
	}
	// If we transition the service from Loadbalancer to ClusterIP - it should get deleted.
	svcExample := (FakeService{
		Name:         SINGLEPORTSVC,
		Namespace:    NAMESPACE,
		Type:         corev1.ServiceTypeClusterIP,
		ServicePorts: []Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: 8080}},
	}).Service()
	svcExample.ResourceVersion = "2"
	_, err := KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", NAMESPACE, SINGLEPORTSVC)}
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 15*time.Second).Should(gomega.Equal(false))
	// If we transition the service from clusterIP to Loadbalancer - vs should get ceated
	svcExample = (FakeService{
		Name:         SINGLEPORTSVC,
		Namespace:    NAMESPACE,
		Type:         corev1.ServiceTypeLoadBalancer,
		ServicePorts: []Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: 8080, NodePort: 31031}},
	}).Service()
	svcExample.ResourceVersion = "3"
	_, err = KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}

	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 15*time.Second).Should(gomega.Equal(true))
	TearDownTestForSvcLB(t, g)
}

// TestSinglePortL4SvcNodePort tests L4 service with single port
func TestSinglePortL4SvcNodePortWithNodeSelector(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	SetNodePortMode()
	defer SetClusterIPMode()
	// Add node filter labels
	os.Setenv("NODE_KEY", "my-node")

	// Add node
	nodeIP1 := "10.1.1.2"
	CreateNode(t, "testNode1", nodeIP1)
	defer DeleteNode(t, "testNode1")

	SetUpTestForSvcLB(t)
	found, aviModel := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
	if !found {
		t.Fatalf("Couldn't find model %v", SINGLEPORTMODEL)
	} else {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(nodes).To(gomega.HaveLen(1))
		// Check for the pools
		g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
		g.Expect(nodes[0].PoolRefs[0].Servers).To(gomega.HaveLen(0))
	}
	TearDownTestForSvcLB(t, g)

	// Reset the node filter labels, now all the nodes should get selected for backend server which is 1 in test case
	os.Setenv("NODE_KEY", "")
	SetUpTestForSvcLB(t)
	found, aviModel = objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
	if !found {
		t.Fatalf("Couldn't find model %v", SINGLEPORTMODEL)
	} else {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(nodes).To(gomega.HaveLen(1))
		// Check for the pools
		g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
		// there should be one backend server
		g.Expect(nodes[0].PoolRefs[0].Servers).To(gomega.HaveLen(1))
	}

	TearDownTestForSvcLB(t, g)
}

// TestMultiPortL4SvcNodePort tests L4 service with multiple port
func TestMultiPortL4SvcNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := fmt.Sprintf("%s/cluster--%s-%s", AVINAMESPACE, NAMESPACE, MULTIPORTSVC)

	SetUpTestForSvcLBMultiport(t)
	SetNodePortMode()
	defer SetClusterIPMode()
	nodeIP := "10.1.1.2"
	nodePort := int32(31030)
	CreateNode(t, "testNode1", nodeIP)
	defer DeleteNode(t, "testNode1")

	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if !found {
		t.Fatalf("Couldn't find model %v", modelName)
	} else {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(nodes).To(gomega.HaveLen(1))
		g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, MULTIPORTSVC)))
		g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
		g.Expect(nodes[0].EastWest).To(gomega.Equal(false))
		g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

		// Check for the pools
		g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(3))
		for _, node := range nodes[0].PoolRefs {
			// Since there is single node, each pool will have a single server entry which is node ip
			g.Expect(node.Servers).To(gomega.HaveLen(1))
			g.Expect(nodes[0].PoolRefs[0].Port).To(gomega.Equal(nodePort))
			g.Expect(node.Servers[0].Ip.Addr).To(gomega.Equal(&nodeIP))
		}
		g.Expect(nodes[0].ApplicationProfile).To(gomega.Equal(utils.DEFAULT_L4_APP_PROFILE))
		g.Expect(nodes[0].NetworkProfile).To(gomega.Equal(utils.TCP_NW_FAST_PATH))
		g.Expect(nodes[0].L4PolicyRefs).To(gomega.HaveLen(1))
	}

	TearDownTestForSvcLBMultiport(t, g)
}

// TestMultiVSIngressInNodePort tests the multiple vs ingresses backed by a nodeport service.
func TestMultiVSIngressInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	SetNodePortMode()
	defer SetClusterIPMode()
	model_Name := "admin/cluster--Shared-L7-6"
	SetUpTestForIngressInNodePortMode(t, model_Name)
	nodeIP := "10.1.1.2"
	nodePort := int32(31030)
	CreateNode(t, "testNode1", nodeIP)
	defer DeleteNode(t, "testNode1")

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
				// check if the port is set to nodeport and backed serverIP is pointing to nodeIP
				g.Expect(pool.Port).To(gomega.Equal(nodePort))
				g.Expect(pool.Servers[0].Ip.Addr).To(gomega.Equal(&nodeIP))
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

	TearDownTestForIngressInNodePortMode(t, model_Name)
}

func TestNoHostIngressInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	SetNodePortMode()
	defer SetClusterIPMode()
	model_Name := "admin/cluster--Shared-L7-6"
	SetUpTestForIngressInNodePortMode(t, model_Name)
	nodeIP := "10.1.1.2"
	CreateNode(t, "testNode1", nodeIP)
	defer DeleteNode(t, "testNode1")

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

	TearDownTestForIngressInNodePortMode(t, model_Name)
}

func TestNoHostIngressInNodePortWithMultiTenantEnabled(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// enable tenants per cluster
	SetAkoTenant()
	defer ResetAkoTenant()

	// set nodeport mode
	SetNodePortMode()
	defer SetClusterIPMode()
	modelName := fmt.Sprintf("%s/cluster--Shared-L7-6", AKOTENANT)
	SetUpTestForIngressInNodePortMode(t, modelName)
	nodeIP := "10.1.1.2"
	CreateNode(t, "testNode1", nodeIP)
	defer DeleteNode(t, "testNode1")

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

	PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		// Tenant should be akotenant instead of admin
		g.Expect(nodes[0].Tenant).To(gomega.Equal(AKOTENANT))
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
		t.Fatalf("Could not find model: %s", modelName)
	}

	err = KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "ingress-nohost", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngressInNodePortMode(t, modelName)
}
