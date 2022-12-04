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
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// TestNodeAddInNodePortMode tests if VRF creation is skipped in NodePort mode for node addition
func TestNodeAddInNodePortMode(t *testing.T) {
	SetNodePortMode()
	defer SetClusterIPMode()
	nodeIP := "10.1.1.2"
	CreateNode(t, "testNode1", nodeIP)
	defer DeleteNode(t, "testNode1")
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
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, SINGLEPORTSVC)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

	// Check for the pools
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PoolRefs[0].Port).To(gomega.Equal(nodePort))
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&nodeIP))
	g.Expect(nodes[0].L4PolicyRefs).To(gomega.HaveLen(1))
	// If we transition the service from Loadbalancer to ClusterIP - it should get deleted.
	svcExample := (FakeService{
		Name:         SINGLEPORTSVC,
		Namespace:    NAMESPACE,
		Type:         corev1.ServiceTypeClusterIP,
		ServicePorts: []Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
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
		ServicePorts: []Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080), NodePort: 31031}},
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

func TestSinglePortL4SvcSkipNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	SetNodePortMode()
	defer SetClusterIPMode()
	nodeIP := "10.1.1.2"
	nodePort := int32(31030)
	CreateNode(t, "testNode1", nodeIP)
	defer DeleteNode(t, "testNode1")
	SetUpTestForSvcLB(t)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, SINGLEPORTSVC)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

	// Check for the pools
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PoolRefs[0].Port).To(gomega.Equal(nodePort))
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&nodeIP))
	g.Expect(nodes[0].L4PolicyRefs).To(gomega.HaveLen(1))

	skipNodePort := make(map[string]string)
	skipNodePort["skipnodeport.ako.vmware.com/enabled"] = "true"

	svcExample := (FakeService{
		Name:         SINGLEPORTSVC,
		Namespace:    NAMESPACE,
		Type:         corev1.ServiceTypeLoadBalancer,
		ServicePorts: []Serviceport{{PortName: "foo0", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080), NodePort: 31031}},
		Annotations:  skipNodePort,
	}).Service()
	svcExample.ResourceVersion = "3"
	_, err := KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	address := "1.1.1.1"
	g.Eventually(func() *string {
		_, model := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
		nodes := model.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].PoolRefs[0].Servers[0].Ip.Addr
	}, 25*time.Second).Should(gomega.Equal(&address))
	// Reset the annotation
	skipNodePort = nil
	svcExample = (FakeService{
		Name:         SINGLEPORTSVC,
		Namespace:    NAMESPACE,
		Type:         corev1.ServiceTypeLoadBalancer,
		ServicePorts: []Serviceport{{PortName: "foo0", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080), NodePort: 31031}},
		Annotations:  skipNodePort,
	}).Service()
	svcExample.ResourceVersion = "4"
	_, err = KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	g.Eventually(func() *string {
		_, model := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
		nodes := model.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].PoolRefs[0].Servers[0].Ip.Addr
	}, 25*time.Second).Should(gomega.Equal(&nodeIP))

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
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	// Check for the pools
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PoolRefs[0].Servers).To(gomega.HaveLen(0))

	TearDownTestForSvcLB(t, g)
	os.Setenv("NODE_KEY", "")
	//Commenting out this code: As nodes are now filtered out during ako boot.
	//We need to take care this testing in FT as it requirs AKO reboot to re-populate all nodes.
	/*
		// Reset the node filter labels, now all the nodes should get selected for backend server which is 1 in test case
		os.Setenv("NODE_KEY", "")
		SetUpTestForSvcLB(t)
		g.Eventually(func() bool {
			found, _ := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
			return found
		}, 10*time.Second).Should(gomega.Equal(true))
		_, aviModel = objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(nodes).To(gomega.HaveLen(1))
		// Check for the pools
		g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
		// there should be one backend server
		g.Expect(nodes[0].PoolRefs[0].Servers).To(gomega.HaveLen(1))

		TearDownTestForSvcLB(t, g)
	*/
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

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, MULTIPORTSVC)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
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

	TearDownTestForSvcLBMultiport(t, g)
}
