package oshiftroutetests

import (
	"ako/internal/cache"
	"ako/internal/objects"
	"testing"
	"time"

	"github.com/onsi/gomega"
	routev1 "github.com/openshift/api/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func SetupDomain() {
	mcache := cache.SharedAviObjCache()
	cloudObj := &cache.AviCloudPropertyCache{Name: "Default-Cloud", VType: "mock"}
	subdomains := []string{"avi.internal", ".com"}
	cloudObj.NSIpamDNS = subdomains
	mcache.CloudKeyCache.AviCacheAdd("Default-Cloud", cloudObj)
}

func TearDownRouteForRestCheck(t *testing.T, modelName string) {
	err := OshiftClient.RouteV1().Routes(DefaultNamespace).Delete(DefaultRouteName, nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the route %v", err)
	}
	TearDownTestForRoute(t, modelName)
}

func TestInsecureRouteStatusCheck(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultModelName)
	routeExample := FakeRoute{}.Route()
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample)

	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	var route *routev1.Route
	g.Eventually(func() string {
		route, _ = OshiftClient.RouteV1().Routes("default").Get(DefaultRouteName, metav1.GetOptions{})
		if (len(route.Status.Ingress)) != 1 {
			return ""
		}
		return route.Status.Ingress[0].Host
	}, 30*time.Second).Should(gomega.Equal(DefaultHostname))

	g.Eventually(func() string {
		route, _ = OshiftClient.RouteV1().Routes("default").Get(DefaultRouteName, metav1.GetOptions{})
		if (len(route.Status.Ingress)) != 1 {
			return ""
		}
		return route.Status.Ingress[0].Conditions[0].Message
	}, 30*time.Second).Should(gomega.Equal("10.250.250.10"))

	TearDownRouteForRestCheck(t, DefaultModelName)
}

func TestRouteStatusUpdatePath(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultModelName)
	routeExample := FakeRoute{}.Route()
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample)

	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Path: "/bar"}.Route()
	_, err = OshiftClient.RouteV1().Routes(DefaultNamespace).Update(routeExample)

	if err != nil {
		t.Fatalf("error in updating route: %v", err)
	}

	var route *routev1.Route
	g.Eventually(func() string {
		route, _ = OshiftClient.RouteV1().Routes("default").Get(DefaultRouteName, metav1.GetOptions{})
		if (len(route.Status.Ingress)) != 1 {
			return ""
		}
		return route.Status.Ingress[0].Host
	}, 30*time.Second).Should(gomega.Equal(DefaultHostname))
	g.Expect(route.Status.Ingress[0].Conditions[0].Message).Should(gomega.Equal("10.250.250.10"))

	TearDownRouteForRestCheck(t, DefaultModelName)
}

func TestRouteStatusUpdateHost(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultModelName)
	routeExample := FakeRoute{}.Route()
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample)

	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Hostname: "bar.com"}.Route()
	_, err = OshiftClient.RouteV1().Routes(DefaultNamespace).Update(routeExample)

	if err != nil {
		t.Fatalf("error in updating route: %v", err)
	}

	var route *routev1.Route
	g.Eventually(func() string {
		route, _ = OshiftClient.RouteV1().Routes("default").Get(DefaultRouteName, metav1.GetOptions{})
		if (len(route.Status.Ingress)) != 1 {
			return ""
		}
		return route.Status.Ingress[0].Host
	}, 30*time.Second).Should(gomega.Equal("bar.com"))
	g.Expect(route.Status.Ingress[0].Conditions[0].Message).Should(gomega.Equal("10.250.250.11"))

	TearDownRouteForRestCheck(t, DefaultModelName)
	objects.SharedAviGraphLister().Delete("admin/cluster--Shared-L7-1")
}

func TestMultiRouteStatusSameHost(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultModelName)
	routeExample1 := FakeRoute{Path: "/foo", ServiceName: "avisvc"}.Route()
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample1)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample2 := FakeRoute{Name: "bar", Path: "/bar", ServiceName: "avisvc"}.Route()
	_, err = OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample2)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	var route *routev1.Route
	g.Eventually(func() string {
		route, _ = OshiftClient.RouteV1().Routes("default").Get(DefaultRouteName, metav1.GetOptions{})
		if (len(route.Status.Ingress)) != 1 {
			return ""
		}
		return route.Status.Ingress[0].Host
	}, 30*time.Second).Should(gomega.Equal(DefaultHostname))
	g.Expect(route.Status.Ingress[0].Conditions[0].Message).Should(gomega.Equal("10.250.250.10"))

	g.Eventually(func() string {
		route, _ = OshiftClient.RouteV1().Routes("default").Get("bar", metav1.GetOptions{})
		if (len(route.Status.Ingress)) != 1 {
			return ""
		}
		return route.Status.Ingress[0].Host
	}, 30*time.Second).Should(gomega.Equal(DefaultHostname))
	g.Expect(route.Status.Ingress[0].Conditions[0].Message).Should(gomega.Equal("10.250.250.10"))

	err = OshiftClient.RouteV1().Routes(DefaultNamespace).Delete("bar", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the route %v", err)
	}

	TearDownRouteForRestCheck(t, DefaultModelName)
	objects.SharedAviGraphLister().Delete("admin/cluster--Shared-L7-1")
}
