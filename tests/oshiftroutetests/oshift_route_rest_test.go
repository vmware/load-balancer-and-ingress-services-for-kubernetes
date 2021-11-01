package oshiftroutetests

import (
	"context"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	utils "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/testlib"

	"github.com/onsi/gomega"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TearDownRouteForRestCheck(t *testing.T, modelName string) {
	testlib.DeleteObject(t, lib.Route, defaultRouteName, defaultNamespace)
	TearDownTestForRoute(t, modelName)
}

func TestInsecureRouteStatusCheck(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{}.Route()
	_, err := utils.GetInformers().OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})

	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	var route *routev1.Route
	g.Eventually(func() string {
		route, _ = utils.GetInformers().OshiftClient.RouteV1().Routes("default").Get(context.TODO(), defaultRouteName, metav1.GetOptions{})
		if (len(route.Status.Ingress)) != 1 {
			return ""
		}
		return route.Status.Ingress[0].Host
	}, 30*time.Second).Should(gomega.Equal(defaultHostname))

	g.Eventually(func() string {
		route, _ = utils.GetInformers().OshiftClient.RouteV1().Routes("default").Get(context.TODO(), defaultRouteName, metav1.GetOptions{})
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
	_, err := utils.GetInformers().OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})

	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Path: "/bar"}.Route()
	testlib.UpdateObjectOrFail(t, lib.Route, routeExample)

	var route *routev1.Route
	g.Eventually(func() string {
		route, _ = utils.GetInformers().OshiftClient.RouteV1().Routes("default").Get(context.TODO(), defaultRouteName, metav1.GetOptions{})
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
	_, err := utils.GetInformers().OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})

	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Hostname: "bar.com"}.Route()
	testlib.UpdateObjectOrFail(t, lib.Route, routeExample)

	var route *routev1.Route
	g.Eventually(func() string {
		route, _ = utils.GetInformers().OshiftClient.RouteV1().Routes("default").Get(context.TODO(), defaultRouteName, metav1.GetOptions{})
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
	_, err := utils.GetInformers().OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample2 := FakeRoute{Name: "bar", Path: "/bar", ServiceName: "avisvc"}.Route()
	_, err = utils.GetInformers().OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	var route *routev1.Route
	g.Eventually(func() string {
		route, _ = utils.GetInformers().OshiftClient.RouteV1().Routes("default").Get(context.TODO(), defaultRouteName, metav1.GetOptions{})
		if (len(route.Status.Ingress)) != 1 {
			return ""
		}
		return route.Status.Ingress[0].Host
	}, 30*time.Second).Should(gomega.Equal(defaultHostname))
	g.Expect(route.Status.Ingress[0].Conditions[0].Message).Should(gomega.Equal("10.250.250.10"))

	g.Eventually(func() string {
		route, _ = utils.GetInformers().OshiftClient.RouteV1().Routes("default").Get(context.TODO(), "bar", metav1.GetOptions{})
		if (len(route.Status.Ingress)) != 1 {
			return ""
		}
		return route.Status.Ingress[0].Host
	}, 30*time.Second).Should(gomega.Equal(defaultHostname))
	g.Expect(route.Status.Ingress[0].Conditions[0].Message).Should(gomega.Equal("10.250.250.10"))

	testlib.DeleteObject(t, lib.Route, "bar", defaultNamespace)
	TearDownRouteForRestCheck(t, defaultModelName)
	objects.SharedAviGraphLister().Delete("admin/cluster--Shared-L7-1")
}

func TestRouteAlternateBackend(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	testlib.CreateSVC(t, "default", "absvc2", corev1.ServiceTypeClusterIP, false)
	testlib.CreateEP(t, "default", "absvc2", false, false, "3.3.3")
	routeExample := FakeRoute{}.ABRoute()
	_, err := utils.GetInformers().OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	var route *routev1.Route
	g.Eventually(func() string {
		route, _ = utils.GetInformers().OshiftClient.RouteV1().Routes("default").Get(context.TODO(), defaultRouteName, metav1.GetOptions{})
		if (len(route.Status.Ingress)) != 1 {
			return ""
		}
		return route.Status.Ingress[0].Host
	}, 60*time.Second).Should(gomega.Equal(defaultHostname))
	g.Expect(route.Status.Ingress[0].Conditions[0].Message).Should(gomega.Equal("10.250.250.10"))

	TearDownRouteForRestCheck(t, defaultModelName)
	objects.SharedAviGraphLister().Delete("admin/cluster--Shared-L7-1")
	testlib.DeleteObject(t, lib.Service, "absvc2", "default")
	testlib.DeleteObject(t, lib.Endpoint, "absvc2", "default")
}

func TestRouteWrongAlternateBackend(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{Backend2: "avisvc"}.ABRoute()
	_, err := utils.GetInformers().OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	var route *routev1.Route
	g.Eventually(func() string {
		route, _ = utils.GetInformers().OshiftClient.RouteV1().Routes("default").Get(context.TODO(), defaultRouteName, metav1.GetOptions{})
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
	testlib.CreateSVC(t, "default", "absvc2", corev1.ServiceTypeClusterIP, false)
	testlib.CreateEP(t, "default", "absvc2", false, false, "3.3.3")
	routeExample := FakeRoute{Path: "/foo"}.PassthroughABRoute()
	_, err := utils.GetInformers().OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	var route *routev1.Route
	g.Eventually(func() string {
		route, _ = utils.GetInformers().OshiftClient.RouteV1().Routes("default").Get(context.TODO(), defaultRouteName, metav1.GetOptions{})
		if (len(route.Status.Ingress)) != 1 {
			return ""
		}
		return route.Status.Ingress[0].Host
	}, 60*time.Second).Should(gomega.Equal(defaultHostname))
	g.Expect(route.Status.Ingress[0].Conditions[0].Message).Should(gomega.Equal("10.250.250.1"))

	TearDownRouteForRestCheck(t, DefaultPassthroughModel)
	testlib.DeleteObject(t, lib.Service, "absvc2", "default")
	testlib.DeleteObject(t, lib.Endpoint, "absvc2", "default")
}

func TestPassthroughRouteWrongAlternateBackend(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultPassthroughModel)
	routeExample := FakeRoute{Backend2: "avisvc"}.ABRoute()
	_, err := utils.GetInformers().OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	var route *routev1.Route
	g.Eventually(func() string {
		route, _ = utils.GetInformers().OshiftClient.RouteV1().Routes("default").Get(context.TODO(), defaultRouteName, metav1.GetOptions{})
		if (len(route.Status.Ingress)) != 1 {
			return ""
		}
		return route.Status.Ingress[0].Host
	}, 60*time.Second).Should(gomega.Equal(defaultHostname))

	g.Expect(route.Status.Ingress[0].Conditions[0].Message).Should(gomega.Equal(""))

	TearDownRouteForRestCheck(t, DefaultPassthroughModel)
	objects.SharedAviGraphLister().Delete(DefaultPassthroughModel)
}
