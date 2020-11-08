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

package utils

import (
	"encoding/json"
	"hash/fnv"
	"math/rand"
	"net"
	"net/url"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	oshiftclientset "github.com/openshift/client-go/route/clientset/versioned"
	oshiftinformers "github.com/openshift/client-go/route/informers/externalversions"
	corev1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	networking "k8s.io/api/networking/v1beta1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

var CtrlVersion string
var runtimeScheme = k8sruntime.NewScheme()

func init() {
	//Setting the package-wide version
	CtrlVersion = os.Getenv("CTRL_VERSION")
	extensions.AddToScheme(runtimeScheme)
	networking.AddToScheme(runtimeScheme)
}

func IsV4(addr string) bool {
	ip := net.ParseIP(addr)
	v4 := ip.To4()
	if v4 == nil {
		return false
	} else {
		return true
	}
}

/*
 * Port name is either "http" or "http-suffix"
 * Following Istio named port convention
 * https://istio.io/docs/setup/kubernetes/spec-requirements/
 * TODO: Define matching ports in configmap and make it configurable
 */

func IsSvcHttp(svc_name string, port int32) bool {
	if svc_name == "http" {
		return true
	} else if strings.HasPrefix(svc_name, "http-") {
		return true
	} else if (port == 80) || (port == 443) || (port == 8080) || (port == 8443) {
		return true
	} else {
		return false
	}
}

func AviUrlToObjType(aviurl string) (string, error) {
	url, err := url.Parse(aviurl)
	if err != nil {
		AviLog.Warnf("aviurl %v parse error", aviurl)
		return "", err
	}

	path := url.EscapedPath()

	elems := strings.Split(path, "/")
	return elems[2], nil
}

/*
 * Hash key to pick workqueue & GoRoutine. Hash needs to ensure that K8S
 * objects that map to the same Avi objects hash to the same wq. E.g.
 * Routes that share the same "host" should hash to the same wq, so "host"
 * is the hash key for Routes. For objects like Service, it can be ns:name
 */

func CrudHashKey(obj_type string, obj interface{}) string {
	var ns, name string
	switch obj_type {
	case "Endpoints":
		ep := obj.(*corev1.Endpoints)
		ns = ep.Namespace
		name = ep.Name
	case "Service":
		svc := obj.(*corev1.Service)
		ns = svc.Namespace
		name = svc.Name
	case "Ingress":
		ing := obj.(*extensions.Ingress)
		ns = ing.Namespace
		name = ing.Name
	default:
		AviLog.Errorf("Unknown obj_type %s obj %v", obj_type, obj)
		return ":"
	}
	return ns + ":" + name
}

func Bkt(key string, num_workers uint32) uint32 {
	bkt := Hash(key) & (num_workers - 1)
	return bkt
}

// DeepCopy deepcopies a to b using json marshaling
func DeepCopy(a, b interface{}) {
	byt, _ := json.Marshal(a)
	json.Unmarshal(byt, b)
}

func Hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

var letters = []rune("abcdefghijklmnopqrstuvwxyz1234567890")

func RandomSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

var informer sync.Once
var informerInstance *Informers

func instantiateInformers(kubeClient KubeClientIntf, registeredInformers []string, ocs oshiftclientset.Interface, namespace string, adv_l4 bool) *Informers {
	cs := kubeClient.ClientSet
	var kubeInformerFactory, configMapInformerFactory kubeinformers.SharedInformerFactory
	if namespace == "" {
		kubeInformerFactory = kubeinformers.NewSharedInformerFactoryWithOptions(cs, InformerDefaultResync)
	} else {
		// The informer factory only allows to initialize 1 namespace filter. Not a set of namespaces.
		kubeInformerFactory = kubeinformers.NewSharedInformerFactoryWithOptions(cs, InformerDefaultResync, kubeinformers.WithNamespace(namespace))
		AviLog.Infof("Initialized informer factory for namespace :%s", namespace)
	}
	// We listen to configmaps only in `avi-system or vmware-system-ako`
	if adv_l4 {
		// Advanced L4 is for vmware-system-ako
		configmapns := VMWARE_SYSTEM_AKO
		configMapInformerFactory = kubeinformers.NewSharedInformerFactoryWithOptions(cs, InformerDefaultResync, kubeinformers.WithNamespace(configmapns))
		AviLog.Infof("Initializing configmap informer in %v", configmapns)
	} else {
		// Regular AKO
		configmapns := AKO_DEFAULT_NS
		configMapInformerFactory = kubeinformers.NewSharedInformerFactoryWithOptions(cs, InformerDefaultResync, kubeinformers.WithNamespace(configmapns))
		AviLog.Infof("Initializing configmap informer in %v", configmapns)
	}
	informers := &Informers{}
	informers.KubeClientIntf = kubeClient
	for _, informer := range registeredInformers {
		switch informer {
		case ServiceInformer:
			informers.ServiceInformer = kubeInformerFactory.Core().V1().Services()
		case NSInformer:
			informers.NSInformer = kubeInformerFactory.Core().V1().Namespaces()
		case PodInformer:
			informers.PodInformer = kubeInformerFactory.Core().V1().Pods()
		case EndpointInformer:
			informers.EpInformer = kubeInformerFactory.Core().V1().Endpoints()
		case SecretInformer:
			informers.SecretInformer = kubeInformerFactory.Core().V1().Secrets()
		case NodeInformer:
			informers.NodeInformer = kubeInformerFactory.Core().V1().Nodes()
		case ConfigMapInformer:
			informers.ConfigMapInformer = configMapInformerFactory.Core().V1().ConfigMaps()
		case IngressInformer:
			ingressAPI := GetIngressApi(cs)
			if ingressAPI == ExtV1IngressInformer {
				inginformer, _ := kubeInformerFactory.ForResource(ExtensionsIngress)
				informers.IngressInformer = inginformer
				informers.IngressVersion = ExtV1IngressInformer
			} else {
				inginformer, _ := kubeInformerFactory.ForResource(NetworkingIngress)
				informers.IngressInformer = inginformer
				informers.IngressVersion = CoreV1IngressInformer
			}
		case RouteInformer:
			if ocs != nil {
				oshiftInformerFactory := oshiftinformers.NewSharedInformerFactory(ocs, time.Second*30)
				informers.RouteInformer = oshiftInformerFactory.Route().V1().Routes()
				informers.OshiftClient = ocs
			}
		}
	}
	return informers
}

/*
 * Returns a set of informers. By default the informer set would be instantiated once and reused for subsequent calls.
 * Extra arguments can be passed in form of key value pairs.
 * "instanciateOnce" <bool> : If false, then a new set of informers would be returned for each call.
 * "oshiftclient" <oshiftclientset.Interface> : Informer for openshift route has to be registered using openshiftclient
 */

func NewInformers(kubeClient KubeClientIntf, registeredInformers []string, args ...map[string]interface{}) *Informers {
	var oshiftclient oshiftclientset.Interface
	var instantiateOnce, ok, adv_l4 bool = true, true, false
	var namespace string
	if len(args) > 0 {
		for k, v := range args[0] {
			switch k {
			case INFORMERS_INSTANTIATE_ONCE:
				instantiateOnce, ok = v.(bool)
				if !ok {
					AviLog.Warnf("arg instantiateOnce is not of type bool")
				}
			case INFORMERS_ADVANCED_L4:
				adv_l4, ok = v.(bool)
				if !ok {
					AviLog.Infof("Running AKO in avi-system namespace")
				}
			case INFORMERS_OPENSHIFT_CLIENT:
				oshiftclient, ok = v.(oshiftclientset.Interface)
				if !ok {
					AviLog.Warnf("arg oshiftclient is not of type oshiftclientset.Interface")
				}
			case INFORMERS_NAMESPACE:
				namespace, ok = v.(string)
				if !ok {
					AviLog.Warnf("arg namespace is not of type string")
				}
			default:
				AviLog.Warnf("Unknown Key %s in args", k)
			}
		}
	}
	if !instantiateOnce {
		return instantiateInformers(kubeClient, registeredInformers, oshiftclient, namespace, adv_l4)
	}
	informer.Do(func() {
		informerInstance = instantiateInformers(kubeClient, registeredInformers, oshiftclient, namespace, adv_l4)
	})
	return informerInstance
}

func GetInformers() *Informers {
	if informerInstance == nil {
		AviLog.Fatal("Cannot retrieve the informers since it's not initialized yet.")
		return nil
	}
	return informerInstance
}

func Stringify(serialize interface{}) string {
	json_marshalled, _ := json.Marshal(serialize)
	return string(json_marshalled)
}

func ExtractNamespaceObjectName(key string) (string, string) {
	segments := strings.Split(key, "/")
	if len(segments) == 2 {
		return segments[0], segments[1]
	}
	return "", ""
}

func HasElem(s interface{}, elem interface{}) bool {
	arrV := reflect.ValueOf(s)

	if arrV.Kind() == reflect.Slice {
		for i := 0; i < arrV.Len(); i++ {
			// XXX - panics if slice element points to an unexported struct field
			// see https://golang.org/pkg/reflect/#Value.Interface
			if arrV.Index(i).Interface() == elem {
				return true
			}
		}
	}

	return false
}

func ObjKey(obj interface{}) string {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		AviLog.Warn(err)
	}

	return key
}

func Remove(arr []string, item string) []string {
	for i, v := range arr {
		if v == item {
			return append(arr[:i], arr[i+1:]...)
		}
	}
	return arr
}

var GlobalK8NSObj *K8ValidNamespaces = &K8ValidNamespaces{}

func GetGlobalK8NSObj() *K8ValidNamespaces {
	return GlobalK8NSObj
}

func IsNSPresent(namespace string, obj *K8ValidNamespaces) bool {
	obj.validNSList.lock.RLock()
	defer obj.validNSList.lock.RUnlock()
	_, flag := obj.validNSList.nsList[namespace]
	AviLog.Debugf("Is Namespace [%s] valid? -->[%v]", namespace, flag)
	return flag
}

func InitializeNSSync(label string) {
	AviLog.Debugf("In InitializeNSSync function, namespace label to check is: %s\n", label)
	if label != "" {
		segments := strings.Split(label, "=")
		seg0 := strings.Trim(segments[0], " ")
		seg1 := strings.Trim(segments[1], " ")
		if len(segments) != 2 || seg0 == "" || seg1 == "" {
			AviLog.Error("Error occured while parsing labels.")
			return
		}
		GlobalK8NSObj.EnableMigration = true
		GlobalK8NSObj.nsFilter.key = seg0
		GlobalK8NSObj.nsFilter.value = seg1
		GlobalK8NSObj.validNSList.nsList = make(map[string]bool)
	}
}

//Get namespace label filter key and value
func GetNSFilter(obj *K8ValidNamespaces) (string, string) {
	var key string
	var value string
	if obj.nsFilter.key != "" {
		key = obj.nsFilter.key
	}
	if obj.nsFilter.value != "" {
		value = obj.nsFilter.value
	}
	return key, value
}

func AddNamespace(namespace string, obj *K8ValidNamespaces) {
	obj.validNSList.lock.Lock()
	defer obj.validNSList.lock.Unlock()
	obj.validNSList.nsList[namespace] = true
}

func DeleteNamespace(namespace string, obj *K8ValidNamespaces) {
	obj.validNSList.lock.Lock()
	defer obj.validNSList.lock.Unlock()
	delete(obj.validNSList.nsList, namespace)
}

func NSFilterFunction(namespace string, obj *K8ValidNamespaces, nsLabels map[string]string, nonNSK8ResFlag bool) bool {
	//Return true if there is no migration labels mentioned
	if !obj.EnableMigration {
		return true
	}
	//For k8 resources other than namespace check NS already present or not
	if nonNSK8ResFlag && IsNSPresent(namespace, obj) {
		return true
	}

	//Following code will be called for Namespace case only from nsevent handler
	if len(nsLabels) != 0 {
		// if namespace have labels
		nsKey, nsValue := GetNSFilter(obj)
		val, ok := nsLabels[nsKey]
		if ok && val == nsValue {
			AviLog.Debugf("Namespace filter passed for namespace: [%s]", namespace)
			return true
		}
	}
	return false
}
func retrieveNSList(nsList map[string]bool) []string {
	var namespaces []string
	for k := range nsList {
		namespaces = append(namespaces, k)
	}
	return namespaces
}
func GetAllNamespacesToSync(obj *K8ValidNamespaces) []string {
	var nsList []string
	if !obj.EnableMigration {
		return nsList
	}
	obj.validNSList.lock.RLock()
	defer obj.validNSList.lock.RUnlock()
	lst := retrieveNSList(obj.validNSList.nsList)
	nsList = append(nsList, lst...)

	return nsList
}
