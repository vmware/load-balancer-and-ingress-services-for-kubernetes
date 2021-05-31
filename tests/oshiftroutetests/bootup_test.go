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

package oshiftroutetests

import (
	"context"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/rest"

	"github.com/onsi/gomega"
	routev1 "github.com/openshift/api/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Commenting out this test for now, because this is failing frequently. Couple of things to look into:
// 1. Fake AVI controller where we are returning objects for various requests.
// 2. Cache population of VS objects - this might be different because ideally this test should run during boot up, which is not the case here.
func TestBootupRouteStatusPersistence(t *testing.T) {
	// create route, sync route and check for status, remove status
	// call SyncObjectStatuses to check if status remains the same

	g := gomega.NewGomegaWithT(t)

	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{}.Route()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample.ResourceVersion = "2"
	routeExample.Status.Ingress = []routev1.RouteIngress{}
	if _, err := OshiftClient.RouteV1().Routes(defaultNamespace).UpdateStatus(context.TODO(), routeExample, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress Status: %v", err)
	}

	aviRestClientPool := cache.SharedAVIClients()
	aviObjCache := cache.SharedAviObjCache()
	restlayer := rest.NewRestOperations(aviObjCache, aviRestClientPool)
	restlayer.SyncObjectStatuses()

	var route *routev1.Route
	g.Eventually(func() string {
		route, _ = OshiftClient.RouteV1().Routes("default").Get(context.TODO(), defaultRouteName, metav1.GetOptions{})
		if (len(route.Status.Ingress)) != 1 {
			return ""
		}
		return route.Status.Ingress[0].Host
	}, 30*time.Second).Should(gomega.Equal(defaultHostname))
	TearDownRouteForRestCheck(t, DefaultPassthroughModel)
}
