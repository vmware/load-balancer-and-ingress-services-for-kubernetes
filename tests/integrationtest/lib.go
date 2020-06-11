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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"os"

	v1 "k8s.io/api/core/v1"
	extensionv1beta1 "k8s.io/api/extensions/v1beta1"

	"ako/pkg/cache"
	"ako/pkg/k8s"
	avinodes "ako/pkg/nodes"
	"ako/pkg/objects"

	"github.com/avinetworks/container-lib/utils"
	corev1 "k8s.io/api/core/v1"

	"github.com/avinetworks/sdk/go/models"
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
	SINGLEPORTMODEL = "admin/cluster--red-ns-testsvc"      // single port model name
	MULTIPORTMODEL  = "admin/cluster--red-ns-testsvcmulti" // multi port model name
	RANDOMUUID      = "random-uuid"                        // random avi object uuid
)

var KubeClient *k8sfake.Clientset
var ctrl *k8s.AviController

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
	KubeClient.CoreV1().Secrets(namespace).Create(fakeSecret)
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
	for secret, hosts := range ing.TlsSecretDNS {
		ingress.Spec.TLS = append(ingress.Spec.TLS, extensionv1beta1.IngressTLS{
			Hosts:      hosts,
			SecretName: secret,
		})
	}
	for i, _ := range ing.Ips {
		hostname := ""
		if len(ing.HostNames) >= i+1 {
			hostname = ing.HostNames[i]
		}
		ingress.Status.LoadBalancer.Ingress = append(ingress.Status.LoadBalancer.Ingress, v1.LoadBalancerIngress{
			IP:       ing.Ips[i],
			Hostname: hostname,
		})
	}
	return ingress
}

func (ing FakeIngress) SecureIngress() *extensionv1beta1.Ingress {
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

func (ing FakeIngress) IngressNoHost() *extensionv1beta1.Ingress {
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
	for _, path := range ing.Paths {
		ingress.Spec.Rules = append(ingress.Spec.Rules, extensionv1beta1.IngressRule{
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

	for secret, hosts := range ing.TlsSecretDNS {
		ingress.Spec.TLS = append(ingress.Spec.TLS, extensionv1beta1.IngressTLS{
			Hosts:      hosts,
			SecretName: secret,
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
	Type         corev1.ServiceType
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
		ports = append(ports, corev1.ServicePort{
			Name:       svcport.PortName,
			Port:       svcport.PortNumber,
			Protocol:   svcport.Protocol,
			TargetPort: intstr.FromInt(svcport.TargetPort),
		})
	}
	svcExample := &corev1.Service{
		Spec: corev1.ServiceSpec{
			Type:  svc.Type,
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
		servicePorts = append(servicePorts, Serviceport{
			PortName:   fmt.Sprintf("foo%d", i),
			PortNumber: int32(mPort),
			Protocol:   "TCP",
			TargetPort: mPort,
		})
	}

	svcExample := (FakeService{Name: Name, Namespace: ns, Type: Type, ServicePorts: servicePorts}).Service()
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
	_, err := KubeClient.CoreV1().Endpoints(ns).Create(epExample)
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
	_, err := KubeClient.CoreV1().Endpoints(ns).Update(epExample)
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

var AviFakeClientInstance *httptest.Server
var FakeServerMiddleware InjectFault
var FakeAviObjects = []string{
	"cloud",
	"ipamdnsproviderprofile",
	"network",
	"pool",
	"poolgroup",
	"virtualservice",
	"vrfcontext",
	"vsdatascriptset",
}

type InjectFault func(w http.ResponseWriter, r *http.Request)

func AddMiddleware(exec InjectFault) {
	FakeServerMiddleware = exec
}

func ResetMiddleware() {
	FakeServerMiddleware = nil
}

func NewAviFakeClientInstance() {
	if AviFakeClientInstance == nil {
		AviFakeClientInstance = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			url := r.URL.EscapedPath()
			utils.AviLog.Infof("[fakeAPI]: %s %s\n", r.Method, url)

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
		k8s.PopulateCache()
	}
}

func NormalControllerServer(w http.ResponseWriter, r *http.Request) {
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
				shardVSNum = strings.Split(parentVSName, "cluster-")[1]

				rData["vh_parent_vs_ref"] = fmt.Sprintf("https://localhost/api/virtualservice/virtualservice-%s-%s#%s", parentVSName, RANDOMUUID, parentVSName)
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
		}

		finalResponse, _ = json.Marshal([]interface{}{resp["data"]})
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, string(finalResponse))

	} else if r.Method == "PUT" {
		data, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(data, &resp)
		resp["uuid"] = strings.Split(strings.Trim(url, "/"), "/")[2]
		finalResponse, _ = json.Marshal(resp)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, string(finalResponse))

	} else if r.Method == "DELETE" {
		w.WriteHeader(http.StatusNoContent)
		fmt.Fprintln(w, string(finalResponse))

	} else if r.Method == "GET" && inArray(FakeAviObjects, object[1]) {
		FeedMockCollectionData(w, r)

	} else if strings.Contains(url, "login") {
		// This is used for /login --> first request to controller
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"success": "true"}`)
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
func FeedMockCollectionData(w http.ResponseWriter, r *http.Request) {
	mockFilePath := "../avimockobjects"
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
		fmt.Fprintln(w, string(data))
	} else if strings.Contains(url, "login") {
		// This is used for /login --> first request to controller
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"success": "true"}`)
	}
}

//UpdateIngress wrapper over ingress update call.
//internally calls Ingress() for fakeIngress object
//performs a get for ingress object so it will update only if ingress exists
func (ing FakeIngress) UpdateIngress() (*extensionv1beta1.Ingress, error) {

	//check if resource already exists
	ingress, err := KubeClient.ExtensionsV1beta1().Ingresses(ing.Namespace).Get(ing.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	//increment resource version
	newIngress := ing.IngressMultiPath() //Maybe we should replace Ingress() with IngressMultiPath() completely
	rv, _ := strconv.Atoi(ingress.ResourceVersion)
	newIngress.ResourceVersion = strconv.Itoa(rv + 1)

	//update ingress resource
	updatedIngress, err := KubeClient.ExtensionsV1beta1().Ingresses(newIngress.Namespace).Update(newIngress)
	return updatedIngress, err
}
