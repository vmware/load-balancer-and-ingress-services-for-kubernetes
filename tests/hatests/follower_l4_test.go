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
	"fmt"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func SetUpTestForSvcLB(t *testing.T) {
	objects.SharedAviGraphLister().Delete(integrationtest.SINGLEPORTMODEL)
	integrationtest.CreateSVC(t, integrationtest.NAMESPACE, integrationtest.SINGLEPORTSVC, corev1.ProtocolTCP, corev1.ServiceTypeLoadBalancer, false)
	integrationtest.CreateEP(t, integrationtest.NAMESPACE, integrationtest.SINGLEPORTSVC, false, false, "1.1.1")
	integrationtest.PollForCompletion(t, integrationtest.SINGLEPORTMODEL, 5)
}

func TearDownTestForSvcLB(t *testing.T, g *gomega.GomegaWithT) {
	objects.SharedAviGraphLister().Delete(integrationtest.SINGLEPORTMODEL)
	integrationtest.DelSVC(t, integrationtest.NAMESPACE, integrationtest.SINGLEPORTSVC)
	integrationtest.DelEP(t, integrationtest.NAMESPACE, integrationtest.SINGLEPORTSVC)
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: integrationtest.AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", integrationtest.SINGLEPORTSVC, integrationtest.NAMESPACE)}
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(false))
}

func TestL4Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForSvcLB(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(integrationtest.SINGLEPORTMODEL)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(integrationtest.SINGLEPORTMODEL)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", integrationtest.NAMESPACE, integrationtest.SINGLEPORTSVC)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(integrationtest.AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

	// Check for the pools
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	address := "1.1.1.1"
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&address))

	// If we transition the service from Loadbalancer to ClusterIP - it should get deleted.
	svcExample := (integrationtest.FakeService{
		Name:         integrationtest.SINGLEPORTSVC,
		Namespace:    integrationtest.NAMESPACE,
		Type:         corev1.ServiceTypeClusterIP,
		ServicePorts: []integrationtest.Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	svcExample.ResourceVersion = "2"
	_, err := KubeClient.CoreV1().Services(integrationtest.NAMESPACE).Update(context.TODO(), svcExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: integrationtest.AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", integrationtest.NAMESPACE, integrationtest.SINGLEPORTSVC)}
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 15*time.Second).Should(gomega.Equal(false))
	// If we transition the service from clusterIP to Loadbalancer - vs should get created
	svcExample = (integrationtest.FakeService{
		Name:         integrationtest.SINGLEPORTSVC,
		Namespace:    integrationtest.NAMESPACE,
		Type:         corev1.ServiceTypeLoadBalancer,
		ServicePorts: []integrationtest.Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	svcExample.ResourceVersion = "3"
	_, err = KubeClient.CoreV1().Services(integrationtest.NAMESPACE).Update(context.TODO(), svcExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}

	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 15*time.Second).Should(gomega.Equal(true))
	TearDownTestForSvcLB(t, g)
}
