package oshiftroutetests

import (
	"context"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/onsi/gomega"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TearDownRouteForRestCheck(t *testing.T, modelName string) {
	err := OshiftClient.RouteV1().Routes(defaultNamespace).Delete(context.TODO(), defaultRouteName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the route %v", err)
	}
	TearDownTestForRoute(t, modelName)
}

func TestInsecureRouteStatusCheck(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{}.Route()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})

	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	var route *routev1.Route
	g.Eventually(func() string {
		route, _ = OshiftClient.RouteV1().Routes("default").Get(context.TODO(), defaultRouteName, metav1.GetOptions{})
		if (len(route.Status.Ingress)) != 1 {
			return ""
		}
		return route.Status.Ingress[0].Host
	}, 30*time.Second).Should(gomega.Equal(defaultHostname))

	g.Eventually(func() string {
		route, _ = OshiftClient.RouteV1().Routes("default").Get(context.TODO(), defaultRouteName, metav1.GetOptions{})
		if (len(route.Status.Ingress)) != 1 {
			return ""
		}
		if route.Status.Ingress[0].Conditions[0].LastTransitionTime == nil {
			return ""
		}
		return route.Status.Ingress[0].Conditions[0].Message
	}, 30*time.Second).Should(gomega.Equal("10.250.250.10"))

	TearDownRouteForRestCheck(t, defaultModelName)
}

func TestRouteStatusUpdatePath(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{}.Route()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})

	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Path: "/bar"}.Route()
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})

	if err != nil {
		t.Fatalf("error in updating route: %v", err)
	}

	var route *routev1.Route
	g.Eventually(func() string {
		route, _ = OshiftClient.RouteV1().Routes("default").Get(context.TODO(), defaultRouteName, metav1.GetOptions{})
		if (len(route.Status.Ingress)) != 1 {
			return ""
		}
		return route.Status.Ingress[0].Host
	}, 30*time.Second).Should(gomega.Equal(defaultHostname))
	g.Expect(route.Status.Ingress[0].Conditions[0].Message).Should(gomega.Equal("10.250.250.10"))

	TearDownRouteForRestCheck(t, defaultModelName)
}

func TestRouteStatusUpdateHost(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{}.Route()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})

	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Hostname: "bar.com"}.Route()
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})

	if err != nil {
		t.Fatalf("error in updating route: %v", err)
	}

	var route *routev1.Route
	g.Eventually(func() string {
		route, _ = OshiftClient.RouteV1().Routes("default").Get(context.TODO(), defaultRouteName, metav1.GetOptions{})
		if (len(route.Status.Ingress)) != 1 {
			return ""
		}
		return route.Status.Ingress[0].Host
	}, 30*time.Second).Should(gomega.Equal("bar.com"))
	g.Expect(route.Status.Ingress[0].Conditions[0].Message).Should(gomega.Equal("10.250.250.11"))

	TearDownRouteForRestCheck(t, defaultModelName)
	objects.SharedAviGraphLister().Delete("admin/cluster--Shared-L7-1")
}

func TestMultiRouteStatusSameHost(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	routeExample1 := FakeRoute{Path: "/foo", ServiceName: "avisvc"}.Route()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample2 := FakeRoute{Name: "bar", Path: "/bar", ServiceName: "avisvc"}.Route()
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	var route *routev1.Route
	g.Eventually(func() string {
		route, _ = OshiftClient.RouteV1().Routes("default").Get(context.TODO(), defaultRouteName, metav1.GetOptions{})
		if (len(route.Status.Ingress)) != 1 {
			return ""
		}
		return route.Status.Ingress[0].Host
	}, 30*time.Second).Should(gomega.Equal(defaultHostname))
	g.Expect(route.Status.Ingress[0].Conditions[0].Message).Should(gomega.Equal("10.250.250.10"))

	g.Eventually(func() string {
		route, _ = OshiftClient.RouteV1().Routes("default").Get(context.TODO(), "bar", metav1.GetOptions{})
		if (len(route.Status.Ingress)) != 1 {
			return ""
		}
		return route.Status.Ingress[0].Host
	}, 30*time.Second).Should(gomega.Equal(defaultHostname))
	g.Expect(route.Status.Ingress[0].Conditions[0].Message).Should(gomega.Equal("10.250.250.10"))

	err = OshiftClient.RouteV1().Routes(defaultNamespace).Delete(context.TODO(), "bar", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the route %v", err)
	}

	TearDownRouteForRestCheck(t, defaultModelName)
	objects.SharedAviGraphLister().Delete("admin/cluster--Shared-L7-1")
}

func TestRouteAlternateBackend(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	integrationtest.CreateSVC(t, "default", "absvc2", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, "default", "absvc2", false, false, "3.3.3")
	routeExample := FakeRoute{}.ABRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	var route *routev1.Route
	g.Eventually(func() string {
		route, _ = OshiftClient.RouteV1().Routes("default").Get(context.TODO(), defaultRouteName, metav1.GetOptions{})
		if (len(route.Status.Ingress)) != 1 {
			return ""
		}
		return route.Status.Ingress[0].Host
	}, 60*time.Second).Should(gomega.Equal(defaultHostname))
	g.Expect(route.Status.Ingress[0].Conditions[0].Message).Should(gomega.Equal("10.250.250.10"))

	TearDownRouteForRestCheck(t, defaultModelName)
	objects.SharedAviGraphLister().Delete("admin/cluster--Shared-L7-1")
	integrationtest.DelSVC(t, "default", "absvc2")
	integrationtest.DelEPS(t, "default", "absvc2")
}

func TestRouteWrongAlternateBackend(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{Backend2: "avisvc"}.ABRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	var route *routev1.Route
	g.Eventually(func() string {
		route, _ = OshiftClient.RouteV1().Routes("default").Get(context.TODO(), defaultRouteName, metav1.GetOptions{})
		if (len(route.Status.Ingress)) != 1 {
			return ""
		}
		return route.Status.Ingress[0].Host
	}, 60*time.Second).Should(gomega.Equal(defaultHostname))

	g.Expect(route.Status.Ingress[0].Conditions[0].Message).Should(gomega.Equal(""))

	TearDownRouteForRestCheck(t, defaultModelName)
	objects.SharedAviGraphLister().Delete("admin/cluster--Shared-L7-1")
}

func TestPassthroughRouteAlternateBackend(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultPassthroughModel)
	integrationtest.CreateSVC(t, "default", "absvc2", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, "default", "absvc2", false, false, "3.3.3")
	routeExample := FakeRoute{Path: "/foo"}.PassthroughABRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	var route *routev1.Route
	g.Eventually(func() string {
		route, _ = OshiftClient.RouteV1().Routes("default").Get(context.TODO(), defaultRouteName, metav1.GetOptions{})
		if (len(route.Status.Ingress)) != 1 {
			return ""
		}
		return route.Status.Ingress[0].Host
	}, 60*time.Second).Should(gomega.Equal(defaultHostname))
	g.Expect(route.Status.Ingress[0].Conditions[0].Message).Should(gomega.Equal("10.250.250.1"))

	TearDownRouteForRestCheck(t, DefaultPassthroughModel)
	integrationtest.DelSVC(t, "default", "absvc2")
	integrationtest.DelEPS(t, "default", "absvc2")
}

func TestPassthroughRouteWrongAlternateBackend(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultPassthroughModel)
	routeExample := FakeRoute{Backend2: "avisvc"}.ABRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	var route *routev1.Route
	g.Eventually(func() string {
		route, _ = OshiftClient.RouteV1().Routes("default").Get(context.TODO(), defaultRouteName, metav1.GetOptions{})
		if (len(route.Status.Ingress)) != 1 {
			return ""
		}
		return route.Status.Ingress[0].Host
	}, 60*time.Second).Should(gomega.Equal(defaultHostname))

	g.Expect(route.Status.Ingress[0].Conditions[0].Message).Should(gomega.Equal(""))

	TearDownRouteForRestCheck(t, DefaultPassthroughModel)
	objects.SharedAviGraphLister().Delete(DefaultPassthroughModel)
}
