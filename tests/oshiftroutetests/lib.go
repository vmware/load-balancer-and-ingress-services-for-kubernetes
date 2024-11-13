/*
 * Copyright 2022-2023 VMware, Inc.
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
	"strings"
	"testing"
	"time"

	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/onsi/gomega"
	routev1 "github.com/openshift/api/route/v1"
	oshiftfake "github.com/openshift/client-go/route/clientset/versioned/fake"
	"github.com/vmware/alb-sdk/go/models"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	defaultRouteName = "foo"
	defaultNamespace = "default"
	defaultHostname  = "foo.com"
	defaultService   = "avisvc"
	defaultModelName = "admin/cluster--Shared-L7-0"
	defaultKey       = "app"
	defaultValue     = "migrate"
	OshiftClient     *oshiftfake.Clientset
	defaultSubdomain = "foo"
)

// Candidate to move to lib
type FakeRoute struct {
	Name        string
	Namespace   string
	Hostname    string
	Path        string
	ServiceName string
	Backend2    string
	TargetPort  int
	Subdomain   string
}

func (rt FakeRoute) Route() *routev1.Route {
	if rt.Name == "" {
		rt.Name = defaultRouteName
	}
	if rt.Namespace == "" {
		rt.Namespace = defaultNamespace
	}
	if rt.Hostname == "" {
		rt.Hostname = defaultHostname
	}
	if rt.ServiceName == "" {
		rt.ServiceName = defaultService
	}
	weight := int32(100)
	routeExample := &routev1.Route{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       rt.Namespace,
			Name:            rt.Name,
			ResourceVersion: "1",
		},
		Spec: routev1.RouteSpec{
			Host: rt.Hostname,
			To: routev1.RouteTargetReference{
				Kind:   "Service",
				Name:   rt.ServiceName,
				Weight: &weight,
			},
		},
	}
	if rt.Path != "" {
		routeExample.Spec.Path = rt.Path
	}
	if rt.TargetPort != 0 {
		port := &routev1.RoutePort{
			TargetPort: intstr.FromInt(rt.TargetPort),
		}
		routeExample.Spec.Port = port
	}
	return routeExample
}

func (rt FakeRoute) RouteWithSubdomainAndNoHost() *routev1.Route {
	if rt.Name == "" {
		rt.Name = defaultRouteName
	}
	if rt.Namespace == "" {
		rt.Namespace = defaultNamespace
	}
	if rt.Subdomain == "" {
		rt.Subdomain = defaultSubdomain
	}
	if rt.ServiceName == "" {
		rt.ServiceName = defaultService
	}
	weight := int32(100)
	routeExample := &routev1.Route{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       rt.Namespace,
			Name:            rt.Name,
			ResourceVersion: "1",
		},
		Spec: routev1.RouteSpec{
			Subdomain: rt.Subdomain,
			To: routev1.RouteTargetReference{
				Kind:   "Service",
				Name:   rt.ServiceName,
				Weight: &weight,
			},
		},
	}
	if rt.Path != "" {
		routeExample.Spec.Path = rt.Path
	}
	if rt.TargetPort != 0 {
		port := &routev1.RoutePort{
			TargetPort: intstr.FromInt(rt.TargetPort),
		}
		routeExample.Spec.Port = port
	}
	return routeExample
}

func (rt FakeRoute) ABRoute(ratio ...int) *routev1.Route {
	routeExample := rt.Route()
	if rt.Backend2 == "" {
		rt.Backend2 = "absvc2"
	}
	weight2 := int32(200)
	if len(ratio) > 0 {
		weight2 = int32(ratio[0])
	}
	backend2 := routev1.RouteTargetReference{
		Kind:   "Service",
		Name:   rt.Backend2,
		Weight: &weight2,
	}
	routeExample.Spec.AlternateBackends = append(routeExample.Spec.AlternateBackends, backend2)
	return routeExample
}

func (rt FakeRoute) SecureRoute() *routev1.Route {
	routeExample := rt.Route()
	routeExample.Spec.TLS = &routev1.TLSConfig{
		Certificate:   "cert",
		CACertificate: "cacert",
		Key:           "key",
		Termination:   routev1.TLSTerminationEdge,
	}
	return routeExample
}

func (rt FakeRoute) SecureRouteWithSubdomainNoHost() *routev1.Route {
	routeExample := rt.RouteWithSubdomainAndNoHost()
	routeExample.Spec.TLS = &routev1.TLSConfig{
		Certificate:   "cert",
		CACertificate: "cacert",
		Key:           "key",
		Termination:   routev1.TLSTerminationEdge,
	}
	return routeExample
}

func (rt FakeRoute) SecureABRoute(ratio ...int) *routev1.Route {
	var routeExample *routev1.Route
	if len(ratio) > 0 {
		routeExample = rt.ABRoute(ratio[0])
	} else {
		routeExample = rt.ABRoute()
	}
	routeExample.Spec.TLS = &routev1.TLSConfig{
		Certificate:   "cert",
		CACertificate: "cacert",
		Key:           "key",
		Termination:   routev1.TLSTerminationEdge,
	}
	return routeExample
}

func (rt FakeRoute) SecureRouteNoCertKey() *routev1.Route {
	routeExample := rt.Route()
	routeExample.Spec.TLS = &routev1.TLSConfig{
		Termination: routev1.TLSTerminationEdge,
	}
	return routeExample
}

func (rt FakeRoute) SecureABRouteNoCertKey(ratio ...int) *routev1.Route {
	var routeExample *routev1.Route
	if len(ratio) > 0 {
		routeExample = rt.ABRoute(ratio[0])
	} else {
		routeExample = rt.ABRoute()
	}
	routeExample.Spec.TLS = &routev1.TLSConfig{
		Termination: routev1.TLSTerminationEdge,
	}
	return routeExample
}

func AddLabelToNamespace(key, value, namespace, modelName string, t *testing.T) {
	nsLabel := map[string]string{
		key: value,
	}
	integrationtest.AddNamespace(t, namespace, nsLabel)
}

func SetUpTestForRoute(t *testing.T, modelName string, models ...string) {
	AddLabelToNamespace(defaultKey, defaultValue, defaultNamespace, modelName, t)
	objects.SharedAviGraphLister().Delete(modelName)
	for _, model := range models {
		objects.SharedAviGraphLister().Delete(model)
	}

	integrationtest.CreateSVC(t, defaultNamespace, "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, defaultNamespace, "avisvc", false, false, "1.1.1")
	integrationtest.PollForCompletion(t, modelName, 5)
}

func TearDownTestForRoute(t *testing.T, modelName string) {
	objects.SharedAviGraphLister().Delete(modelName)
	integrationtest.DelSVC(t, "default", "avisvc")
	integrationtest.DelEPorEPS(t, "default", "avisvc")
}

// TO DO (Aakash) : Rename function to DeleteRouteAndVerify
func VerifyRouteDeletion(t *testing.T, g *gomega.WithT, aviModel interface{}, poolCount int, nsname ...string) {
	namespace, name := defaultNamespace, defaultRouteName
	if len(nsname) > 0 {
		namespace, name = strings.Split(nsname[0], "/")[0], strings.Split(nsname[0], "/")[1]
	}

	err := OshiftClient.RouteV1().Routes(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the route %v", err)
	}
	var nodes []*avinodes.AviVsNode
	g.Eventually(func() []*avinodes.AviPoolNode {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].PoolRefs
	}, 50*time.Second).Should(gomega.HaveLen(poolCount))

	g.Eventually(func() []*models.PoolGroupMember {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].PoolGroupRefs[0].Members
	}, 50*time.Second).Should(gomega.HaveLen(poolCount))
}

func ValidateModelCommon(t *testing.T, g *gomega.GomegaWithT) interface{} {

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(defaultModelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(defaultModelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()

	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))

	g.Expect(nodes[0].SharedVS).To(gomega.Equal(true))
	dsNodes := aviModel.(*avinodes.AviObjectGraph).GetAviHTTPDSNode()
	g.Expect(len(dsNodes)).To(gomega.Equal(1))

	g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))

	return aviModel
}

// TO DO (Aakash) : Rename function to DeleteSecureRouteAndVerify
func VerifySecureRouteDeletion(t *testing.T, g *gomega.WithT, modelName string, poolCount, snicount int, nsname ...string) {
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	VerifyRouteDeletion(t, g, aviModel, poolCount, nsname...)
	g.Eventually(func() int {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].SniNodes)
	}, 20*time.Second).Should(gomega.Equal(snicount))
}

func VerifySniNodeNoCA(g *gomega.WithT, sniVS *avinodes.AviVsNode) {
	g.Expect(sniVS.SSLKeyCertRefs).To(gomega.HaveLen(1))
	g.Expect(sniVS.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(sniVS.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(sniVS.HttpPolicyRefs).To(gomega.HaveLen(1))
}

func VerifySniNode(g *gomega.WithT, sniVS *avinodes.AviVsNode) {
	g.Expect(sniVS.CACertRefs).To(gomega.HaveLen(1))
	VerifySniNodeNoCA(g, sniVS)
}

func ValidateSniModel(t *testing.T, g *gomega.GomegaWithT, modelName string, redirect ...bool) interface{} {
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 50*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)

	g.Eventually(func() int {
		return len(aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes)
	}, 50*time.Second).Should(gomega.Equal(1))
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()

	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))

	g.Expect(nodes[0].SharedVS).To(gomega.Equal(true))
	redirectPol := 0
	if len(redirect) > 0 {
		if redirect[0] == true {
			redirectPol = 1
		}
	}
	g.Expect(nodes[0].HttpPolicyRefs).To(gomega.HaveLen(redirectPol))
	dsNodes := aviModel.(*avinodes.AviObjectGraph).GetAviHTTPDSNode()
	g.Expect(len(dsNodes)).To(gomega.Equal(1))

	return aviModel
}
