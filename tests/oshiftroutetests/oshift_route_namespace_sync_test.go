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

package oshiftroutetests

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"
)

func SetupRouteNamespaceSync(key, value string) {
	os.Setenv("NAMESPACE_SYNC_LABEL_KEY", key)
	os.Setenv("NAMESPACE_SYNC_LABEL_VALUE", value)
	ctrl.InitializeNamespaceSync()
}

func SetupRoute(t *testing.T, modelName, namespace string) {

	objects.SharedAviGraphLister().Delete(modelName)
	integrationtest.CreateSVC(t, namespace, "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, namespace, "avisvc", false, false, "1.1.1")
	integrationtest.PollForCompletion(t, modelName, 5)

	routeExample := FakeRoute{Namespace: namespace, Path: "/foo"}.Route()
	routeExample.ResourceVersion = "1"
	_, err := OshiftClient.RouteV1().Routes(namespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})

	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 15)
}
func UpdateRoute(t *testing.T, modelName, namespace string) {

	routeExample := FakeRoute{Namespace: namespace, Path: "/bar"}.Route()
	routeExample.ResourceVersion = "2"
	_, err := OshiftClient.RouteV1().Routes(namespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating route: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 15)

}
func TearDownTest(t *testing.T, modelName, namespace string) {
	if err := OshiftClient.RouteV1().Routes(namespace).Delete(context.TODO(), defaultRouteName, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("Couldn't delete route, err: %v", err)
	}

	objects.SharedAviGraphLister().Delete(modelName)
	integrationtest.DelSVC(t, namespace, "avisvc")
	integrationtest.DelEPS(t, namespace, "avisvc")
	integrationtest.DeleteNamespace(namespace)
	integrationtest.PollForCompletion(t, modelName, 10)
}

func VerifyModelDeleted(g *gomega.WithT, modelName string) {
	g.Eventually(func() interface{} {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if aviModel != nil {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return nodes[0].PoolRefs
		}
		return aviModel
	}, 30*time.Second).Should(gomega.BeNil())
}

func TestNSSyncFeatureWithCorrectEnvParameters(t *testing.T) {

	g := gomega.NewGomegaWithT(t)
	var nsLabel map[string]string
	nsLabel = map[string]string{
		"app": "migrate",
	}

	var found bool
	//Valid Namespace
	namespace1 := "routens"
	err := integrationtest.AddNamespace(t, namespace1, nsLabel)

	if err != nil {
		t.Fatal("Error while adding namespace")
	}
	modelName1 := "admin/cluster--Shared-L7-0"
	SetupRoute(t, modelName1, namespace1)
	g.Eventually(func() bool {
		found, _ = objects.SharedAviGraphLister().Get(modelName1)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))

	mcache := cache.SharedAviObjCache()

	poolName := fmt.Sprintf("cluster--foo.com_foo-%s-foo-avisvc", namespace1)
	poolKey := cache.NamespaceName{Namespace: "admin", Name: poolName}

	g.Eventually(func() bool {
		_, found := mcache.PoolCache.AviCacheGet(poolKey)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))

	UpdateRoute(t, modelName1, namespace1)

	g.Eventually(func() error {
		_, err := OshiftClient.RouteV1().Routes(namespace1).Get(context.TODO(), defaultRouteName, metav1.GetOptions{})
		return err
	}, 30*time.Second).Should(gomega.BeNil())

	poolName = fmt.Sprintf("cluster--foo.com_bar-%s-foo-avisvc", namespace1)
	poolKey = cache.NamespaceName{Namespace: "admin", Name: poolName}

	g.Eventually(func() bool {
		_, found := mcache.PoolCache.AviCacheGet(poolKey)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))

	TearDownTest(t, modelName1, namespace1)
	VerifyModelDeleted(g, modelName1)

	//Invalid Namespace
	utils.AviLog.Debug("Adding namespace with wrong label")

	namespace := "greenroutens"
	nsLabel = map[string]string{
		"app": "migrate1",
	}

	err = integrationtest.AddNamespace(t, namespace, nsLabel)
	modelName := "admin/cluster--Shared-L7-0"
	if err != nil {
		t.Fatal("Error while adding namespace")
	}

	SetupRoute(t, modelName, namespace)
	g.Eventually(func() bool {
		found, _ = objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(false))

	TearDownTest(t, modelName, namespace)
	VerifyModelDeleted(g, modelName)
}

func checkNSTransition(t *testing.T, oldLabels, newLabels map[string]string, oldFlag, newFlag bool, namespace, modelName string) {

	g := gomega.NewGomegaWithT(t)
	var found bool

	err := integrationtest.AddNamespace(t, namespace, oldLabels)
	if err != nil {
		t.Fatal("Error while adding namespace")
	}

	SetupRoute(t, modelName, namespace)
	time.Sleep(time.Second * 20)
	poolName := fmt.Sprintf("cluster--foo.com_foo-%s-foo-avisvc", namespace)

	mcache := cache.SharedAviObjCache()
	poolKey := cache.NamespaceName{Namespace: "admin", Name: poolName}
	if !oldFlag {
		g.Eventually(func() bool {
			found, _ = objects.SharedAviGraphLister().Get(modelName)

			return found
		}, 30*time.Second).Should(gomega.Equal(oldFlag))
	} else {
		g.Eventually(func() bool {
			_, found := mcache.PoolCache.AviCacheGet(poolKey)
			return found
		}, 30*time.Second).Should(gomega.Equal(oldFlag))
	}

	err = integrationtest.UpdateNamespace(t, namespace, newLabels)
	integrationtest.PollForCompletion(t, modelName, 5)
	if err != nil {
		t.Fatal("Error occurred while updating namespace")
	}

	g.Eventually(func() bool {
		_, found := mcache.PoolCache.AviCacheGet(poolKey)
		return found
	}, 30*time.Second).Should(gomega.Equal(newFlag))

	TearDownTest(t, modelName, namespace)
	VerifyModelDeleted(g, modelName)

}

func TestNSTransitionValidToInvalid(t *testing.T) {

	oldLabels := map[string]string{
		"app": "migrate",
	}
	newLabels := map[string]string{
		"app": "migrate2",
	}
	namespace := "routebluemigns"
	modelName := "admin/cluster--Shared-L7-0"
	checkNSTransition(t, oldLabels, newLabels, true, false, namespace, modelName)
}

func TestNSTransitionInvalidToValid(t *testing.T) {

	oldLabels := map[string]string{
		"app": "migrate2",
	}
	newLabels := map[string]string{
		"app": "migrate",
	}
	namespace := "routepurplemigns"
	modelName := "admin/cluster--Shared-L7-0"

	checkNSTransition(t, oldLabels, newLabels, false, true, namespace, modelName)
}

func TestNSTransitionInvalidToInvalid(t *testing.T) {

	oldLabels := map[string]string{
		"app": "migrate2",
	}
	newLabels := map[string]string{
		"app": "migrate1",
	}
	namespace := "routemagentamigns"
	modelName := "admin/cluster--Shared-L7-0"

	checkNSTransition(t, oldLabels, newLabels, false, false, namespace, modelName)
}
