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

	v1 "k8s.io/api/core/v1"
	extensionv1beta1 "k8s.io/api/extensions/v1beta1"

	avinodes "gitlab.eng.vmware.com/orion/akc/pkg/nodes"
	"gitlab.eng.vmware.com/orion/akc/pkg/objects"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Fake ingress
type fakeIngress struct {
	dnsnames    []string
	tlsdnsnames [][]string
	ips         []string
	hostnames   []string
	namespace   string
	name        string
	annotations map[string]string
	serviceName string
}

func (ing fakeIngress) Ingress() *extensionv1beta1.Ingress {
	ingress := &extensionv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   ing.namespace,
			Name:        ing.name,
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
	for _, dnsname := range ing.dnsnames {
		ingress.Spec.Rules = append(ingress.Spec.Rules, extensionv1beta1.IngressRule{
			Host: dnsname,
			IngressRuleValue: extensionv1beta1.IngressRuleValue{
				HTTP: &extensionv1beta1.HTTPIngressRuleValue{
					Paths: []extensionv1beta1.HTTPIngressPath{extensionv1beta1.HTTPIngressPath{
						Path: "/foo",
						Backend: extensionv1beta1.IngressBackend{ServiceName: ing.serviceName, ServicePort: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: 8080,
						}},
					},
					},
				},
			},
		})
	}
	for _, hosts := range ing.tlsdnsnames {
		ingress.Spec.TLS = append(ingress.Spec.TLS, extensionv1beta1.IngressTLS{
			Hosts: hosts,
		})
	}
	for _, ip := range ing.ips {
		ingress.Status.LoadBalancer.Ingress = append(ingress.Status.LoadBalancer.Ingress, v1.LoadBalancerIngress{
			IP: ip,
		})
	}
	for _, hostname := range ing.hostnames {
		ingress.Status.LoadBalancer.Ingress = append(ingress.Status.LoadBalancer.Ingress, v1.LoadBalancerIngress{
			Hostname: hostname,
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

func pollForCompletion(t *testing.T, key string, counter int) interface{} {
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

type fakeService struct {
	namespace    string
	name         string
	annotations  map[string]string
	servicePorts []serviceport
}

type serviceport struct {
	portName   string
	portNumber int32
	protocol   v1.Protocol
	targetPort int
}

func (svc fakeService) Service() *corev1.Service {
	var ports []corev1.ServicePort
	for _, svcport := range svc.servicePorts {
		ports = append(ports, corev1.ServicePort{Name: svcport.portName, Port: svcport.portNumber, Protocol: svcport.protocol, TargetPort: intstr.FromInt(svcport.targetPort)})
	}
	svcExample := &corev1.Service{
		Spec: corev1.ServiceSpec{
			Type:  corev1.ServiceTypeClusterIP,
			Ports: ports,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: svc.namespace,
			Name:      svc.name,
		},
	}
	return svcExample
}

type fakeNode struct {
	name    string
	podCIDR string
	nodeIP  string
	version string
}

func (node fakeNode) Node() *corev1.Node {
	nodeExample := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:            node.name,
			ResourceVersion: node.version,
		},
		Spec: corev1.NodeSpec{
			PodCIDR: node.podCIDR,
		},
		Status: corev1.NodeStatus{
			Addresses: []corev1.NodeAddress{
				{
					Type:    "Internal",
					Address: node.nodeIP,
				},
			},
		},
	}
	return nodeExample
}
