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
	"testing"
	"time"

	"github.com/onsi/gomega"
	avinodes "gitlab.eng.vmware.com/orion/akc/pkg/nodes"
	"gitlab.eng.vmware.com/orion/akc/pkg/objects"
)

func TestNodeAdd(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/global"
	nodeip := "10.1.1.2"
	objects.SharedAviGraphLister().Delete(modelName)
	nodeExample := (fakeNode{
		name:    "testNode1",
		podCIDR: "10.244.0.0/24",
		version: "1",
		nodeIP:  nodeip,
	}).Node()

	_, err := kubeClient.CoreV1().Nodes().Create(nodeExample)
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	pollForCompletion(t, modelName, 5)
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
	nodeExample := (fakeNode{
		name:    "testNode1",
		podCIDR: "10.244.0.0/24",
		version: "1",
		nodeIP:  nodeip,
	}).Node()

	nodeExample.ObjectMeta.ResourceVersion = "2"
	nodeExample.Spec.PodCIDR = "10.245.0.0/24"

	_, err := kubeClient.CoreV1().Nodes().Update(nodeExample)
	if err != nil {
		t.Fatalf("error in updating Node: %v", err)
	}

	pollForCompletion(t, modelName, 5)
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
	err := kubeClient.CoreV1().Nodes().Delete(nodeName, nil)
	if err != nil {
		t.Fatalf("error in deleting Node: %v", err)
	}
	pollForCompletion(t, modelName, 5)
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
	nodeExample := (fakeNode{
		name:    "testNodeInvalid",
		version: "1",
		nodeIP:  nodeip,
	}).Node()

	_, err := kubeClient.CoreV1().Nodes().Create(nodeExample)
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	pollForCompletion(t, modelName, 5)
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
	nodeExample1 := (fakeNode{
		name:    "testNode1",
		podCIDR: "10.244.1.0/24",
		version: "1",
		nodeIP:  nodeip1,
	}).Node()
	nodeExample2 := (fakeNode{
		name:    "testNode2",
		podCIDR: "10.244.2.0/24",
		version: "1",
		nodeIP:  nodeip2,
	}).Node()

	_, err := kubeClient.CoreV1().Nodes().Create(nodeExample1)
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}
	_, err = kubeClient.CoreV1().Nodes().Create(nodeExample2)
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	time.Sleep(5)
	pollForCompletion(t, modelName, 5)
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
