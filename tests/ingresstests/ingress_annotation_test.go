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

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var DefaultPassthroughModel = "admin/cluster--Shared-Passthrough-0"
var passthroughIngressName = "foo"

func ValidatePassthroughModel(t *testing.T, g *gomega.WithT, modelName string) interface{} {

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 60*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()

	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].HTTPDSrefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))

	return aviModel
}

func VerifyPassthroughIngressDeletion(t *testing.T, g *gomega.WithT, modelName string, poolCount, childcount int) {
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)

	var nodes []*avinodes.AviVsNode
	g.Eventually(func() []*avinodes.AviPoolNode {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].PoolRefs
	}, 60*time.Second).Should(gomega.HaveLen(poolCount))

	g.Eventually(func() []*avinodes.AviPoolGroupNode {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].PoolGroupRefs
	}, 10*time.Second).Should(gomega.HaveLen(poolCount))

	g.Eventually(func() int {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(nodes[0].PassthroughChildNodes) == 0 {
			return 0
		}
		return len(nodes[0].PassthroughChildNodes[0].HttpPolicySetRefs)
	}, 60*time.Second).Should(gomega.Equal(childcount))
}

func VerifyPasthrough(t *testing.T, g *gomega.WithT, vs *avinodes.AviVsNode, svcName string) {

	g.Eventually(func() int {
		if len(vs.HTTPDSrefs) < 1 {
			return 0
		}
		return len(vs.HTTPDSrefs[0].PoolGroupRefs)
	}, 60*time.Second).Should(gomega.Equal(1))

	g.Expect(vs.HTTPDSrefs[0].PoolGroupRefs[0]).To(gomega.Equal("cluster--foo.com"))

	g.Expect(vs.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(vs.PoolGroupRefs[0].Name).To(gomega.Equal("cluster--foo.com"))

	g.Eventually(func() int {
		return len(vs.PoolGroupRefs[0].Members)
	}, 60*time.Second).Should(gomega.Equal(1))
	g.Expect(*vs.PoolGroupRefs[0].Members[0].PoolRef).To(gomega.Equal("/api/pool?name=cluster--foo.com-" + svcName))

	g.Expect(vs.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(vs.PoolRefs[0].Name).To(gomega.Equal("cluster--foo.com-" + svcName))

	g.Eventually(func() int {
		return len(vs.PoolRefs[0].Servers)
	}, 60*time.Second).Should(gomega.Equal(1))

	g.Expect(vs.VSVIPRefs[0].FQDNs[0]).To(gomega.Equal("foo.com"))
}

func TestPassthroughIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, DefaultPassthroughModel)
	ingrFake := (integrationtest.FakeIngress{
		Name:        passthroughIngressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
	}).Ingress()
	ann := make(map[string]string)
	ann[lib.PassthroughAnnotation] = "true"
	ingrFake.SetAnnotations(ann)
	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, DefaultPassthroughModel, 5)

	aviModel := ValidatePassthroughModel(t, g, DefaultPassthroughModel)
	graph := aviModel.(*avinodes.AviObjectGraph)
	vs := graph.GetAviVS()[0]
	VerifyPasthrough(t, g, vs, svcName)

	g.Expect(vs.PassthroughChildNodes).To(gomega.HaveLen(1))
	passInsecureNode := vs.PassthroughChildNodes[0]
	g.Expect(passInsecureNode.Name).To(gomega.Equal("cluster--Shared-Passthrough-0-insecure"))
	g.Expect(passInsecureNode.HttpPolicyRefs[0].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))
	g.Expect(passInsecureNode.HttpPolicyRefs[0].RedirectPorts[0].StatusCode).To(gomega.Equal("HTTP_REDIRECT_STATUS_CODE_302"))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), passthroughIngressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error in deleting Ingress: %v", err)
	}
	VerifyPassthroughIngressDeletion(t, g, DefaultPassthroughModel, 0, 0)
	TearDownTestForIngress(t, svcName, DefaultPassthroughModel)
}

func TestPassthroughIngressUpdateHostname(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, DefaultPassthroughModel)
	ingrFake := (integrationtest.FakeIngress{
		Name:        passthroughIngressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
	}).Ingress()
	ann := make(map[string]string)
	ann[lib.PassthroughAnnotation] = "true"
	ingrFake.SetAnnotations(ann)
	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, DefaultPassthroughModel, 5)

	aviModel := ValidatePassthroughModel(t, g, DefaultPassthroughModel)
	graph := aviModel.(*avinodes.AviObjectGraph)
	vs := graph.GetAviVS()[0]
	VerifyPasthrough(t, g, vs, svcName)

	g.Expect(vs.PassthroughChildNodes).To(gomega.HaveLen(1))
	passInsecureNode := vs.PassthroughChildNodes[0]
	g.Expect(passInsecureNode.Name).To(gomega.Equal("cluster--Shared-Passthrough-0-insecure"))
	g.Expect(passInsecureNode.HttpPolicyRefs[0].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))
	g.Expect(passInsecureNode.HttpPolicyRefs[0].RedirectPorts[0].StatusCode).To(gomega.Equal("HTTP_REDIRECT_STATUS_CODE_302"))

	ingrFake = (integrationtest.FakeIngress{
		Name:        passthroughIngressName,
		Namespace:   "default",
		DnsNames:    []string{"bar.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
	}).Ingress()
	ann = make(map[string]string)
	ann[lib.PassthroughAnnotation] = "true"
	ingrFake.SetAnnotations(ann)
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	g.Eventually(func() string {
		_, aviModel = objects.SharedAviGraphLister().Get(DefaultPassthroughModel)
		graph = aviModel.(*avinodes.AviObjectGraph)
		vs = graph.GetAviVS()[0]
		if len(vs.HTTPDSrefs[0].PoolGroupRefs) != 1 {
			return ""
		}
		return vs.HTTPDSrefs[0].PoolGroupRefs[0]
	}, 60*time.Second).Should(gomega.Equal("cluster--bar.com"))

	pg := vs.PoolGroupRefs[0]
	g.Expect(pg.Members).To(gomega.HaveLen(1))
	g.Expect(*pg.Members[0].PoolRef).To(gomega.Equal("/api/pool?name=cluster--bar.com-" + svcName))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), passthroughIngressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error in deleting Ingress: %v", err)
	}
	VerifyPassthroughIngressDeletion(t, g, DefaultPassthroughModel, 0, 0)
	TearDownTestForIngress(t, svcName, DefaultPassthroughModel)
}

func TestPassthroughIngressRemoveAnnotation(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, DefaultPassthroughModel)
	ingrFake := (integrationtest.FakeIngress{
		Name:        passthroughIngressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
	}).Ingress()
	ann := make(map[string]string)
	ann[lib.PassthroughAnnotation] = "true"
	ingrFake.SetAnnotations(ann)
	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, DefaultPassthroughModel, 5)

	aviModel := ValidatePassthroughModel(t, g, DefaultPassthroughModel)
	graph := aviModel.(*avinodes.AviObjectGraph)
	vs := graph.GetAviVS()[0]
	VerifyPasthrough(t, g, vs, svcName)

	ann = make(map[string]string)
	ingrFake.SetAnnotations(ann)
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}
	VerifyPassthroughIngressDeletion(t, g, DefaultPassthroughModel, 0, 0)

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), passthroughIngressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	TearDownTestForIngress(t, svcName, DefaultPassthroughModel)
}

func TestPassthroughIngressAddAnnotation(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, DefaultPassthroughModel)
	ingrFake := (integrationtest.FakeIngress{
		Name:        passthroughIngressName,
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

	ann := make(map[string]string)
	ann[lib.PassthroughAnnotation] = "true"
	ingrFake.SetAnnotations(ann)
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, DefaultPassthroughModel, 5)

	aviModel := ValidatePassthroughModel(t, g, DefaultPassthroughModel)
	graph := aviModel.(*avinodes.AviObjectGraph)
	vs := graph.GetAviVS()[0]
	VerifyPasthrough(t, g, vs, svcName)

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), passthroughIngressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	VerifyPassthroughIngressDeletion(t, g, DefaultPassthroughModel, 0, 0)
	TearDownTestForIngress(t, svcName, DefaultPassthroughModel)
}

func TestPassthroughMultipleIngresses(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, DefaultPassthroughModel)
	ingrFake1 := (integrationtest.FakeIngress{
		Name:        passthroughIngressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
	}).Ingress()
	ann := make(map[string]string)
	ann[lib.PassthroughAnnotation] = "true"
	ingrFake1.SetAnnotations(ann)
	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	passthroughIngressName2 := "bar"
	ingrFake2 := (integrationtest.FakeIngress{
		Name:        passthroughIngressName2,
		Namespace:   "default",
		DnsNames:    []string{"bar.com"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
	}).Ingress()
	ingrFake2.SetAnnotations(ann)
	_, err = KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, DefaultPassthroughModel, 5)

	_, aviModel := objects.SharedAviGraphLister().Get(DefaultPassthroughModel)
	graph := aviModel.(*avinodes.AviObjectGraph)
	vs := graph.GetAviVS()[0]

	g.Eventually(func() int {
		return len(vs.HTTPDSrefs[0].PoolGroupRefs)
	}, 60*time.Second).Should(gomega.Equal(2))

	for _, pgname := range vs.HTTPDSrefs[0].PoolGroupRefs {
		if pgname != "cluster--foo.com" && pgname != "cluster--bar.com" {
			t.Fatalf("Unexpected pg ref in datascript: %s", pgname)
		}
	}

	g.Expect(vs.PoolGroupRefs).To(gomega.HaveLen(2))
	for _, pg := range vs.PoolGroupRefs {
		if pg.Name == "cluster--foo.com" {
			g.Expect(pg.Members).To(gomega.HaveLen(1))
			g.Expect(*pg.Members[0].PoolRef).To(gomega.Equal("/api/pool?name=cluster--foo.com-" + svcName))
		} else if pg.Name == "cluster--bar.com" {
			g.Expect(pg.Members).To(gomega.HaveLen(1))
			g.Expect(*pg.Members[0].PoolRef).To(gomega.Equal("/api/pool?name=cluster--bar.com-" + svcName))
		} else {
			t.Fatalf("Unexpected PG: %s", pg.Name)
		}
	}

	g.Expect(vs.PoolRefs).To(gomega.HaveLen(2))
	for _, pool := range vs.PoolRefs {
		if pool.Name == "cluster--foo.com-"+svcName || pool.Name == "cluster--bar.com-"+svcName {
			g.Expect(pool.Servers).To(gomega.HaveLen(1))
		} else {
			t.Fatalf("Unexpected Pool: %s", pool.Name)
		}
	}

	g.Expect(vs.PassthroughChildNodes).To(gomega.HaveLen(1))
	passInsecureNode := vs.PassthroughChildNodes[0]
	g.Expect(passInsecureNode.Name).To(gomega.Equal("cluster--Shared-Passthrough-0-insecure"))
	for _, redir := range passInsecureNode.HttpPolicyRefs {
		if redir.RedirectPorts[0].Hosts[0] != "foo.com" && redir.RedirectPorts[0].Hosts[0] != "bar.com" {
			t.Fatalf("unexpected redirect policy for: %s", redir.RedirectPorts[0].Hosts[0])
		}
		g.Expect(redir.RedirectPorts[0].StatusCode).To(gomega.Equal("HTTP_REDIRECT_STATUS_CODE_302"))
	}

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), passthroughIngressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error in deleting Ingress: %v", err)
	}
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), passthroughIngressName2, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error in deleting Ingress: %v", err)
	}
	VerifyPassthroughIngressDeletion(t, g, DefaultPassthroughModel, 0, 0)
	TearDownTestForIngress(t, svcName, DefaultPassthroughModel)
}

func TestAddIngressDefaultCert(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	ingName := objNameMap.GenerateName("foo-with-targets")
	SetUpTestForIngress(t, svcName, modelName)
	integrationtest.AddSecret(lib.DefaultRouteCert, utils.GetAKONamespace(), "tlsCert", "tlsKey")

	ingrFake := (integrationtest.FakeIngress{
		Name:        ingName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
	}).Ingress()

	ann := make(map[string]string)
	ann[lib.DefaultSecretEnabled] = "true"
	ingrFake.SetAnnotations(ann)

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	var aviModel interface{}
	var nodes []*avinodes.AviVsNode
	var found bool
	g.Eventually(func() bool {
		found, aviModel = objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(nodes) != 1 {
			return false
		}
		if len(nodes[0].SniNodes) != 1 {
			return false
		}
		return true

	}, 40*time.Second).Should(gomega.Equal(true))

	g.Expect(nodes[0].SniNodes[0].Name).To(gomega.Equal("cluster--foo.com"))
	g.Expect(nodes[0].SniNodes[0].PoolGroupRefs[0].Name).To(gomega.Equal("cluster--default-foo.com_foo-" + ingName))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--default-foo.com_foo-" + ingName))
	g.Expect(nodes[0].SniNodes[0].SSLKeyCertRefs[0].Name).To(gomega.Equal("cluster--" + integrationtest.DefaultRouteCert))
	g.Expect(nodes[0].SniNodes[0].HttpPolicyRefs[0].Name).To(gomega.Equal("cluster--default-foo.com"))
	g.Expect(nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[0].Name).To(gomega.Equal("cluster--default-foo.com_foo-" + ingName))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't Delete the Ingress %v", err)
	}
	err = KubeClient.CoreV1().Secrets(utils.GetAKONamespace()).Delete(context.TODO(), lib.DefaultRouteCert, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't Delete the secret %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)
	TearDownTestForIngress(t, svcName, modelName)
}

func TestAddIngressDefaultCertRemoveAnnotation(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	ingName := objNameMap.GenerateName("foo-with-targets")
	SetUpTestForIngress(t, svcName, modelName)
	integrationtest.AddSecret(lib.DefaultRouteCert, utils.GetAKONamespace(), "tlsCert", "tlsKey")

	ingrFake := (integrationtest.FakeIngress{
		Name:        ingName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
	}).Ingress()

	ann := make(map[string]string)
	ann[lib.DefaultSecretEnabled] = "true"
	ingrFake.SetAnnotations(ann)

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	var aviModel interface{}
	var nodes []*avinodes.AviVsNode
	var found bool
	g.Eventually(func() bool {
		found, aviModel = objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(nodes) != 1 {
			return false
		}
		if len(nodes[0].SniNodes) != 1 {
			return false
		}
		return true

	}, 40*time.Second).Should(gomega.Equal(true))

	ann = make(map[string]string)
	ingrFake.SetAnnotations(ann)
	ingrFake.ResourceVersion = "2"

	_, err = KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	g.Eventually(func() int {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].SniNodes)
	}, 40*time.Second).Should(gomega.Equal(0))

	g.Expect(nodes[0].PoolGroupRefs[0].Name).To(gomega.Equal("cluster--Shared-L7-0"))
	g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--foo.com_foo-default-" + ingName))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't Delete the Ingress %v", err)
	}
	err = KubeClient.CoreV1().Secrets(utils.GetAKONamespace()).Delete(context.TODO(), lib.DefaultRouteCert, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't Delete the secret %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)
	TearDownTestForIngress(t, svcName, modelName)
}

func TestAddIngressDefaultCertAddAnnotation(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	ingName := objNameMap.GenerateName("foo-with-targets")
	SetUpTestForIngress(t, svcName, modelName)
	integrationtest.AddSecret(lib.DefaultRouteCert, utils.GetAKONamespace(), "tlsCert", "tlsKey")

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

	ann := make(map[string]string)
	ann[lib.DefaultSecretEnabled] = "true"
	ingrFake.SetAnnotations(ann)
	ingrFake.ResourceVersion = "2"

	_, err = KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	var aviModel interface{}
	var nodes []*avinodes.AviVsNode
	var found bool
	g.Eventually(func() bool {
		found, aviModel = objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(nodes) != 1 {
			return false
		}
		if len(nodes[0].SniNodes) != 1 {
			return false
		}
		return true
	}, 40*time.Second).Should(gomega.Equal(true))

	g.Expect(nodes[0].SniNodes[0].Name).To(gomega.Equal("cluster--foo.com"))
	g.Expect(nodes[0].SniNodes[0].PoolGroupRefs[0].Name).To(gomega.Equal("cluster--default-foo.com_foo-" + ingName))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--default-foo.com_foo-" + ingName))
	g.Expect(nodes[0].SniNodes[0].SSLKeyCertRefs[0].Name).To(gomega.Equal("cluster--" + integrationtest.DefaultRouteCert))
	g.Expect(nodes[0].SniNodes[0].HttpPolicyRefs[0].Name).To(gomega.Equal("cluster--default-foo.com"))
	g.Expect(nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[0].Name).To(gomega.Equal("cluster--default-foo.com_foo-" + ingName))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't Delete the Ingress %v", err)
	}
	err = KubeClient.CoreV1().Secrets(utils.GetAKONamespace()).Delete(context.TODO(), lib.DefaultRouteCert, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't Delete the secret %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)
	TearDownTestForIngress(t, svcName, modelName)
}

// TestIngressAnnotationAddDefaultCert first adds an Ingress with default secret annotation, then adds the secret and verifies the model graph.
func TestIngressAnnotationAddDefaultCert(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	ingName := objNameMap.GenerateName("foo-with-targets")
	SetUpTestForIngress(t, svcName, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        ingName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
	}).Ingress()

	ann := make(map[string]string)
	ann[lib.DefaultSecretEnabled] = "true"
	ingrFake.SetAnnotations(ann)

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.AddSecret(lib.DefaultRouteCert, utils.GetAKONamespace(), "tlsCert", "tlsKey")

	var aviModel interface{}
	var nodes []*avinodes.AviVsNode
	var found bool
	g.Eventually(func() bool {
		found, aviModel = objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(nodes) != 1 {
			return false
		}
		if len(nodes[0].SniNodes) != 1 {
			return false
		}
		return true
	}, 40*time.Second).Should(gomega.Equal(true))

	g.Expect(nodes[0].SniNodes[0].Name).To(gomega.Equal("cluster--foo.com"))
	g.Expect(nodes[0].SniNodes[0].PoolGroupRefs[0].Name).To(gomega.Equal("cluster--default-foo.com_foo-" + ingName))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--default-foo.com_foo-" + ingName))
	g.Expect(nodes[0].SniNodes[0].SSLKeyCertRefs[0].Name).To(gomega.Equal("cluster--" + integrationtest.DefaultRouteCert))
	g.Expect(nodes[0].SniNodes[0].HttpPolicyRefs[0].Name).To(gomega.Equal("cluster--default-foo.com"))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't Delete the Ingress %v", err)
	}
	err = KubeClient.CoreV1().Secrets(utils.GetAKONamespace()).Delete(context.TODO(), lib.DefaultRouteCert, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't Delete the secret %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)
	TearDownTestForIngress(t, svcName, modelName)
}
