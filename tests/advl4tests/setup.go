package advl4tests

import (
	"context"
	"testing"
	"time"

	"github.com/onsi/gomega"
	advl4v1alpha1pre1 "github.com/vmware-tanzu/service-apis/apis/v1alpha1pre1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha1/clientset/versioned/fake"
	v1beta1crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1beta1/clientset/versioned/fake"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"
	advl4fake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/service-apis/client/clientset/versioned/fake"
)

var KubeClient *k8sfake.Clientset
var AdvL4Client *advl4fake.Clientset
var CRDClient *crdfake.Clientset
var V1beta1CRDClient *v1beta1crdfake.Clientset
var ctrl *k8s.AviController

type FakeGateway struct {
	Name      string
	Namespace string
	GWClass   string
	IPAddress string
	Listeners []FakeGWListener
}

type FakeGWListener struct {
	Port     int32
	Protocol string
	Labels   map[string]string
}

func (gw FakeGateway) Gateway() *advl4v1alpha1pre1.Gateway {
	var fakeListeners []advl4v1alpha1pre1.Listener
	for _, listener := range gw.Listeners {
		fakeListeners = append(fakeListeners, advl4v1alpha1pre1.Listener{
			Port:     listener.Port,
			Protocol: advl4v1alpha1pre1.ProtocolType(listener.Protocol),
			Routes: advl4v1alpha1pre1.RouteBindingSelector{
				Resource: "services",
				RouteSelector: metav1.LabelSelector{
					MatchLabels: listener.Labels,
				},
			},
		})
	}

	gateway := &advl4v1alpha1pre1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: gw.Namespace,
			Name:      gw.Name,
		},
		Spec: advl4v1alpha1pre1.GatewaySpec{
			Class:     gw.GWClass,
			Listeners: fakeListeners,
		},
	}

	if gw.IPAddress != "" {
		gateway.Spec.Addresses = []advl4v1alpha1pre1.GatewayAddress{{
			Type:  advl4v1alpha1pre1.IPAddressType,
			Value: gw.IPAddress,
		}}
	}

	return gateway
}

func SetupGateway(t *testing.T, gwname, namespace, gwclass string, proxyAnnotate bool) {
	gateway := FakeGateway{
		Name:      gwname,
		Namespace: namespace,
		GWClass:   gwclass,
		Listeners: []FakeGWListener{{
			Port:     int32(8081),
			Protocol: "TCP",
			Labels: map[string]string{
				lib.GatewayNameLabelKey:      gwname,
				lib.GatewayNamespaceLabelKey: namespace,
				lib.GatewayTypeLabelKey:      "direct",
			},
		}},
	}

	gwCreate := gateway.Gateway()
	if proxyAnnotate {
		ann := map[string]string{lib.GwProxyProtocolEnableAnnotation: "true"}
		gwCreate.SetAnnotations(ann)
	}
	if _, err := lib.AKOControlConfig().AdvL4Clientset().NetworkingV1alpha1pre1().Gateways(namespace).Create(context.TODO(), gwCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Gateway: %v", err)
	}
}

func TeardownGateway(t *testing.T, gwname, namespace string) {
	if err := lib.AKOControlConfig().AdvL4Clientset().NetworkingV1alpha1pre1().Gateways(namespace).Delete(context.TODO(), gwname, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("error in deleting Gateway: %v", err)
	}
}

type FakeGWClass struct {
	Name       string
	Controller string
}

func (gwclass FakeGWClass) GatewayClass() *advl4v1alpha1pre1.GatewayClass {
	gatewayclass := &advl4v1alpha1pre1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: gwclass.Name,
		},
		Spec: advl4v1alpha1pre1.GatewayClassSpec{
			Controller: gwclass.Controller,
		},
	}

	return gatewayclass
}

func SetupGatewayClass(t *testing.T, gwclassName, controller string) {
	gatewayclass := FakeGWClass{
		Name:       gwclassName,
		Controller: controller,
	}

	gwClassCreate := gatewayclass.GatewayClass()
	if _, err := lib.AKOControlConfig().AdvL4Clientset().NetworkingV1alpha1pre1().GatewayClasses().Create(context.TODO(), gwClassCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding GatewayClass: %v", err)
	}
}

func TeardownGatewayClass(t *testing.T, gwClassName string) {
	if err := lib.AKOControlConfig().AdvL4Clientset().NetworkingV1alpha1pre1().GatewayClasses().Delete(context.TODO(), gwClassName, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("error in deleting GatewayClass: %v", err)
	}
}

func SetupAdvLBService(t *testing.T, svcname, namespace, gwname, gwnamespace string) {
	svc := integrationtest.FakeService{
		Name:      svcname,
		Namespace: namespace,
		Labels: map[string]string{
			lib.GatewayNameLabelKey:      gwname,
			lib.GatewayNamespaceLabelKey: gwnamespace,
			lib.GatewayTypeLabelKey:      "direct",
		},
		Type:         corev1.ServiceTypeLoadBalancer,
		ServicePorts: []integrationtest.Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8081, TargetPort: intstr.FromInt(8081)}},
	}

	svcCreate := svc.Service()
	if _, err := KubeClient.CoreV1().Services(namespace).Create(context.TODO(), svcCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	integrationtest.CreateEP(t, namespace, svcname, false, true, "1.1.1")
}

func SetupAdvLBServiceWithLoadBalancerClass(t *testing.T, svcname, namespace, gwname, gwnamespace, LBClass string) {
	svc := integrationtest.FakeService{
		Name:      svcname,
		Namespace: namespace,
		Labels: map[string]string{
			lib.GatewayNameLabelKey:      gwname,
			lib.GatewayNamespaceLabelKey: gwnamespace,
			lib.GatewayTypeLabelKey:      "direct",
		},
		Type:              corev1.ServiceTypeLoadBalancer,
		ServicePorts:      []integrationtest.Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8081, TargetPort: intstr.FromInt(8081)}},
		LoadBalancerClass: LBClass,
	}

	svcCreate := svc.Service()
	if _, err := KubeClient.CoreV1().Services(namespace).Create(context.TODO(), svcCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	integrationtest.CreateEP(t, namespace, svcname, false, true, "1.1.1")
}

func TeardownAdvLBService(t *testing.T, svcname, namespace string) {
	if err := KubeClient.CoreV1().Services(namespace).Delete(context.TODO(), svcname, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("error in deleting AdvLB Service: %v", err)
	}
	integrationtest.DelEP(t, namespace, svcname)
}

func VerifyGatewayVSNodeDeletion(g *gomega.WithT, modelName string) {
	g.Eventually(func() interface{} {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		return aviModel
	}, 30*time.Second).Should(gomega.BeNil())
}
