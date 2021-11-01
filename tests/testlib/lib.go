/*
 * Copyright 2019-2020 VMware, Inc.
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

package testlib

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/vmware-tanzu/service-apis/apis/v1alpha1pre1"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/api"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/onsi/gomega"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/vmware/alb-sdk/go/models"
	corev1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	servicesapi "sigs.k8s.io/service-apis/apis/v1alpha1"
)

// constants to be used for creating K8s objs and verifying Avi objs
const (
	SINGLEPORTSVC       = "testsvc"                            // single port service name
	MULTIPORTSVC        = "testsvcmulti"                       // multi port service name
	NAMESPACE           = "red-ns"                             // namespace
	AVINAMESPACE        = "admin"                              // avi namespace
	AKOTENANT           = "akotenant"                          // ako tenant where TENANTS_PER_CLUSTER is enabled
	SINGLEPORTMODEL     = "admin/cluster--red-ns-testsvc"      // single port model name
	MULTIPORTMODEL      = "admin/cluster--red-ns-testsvcmulti" // multi port model name
	RANDOMUUID          = "random-uuid"                        // random avi object uuid
	DefaultIngressClass = "avi-lb"
)

var AllModels = []string{
	"admin/cluster--Shared-L7-0",
	"admin/cluster--Shared-L7-1",
	"admin/cluster--Shared-L7-2",
	"admin/cluster--Shared-L7-3",
	"admin/cluster--Shared-L7-4",
	"admin/cluster--Shared-L7-5",
	"admin/cluster--Shared-L7-6",
	"admin/cluster--Shared-L7-7",
	"admin/cluster--Shared-L7-EVH-0",
	"admin/cluster--Shared-L7-EVH-1",
	"admin/cluster--Shared-L7-EVH-2",
	"admin/cluster--Shared-L7-EVH-3",
	"admin/cluster--Shared-L7-EVH-4",
	"admin/cluster--Shared-L7-EVH-5",
	"admin/cluster--Shared-L7-EVH-6",
	"admin/cluster--Shared-L7-EVH-7",
}

// Config Map
func AddConfigMap() {
	aviCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: utils.GetAKONamespace(),
			Name:      lib.AviConfigMap,
		},
	}
	utils.GetInformers().ClientSet.CoreV1().ConfigMaps(utils.GetAKONamespace()).Create(context.TODO(), aviCM, metav1.CreateOptions{})
}

// Namespace
type FakeNamespace struct {
	Name   string
	Labels map[string]string
}

func (namespace FakeNamespace) Namespace() *corev1.Namespace {
	FakeNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   namespace.Name,
			Labels: namespace.Labels,
		},
	}
	return FakeNamespace
}

func AddNamespace(t *testing.T, nsName string, labels map[string]string) error {
	nsMetaOptions := (FakeNamespace{
		Name:   nsName,
		Labels: labels,
	}).Namespace()
	nsMetaOptions.ResourceVersion = "1"
	ns, err := utils.GetInformers().ClientSet.CoreV1().Namespaces().Get(context.TODO(), nsName, metav1.GetOptions{})
	if err != nil {
		_, err = utils.GetInformers().ClientSet.CoreV1().Namespaces().Create(context.TODO(), nsMetaOptions, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Error occurred while Adding namespace: %v", err)
		}
		return nil
	}
	nsLabels := ns.GetLabels()
	if len(nsLabels) == 0 {
		UpdateNamespace(t, nsName, labels)
	}
	return nil
}

func UpdateNamespace(t *testing.T, nsName string, labels map[string]string) {
	nsMetaOptions := (FakeNamespace{
		Name:   nsName,
		Labels: labels,
	}).Namespace()
	nsMetaOptions.ResourceVersion = "2"
	UpdateObjectOrFail(t, lib.Namespace, nsMetaOptions)
}

func WaitTillNamespaceDelete(nsName string, retry_count int) {
	_, err := utils.GetInformers().ClientSet.CoreV1().Namespaces().Get(context.TODO(), nsName, metav1.GetOptions{})
	if err == nil {
		//NS still exists
		if retry_count > 0 {
			time.Sleep(time.Second * 1)
			WaitTillNamespaceDelete(nsName, retry_count-1)
		}
	}
}

// Secret
type FakeSecret struct {
	Cert      string
	Key       string
	Name      string
	Namespace string
}

func (secret FakeSecret) Secret() *corev1.Secret {
	data := map[string][]byte{
		"tls.crt": []byte(secret.Cert),
		"tls.key": []byte(secret.Key),
	}
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: secret.Namespace,
			Name:      secret.Name,
		},
		Data: data,
	}
}

func AddSecret(secretName string, namespace string, cert string, key string) {
	fakeSecret := (FakeSecret{
		Cert:      cert,
		Key:       key,
		Namespace: namespace,
		Name:      secretName,
	}).Secret()
	utils.GetInformers().ClientSet.CoreV1().Secrets(namespace).Create(context.TODO(), fakeSecret, metav1.CreateOptions{})
}

// Ingress
type FakeIngress struct {
	DnsNames     []string
	Paths        []string
	Ips          []string
	HostNames    []string
	Namespace    string
	Name         string
	ClassName    string
	annotations  map[string]string
	ServiceName  string
	TlsSecretDNS map[string][]string
}

func (ing FakeIngress) Ingress(multiport ...bool) *networking.Ingress {
	ingress := &networking.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   ing.Namespace,
			Name:        ing.Name,
			Annotations: ing.annotations,
		},
		Spec: networking.IngressSpec{
			Rules: []networking.IngressRule{},
		},
		Status: networking.IngressStatus{
			LoadBalancer: corev1.LoadBalancerStatus{
				Ingress: []corev1.LoadBalancerIngress{},
			},
		},
	}
	for i, dnsName := range ing.DnsNames {
		path := "/foo"
		if len(ing.Paths) > i {
			path = ing.Paths[i]
		}
		if len(multiport) > 0 {
			ingress.Spec.Rules = append(ingress.Spec.Rules, networking.IngressRule{
				Host: dnsName,
				IngressRuleValue: networking.IngressRuleValue{
					HTTP: &networking.HTTPIngressRuleValue{
						Paths: []networking.HTTPIngressPath{{
							Path: "/foo",
							Backend: networking.IngressBackend{Service: &networking.IngressServiceBackend{
								Name: ing.ServiceName,
								Port: networking.ServiceBackendPort{Name: "foo0"},
							}}}}}}},
			)
			ingress.Spec.Rules = append(ingress.Spec.Rules, networking.IngressRule{
				Host: dnsName,
				IngressRuleValue: networking.IngressRuleValue{
					HTTP: &networking.HTTPIngressRuleValue{
						Paths: []networking.HTTPIngressPath{{
							Path: "/bar",
							Backend: networking.IngressBackend{Service: &networking.IngressServiceBackend{
								Name: ing.ServiceName,
								Port: networking.ServiceBackendPort{Name: "foo1"},
							}}}}}}},
			)
		} else {
			ingress.Spec.Rules = append(ingress.Spec.Rules, networking.IngressRule{
				Host: dnsName,
				IngressRuleValue: networking.IngressRuleValue{
					HTTP: &networking.HTTPIngressRuleValue{
						Paths: []networking.HTTPIngressPath{{
							Path: path,
							Backend: networking.IngressBackend{Service: &networking.IngressServiceBackend{
								Name: ing.ServiceName,
								Port: networking.ServiceBackendPort{Number: 8080},
							}}}}}}},
			)
		}
	}
	for secret, hosts := range ing.TlsSecretDNS {
		ingress.Spec.TLS = append(ingress.Spec.TLS, networking.IngressTLS{
			Hosts:      hosts,
			SecretName: secret,
		})
	}
	for i := range ing.Ips {
		hostname := ""
		if len(ing.HostNames) >= i+1 {
			hostname = ing.HostNames[i]
		}
		ingress.Status.LoadBalancer.Ingress = append(ingress.Status.LoadBalancer.Ingress, corev1.LoadBalancerIngress{
			IP:       ing.Ips[i],
			Hostname: hostname,
		})
	}
	if ing.ClassName != "" {
		ingress.Spec.IngressClassName = &ing.ClassName
	}
	return ingress
}

func (ing FakeIngress) IngressMultiPort() *networking.Ingress {
	ingress := &networking.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   ing.Namespace,
			Name:        ing.Name,
			Annotations: ing.annotations,
		},
		Spec: networking.IngressSpec{
			Rules: []networking.IngressRule{},
		},
		Status: networking.IngressStatus{
			LoadBalancer: corev1.LoadBalancerStatus{
				Ingress: []corev1.LoadBalancerIngress{},
			},
		},
	}
	for _, dnsName := range ing.DnsNames {
		ingress.Spec.Rules = append(ingress.Spec.Rules, networking.IngressRule{
			Host: dnsName,
			IngressRuleValue: networking.IngressRuleValue{
				HTTP: &networking.HTTPIngressRuleValue{
					Paths: []networking.HTTPIngressPath{{
						Path: "/foo",
						Backend: networking.IngressBackend{Service: &networking.IngressServiceBackend{
							Name: ing.ServiceName,
							Port: networking.ServiceBackendPort{Number: 8080},
						}}}}}}},
		)
		ingress.Spec.Rules = append(ingress.Spec.Rules, networking.IngressRule{
			Host: dnsName,
			IngressRuleValue: networking.IngressRuleValue{
				HTTP: &networking.HTTPIngressRuleValue{
					Paths: []networking.HTTPIngressPath{{
						Path: "/bar",
						Backend: networking.IngressBackend{Service: &networking.IngressServiceBackend{
							Name: ing.ServiceName,
							Port: networking.ServiceBackendPort{Name: "foo1"},
						}}}}}}},
		)

	}
	for secret, hosts := range ing.TlsSecretDNS {
		ingress.Spec.TLS = append(ingress.Spec.TLS, networking.IngressTLS{
			Hosts:      hosts,
			SecretName: secret,
		})
	}
	for i := range ing.Ips {
		hostname := ""
		if len(ing.HostNames) >= i+1 {
			hostname = ing.HostNames[i]
		}
		ingress.Status.LoadBalancer.Ingress = append(ingress.Status.LoadBalancer.Ingress, corev1.LoadBalancerIngress{
			IP:       ing.Ips[i],
			Hostname: hostname,
		})
	}
	if ing.ClassName != "" {
		ingress.Spec.IngressClassName = &ing.ClassName
	}
	return ingress
}

func (ing FakeIngress) SecureIngress() *networking.Ingress {
	ingress := &networking.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   ing.Namespace,
			Name:        ing.Name,
			Annotations: ing.annotations,
		},
		Spec: networking.IngressSpec{
			Rules: []networking.IngressRule{},
		},
		Status: networking.IngressStatus{
			LoadBalancer: corev1.LoadBalancerStatus{
				Ingress: []corev1.LoadBalancerIngress{},
			},
		},
	}
	for i, dnsName := range ing.DnsNames {
		path := "/foo"
		if len(ing.Paths) > i {
			path = ing.Paths[i]
		}
		ingress.Spec.Rules = append(ingress.Spec.Rules, networking.IngressRule{
			Host: dnsName,
			IngressRuleValue: networking.IngressRuleValue{
				HTTP: &networking.HTTPIngressRuleValue{
					Paths: []networking.HTTPIngressPath{{
						Path: path,
						Backend: networking.IngressBackend{Service: &networking.IngressServiceBackend{
							Name: ing.ServiceName,
							Port: networking.ServiceBackendPort{Number: 8080},
						}}}}}}},
		)
	}

	for _, ip := range ing.Ips {
		ingress.Status.LoadBalancer.Ingress = append(ingress.Status.LoadBalancer.Ingress, corev1.LoadBalancerIngress{
			IP: ip,
		})
	}
	for _, hostName := range ing.HostNames {
		ingress.Status.LoadBalancer.Ingress = append(ingress.Status.LoadBalancer.Ingress, corev1.LoadBalancerIngress{
			Hostname: hostName,
		})
	}
	if ing.ClassName != "" {
		ingress.Spec.IngressClassName = &ing.ClassName
	}
	return ingress
}

func (ing FakeIngress) IngressNoHost() *networking.Ingress {
	ingress := &networking.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   ing.Namespace,
			Name:        ing.Name,
			Annotations: ing.annotations,
		},
		Spec: networking.IngressSpec{
			Rules: []networking.IngressRule{},
		},
		Status: networking.IngressStatus{
			LoadBalancer: corev1.LoadBalancerStatus{
				Ingress: []corev1.LoadBalancerIngress{},
			},
		},
	}
	for _, path := range ing.Paths {
		ingress.Spec.Rules = append(ingress.Spec.Rules, networking.IngressRule{
			IngressRuleValue: networking.IngressRuleValue{
				HTTP: &networking.HTTPIngressRuleValue{
					Paths: []networking.HTTPIngressPath{{
						Path: path,
						Backend: networking.IngressBackend{Service: &networking.IngressServiceBackend{
							Name: ing.ServiceName,
							Port: networking.ServiceBackendPort{Number: 8080},
						}}}}}}},
		)
	}
	for _, ip := range ing.Ips {
		ingress.Status.LoadBalancer.Ingress = append(ingress.Status.LoadBalancer.Ingress, corev1.LoadBalancerIngress{
			IP: ip,
		})
	}
	for _, hostName := range ing.HostNames {
		ingress.Status.LoadBalancer.Ingress = append(ingress.Status.LoadBalancer.Ingress, corev1.LoadBalancerIngress{
			Hostname: hostName,
		})
	}
	if ing.ClassName != "" {
		ingress.Spec.IngressClassName = &ing.ClassName
	}
	return ingress
}

func (ing FakeIngress) IngressOnlyHostNoBackend() *networking.Ingress {
	ingress := &networking.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   ing.Namespace,
			Name:        ing.Name,
			Annotations: ing.annotations,
		},
		Spec: networking.IngressSpec{
			Rules: nil,
		},
	}
	ingress.Spec.Rules = append(ingress.Spec.Rules, networking.IngressRule{
		IngressRuleValue: networking.IngressRuleValue{
			HTTP: nil,
		},
	})
	if ing.ClassName != "" {
		ingress.Spec.IngressClassName = &ing.ClassName
	}

	return ingress
}

func (ing FakeIngress) IngressMultiPath() *networking.Ingress {
	ingress := &networking.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   ing.Namespace,
			Name:        ing.Name,
			Annotations: ing.annotations,
		},
		Spec: networking.IngressSpec{
			Rules: []networking.IngressRule{},
		},
		Status: networking.IngressStatus{
			LoadBalancer: corev1.LoadBalancerStatus{
				Ingress: []corev1.LoadBalancerIngress{},
			},
		},
	}
	if ing.ClassName != "" {
		ingress.Spec.IngressClassName = &ing.ClassName
	}
	for _, dnsName := range ing.DnsNames {
		var ingrPaths []networking.HTTPIngressPath
		for _, path := range ing.Paths {
			ingrPath := networking.HTTPIngressPath{
				Path: path,
				Backend: networking.IngressBackend{Service: &networking.IngressServiceBackend{
					Name: ing.ServiceName,
					Port: networking.ServiceBackendPort{Number: 8080},
				}},
			}
			ingrPaths = append(ingrPaths, ingrPath)
		}
		ingress.Spec.Rules = append(ingress.Spec.Rules, networking.IngressRule{
			Host: dnsName,
			IngressRuleValue: networking.IngressRuleValue{
				HTTP: &networking.HTTPIngressRuleValue{
					Paths: ingrPaths,
				},
			},
		})
	}

	for secret, hosts := range ing.TlsSecretDNS {
		ingress.Spec.TLS = append(ingress.Spec.TLS, networking.IngressTLS{
			Hosts:      hosts,
			SecretName: secret,
		})
	}
	for _, ip := range ing.Ips {
		ingress.Status.LoadBalancer.Ingress = append(ingress.Status.LoadBalancer.Ingress, corev1.LoadBalancerIngress{
			IP: ip,
		})
	}
	for _, hostName := range ing.HostNames {
		ingress.Status.LoadBalancer.Ingress = append(ingress.Status.LoadBalancer.Ingress, corev1.LoadBalancerIngress{
			Hostname: hostName,
		})
	}
	return ingress
}

// Service
type FakeService struct {
	Namespace      string
	Name           string
	Labels         map[string]string
	Type           corev1.ServiceType
	LoadBalancerIP string
	ServicePorts   []Serviceport
	Selectors      map[string]string
	Annotations    map[string]string
}

type Serviceport struct {
	PortName   string
	PortNumber int32
	NodePort   int32
	Protocol   corev1.Protocol
	TargetPort int
}

func (svc FakeService) Service() *corev1.Service {
	var ports []corev1.ServicePort
	for _, svcport := range svc.ServicePorts {
		ports = append(ports, corev1.ServicePort{
			Name:       svcport.PortName,
			Port:       svcport.PortNumber,
			Protocol:   svcport.Protocol,
			TargetPort: intstr.FromInt(svcport.TargetPort),
			NodePort:   svcport.NodePort,
		})
	}
	svcExample := &corev1.Service{
		Spec: corev1.ServiceSpec{
			Type:           svc.Type,
			Ports:          ports,
			LoadBalancerIP: svc.LoadBalancerIP,
			Selector:       svc.Selectors,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   svc.Namespace,
			Name:        svc.Name,
			Labels:      svc.Labels,
			Annotations: svc.Annotations,
		},
	}
	return svcExample
}

/*
CreateSVC creates a sample service of type: Type
if multiPort: True, the service gets created with 3 ports as follows
ServicePorts: [
	{Name: "foo0", Port: 8080, Protocol: "TCP", TargetPort: 8080},
	{Name: "foo1", Port: 8081, Protocol: "TCP", TargetPort: 8081},
	{Name: "foo2", Port: 8082, Protocol: "TCP", TargetPort: 8082},
]
*/
func CreateSVC(t *testing.T, ns string, Name string, Type corev1.ServiceType, multiPort bool) {
	selectors := make(map[string]string)
	CreateServiceWithSelectors(t, ns, Name, Type, multiPort, selectors)
}

func CreateServiceWithSelectors(t *testing.T, ns string, Name string, Type corev1.ServiceType, multiPort bool, selectors map[string]string) {
	svcExample := ConstructService(ns, Name, Type, multiPort, selectors)
	_, err := utils.GetInformers().ClientSet.CoreV1().Services(ns).Create(context.TODO(), svcExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
}

func UpdateSVC(t *testing.T, ns string, Name string, Type corev1.ServiceType, multiPort bool) {
	selectors := make(map[string]string)
	UpdateServiceWithSelectors(t, ns, Name, Type, multiPort, selectors)
}

func UpdateServiceWithSelectors(t *testing.T, ns string, Name string, Type corev1.ServiceType, multiPort bool, selectors map[string]string) {
	svcExample := ConstructService(ns, Name, Type, multiPort, selectors)
	svcExample.ResourceVersion = "2"
	UpdateObjectOrFail(t, lib.Service, svcExample)
}

func ConstructService(ns string, Name string, Type corev1.ServiceType, multiPort bool, selectors map[string]string) *corev1.Service {
	var servicePorts []Serviceport
	numPorts := 1
	if multiPort {
		numPorts = 3
	}

	for i := 0; i < numPorts; i++ {
		mPort := 8080 + i
		sp := Serviceport{
			PortName:   fmt.Sprintf("foo%d", i),
			PortNumber: int32(mPort),
			Protocol:   "TCP",
			TargetPort: mPort,
		}
		if Type != corev1.ServiceTypeClusterIP {
			// set nodeport value in case of LoadBalancer and NodePort service type
			nodePort := 31030 + i
			sp.NodePort = int32(nodePort)
		}
		servicePorts = append(servicePorts, sp)
	}

	svcExample := (FakeService{Name: Name, Namespace: ns, Type: Type, ServicePorts: servicePorts, Selectors: selectors}).Service()
	return svcExample
}

// Node
type FakeNode struct {
	Name               string
	PodCIDR            string
	NodeIP             string
	Version            string
	PodCIDRsAnnotation string
}

func (node FakeNode) Node() *corev1.Node {
	nodeExample := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:            node.Name,
			ResourceVersion: node.Version,
		},
		Spec: corev1.NodeSpec{
			PodCIDR: node.PodCIDR,
		},
		Status: corev1.NodeStatus{
			Addresses: []corev1.NodeAddress{
				{
					Type:    "InternalIP",
					Address: node.NodeIP,
				},
			},
		},
	}
	if node.PodCIDRsAnnotation != "" {
		nodeExample.Annotations = map[string]string{
			lib.StaticRouteAnnotation: node.PodCIDRsAnnotation,
		}
	}
	return nodeExample
}

func CreateNode(t *testing.T, nodeName string, nodeIP string) {
	modelName := "admin/global"
	objects.SharedAviGraphLister().Delete(modelName)
	nodeExample := (FakeNode{
		Name:    nodeName,
		PodCIDR: "10.244.0.0/24",
		Version: "1",
		NodeIP:  nodeIP,
	}).Node()

	_, err := utils.GetInformers().ClientSet.CoreV1().Nodes().Create(context.TODO(), nodeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	PollForCompletion(t, modelName, 5)
}

func DeleteNode(t *testing.T, nodeName string) {
	modelName := "admin/global"
	objects.SharedAviGraphLister().Delete(modelName)
	DeleteObject(t, lib.Node, nodeName)
	PollForCompletion(t, modelName, 5)
}

/*
CreateEP creates a sample Endpoint object
if multiPort: False and multiAddress: False
	1.1.1.1:8080
if multiPort: True and multiAddress: False
	1.1.1.1:8080,
	1.1.1.2:8081,
	1.1.1.3:8082
if multiPort: False and multiAddress: True
	1.1.1.1:8080, 1.1.1.2:8080, 1.1.1.2:8080
if multiPort: True and multiAddress: True
	1.1.1.1:8080, 1.1.1.2:8080, 1.1.1.3:8080,
	1.1.1.4:8081, 1.1.1.5:8081,
	1.1.1.6:8082
*/
func CreateEP(t *testing.T, ns string, Name string, multiPort bool, multiAddress bool, addressPrefix string) {
	if addressPrefix == "" {
		addressPrefix = "1.1.1"
	}
	var endpointSubsets []corev1.EndpointSubset
	numPorts, numAddresses, addressStart := 1, 1, 0
	if multiPort {
		numPorts = 3
	}
	if multiAddress {
		numAddresses, addressStart = 3, 0
	}

	for i := 0; i < numPorts; i++ {
		mPort := 8080 + i
		var epAddresses []corev1.EndpointAddress
		for j := 0; j < numAddresses; j++ {
			epAddresses = append(epAddresses, corev1.EndpointAddress{IP: fmt.Sprintf("%s.%d", addressPrefix, addressStart+j+i+1)})
		}
		numAddresses = numAddresses - 1
		addressStart = addressStart + numAddresses
		endpointSubsets = append(endpointSubsets, corev1.EndpointSubset{
			Addresses: epAddresses,
			Ports: []corev1.EndpointPort{{
				Name:     fmt.Sprintf("foo%d", i),
				Port:     int32(mPort),
				Protocol: "TCP",
			}},
		})
	}

	epExample := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: Name},
		Subsets:    endpointSubsets,
	}
	_, err := utils.GetInformers().ClientSet.CoreV1().Endpoints(ns).Create(context.TODO(), epExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in creating Endpoint: %v", err)
	}
}

func ScaleCreateEP(t *testing.T, ns string, Name string) {
	epExample := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      Name,
		},
		Subsets: []corev1.EndpointSubset{{
			Addresses: []corev1.EndpointAddress{{IP: "1.2.3.4"}, {IP: "1.2.3.5"}},
			Ports:     []corev1.EndpointPort{{Name: "foo", Port: 8080, Protocol: "TCP"}},
		}},
	}
	epExample.ResourceVersion = "2"
	_, err := utils.GetInformers().ClientSet.CoreV1().Endpoints(ns).Update(context.TODO(), epExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in creating Endpoint: %v", err)
	}
}

// HostRule
type FakeHostRule struct {
	Name               string
	Namespace          string
	Fqdn               string
	SslKeyCertificate  string
	SslProfile         string
	WafPolicy          string
	ApplicationProfile string
	EnableVirtualHost  bool
	AnalyticsProfile   string
	ErrorPageProfile   string
	Datascripts        []string
	HttpPolicySets     []string
	GslbFqdn           string
}

func (hr FakeHostRule) HostRule() *akov1alpha1.HostRule {
	enable := true
	hostrule := &akov1alpha1.HostRule{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: hr.Namespace,
			Name:      hr.Name,
		},
		Spec: akov1alpha1.HostRuleSpec{
			VirtualHost: akov1alpha1.HostRuleVirtualHost{
				Fqdn: hr.Fqdn,
				TLS: akov1alpha1.HostRuleTLS{
					SSLKeyCertificate: akov1alpha1.HostRuleSecret{
						Name: hr.SslKeyCertificate,
						Type: "ref",
					},
					SSLProfile:  hr.SslProfile,
					Termination: "edge",
				},
				HTTPPolicy: akov1alpha1.HostRuleHTTPPolicy{
					PolicySets: hr.HttpPolicySets,
					Overwrite:  false,
				},
				WAFPolicy:          hr.WafPolicy,
				ApplicationProfile: hr.ApplicationProfile,
				AnalyticsProfile:   hr.AnalyticsProfile,
				ErrorPageProfile:   hr.ErrorPageProfile,
				Datascripts:        hr.Datascripts,
				EnableVirtualHost:  &enable,
				Gslb: akov1alpha1.HostRuleGSLB{
					Fqdn: hr.GslbFqdn,
				},
			},
		},
	}

	return hostrule
}

func SetupHostRule(t *testing.T, hrname, fqdn string, secure bool, gslbHost ...string) {
	hostrule := FakeHostRule{
		Name:               hrname,
		Namespace:          "default",
		Fqdn:               fqdn,
		WafPolicy:          "thisisaviref-waf",
		ApplicationProfile: "thisisaviref-appprof",
		AnalyticsProfile:   "thisisaviref-analyticsprof",
		ErrorPageProfile:   "thisisaviref-errorprof",
		Datascripts:        []string{"thisisaviref-ds2", "thisisaviref-ds1"},
		HttpPolicySets:     []string{"thisisaviref-httpps2", "thisisaviref-httpps1"},
		GslbFqdn:           "bar.com",
	}
	if len(gslbHost) > 0 {
		// It's assumed that the update case updates the gslb fqdn else bar.com is used.
		hostrule.GslbFqdn = gslbHost[0]
		hrUpdate := hostrule.HostRule()
		hrUpdate.ResourceVersion = "2"
		if _, err := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{}); err != nil {
			t.Fatalf("error in updating HostRule: %v", err)
		}
		return
	}
	if secure {
		hostrule.SslKeyCertificate = "thisisaviref-sslkey"
		hostrule.SslProfile = "thisisaviref-sslprof"
	}

	hrCreate := hostrule.HostRule()
	if _, err := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().HostRules("default").Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
}

func TeardownHostRule(t *testing.T, g *gomega.WithT, vskey cache.NamespaceName, hrname string) {
	DeleteObject(t, lib.HostRule, hrname, "default")
	VerifyMetadataHostRule(t, g, vskey, "default/"+hrname, false)
}

// HTTPRule
type FakeHTTPRule struct {
	Name           string
	Namespace      string
	Fqdn           string
	PathProperties []FakeHTTPRulePath
}

type FakeHTTPRulePath struct {
	Path           string
	SslProfile     string
	DestinationCA  string
	HealthMonitors []string
	LbAlgorithm    string
	Hash           string
}

func (rr FakeHTTPRule) HTTPRule() *akov1alpha1.HTTPRule {
	var rrPaths []akov1alpha1.HTTPRulePaths
	for _, p := range rr.PathProperties {
		rrPaths = append(rrPaths, akov1alpha1.HTTPRulePaths{
			Target:         p.Path,
			HealthMonitors: p.HealthMonitors,
			TLS: akov1alpha1.HTTPRuleTLS{
				Type:          "reencrypt",
				SSLProfile:    p.SslProfile,
				DestinationCA: p.DestinationCA,
			},
			LoadBalancerPolicy: akov1alpha1.HTTPRuleLBPolicy{
				Algorithm: p.LbAlgorithm,
				Hash:      p.Hash,
			},
		})
	}
	return &akov1alpha1.HTTPRule{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: rr.Namespace,
			Name:      rr.Name,
		},
		Spec: akov1alpha1.HTTPRuleSpec{
			Fqdn:  rr.Fqdn,
			Paths: rrPaths,
		},
	}
}

func SetupHTTPRule(t *testing.T, rrname, fqdn, path string) {
	httprule := FakeHTTPRule{
		Name:      rrname,
		Namespace: "default",
		Fqdn:      fqdn,
		PathProperties: []FakeHTTPRulePath{{
			Path:           path,
			SslProfile:     "thisisaviref-sslprofile",
			DestinationCA:  "httprule-destinationCA",
			LbAlgorithm:    "LB_ALGORITHM_CONSISTENT_HASH",
			Hash:           "LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS",
			HealthMonitors: []string{"thisisaviref-hm2", "thisisaviref-hm1"},
		}},
	}

	rrCreate := httprule.HTTPRule()
	if _, err := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().HTTPRules("default").Create(context.TODO(), rrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HTTPRule: %v", err)
	}
}

func VerifyMetadataHostRule(t *testing.T, g *gomega.WithT, vsKey cache.NamespaceName, hrnsname string, active bool) {
	mcache := cache.SharedAviObjCache()
	wait.Poll(2*time.Second, 50*time.Second, func() (bool, error) {
		sniCache, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		if active && !found {
			t.Logf("SNI Cache not found.")
			return false, nil
		}

		if !active && !found {
			return true, nil
		}

		sniCacheObj, ok := sniCache.(*cache.AviVsCache)
		if !ok {
			t.Logf("Unable to cast SNI Cache to AviVsCache.")
			return false, nil
		}

		if active {
			if sniCacheObj.ServiceMetadataObj.CRDStatus.Value != hrnsname {
				t.Logf("Expected CRD ServiceMetadata Value to be %s, found %s", hrnsname, sniCacheObj.ServiceMetadataObj.CRDStatus.Value)
				return false, nil
			}

			if sniCacheObj.ServiceMetadataObj.CRDStatus.Status != lib.CRDActive {
				t.Logf("Expected CRD ServiceMetadata Status to be %s, found %s", lib.CRDActive, sniCacheObj.ServiceMetadataObj.CRDStatus.Status)
				return false, nil
			}
		}

		if !active && (sniCacheObj.ServiceMetadataObj.CRDStatus.Status == lib.CRDActive) {
			t.Logf("Expected CRD ServiceMetadata Status to be empty/inactive, found %s", sniCacheObj.ServiceMetadataObj.CRDStatus.Status)
			return false, nil
		}

		return true, nil
	})
}

func VerifyMetadataHTTPRule(t *testing.T, g *gomega.WithT, poolKey cache.NamespaceName, httpruleNSNamePath string, active bool) {
	mcache := cache.SharedAviObjCache()
	wait.Poll(2*time.Second, 50*time.Second, func() (bool, error) {
		poolCache, found := mcache.PoolCache.AviCacheGet(poolKey)
		if !found {
			t.Logf("Pool Cache not found.")
			return false, nil
		}

		if !active && !found {
			return true, nil
		}

		poolCacheObj, ok := poolCache.(*cache.AviPoolCache)
		if !ok {
			t.Logf("Unable to cast Pool Cache to AviPoolCache.")
			return false, nil
		}

		if active {
			if poolCacheObj.ServiceMetadataObj.CRDStatus.Value != httpruleNSNamePath {
				t.Logf("Expected CRD ServiceMetadata Value to be %s, found %s", httpruleNSNamePath, poolCacheObj.ServiceMetadataObj.CRDStatus.Value)
				return false, nil
			}

			if poolCacheObj.ServiceMetadataObj.CRDStatus.Status != lib.CRDActive {
				t.Logf("Expected CRD ServiceMetadata Status to be %s, found %s", lib.CRDActive, poolCacheObj.ServiceMetadataObj.CRDStatus.Status)
				return false, nil
			}
		}

		if !active && (poolCacheObj.ServiceMetadataObj.CRDStatus.Status == lib.CRDActive) {
			t.Logf("Expected CRD ServiceMetadata Status to be empty/inactive, found %s", poolCacheObj.ServiceMetadataObj.CRDStatus.Status)
			return false, nil
		}

		return true, nil
	})
}

// Ingress Class
type FakeIngressClass struct {
	Name            string
	Controller      string
	AviInfraSetting string
	Default         bool
}

func (ingclass FakeIngressClass) IngressClass() *networking.IngressClass {
	ingressclass := &networking.IngressClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: ingclass.Name,
		},
		Spec: networking.IngressClassSpec{
			Controller: ingclass.Controller,
		},
	}

	if ingclass.Default {
		ingressclass.Annotations = map[string]string{lib.DefaultIngressClassAnnotation: "true"}
	} else {
		ingressclass.Annotations = map[string]string{lib.DefaultIngressClassAnnotation: "false"}
	}

	if ingclass.AviInfraSetting != "" {
		akoGroup := lib.AkoGroup
		ingressclass.Spec.Parameters = &networking.IngressClassParametersReference{
			APIGroup: &akoGroup,
			Kind:     lib.AviInfraSetting,
			Name:     ingclass.AviInfraSetting,
		}
	}

	return ingressclass
}

func SetupIngressClass(t *testing.T, ingclassName, controller, infraSetting string) {
	ingclass := FakeIngressClass{
		Name:            ingclassName,
		Controller:      controller,
		Default:         false,
		AviInfraSetting: infraSetting,
	}

	ingClassCreate := ingclass.IngressClass()
	if _, err := utils.GetInformers().ClientSet.NetworkingV1().IngressClasses().Create(context.TODO(), ingClassCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding IngressClass: %v", err)
	}
}

// Ingress Class
func AddDefaultIngressClass() {
	aviIngressClass := &networking.IngressClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: DefaultIngressClass,
			Annotations: map[string]string{
				lib.DefaultIngressClassAnnotation: "true",
			},
		},
		Spec: networking.IngressClassSpec{
			Controller: lib.AviIngressController,
		},
	}

	utils.GetInformers().ClientSet.NetworkingV1().IngressClasses().Create(context.TODO(), aviIngressClass, metav1.CreateOptions{})
}

func RemoveDefaultIngressClass() {
	utils.GetInformers().ClientSet.NetworkingV1().IngressClasses().Delete(context.TODO(), DefaultIngressClass, metav1.DeleteOptions{})
}

// AviInfraSetting
type FakeAviInfraSetting struct {
	Name           string
	SeGroupName    string
	Networks       []string
	EnableRhi      bool
	EnablePublicIP bool
	ShardSize      string
	BGPPeerLabels  []string
}

func (infraSetting FakeAviInfraSetting) AviInfraSetting() *akov1alpha1.AviInfraSetting {
	setting := &akov1alpha1.AviInfraSetting{
		ObjectMeta: metav1.ObjectMeta{
			Name: infraSetting.Name,
		},
		Spec: akov1alpha1.AviInfraSettingSpec{
			SeGroup: akov1alpha1.AviInfraSettingSeGroup{
				Name: infraSetting.SeGroupName,
			},
			Network: akov1alpha1.AviInfraSettingNetwork{
				EnableRhi:      &infraSetting.EnableRhi,
				BgpPeerLabels:  infraSetting.BGPPeerLabels,
				EnablePublicIP: &infraSetting.EnablePublicIP,
			},
		},
	}

	for _, networkName := range infraSetting.Networks {
		setting.Spec.Network.VipNetworks = append(setting.Spec.Network.VipNetworks, akov1alpha1.AviInfraSettingVipNetwork{
			NetworkName: networkName,
		})
	}

	if infraSetting.ShardSize != "" {
		setting.Spec.L7Settings.ShardSize = infraSetting.ShardSize
	}

	return setting
}

func SetupAviInfraSetting(t *testing.T, infraSettingName, shardSize string) {
	setting := FakeAviInfraSetting{
		Name:          infraSettingName,
		SeGroupName:   "thisisaviref-" + infraSettingName + "-seGroup",
		Networks:      []string{"thisisaviref-" + infraSettingName + "-networkName"},
		EnableRhi:     true,
		BGPPeerLabels: []string{"peer1", "peer2"},
		ShardSize:     shardSize,
	}
	settingCreate := setting.AviInfraSetting()
	if _, err := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().AviInfraSettings().Create(context.TODO(), settingCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding AviInfraSetting: %v", err)
	}
}

// AKO API Server
func InitializeFakeAKOAPIServer() *api.FakeApiServer {
	utils.AviLog.Infof("Initializing Fake AKO API server")
	akoApi := &api.FakeApiServer{
		Port: "54321",
	}

	akoApi.InitApi()
	lib.SetApiServerInstance(akoApi)
	return akoApi
}

func ClearAllCache(cacheObj *cache.AviObjCache) {
	var keys []cache.NamespaceName
	keys = cacheObj.PgCache.AviGetAllKeys()
	for _, k := range keys {
		cacheObj.PgCache.AviCacheDelete(k)
	}
	keys = cacheObj.PoolCache.AviGetAllKeys()
	for _, k := range keys {
		cacheObj.PoolCache.AviCacheDelete(k)
	}
	keys = cacheObj.VSVIPCache.AviGetAllKeys()
	for _, k := range keys {
		cacheObj.VSVIPCache.AviCacheDelete(k)
	}
	keys = cacheObj.VsCacheMeta.AviGetAllKeys()
	for _, k := range keys {
		cacheObj.VsCacheMeta.AviCacheDelete(k)
	}
	keys = cacheObj.VsCacheLocal.AviGetAllKeys()
	for _, k := range keys {
		cacheObj.VsCacheLocal.AviCacheDelete(k)
	}
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
		if !found || aviModel == nil {
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

func SetAkoTenant() {
	os.Setenv("TENANTS_PER_CLUSTER", "true")
	os.Setenv("TENANT_NAME", AKOTENANT)
}

func ResetAkoTenant() {
	os.Setenv("TENANTS_PER_CLUSTER", "false")
	os.Setenv("TENANT_NAME", "admin")
}

func SetNodePortMode() {
	os.Setenv("SERVICE_TYPE", "NodePort")
}

func SetClusterIPMode() {
	os.Setenv("SERVICE_TYPE", "ClusterIP")
}

func GetShardVSNumber(s string) string {
	var vsNum uint32
	shardSize := lib.GetshardSize()
	if shardSize != 0 {
		vsNum = utils.Bkt(s, shardSize)
	} else {
		return ""
	}
	vsNumber := fmt.Sprint(vsNum)
	return vsNumber
}

func UpdateObjectOrFail(t *testing.T, objType string, obj runtime.Object) {
	switch objType {
	case lib.AdvancedL4GatewayClass:
		if _, err := lib.AKOControlConfig().AdvL4Clientset().NetworkingV1alpha1pre1().GatewayClasses().Update(context.TODO(), obj.(*v1alpha1pre1.GatewayClass), metav1.UpdateOptions{}); err != nil {
			t.Fatalf("error in updating GatewayClass: %v", err)
		}
	case lib.AdvancedL4Gateway:
		gw := obj.(*v1alpha1pre1.Gateway)
		if _, err := lib.AKOControlConfig().AdvL4Clientset().NetworkingV1alpha1pre1().Gateways(gw.Namespace).Update(context.TODO(), gw, metav1.UpdateOptions{}); err != nil {
			t.Fatalf("error in updating Gateway: %v", err)
		}
	case lib.GatewayClass:
		if _, err := lib.AKOControlConfig().ServicesAPIClientset().NetworkingV1alpha1().GatewayClasses().Update(context.TODO(), obj.(*servicesapi.GatewayClass), metav1.UpdateOptions{}); err != nil {
			t.Fatalf("error in updating GatewayClass: %v", err)
		}
	case lib.Gateway:
		gw := obj.(*servicesapi.Gateway)
		if _, err := lib.AKOControlConfig().ServicesAPIClientset().NetworkingV1alpha1().Gateways(gw.Namespace).Update(context.TODO(), gw, metav1.UpdateOptions{}); err != nil {
			t.Fatalf("error in updating Gateway: %v", err)
		}
	case lib.Service:
		svc := obj.(*corev1.Service)
		if _, err := utils.GetInformers().ClientSet.CoreV1().Services(svc.Namespace).Update(context.TODO(), svc, metav1.UpdateOptions{}); err != nil {
			t.Fatalf("error in updating Service: %v", err)
		}
	case lib.Endpoint:
		ep := obj.(*corev1.Endpoints)
		if _, err := utils.GetInformers().ClientSet.CoreV1().Endpoints(ep.Namespace).Update(context.TODO(), ep, metav1.UpdateOptions{}); err != nil {
			t.Fatalf("error in updating Endpoint: %v", err)
		}
	case lib.Secret:
		secret := obj.(*corev1.Secret)
		if _, err := utils.GetInformers().ClientSet.CoreV1().Secrets(secret.Namespace).Update(context.TODO(), secret, metav1.UpdateOptions{}); err != nil {
			t.Fatalf("error in updating Endpoint: %v", err)
		}
	case lib.Node:
		if _, err := utils.GetInformers().ClientSet.CoreV1().Nodes().Update(context.TODO(), obj.(*corev1.Node), metav1.UpdateOptions{}); err != nil {
			t.Fatalf("error in updating Endpoint: %v", err)
		}
	case lib.Pod:
		pod := obj.(*corev1.Pod)
		if _, err := utils.GetInformers().ClientSet.CoreV1().Pods(pod.Namespace).Update(context.TODO(), pod, metav1.UpdateOptions{}); err != nil {
			t.Fatalf("error in updating Endpoint: %v", err)
		}
	case lib.Namespace:
		if _, err := utils.GetInformers().ClientSet.CoreV1().Namespaces().Update(context.TODO(), obj.(*corev1.Namespace), metav1.UpdateOptions{}); err != nil {
			t.Fatalf("error in updating Endpoint: %v", err)
		}
	case lib.IngressClass:
		if _, err := utils.GetInformers().ClientSet.NetworkingV1().IngressClasses().Update(context.TODO(), obj.(*networking.IngressClass), metav1.UpdateOptions{}); err != nil {
			t.Fatalf("error in updating Service: %v", err)
		}
	case lib.Ingress:
		ing := obj.(*networking.Ingress)
		if _, err := utils.GetInformers().ClientSet.NetworkingV1().Ingresses(ing.Namespace).Update(context.TODO(), ing, metav1.UpdateOptions{}); err != nil {
			t.Fatalf("error in updating Service: %v", err)
		}
	case lib.Route:
		route := obj.(*routev1.Route)
		if _, err := utils.GetInformers().OshiftClient.RouteV1().Routes(route.Namespace).Update(context.TODO(), route, metav1.UpdateOptions{}); err != nil {
			t.Fatalf("error in updating route: %v", err)
		}
	case lib.HostRule:
		hostrule := obj.(*akov1alpha1.HostRule)
		if _, err := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().HostRules(hostrule.Namespace).Update(context.TODO(), hostrule, metav1.UpdateOptions{}); err != nil {
			t.Fatalf("error in updating HostRule: %v", err)
		}
	case lib.HTTPRule:
		httprule := obj.(*akov1alpha1.HTTPRule)
		if _, err := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().HTTPRules(httprule.Namespace).Update(context.TODO(), httprule, metav1.UpdateOptions{}); err != nil {
			t.Fatalf("error in updating HostRule: %v", err)
		}
	case lib.AviInfraSetting:
		if _, err := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().AviInfraSettings().Update(context.TODO(), obj.(*akov1alpha1.AviInfraSetting), metav1.UpdateOptions{}); err != nil {
			t.Fatalf("error in updating HostRule: %v", err)
		}
	default:
		t.Fatalf("Unrecognized object Type: %s", objType)
	}
}

func DeleteObject(t *testing.T, objType string, objName string, objNamespace ...string) {
	switch objType {
	case lib.AdvancedL4GatewayClass:
		if err := lib.AKOControlConfig().AdvL4Clientset().NetworkingV1alpha1pre1().GatewayClasses().Delete(context.TODO(), objName, metav1.DeleteOptions{}); err != nil {
			t.Logf("error in deleting GatewayClass: %v", err)
		}
	case lib.AdvancedL4Gateway:
		if err := lib.AKOControlConfig().AdvL4Clientset().NetworkingV1alpha1pre1().Gateways(objNamespace[0]).Delete(context.TODO(), objName, metav1.DeleteOptions{}); err != nil {
			t.Logf("error in deleting Gateway: %v", err)
		}
	case lib.GatewayClass:
		if err := lib.AKOControlConfig().ServicesAPIClientset().NetworkingV1alpha1().GatewayClasses().Delete(context.TODO(), objName, metav1.DeleteOptions{}); err != nil {
			t.Logf("error in deleting GatewayClass: %v", err)
		}
	case lib.Gateway:
		if err := lib.AKOControlConfig().ServicesAPIClientset().NetworkingV1alpha1().Gateways(objNamespace[0]).Delete(context.TODO(), objName, metav1.DeleteOptions{}); err != nil {
			t.Logf("error in deleting Gateway: %v", err)
		}
	case lib.Service:
		if err := utils.GetInformers().ClientSet.CoreV1().Services(objNamespace[0]).Delete(context.TODO(), objName, metav1.DeleteOptions{}); err != nil {
			t.Logf("error in deleting Service: %v", err)
		}
	case lib.Endpoint:
		if err := utils.GetInformers().ClientSet.CoreV1().Endpoints(objNamespace[0]).Delete(context.TODO(), objName, metav1.DeleteOptions{}); err != nil {
			t.Logf("error in deleting Endpoint: %v", err)
		}
	case lib.Secret:
		if err := utils.GetInformers().ClientSet.CoreV1().Secrets(objNamespace[0]).Delete(context.TODO(), objName, metav1.DeleteOptions{}); err != nil {
			t.Logf("error in deleting Secret: %v", err)
		}
	case lib.Node:
		if err := utils.GetInformers().ClientSet.CoreV1().Nodes().Delete(context.TODO(), objName, metav1.DeleteOptions{}); err != nil {
			t.Logf("error in deleting Node: %v", err)
		}
	case lib.ConfigMap:
		if err := utils.GetInformers().ClientSet.CoreV1().ConfigMaps(objNamespace[0]).Delete(context.TODO(), objName, metav1.DeleteOptions{}); err != nil {
			t.Logf("error in deleting ConfigMap: %v", err)
		}
	case lib.Pod:
		if err := utils.GetInformers().ClientSet.CoreV1().Pods(objNamespace[0]).Delete(context.TODO(), objName, metav1.DeleteOptions{}); err != nil {
			t.Logf("error in deleting Pod: %v", err)
		}
	case lib.Namespace:
		if err := utils.GetInformers().ClientSet.CoreV1().Namespaces().Delete(context.TODO(), objName, metav1.DeleteOptions{}); err != nil {
			t.Logf("error in deleting Namespace: %v", err)
		}
	case lib.IngressClass:
		if err := utils.GetInformers().ClientSet.NetworkingV1().IngressClasses().Delete(context.TODO(), objName, metav1.DeleteOptions{}); err != nil {
			t.Logf("error in deleting IngressClass: %v", err)
		}
	case lib.Ingress:
		if err := utils.GetInformers().ClientSet.NetworkingV1().Ingresses(objNamespace[0]).Delete(context.TODO(), objName, metav1.DeleteOptions{}); err != nil {
			t.Logf("error in deleting Ingress: %v", err)
		}
	case lib.Route:
		if err := utils.GetInformers().OshiftClient.RouteV1().Routes(objNamespace[0]).Delete(context.TODO(), objName, metav1.DeleteOptions{}); err != nil {
			t.Logf("error in deleting Route: %v", err)
		}
	case lib.HostRule:
		if err := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().HostRules(objNamespace[0]).Delete(context.TODO(), objName, metav1.DeleteOptions{}); err != nil {
			t.Logf("error in deleting HostRule: %v", err)
		}
	case lib.HTTPRule:
		if err := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().HTTPRules(objNamespace[0]).Delete(context.TODO(), objName, metav1.DeleteOptions{}); err != nil {
			t.Logf("error in deleting HttpRule: %v", err)
		}
	case lib.AviInfraSetting:
		if err := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().AviInfraSettings().Delete(context.TODO(), objName, metav1.DeleteOptions{}); err != nil {
			t.Logf("error in deleting AviInfraSetting: %v", err)
		}
	default:
		t.Fatalf("Unrecognized object Type: %s", objType)
	}
}
