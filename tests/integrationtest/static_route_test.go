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

package integrationtest

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"

	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const MODEL_GLOBAL = "admin/global"

func TestNodeAdd(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_GLOBAL
	nodeip := "10.1.1.2"
	objects.SharedAviGraphLister().Delete(modelName)
	nodeExample := (FakeNode{
		Name:     "testNode1",
		PodCIDR:  "10.244.0.0/24",
		PodCIDRs: []string{"10.244.0.0/24"},
		Version:  "1",
		NodeIP:   nodeip,
	}).Node()

	_, err := KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	PollForCompletion(t, modelName, 5)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))

	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(1))
	g.Expect(*(nodes[0].StaticRoutes[0].NextHop.Addr)).To(gomega.Equal(nodeip))
	g.Expect(*(nodes[0].StaticRoutes[0].Prefix.IPAddr.Addr)).To(gomega.Equal("10.244.0.0"))
	g.Expect(*(nodes[0].StaticRoutes[0].Prefix.Mask)).To(gomega.Equal(int32(24)))
}

func TestNodeUpdate(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_GLOBAL
	nodeip := "10.1.1.2"
	objects.SharedAviGraphLister().Delete(modelName)
	nodeExample := (FakeNode{
		Name:     "testNode1",
		PodCIDR:  "10.244.0.0/24",
		PodCIDRs: []string{"10.244.0.0/24"},
		Version:  "1",
		NodeIP:   nodeip,
	}).Node()

	nodeExample.ObjectMeta.ResourceVersion = "2"
	nodeExample.Spec.PodCIDR = "10.245.0.0/24"
	nodeExample.Spec.PodCIDRs = []string{"10.245.0.0/24"}

	_, err := KubeClient.CoreV1().Nodes().Update(context.TODO(), nodeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Node: %v", err)
	}

	PollForCompletion(t, modelName, 5)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))

	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(1))
	g.Expect(*(nodes[0].StaticRoutes[0].NextHop.Addr)).To(gomega.Equal(nodeip))
	g.Expect(*(nodes[0].StaticRoutes[0].Prefix.IPAddr.Addr)).To(gomega.Equal("10.245.0.0"))
	g.Expect(*(nodes[0].StaticRoutes[0].Prefix.Mask)).To(gomega.Equal(int32(24)))
}

func TestNodeDel(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_GLOBAL
	nodeName := "testNode1"
	objects.SharedAviGraphLister().Delete(modelName)
	err := KubeClient.CoreV1().Nodes().Delete(context.TODO(), nodeName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error in deleting Node: %v", err)
	}
	PollForCompletion(t, modelName, 5)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))

	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(0))
}

func TestNodeAddNoPodCIDR(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_GLOBAL
	nodeip := "20.1.1.2"
	objects.SharedAviGraphLister().Delete(modelName)
	nodeExample := (FakeNode{
		Name:    "testNodeInvalid",
		Version: "1",
		NodeIP:  nodeip,
	}).Node()

	_, err := KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	PollForCompletion(t, modelName, 5)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(false))
}

func TestMultiNodeAdd(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_GLOBAL
	nodeip1 := "10.1.1.1"
	nodeip2 := "10.1.1.2"
	objects.SharedAviGraphLister().Delete(modelName)
	nodeExample1 := (FakeNode{
		Name:     "testNode1",
		PodCIDR:  "10.244.1.0/24",
		PodCIDRs: []string{"10.244.1.0/24"},
		Version:  "1",
		NodeIP:   nodeip1,
	}).Node()
	nodeExample2 := (FakeNode{
		Name:     "testNode2",
		PodCIDR:  "10.244.2.0/24",
		PodCIDRs: []string{"10.244.2.0/24"},
		Version:  "1",
		NodeIP:   nodeip2,
	}).Node()

	_, err := KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}
	_, err = KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	PollForCompletion(t, modelName, 5)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))

	nodeIPMap := make(map[string]bool)
	nodeIPMap[nodeip1] = true
	nodeIPMap[nodeip2] = true
	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(2))
	for _, staticRoute := range nodes[0].StaticRoutes {
		g.Expect(*(nodes[0].StaticRoutes[0].Prefix.Mask)).To(gomega.Equal(int32(24)))
		if *(staticRoute.NextHop.Addr) == nodeip1 {
			delete(nodeIPMap, nodeip1)
			g.Expect(*(staticRoute.Prefix.IPAddr.Addr)).To(gomega.Equal("10.244.1.0"))
		} else if *(staticRoute.NextHop.Addr) == nodeip2 {
			delete(nodeIPMap, nodeip2)
			g.Expect(*(staticRoute.Prefix.IPAddr.Addr)).To(gomega.Equal("10.244.2.0"))
		} else {
			t.Fatalf("nodeIP %v did not match with expected IPs", staticRoute.NextHop.Addr)
		}
	}
	g.Expect(len(nodeIPMap)).To(gomega.Equal(0))
	KubeClient.CoreV1().Nodes().Delete(context.TODO(), "testNode1", metav1.DeleteOptions{})
	KubeClient.CoreV1().Nodes().Delete(context.TODO(), "testNode2", metav1.DeleteOptions{})
}

func TestMultiNodeUpdate(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_GLOBAL
	nodeip1 := "10.2.1.1"
	nodeip2 := "10.2.1.2"
	objects.SharedAviGraphLister().Delete(modelName)
	PollForCompletion(t, modelName, 10)
	nodeExample1 := (FakeNode{
		Name:     "testNode3",
		PodCIDR:  "10.244.3.0/24",
		PodCIDRs: []string{"10.244.3.0/24"},
		Version:  "1",
		NodeIP:   nodeip1,
	}).Node()
	nodeExample2 := (FakeNode{
		Name:     "testNode4",
		PodCIDR:  "10.244.4.0/24",
		PodCIDRs: []string{"10.244.4.0/24"},
		Version:  "1",
		NodeIP:   nodeip2,
	}).Node()

	_, err := KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}
	_, err = KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}
	PollForCompletion(t, modelName, 5)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	time.Sleep(10 * time.Second)
	nodeIPMap := make(map[string]bool)
	nodeIPMap[nodeip1] = true
	nodeIPMap[nodeip2] = true
	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(2))
	for _, staticRoute := range nodes[0].StaticRoutes {
		g.Expect(*(nodes[0].StaticRoutes[0].Prefix.Mask)).To(gomega.Equal(int32(24)))
		if *(staticRoute.NextHop.Addr) == nodeip1 {
			delete(nodeIPMap, nodeip1)
			g.Expect(*(staticRoute.Prefix.IPAddr.Addr)).To(gomega.Equal("10.244.3.0"))
		} else if *(staticRoute.NextHop.Addr) == nodeip2 {
			delete(nodeIPMap, nodeip2)
			g.Expect(*(staticRoute.Prefix.IPAddr.Addr)).To(gomega.Equal("10.244.4.0"))
		} else {
			t.Fatalf("nodeIP %v did not match with expected IPs", staticRoute.NextHop.Addr)
		}
	}
	g.Expect(len(nodeIPMap)).To(gomega.Equal(0))

	nodeExample := (FakeNode{
		Name:     "testNode4",
		PodCIDR:  "10.244.4.0/24",
		PodCIDRs: []string{"10.244.4.0/24"},
		Version:  "1",
		NodeIP:   nodeip2,
	}).Node()
	nodeExample.ObjectMeta.ResourceVersion = "2"
	nodeExample.Spec.PodCIDR = "10.245.0.0/24"
	nodeExample.Spec.PodCIDRs = []string{"10.245.0.0/24"}

	_, err = KubeClient.CoreV1().Nodes().Update(context.TODO(), nodeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Node: %v", err)
	}

	PollForCompletion(t, modelName, 10)
	time.Sleep(10 * time.Second)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	//AV-171818-Route ID should not change for update
	g.Expect(*(nodes[0].StaticRoutes[1].RouteID)).Should(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNode4"].RouteIDPrefix + "-0"))
	KubeClient.CoreV1().Nodes().Delete(context.TODO(), "testNode3", metav1.DeleteOptions{})
	KubeClient.CoreV1().Nodes().Delete(context.TODO(), "testNode4", metav1.DeleteOptions{})
	PollForCompletion(t, modelName, 10)
	g.Eventually(func() int {
		num_routes := len(nodes[0].StaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(0))
}

func TestMultiNodeCDC(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_GLOBAL
	nodeip1 := "10.2.1.5"
	nodeip2 := "10.2.1.6"
	nodeip3 := "10.2.1.7"
	objects.SharedAviGraphLister().Delete(modelName)
	PollForCompletion(t, modelName, 5)
	nodeExample1 := (FakeNode{
		Name:     "testNode5",
		PodCIDR:  "10.244.5.0/24",
		PodCIDRs: []string{"10.244.5.0/24"},
		Version:  "1",
		NodeIP:   nodeip1,
	}).Node()
	nodeExample2 := (FakeNode{
		Name:     "testNode6",
		PodCIDR:  "10.244.6.0/24",
		PodCIDRs: []string{"10.244.6.0/24"},
		Version:  "1",
		NodeIP:   nodeip2,
	}).Node()

	_, err := KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}
	_, err = KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}
	PollForCompletion(t, modelName, 5)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	nodeIPMap := make(map[string]bool)
	nodeIPMap[nodeip1] = true
	nodeIPMap[nodeip2] = true
	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(2))
	for _, staticRoute := range nodes[0].StaticRoutes {
		g.Expect(*(nodes[0].StaticRoutes[0].Prefix.Mask)).To(gomega.Equal(int32(24)))
		if *(staticRoute.NextHop.Addr) == nodeip1 {
			delete(nodeIPMap, nodeip1)
			g.Expect(*(staticRoute.Prefix.IPAddr.Addr)).To(gomega.Equal("10.244.5.0"))
		} else if *(staticRoute.NextHop.Addr) == nodeip2 {
			delete(nodeIPMap, nodeip2)
			g.Expect(*(staticRoute.Prefix.IPAddr.Addr)).To(gomega.Equal("10.244.6.0"))
		} else {
			t.Fatalf("nodeIP %v did not match with expected IPs", staticRoute.NextHop.Addr)
		}
	}
	g.Expect(len(nodeIPMap)).To(gomega.Equal(0))
	//Delete node
	KubeClient.CoreV1().Nodes().Delete(context.TODO(), "testNode5", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error in deleting Node: %v", err)
	}
	PollForCompletion(t, modelName, 5)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Eventually(func() int {
		num_static_routes := len(nodes[0].StaticRoutes)
		return num_static_routes
	}, 10*time.Second).Should(gomega.Equal(1))
	//After delete, Now testNode6 should be at index 0.
	g.Expect(*(nodes[0].StaticRoutes[0].NextHop.Addr)).Should(gomega.Equal(nodeip2))
	g.Expect(*(nodes[0].StaticRoutes[0].RouteID)).Should(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNode6"].RouteIDPrefix + "-0"))
	// Add another node
	nodeExample := (FakeNode{
		Name:     "testNode7",
		PodCIDR:  "10.244.7.0/24",
		PodCIDRs: []string{"10.244.7.0/24"},
		Version:  "1",
		NodeIP:   nodeip3,
	}).Node()
	nodeExample.ObjectMeta.ResourceVersion = "1"

	_, err = KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	//PollForCompletion(t, modelName, 10)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	//AV-171818-Route ID should not change for update
	g.Eventually(func() int {
		num_static_routes := len(nodes[0].StaticRoutes)
		return num_static_routes
	}, 10*time.Second).Should(gomega.Equal(2))
	g.Expect(*(nodes[0].StaticRoutes[1].RouteID)).Should(gomega.Equal("cluster-" + nodes[0].NodeStaticRoutes["testNode7"].RouteIDPrefix + "-0"))
	g.Expect(*(nodes[0].StaticRoutes[1].NextHop.Addr)).Should(gomega.Equal(nodeip3))
	KubeClient.CoreV1().Nodes().Delete(context.TODO(), "testNode6", metav1.DeleteOptions{})
	KubeClient.CoreV1().Nodes().Delete(context.TODO(), "testNode7", metav1.DeleteOptions{})
	g.Eventually(func() int {
		num_routes := len(nodes[0].StaticRoutes)
		return num_routes
	}, 10*time.Second).Should(gomega.Equal(0))
}

func TestNodeCIDRInAnnotation(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_GLOBAL
	nodeip := "30.1.1.2"
	objects.SharedAviGraphLister().Delete(modelName)
	nodeExample := (FakeNode{
		Name:               "testNodeAnnotation",
		PodCIDR:            "10.244.0.0/24",
		PodCIDRs:           []string{"10.244.0.0/24"},
		PodCIDRsAnnotation: "192.168.1.0/24, 192.168.2.0/24 ,",
		Version:            "1",
		NodeIP:             nodeip,
	}).Node()

	_, err := KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	PollForCompletion(t, modelName, 5)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))

	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(2))
	g.Expect(*(nodes[0].StaticRoutes[0].NextHop.Addr)).To(gomega.Equal(nodeip))
	g.Expect(*(nodes[0].StaticRoutes[0].Prefix.IPAddr.Addr)).To(gomega.Equal("192.168.1.0"))
	g.Expect(*(nodes[0].StaticRoutes[1].Prefix.IPAddr.Addr)).To(gomega.Equal("192.168.2.0"))
	g.Expect(*(nodes[0].StaticRoutes[0].Prefix.Mask)).To(gomega.Equal(int32(24)))

	nodeExample = (FakeNode{
		Name:               "testNodeAnnotation",
		PodCIDR:            "10.244.0.0/24",
		PodCIDRs:           []string{"10.244.0.0/24"},
		PodCIDRsAnnotation: "  192.168.1.0/24,  192.168.2.0/24   ",
		Version:            "1",
		NodeIP:             nodeip,
	}).Node()

	// Update the annotation to have a single and different CIDR.
	nodeExample.Annotations[lib.StaticRouteAnnotation] = "192.168.3.0/24   "
	nodeExample.ResourceVersion = "2"
	_, err = KubeClient.CoreV1().Nodes().Update(context.TODO(), nodeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Node: %v", err)
	}

	g.Eventually(func() int {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		return len(nodes[0].StaticRoutes)
	}, 10*time.Second).Should(gomega.Equal(1))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(nodes[0].StaticRoutes).To(gomega.HaveLen(1))
	g.Expect(*(nodes[0].StaticRoutes[0].Prefix.IPAddr.Addr)).To(gomega.Equal("192.168.3.0"))

	// Remove the whole annotation for AKO to fallback to PodCIDR field.
	delete(nodeExample.Annotations, lib.StaticRouteAnnotation)
	nodeExample.ResourceVersion = "3"
	_, err = KubeClient.CoreV1().Nodes().Update(context.TODO(), nodeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Node: %v", err)
	}

	g.Eventually(func() string {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
		if len(nodes[0].StaticRoutes) > 0 {
			return *(nodes[0].StaticRoutes[0].Prefix.IPAddr.Addr)
		}
		return ""
	}, 10*time.Second).Should(gomega.Equal("10.244.0.0"))
	KubeClient.CoreV1().Nodes().Delete(context.TODO(), "testNodeAnnotation", metav1.DeleteOptions{})
}

func TestNodeOVNKubernetesAdd(t *testing.T) {
	os.Setenv("CNI_PLUGIN", "ovn-kubernetes")
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_GLOBAL
	nodeip := "10.1.1.2"
	nodeName := "testNode1"
	objects.SharedAviGraphLister().Delete(modelName)
	nodeExample := (FakeNode{
		Name:               nodeName,
		PodCIDRsAnnotation: "192.168.1.0/24",
		Version:            "1",
		NodeIP:             nodeip,
	}).NodeOVN()

	_, err := KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	PollForCompletion(t, modelName, 5)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))

	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(1))
	g.Expect(*(nodes[0].StaticRoutes[0].NextHop.Addr)).To(gomega.Equal(nodeip))
	g.Expect(*(nodes[0].StaticRoutes[0].Prefix.IPAddr.Addr)).To(gomega.Equal("192.168.1.0"))
	g.Expect(*(nodes[0].StaticRoutes[0].Prefix.Mask)).To(gomega.Equal(int32(24)))
}

func TestNodeOVNKubernetesDel(t *testing.T) {
	os.Setenv("CNI_PLUGIN", "ovn-kubernetes")
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_GLOBAL
	nodeName := "testNode1"
	objects.SharedAviGraphLister().Delete(modelName)
	err := KubeClient.CoreV1().Nodes().Delete(context.TODO(), nodeName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error in deleting Node: %v", err)
	}
	PollForCompletion(t, modelName, 5)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))

	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(0))
}
