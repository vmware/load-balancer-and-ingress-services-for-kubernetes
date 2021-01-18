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
	"testing"

	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"

	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNodeAdd(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/global"
	nodeip := "10.1.1.2"
	objects.SharedAviGraphLister().Delete(modelName)
	nodeExample := (FakeNode{
		Name:    "testNode1",
		PodCIDR: "10.244.0.0/24",
		Version: "1",
		NodeIP:  nodeip,
	}).Node()

	_, err := KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if !found {
		t.Fatalf("Model not found for node add %v", modelName)
	}
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
	modelName := "admin/global"
	nodeip := "10.1.1.2"
	objects.SharedAviGraphLister().Delete(modelName)
	nodeExample := (FakeNode{
		Name:    "testNode1",
		PodCIDR: "10.244.0.0/24",
		Version: "1",
		NodeIP:  nodeip,
	}).Node()

	nodeExample.ObjectMeta.ResourceVersion = "2"
	nodeExample.Spec.PodCIDR = "10.245.0.0/24"

	_, err := KubeClient.CoreV1().Nodes().Update(context.TODO(), nodeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Node: %v", err)
	}

	PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if !found {
		t.Fatalf("Model not found for node add %v", modelName)
	}
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
	modelName := "admin/global"
	nodeName := "testNode1"
	objects.SharedAviGraphLister().Delete(modelName)
	err := KubeClient.CoreV1().Nodes().Delete(context.TODO(), nodeName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error in deleting Node: %v", err)
	}
	PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if !found {
		t.Fatalf("Model not found for node add %v", modelName)
	}
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))

	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(0))
}

func TestNodeAddNoPodCIDR(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/global"
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
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if !found {
		t.Fatalf("Model not found for node add %v", modelName)
	}
	g.Expect(aviModel.(*avinodes.AviObjectGraph).IsVrf).To(gomega.Equal(true))
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVRF()
	g.Expect(len(nodes)).To(gomega.Equal(1))

	g.Expect(len(nodes[0].StaticRoutes)).To(gomega.Equal(0))
}

func TestMultiNodeAdd(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/global"
	nodeip1 := "10.1.1.1"
	nodeip2 := "10.1.1.2"
	objects.SharedAviGraphLister().Delete(modelName)
	nodeExample1 := (FakeNode{
		Name:    "testNode1",
		PodCIDR: "10.244.1.0/24",
		Version: "1",
		NodeIP:  nodeip1,
	}).Node()
	nodeExample2 := (FakeNode{
		Name:    "testNode2",
		PodCIDR: "10.244.2.0/24",
		Version: "1",
		NodeIP:  nodeip2,
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
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if !found {
		t.Fatalf("Model not found for node add %v", modelName)
	}
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
}
