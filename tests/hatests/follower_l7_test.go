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

package hatests

import (
	"context"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/ingresstests"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/oshiftroutetests"

	"github.com/onsi/gomega"
	"github.com/vmware/alb-sdk/go/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func VerifyIngressDeletion(t *testing.T, g *gomega.WithT, aviModel interface{}, poolCount int) {
	var nodes []*avinodes.AviVsNode
	g.Eventually(func() []*avinodes.AviPoolNode {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].PoolRefs
	}, 10*time.Second).Should(gomega.HaveLen(poolCount))

	g.Eventually(func() []*models.PoolGroupMember {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].PoolGroupRefs[0].Members
	}, 10*time.Second).Should(gomega.HaveLen(poolCount))
}

func TestL7Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	ingresstests.SetUpTestForIngress(t, modelName)

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

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 50*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-0"}
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 15*time.Second).Should(gomega.Equal(true))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "foo-with-targets", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	ingresstests.TearDownTestForIngress(t, modelName)
}

func TestRoute(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/cluster--Shared-L7-0"
	oshiftroutetests.OshiftClient = OshiftClient
	oshiftroutetests.SetUpTestForRoute(t, modelName)
	routeObj := oshiftroutetests.FakeRoute{Name: "foo", Path: "/foo"}.SecureRoute()
	_, err := OshiftClient.RouteV1().Routes("default").Create(context.TODO(), routeObj, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := oshiftroutetests.ValidateSniModel(t, g, modelName)

	g.Expect(aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes).To(gomega.HaveLen(1))
	sniVS := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
	g.Eventually(func() string {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return sniVS.VHDomainNames[0]
	}, 20*time.Second).Should(gomega.Equal("foo.com"))
	oshiftroutetests.VerifySniNode(g, sniVS)

	oshiftroutetests.VerifySecureRouteDeletion(t, g, modelName, 0, 0)
	oshiftroutetests.TearDownTestForRoute(t, modelName)
}
