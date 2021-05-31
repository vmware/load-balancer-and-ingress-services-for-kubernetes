/*
 * Copyright 2020-2021 VMware, Inc.
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
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/rest"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBootupIngressStatusPersistence(t *testing.T) {
	// create ingress, sync ingress and check for status, remove status
	// call SyncObjectStatuses to check if status remains the same

	g := gomega.NewGomegaWithT(t)

	ingressName := "foo-with-targets"

	modelName := "admin/cluster--Shared-L7-0"
	SetupDomain()
	SetUpTestForIngress(t, integrationtest.AllModels...)
	integrationtest.PollForCompletion(t, modelName, 5)
	ingressObject := integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo", "/xyz"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			"my-secret": {"foo.com"},
		},
	}
	integrationtest.AddSecret("my-secret", "default", "tlsCert", "tlsKey")
	ingrFake := ingressObject.Ingress()
	if _, err := KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-0"}
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 15*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() string {
		ingress, _ := KubeClient.NetworkingV1beta1().Ingresses("default").Get(context.TODO(), ingressName, metav1.GetOptions{})
		if len(ingress.Status.LoadBalancer.Ingress) > 0 {
			return ingress.Status.LoadBalancer.Ingress[0].IP
		}
		return ""
	}, 40*time.Second).Should(gomega.Equal("10.250.250.10"))

	ingrFake.Status.LoadBalancer.Ingress = []corev1.LoadBalancerIngress{}
	if _, err := KubeClient.NetworkingV1beta1().Ingresses("default").UpdateStatus(context.TODO(), ingrFake, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress Status: %v", err)
	}

	aviRestClientPool := cache.SharedAVIClients()
	aviObjCache := cache.SharedAviObjCache()
	restlayer := rest.NewRestOperations(aviObjCache, aviRestClientPool)
	restlayer.SyncObjectStatuses()

	g.Eventually(func() string {
		ingress, _ := KubeClient.NetworkingV1beta1().Ingresses("default").Get(context.TODO(), ingressName, metav1.GetOptions{})
		if len(ingress.Status.LoadBalancer.Ingress) > 0 {
			return ingress.Status.LoadBalancer.Ingress[0].IP
		}
		return ""
	}, 40*time.Second).Should(gomega.Equal("10.250.250.10"))
	TearDownIngressForCacheSyncCheck(t, modelName)
}
