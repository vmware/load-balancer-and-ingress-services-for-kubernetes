/*
* [2013] - [2019] Avi Networks Incorporated
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

package integrationtest

import (
	"testing"
	"time"

	"os"

	v1 "k8s.io/api/core/v1"
	extensionv1beta1 "k8s.io/api/extensions/v1beta1"

	"gitlab.eng.vmware.com/orion/akc/pkg/k8s"
	avinodes "gitlab.eng.vmware.com/orion/akc/pkg/nodes"
	"gitlab.eng.vmware.com/orion/akc/pkg/objects"
	meshutils "gitlab.eng.vmware.com/orion/container-lib/utils"
	corev1 "k8s.io/api/core/v1"

	"github.com/avinetworks/sdk/go/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

var KubeClient *k8sfake.Clientset
var ctrl *k8s.AviController

func SetUp() {
	KubeClient = k8sfake.NewSimpleClientset()
	registeredInformers := []string{meshutils.ServiceInformer, meshutils.EndpointInformer, meshutils.ExtV1IngressInformer, meshutils.SecretInformer, meshutils.NSInformer, meshutils.NodeInformer, meshutils.ConfigMapInformer}
	meshutils.NewInformers(meshutils.KubeClientIntf{KubeClient}, registeredInformers)
	informers := k8s.K8sinformers{Cs: KubeClient}
	os.Setenv("CTRL_USERNAME", "admin")
	os.Setenv("CTRL_PASSWORD", "admin")
	os.Setenv("CTRL_IPADDRESS", "localhost")
	os.Setenv("INGRESS_API", "extensionv1")
	os.Setenv("FULL_SYNC_INTERVAL", "60")
	ctrl = k8s.SharedAviController()
	stopCh := meshutils.SetupSignalHandler()
	k8s.PopulateCache()
	ctrlCh := make(chan struct{})
	ctrl.HandleConfigMap(informers, ctrlCh, stopCh)
	go ctrl.InitController(informers, ctrlCh, stopCh)
	AddConfigMap()
}

func AddConfigMap() {
	aviCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "avi-system",
			Name:      "avi-k8s-config",
		},
	}
	KubeClient.CoreV1().ConfigMaps("avi-system").Create(aviCM)

	PollForSyncStart(ctrl, 10)
}

// Fake ingress
type FakeIngress struct {
	DnsNames    []string
	Paths       []string
	tlsDnsNames [][]string
	Ips         []string
	HostNames   []string
	Namespace   string
	Name        string
	annotations map[string]string
	ServiceName string
}

func (ing FakeIngress) Ingress() *extensionv1beta1.Ingress {
	ingress := &extensionv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   ing.Namespace,
			Name:        ing.Name,
			Annotations: ing.annotations,
		},
		Spec: extensionv1beta1.IngressSpec{
			Rules: []extensionv1beta1.IngressRule{},
		},
		Status: extensionv1beta1.IngressStatus{
			LoadBalancer: v1.LoadBalancerStatus{
				Ingress: []v1.LoadBalancerIngress{},
			},
		},
	}
	for i, dnsName := range ing.DnsNames {
		path := "/foo"
		if len(ing.Paths) > i {
			path = ing.Paths[i]
		}
		ingress.Spec.Rules = append(ingress.Spec.Rules, extensionv1beta1.IngressRule{
			Host: dnsName,
			IngressRuleValue: extensionv1beta1.IngressRuleValue{
				HTTP: &extensionv1beta1.HTTPIngressRuleValue{
					Paths: []extensionv1beta1.HTTPIngressPath{extensionv1beta1.HTTPIngressPath{
						Path: path,
						Backend: extensionv1beta1.IngressBackend{ServiceName: ing.ServiceName, ServicePort: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: 8080,
						}},
					},
					},
				},
			},
		})
	}
	for _, hosts := range ing.tlsDnsNames {
		ingress.Spec.TLS = append(ingress.Spec.TLS, extensionv1beta1.IngressTLS{
			Hosts: hosts,
		})
	}
	for _, ip := range ing.Ips {
		ingress.Status.LoadBalancer.Ingress = append(ingress.Status.LoadBalancer.Ingress, v1.LoadBalancerIngress{
			IP: ip,
		})
	}
	for _, hostName := range ing.HostNames {
		ingress.Status.LoadBalancer.Ingress = append(ingress.Status.LoadBalancer.Ingress, v1.LoadBalancerIngress{
			Hostname: hostName,
		})
	}
	return ingress
}

func (ing FakeIngress) IngressMultiPath() *extensionv1beta1.Ingress {
	ingress := &extensionv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   ing.Namespace,
			Name:        ing.Name,
			Annotations: ing.annotations,
		},
		Spec: extensionv1beta1.IngressSpec{
			Rules: []extensionv1beta1.IngressRule{},
		},
		Status: extensionv1beta1.IngressStatus{
			LoadBalancer: v1.LoadBalancerStatus{
				Ingress: []v1.LoadBalancerIngress{},
			},
		},
	}
	for _, dnsName := range ing.DnsNames {
		var ingrPaths []extensionv1beta1.HTTPIngressPath
		for _, path := range ing.Paths {
			ingrPath := extensionv1beta1.HTTPIngressPath{
				Path: path,
				Backend: extensionv1beta1.IngressBackend{ServiceName: ing.ServiceName, ServicePort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 8080,
				}},
			}
			ingrPaths = append(ingrPaths, ingrPath)
		}
		ingress.Spec.Rules = append(ingress.Spec.Rules, extensionv1beta1.IngressRule{
			Host: dnsName,
			IngressRuleValue: extensionv1beta1.IngressRuleValue{
				HTTP: &extensionv1beta1.HTTPIngressRuleValue{
					Paths: ingrPaths,
				},
			},
		})
	}
	for _, hosts := range ing.tlsDnsNames {
		ingress.Spec.TLS = append(ingress.Spec.TLS, extensionv1beta1.IngressTLS{
			Hosts: hosts,
		})
	}
	for _, ip := range ing.Ips {
		ingress.Status.LoadBalancer.Ingress = append(ingress.Status.LoadBalancer.Ingress, v1.LoadBalancerIngress{
			IP: ip,
		})
	}
	for _, hostName := range ing.HostNames {
		ingress.Status.LoadBalancer.Ingress = append(ingress.Status.LoadBalancer.Ingress, v1.LoadBalancerIngress{
			Hostname: hostName,
		})
	}
	return ingress
}

func DetectModelChecksumChange(t *testing.T, key string, counter int) interface{} {
	// This method detects a change in the checksum and returns.
	count := 0
	initialcs := uint32(0)
	found, aviModel := objects.SharedAviGraphLister().Get(key)
	if found {
		initialcs = aviModel.(*avinodes.AviObjectGraph).GraphChecksum
	}
	for count < counter {
		found, aviModel = objects.SharedAviGraphLister().Get(key)
		if found {
			if initialcs == aviModel.(*avinodes.AviObjectGraph).GraphChecksum {
				count = count + 1
				time.Sleep(1 * time.Second)
			} else {
				return aviModel
			}
		}
	}
	return nil
}

func PollForCompletion(t *testing.T, key string, counter int) interface{} {
	count := 0
	for count < counter {
		found, aviModel := objects.SharedAviGraphLister().Get(key)
		if !found {
			time.Sleep(1 * time.Second)
			count = count + 1
		} else {
			return aviModel
		}
	}
	return nil
}

func PollForSyncStart(ctrl *k8s.AviController, counter int) bool {
	count := 0
	for count < counter {

		if ctrl.DisableSync {
			time.Sleep(1 * time.Second)
			count = count + 1
		} else {
			return true
		}
	}
	return false
}

type FakeService struct {
	Namespace    string
	Name         string
	annotations  map[string]string
	ServicePorts []Serviceport
}

type Serviceport struct {
	PortName   string
	PortNumber int32
	Protocol   v1.Protocol
	TargetPort int
}

func (svc FakeService) Service() *corev1.Service {
	var ports []corev1.ServicePort
	for _, svcport := range svc.ServicePorts {
		ports = append(ports, corev1.ServicePort{Name: svcport.PortName, Port: svcport.PortNumber, Protocol: svcport.Protocol, TargetPort: intstr.FromInt(svcport.TargetPort)})
	}
	svcExample := &corev1.Service{
		Spec: corev1.ServiceSpec{
			Type:  corev1.ServiceTypeClusterIP,
			Ports: ports,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: svc.Namespace,
			Name:      svc.Name,
		},
	}
	return svcExample
}

type fakeNode struct {
	Name    string
	podCIDR string
	nodeIP  string
	version string
}

func (node fakeNode) Node() *corev1.Node {
	nodeExample := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:            node.Name,
			ResourceVersion: node.version,
		},
		Spec: corev1.NodeSpec{
			PodCIDR: node.podCIDR,
		},
		Status: corev1.NodeStatus{
			Addresses: []corev1.NodeAddress{
				{
					Type:    "InternalIP",
					Address: node.nodeIP,
				},
			},
		},
	}
	return nodeExample
}

func GetStaticRoute(nodeAddr, prefixAddr, routeID string, mask int32) *models.StaticRoute {
	nodeAddrType := "V4"
	nexthop := models.IPAddr{
		Addr: &nodeAddr,
		Type: &nodeAddrType,
	}
	prefixAddrType := "V4"
	prefixIP := models.IPAddr{
		Addr: &prefixAddr,
		Type: &prefixAddrType,
	}
	prefix := models.IPAddrPrefix{
		IPAddr: &prefixIP,
		Mask:   &mask,
	}
	staticRoute := models.StaticRoute{
		NextHop: &nexthop,
		Prefix:  &prefix,
		RouteID: &routeID,
	}
	return &staticRoute
}

func CreateSVC(t *testing.T, ns string, Name string) {
	svcExample := (FakeService{
		Name:         Name,
		Namespace:    ns,
		ServicePorts: []Serviceport{{PortName: "foo", Protocol: "TCP", PortNumber: 8080, TargetPort: 8080}},
	}).Service()

	_, err := KubeClient.CoreV1().Services(ns).Create(svcExample)
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
}

func DelSVC(t *testing.T, ns string, Name string) {
	err := KubeClient.CoreV1().Services(ns).Delete(Name, nil)
	if err != nil {
		t.Fatalf("error in deleting Service: %v", err)
	}
}

func CreateEP(t *testing.T, ns string, Name string) {
	epExample := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      Name,
		},
		Subsets: []corev1.EndpointSubset{{
			Addresses: []corev1.EndpointAddress{{IP: "1.2.3.4"}},
			Ports:     []corev1.EndpointPort{{Name: "foo", Port: 8080, Protocol: "TCP"}},
		}},
	}
	_, err := KubeClient.CoreV1().Endpoints(ns).Create(epExample)
	if err != nil {
		t.Fatalf("error in creating Endpoint: %v", err)
	}
}

func DelEP(t *testing.T, ns string, Name string) {
	err := KubeClient.CoreV1().Endpoints(ns).Delete(Name, nil)
	if err != nil {
		t.Fatalf("error in deleting Endpoint: %v", err)
	}
}
