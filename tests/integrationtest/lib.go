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

package integrationtest

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/apis/ako/v1alpha1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/client/v1alpha1/clientset/versioned/fake"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/api"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/avinetworks/sdk/go/models"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

// constants to be used for creating K8s objs and verifying Avi objs
const (
	SINGLEPORTSVC   = "testsvc"                            // single port service name
	MULTIPORTSVC    = "testsvcmulti"                       // multi port service name
	NAMESPACE       = "red-ns"                             // namespace
	AVINAMESPACE    = "admin"                              // avi namespace
	AKOTENANT       = "akotenant"                          // ako tenant where TENANTS_PER_CLUSTER is enabled
	SINGLEPORTMODEL = "admin/cluster--red-ns-testsvc"      // single port model name
	MULTIPORTMODEL  = "admin/cluster--red-ns-testsvcmulti" // multi port model name
	RANDOMUUID      = "random-uuid"                        // random avi object uuid
)

var KubeClient *k8sfake.Clientset
var CRDClient *crdfake.Clientset
var ctrl *k8s.AviController

func AddConfigMap() {
	aviCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "avi-system",
			Name:      "avi-k8s-config",
		},
	}
	KubeClient.CoreV1().ConfigMaps("avi-system").Create(context.TODO(), aviCM, metav1.CreateOptions{})

	PollForSyncStart(ctrl, 10)
}

func AddDefaultIngressClass() {
	aviIngressClass := &networking.IngressClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "avi-lb",
			Annotations: map[string]string{
				lib.DefaultIngressClassAnnotation: "true",
			},
		},
		Spec: networking.IngressClassSpec{
			Controller: "ako.vmware.com/avi-lb",
		},
	}

	KubeClient.NetworkingV1beta1().IngressClasses().Create(context.TODO(), aviIngressClass, metav1.CreateOptions{})
}

//Fake Namespace
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
func AddNamespace(nsName string, labels map[string]string) error {
	nsMetaOptions := (FakeNamespace{
		Name:   nsName,
		Labels: labels,
	}).Namespace()
	nsMetaOptions.ResourceVersion = "1"
	_, err := KubeClient.CoreV1().Namespaces().Get(context.TODO(), nsName, metav1.GetOptions{})
	if err != nil {
		_, err = KubeClient.CoreV1().Namespaces().Create(context.TODO(), nsMetaOptions, metav1.CreateOptions{})
		if err != nil {
			utils.AviLog.Errorf("Error occured while Adding namespace : %v", err)
		}
	}
	return err
}

func UpdateNamespace(nsName string, labels map[string]string) error {
	nsMetaOptions := (FakeNamespace{
		Name:   nsName,
		Labels: labels,
	}).Namespace()
	nsMetaOptions.ResourceVersion = "2"
	_, err := KubeClient.CoreV1().Namespaces().Update(context.TODO(), nsMetaOptions, metav1.UpdateOptions{})
	return err
}

func DeleteNamespace(nsName string) {
	KubeClient.CoreV1().Namespaces().Delete(context.TODO(), nsName, metav1.DeleteOptions{})
}

// Fake Secret
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
	KubeClient.CoreV1().Secrets(namespace).Create(context.TODO(), fakeSecret, metav1.CreateOptions{})
}

// Fake ingress
type FakeIngress struct {
	DnsNames     []string
	Paths        []string
	Ips          []string
	HostNames    []string
	Namespace    string
	Name         string
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
						Paths: []networking.HTTPIngressPath{networking.HTTPIngressPath{
							Path: "/foo",
							Backend: networking.IngressBackend{ServiceName: ing.ServiceName, ServicePort: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: "foo0",
							}},
						},
						}},
				},
			})
			ingress.Spec.Rules = append(ingress.Spec.Rules, networking.IngressRule{
				Host: dnsName,
				IngressRuleValue: networking.IngressRuleValue{
					HTTP: &networking.HTTPIngressRuleValue{
						Paths: []networking.HTTPIngressPath{networking.HTTPIngressPath{
							Path: "/bar",
							Backend: networking.IngressBackend{ServiceName: ing.ServiceName, ServicePort: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: "foo1",
							}},
						},
						}},
				},
			})
		} else {
			ingress.Spec.Rules = append(ingress.Spec.Rules, networking.IngressRule{
				Host: dnsName,
				IngressRuleValue: networking.IngressRuleValue{
					HTTP: &networking.HTTPIngressRuleValue{
						Paths: []networking.HTTPIngressPath{networking.HTTPIngressPath{
							Path: path,
							Backend: networking.IngressBackend{ServiceName: ing.ServiceName, ServicePort: intstr.IntOrString{
								Type:   intstr.Int,
								IntVal: 8080,
							}},
						},
						}},
				},
			})
		}
	}
	for secret, hosts := range ing.TlsSecretDNS {
		ingress.Spec.TLS = append(ingress.Spec.TLS, networking.IngressTLS{
			Hosts:      hosts,
			SecretName: secret,
		})
	}
	for i, _ := range ing.Ips {
		hostname := ""
		if len(ing.HostNames) >= i+1 {
			hostname = ing.HostNames[i]
		}
		ingress.Status.LoadBalancer.Ingress = append(ingress.Status.LoadBalancer.Ingress, corev1.LoadBalancerIngress{
			IP:       ing.Ips[i],
			Hostname: hostname,
		})
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
					Paths: []networking.HTTPIngressPath{networking.HTTPIngressPath{
						Path: path,
						Backend: networking.IngressBackend{ServiceName: ing.ServiceName, ServicePort: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: 8080,
						}},
					},
					}},
			},
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
					Paths: []networking.HTTPIngressPath{networking.HTTPIngressPath{
						Path: path,
						Backend: networking.IngressBackend{ServiceName: ing.ServiceName, ServicePort: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: 8080,
						}},
					},
					}},
			},
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
	for _, dnsName := range ing.DnsNames {
		var ingrPaths []networking.HTTPIngressPath
		for _, path := range ing.Paths {
			ingrPath := networking.HTTPIngressPath{
				Path: path,
				Backend: networking.IngressBackend{ServiceName: ing.ServiceName, ServicePort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 8080,
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
	Namespace      string
	Name           string
	Labels         map[string]string
	Type           corev1.ServiceType
	LoadBalancerIP string
	annotations    map[string]string
	ServicePorts   []Serviceport
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
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: svc.Namespace,
			Name:      svc.Name,
			Labels:    svc.Labels,
		},
	}
	return svcExample
}

type FakeNode struct {
	Name    string
	PodCIDR string
	NodeIP  string
	Version string
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

func CreateNode(t *testing.T, nodeName string, nodeIP string) {
	modelName := "admin/global"
	objects.SharedAviGraphLister().Delete(modelName)
	nodeExample := (FakeNode{
		Name:    nodeName,
		PodCIDR: "10.244.0.0/24",
		Version: "1",
		NodeIP:  nodeIP,
	}).Node()

	_, err := KubeClient.CoreV1().Nodes().Create(context.TODO(), nodeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}

	PollForCompletion(t, modelName, 5)
}

func DeleteNode(t *testing.T, nodeName string) {
	modelName := "admin/global"
	objects.SharedAviGraphLister().Delete(modelName)
	err := KubeClient.CoreV1().Nodes().Delete(context.TODO(), nodeName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error in deleting Node: %v", err)
	}
	PollForCompletion(t, modelName, 5)
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

	svcExample := (FakeService{Name: Name, Namespace: ns, Type: Type, ServicePorts: servicePorts}).Service()
	_, err := KubeClient.CoreV1().Services(ns).Create(context.TODO(), svcExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
}

func DelSVC(t *testing.T, ns string, Name string) {
	err := KubeClient.CoreV1().Services(ns).Delete(context.TODO(), Name, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error in deleting Service: %v", err)
	}
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
	_, err := KubeClient.CoreV1().Endpoints(ns).Create(context.TODO(), epExample, metav1.CreateOptions{})
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
	_, err := KubeClient.CoreV1().Endpoints(ns).Update(context.TODO(), epExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in creating Endpoint: %v", err)
	}
}

func DelEP(t *testing.T, ns string, Name string) {
	err := KubeClient.CoreV1().Endpoints(ns).Delete(context.TODO(), Name, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error in deleting Endpoint: %v", err)
	}
}

func InitializeFakeAKOAPIServer() *api.FakeApiServer {
	utils.AviLog.Infof("Initializing Fake AKO API server")
	akoApi := &api.FakeApiServer{
		Port: "54321",
	}

	akoApi.InitApi()
	lib.SetApiServerInstance(akoApi)
	return akoApi
}

//s: namespace or hostname
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

const defaultMockFilePath = "../avimockobjects"

var AviFakeClientInstance *httptest.Server
var FakeServerMiddleware InjectFault
var FakeAviObjects = []string{
	"cloud",
	"ipamdnsproviderprofile",
	"ipamdnsproviderprofiledomainlist",
	"network",
	"pool",
	"poolgroup",
	"virtualservice",
	"vrfcontext",
	"vsdatascriptset",
	"serviceenginegroup",
	"tenant",
}

type InjectFault func(w http.ResponseWriter, r *http.Request)

func AddMiddleware(exec InjectFault) {
	FakeServerMiddleware = exec
}

func ResetMiddleware() {
	FakeServerMiddleware = nil
}

func NewAviFakeClientInstance(skipCachePopulation ...bool) {
	if AviFakeClientInstance == nil {
		AviFakeClientInstance = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			utils.AviLog.Infof("[fakeAPI]: %s %s\n", r.Method, r.URL)

			if FakeServerMiddleware != nil {
				FakeServerMiddleware(w, r)
				return
			}

			NormalControllerServer(w, r)
		}))

		url := strings.Split(AviFakeClientInstance.URL, "https://")[1]
		os.Setenv("CTRL_USERNAME", "admin")
		os.Setenv("CTRL_PASSWORD", "admin")
		os.Setenv("CTRL_IPADDRESS", url)
		os.Setenv("SHARD_VS_SIZE", "LARGE")
		os.Setenv("FULL_SYNC_INTERVAL", "600")
		os.Setenv("SHARD_VS_SIZE", "LARGE")

		// resets avi client pool instance, allows to connect with the new `ts` server
		cache.AviClientInstance = nil
		if len(skipCachePopulation) == 0 || skipCachePopulation[0] == false {
			k8s.PopulateCache()
		}
	}
}

func NormalControllerServer(w http.ResponseWriter, r *http.Request, args ...string) {
	mockFilePath := defaultMockFilePath
	if len(args) > 0 {
		mockFilePath = args[0]
	}
	url := r.URL.EscapedPath()
	var resp map[string]interface{}
	var finalResponse []byte
	var vipAddress, shardVSNum string
	addrPrefix := "10.250.250"
	object := strings.Split(strings.Trim(url, "/"), "/")

	if strings.Contains(url, "macro") && r.Method == "POST" {
		data, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(data, &resp)
		rData, rModelName := resp["data"].(map[string]interface{}), strings.ToLower(resp["model_name"].(string))
		rName := rData["name"].(string)
		objURL := fmt.Sprintf("https://localhost/api/%s/%s-%s-%s#%s", rModelName, rModelName, rName, RANDOMUUID, rName)

		// adding additional 'uuid' and 'url' (read-only) fields in the response
		rData["url"] = objURL
		rData["uuid"] = fmt.Sprintf("%s-%s-%s", rModelName, rName, RANDOMUUID)

		if rModelName == "virtualservice" {
			// handle sni child, fill in vs parent ref
			if vsType := rData["type"]; vsType == "VS_TYPE_VH_CHILD" {
				parentVSName := strings.Split(rData["vh_parent_vs_uuid"].(string), "name=")[1]
				shardVSNum = strings.Split(parentVSName, "cluster--Shared-L7-")[1]

				rData["vh_parent_vs_ref"] = fmt.Sprintf("https://localhost/api/virtualservice/virtualservice-%s-%s#%s", parentVSName, RANDOMUUID, parentVSName)
				//rData["vsvip_ref"] = fmt.Sprintf("https://localhost/api/vsvip/vsvip-%s-%s#%s", parentVSName, RANDOMUUID, parentVSName)
				vipAddress = fmt.Sprintf("%s.1%s", addrPrefix, shardVSNum)

			} else if strings.Contains(rName, "Shared-L7") {
				shardVSNum = strings.Split(rName, "Shared-L7-")[1]
				vipAddress = fmt.Sprintf("%s.1%s", addrPrefix, shardVSNum)
			} else {
				vipAddress = "10.250.250.250"
			}

			// add vip for status update checks
			// use vh_parent_vs_uuid for sniVS, and name for normal VSes

			rData["vip"] = []interface{}{map[string]interface{}{"ip_address": map[string]string{"addr": vipAddress, "type": "V4"}}}
			rData["vsvip_ref"] = fmt.Sprintf("https://localhost/api/vsvip/vsvip-%s-%s#%s", rName, RANDOMUUID, rName)
		} else if rModelName == "vsvip" {
			if vsType := rData["type"]; vsType == "VS_TYPE_VH_CHILD" {
				parentVSName := strings.Split(rData["vh_parent_vs_uuid"].(string), "name=")[1]
				shardVSNum = strings.Split(parentVSName, "cluster--Shared-L7-")[1]
				vipAddress = fmt.Sprintf("%s.1%s", addrPrefix, shardVSNum)
			} else if strings.Contains(rName, "Shared-L7") {
				shardVSNum = strings.Split(rName, "Shared-L7-")[1]
				vipAddress = fmt.Sprintf("%s.1%s", addrPrefix, shardVSNum)
			} else {
				vipAddress = "10.250.250.250"
			}
			rData["vip"] = []interface{}{map[string]interface{}{"ip_address": map[string]string{"addr": vipAddress, "type": "V4"}}}
		}

		finalResponse, _ = json.Marshal([]interface{}{resp["data"]})
		w.WriteHeader(http.StatusOK)
		w.Write(finalResponse)

	} else if r.Method == "PUT" {
		data, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(data, &resp)
		resp["uuid"] = strings.Split(strings.Trim(url, "/"), "/")[2]
		finalResponse, _ = json.Marshal(resp)
		w.WriteHeader(http.StatusOK)
		w.Write(finalResponse)

	} else if r.Method == "DELETE" {
		w.WriteHeader(http.StatusNoContent)
		w.Write(finalResponse)

	} else if r.Method == "GET" &&
		(strings.Contains(r.URL.RawQuery, "thisisahostruleref") || strings.Contains(r.URL.RawQuery, "thisisahttpruleref")) {
		w.WriteHeader(http.StatusOK)
		data, _ := ioutil.ReadFile(fmt.Sprintf("%s/crd_mock.json", mockFilePath))
		w.Write(data)

	} else if r.Method == "GET" &&
		(strings.Contains(url, "/api/cloud/")) {
		var data []byte
		if strings.HasSuffix(r.URL.RawQuery, "CLOUD_NONE") {
			data, _ = ioutil.ReadFile(fmt.Sprintf("%s/%s_mock.json", mockFilePath, "CLOUD_NONE"))
		} else if strings.HasSuffix(r.URL.RawQuery, "CLOUD_AZURE") {
			data, _ = ioutil.ReadFile(fmt.Sprintf("%s/%s_mock.json", mockFilePath, "CLOUD_AZURE"))
		} else if strings.HasSuffix(r.URL.RawQuery, "CLOUD_AWS") {
			data, _ = ioutil.ReadFile(fmt.Sprintf("%s/%s_mock.json", mockFilePath, "CLOUD_AWS"))
		} else {
			data, _ = ioutil.ReadFile(fmt.Sprintf("%s/%s_mock.json", mockFilePath, "CLOUD_VCENTER"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write(data)

	} else if r.Method == "GET" && inArray(FakeAviObjects, object[1]) {
		FeedMockCollectionData(w, r, mockFilePath)

	} else if strings.Contains(url, "login") {
		// This is used for /login --> first request to controller
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": "true"}`))
	} else if strings.Contains(url, "initial-data") {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"version": {"Version": "20.1.2"}}`))
	} else if strings.Contains(url, "/api/cluster/runtime") {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"node_states": [{"name": "10.79.169.60","role": "CLUSTER_LEADER","up_since": "2020-10-28 04:58:48"}]}`))
	}
}

func inArray(a []string, b string) bool {
	for _, k := range a {
		if k == b {
			return true
		}
	}
	return false
}

// FeedMockCollectionData reads data from avimockobjects/*.json files and returns mock data
// for GET objects list API. GET /api/virtualservice returns from virtualservice_mock.json and so on
func FeedMockCollectionData(w http.ResponseWriter, r *http.Request, mockFilePath string) {
	url := r.URL.EscapedPath() // url = //api/<object>/:objectId
	splitURL := strings.Split(strings.Trim(url, "/"), "/")

	if r.Method == "GET" {
		var data []byte
		if len(splitURL) == 2 {
			data, _ = ioutil.ReadFile(fmt.Sprintf("%s/%s_mock.json", mockFilePath, splitURL[1]))
		} else if len(splitURL) == 3 {
			// with uuid
			data, _ = ioutil.ReadFile(fmt.Sprintf("%s/%s_uuid_mock.json", mockFilePath, splitURL[1]))
		}
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	} else if strings.Contains(url, "login") {
		// This is used for /login --> first request to controller
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": "true"}`))
	}
}

//UpdateIngress wrapper over ingress update call.
//internally calls Ingress() for fakeIngress object
//performs a get for ingress object so it will update only if ingress exists
func (ing FakeIngress) UpdateIngress() (*networking.Ingress, error) {

	//check if resource already exists
	ingress, err := KubeClient.NetworkingV1beta1().Ingresses(ing.Namespace).Get(context.TODO(), ing.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	//increment resource version
	newIngress := ing.IngressMultiPath() //Maybe we should replace Ingress() with IngressMultiPath() completely
	rv, _ := strconv.Atoi(ingress.ResourceVersion)
	newIngress.ResourceVersion = strconv.Itoa(rv + 1)

	//update ingress resource
	updatedIngress, err := KubeClient.NetworkingV1beta1().Ingresses(newIngress.Namespace).Update(context.TODO(), newIngress, metav1.UpdateOptions{})
	return updatedIngress, err
}

// HostRule/HTTPRule lib functions
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
			},
		},
	}

	return hostrule
}

func SetupHostRule(t *testing.T, hrname, fqdn string, secure bool) {
	hostrule := FakeHostRule{
		Name:               hrname,
		Namespace:          "default",
		Fqdn:               fqdn,
		WafPolicy:          "thisisahostruleref-waf",
		ApplicationProfile: "thisisahostruleref-appprof",
		AnalyticsProfile:   "thisisahostruleref-analyticsprof",
		ErrorPageProfile:   "thisisahostruleref-errorprof",
		Datascripts:        []string{"thisisahostruleref-ds2", "thisisahostruleref-ds1"},
		HttpPolicySets:     []string{"thisisahostruleref-httpps2", "thisisahostruleref-httpps1"},
	}
	if secure {
		hostrule.SslKeyCertificate = "thisisahostruleref-sslkey"
		hostrule.SslProfile = "thisisahostruleref-sslprof"
	}

	hrCreate := hostrule.HostRule()
	if _, err := lib.GetCRDClientset().AkoV1alpha1().HostRules("default").Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
}

func TeardownHostRule(t *testing.T, g *gomega.WithT, vskey cache.NamespaceName, hrname string) {
	if err := lib.GetCRDClientset().AkoV1alpha1().HostRules("default").Delete(context.TODO(), hrname, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("error in deleting HostRule: %v", err)
	}
	VerifyMetadataHostRule(g, vskey, "default/"+hrname, false)
}

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
			SslProfile:     "thisisahttpruleref-sslprofile",
			DestinationCA:  "httprule-destinationCA",
			LbAlgorithm:    "LB_ALGORITHM_CONSISTENT_HASH",
			Hash:           "LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS",
			HealthMonitors: []string{"thisisahttpruleref-hm2", "thisisahttpruleref-hm1"},
		}},
	}

	rrCreate := httprule.HTTPRule()
	if _, err := lib.GetCRDClientset().AkoV1alpha1().HTTPRules("default").Create(context.TODO(), rrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HTTPRule: %v", err)
	}
}

func TeardownHTTPRule(t *testing.T, rrname string) {
	if err := lib.GetCRDClientset().AkoV1alpha1().HTTPRules("default").Delete(context.TODO(), rrname, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("error in deleting HTTPRule: %v", err)
	}
}

func VerifyMetadataHostRule(g *gomega.WithT, vsKey cache.NamespaceName, hrnsname string, active bool) {
	mcache := cache.SharedAviObjCache()
	status := "INACTIVE"
	if active {
		status = "ACTIVE"
	}
	g.Eventually(func() bool {
		sniCache, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		sniCacheObj, ok := sniCache.(*cache.AviVsCache)
		if (ok && found && sniCacheObj.ServiceMetadataObj.CRDStatus.Value == hrnsname && sniCacheObj.ServiceMetadataObj.CRDStatus.Status == status) ||
			(ok && found && !active && sniCacheObj.ServiceMetadataObj.CRDStatus.Status == "") ||
			(!active && !found) {
			return true
		}
		return false
	}, 50*time.Second).Should(gomega.Equal(true))
}

func VerifyMetadataHTTPRule(g *gomega.WithT, poolKey cache.NamespaceName, rrnsname string, active bool) {
	mcache := cache.SharedAviObjCache()
	status := "INACTIVE"
	if active {
		status = "ACTIVE"
	}

	g.Eventually(func() bool {
		poolCache, found := mcache.PoolCache.AviCacheGet(poolKey)
		poolCacheObj, ok := poolCache.(*cache.AviPoolCache)
		if (ok && found && poolCacheObj.ServiceMetadataObj.CRDStatus.Value == rrnsname && poolCacheObj.ServiceMetadataObj.CRDStatus.Status == status) ||
			(ok && found && !active && poolCacheObj.ServiceMetadataObj.CRDStatus.Status == "") ||
			(!active && !found) {
			return true
		}
		return false
	}, 50*time.Second).Should(gomega.Equal(true))
}
