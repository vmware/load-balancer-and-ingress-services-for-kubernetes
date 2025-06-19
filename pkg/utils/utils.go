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
	"errors"
	"hash/fnv"
	"math/rand"
	"net"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	oshiftclientset "github.com/openshift/client-go/route/clientset/versioned"
	oshiftinformers "github.com/openshift/client-go/route/informers/externalversions"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	akov1beta1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1beta1"
	// TODO: Check this to convert to v1beta1 in next release. Couldn't conver as MCI and SI uses that.
	akocrd "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha1/clientset/versioned"

	akoinformers "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha1/informers/externalversions"
)

var CtrlVersion string
var runtimeScheme = k8sruntime.NewScheme()

func init() {
	//Setting the package-wide version
	CtrlVersion = os.Getenv("CTRL_VERSION")
	networkingv1.AddToScheme(runtimeScheme)
}

func IsV4(addr string) bool {
	ip := net.ParseIP(addr)
	v4 := ip.To4()
	return v4 != nil
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
		ing := obj.(*networkingv1.Ingress)
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

func instantiateInformers(kubeClient KubeClientIntf, registeredInformers []string, ocs oshiftclientset.Interface, akoClientSet akocrd.Interface, namespace string, akoNSBoundInformer bool) *Informers {
	cs := kubeClient.ClientSet
	var kubeInformerFactory, akoNSInformerFactory kubeinformers.SharedInformerFactory
	if namespace == "" {
		kubeInformerFactory = kubeinformers.NewSharedInformerFactoryWithOptions(cs, InformerDefaultResync)
	} else {
		// The informer factory only allows to initialize 1 namespace filter. Not a set of namespaces.
		kubeInformerFactory = kubeinformers.NewSharedInformerFactoryWithOptions(cs, InformerDefaultResync, kubeinformers.WithNamespace(namespace))
		AviLog.Infof("Initialized informer factory for namespace :%s", namespace)
	}

	// We listen to configmaps only in the namespace in which AKO runs.
	akoNS := GetAKONamespace()

	akoNSInformerFactory = kubeinformers.NewSharedInformerFactoryWithOptions(cs, InformerDefaultResync, kubeinformers.WithNamespace(akoNS))
	AviLog.Infof("Initializing configmap informer in %v", akoNS)

	// To initialize the MCI and SI informers
	akoInformerFactory := akoinformers.NewSharedInformerFactoryWithOptions(akoClientSet, time.Second*30)

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
			if akoNSBoundInformer {
				informers.SecretInformer = akoNSInformerFactory.Core().V1().Secrets()
			} else {
				informers.SecretInformer = kubeInformerFactory.Core().V1().Secrets()
			}
		case NodeInformer:
			informers.NodeInformer = kubeInformerFactory.Core().V1().Nodes()
		case ConfigMapInformer:
			informers.ConfigMapInformer = akoNSInformerFactory.Core().V1().ConfigMaps()
		case IngressInformer:
			informers.IngressInformer = kubeInformerFactory.Networking().V1().Ingresses()
		case IngressClassInformer:
			informers.IngressClassInformer = kubeInformerFactory.Networking().V1().IngressClasses()
		case RouteInformer:
			if ocs != nil {
				oshiftInformerFactory := oshiftinformers.NewSharedInformerFactory(ocs, time.Second*30)
				informers.RouteInformer = oshiftInformerFactory.Route().V1().Routes()
				informers.OshiftClient = ocs
			}
		case MultiClusterIngressInformer:
			informers.MultiClusterIngressInformer = akoInformerFactory.Ako().V1alpha1().MultiClusterIngresses()
		case ServiceImportInformer:
			informers.ServiceImportInformer = akoInformerFactory.Ako().V1alpha1().ServiceImports()
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
	var akoClient akocrd.Interface
	var instantiateOnce, ok, akoNSBoundInformer bool = true, true, false
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
				akoNSBoundInformer, ok = v.(bool)
				if !ok {
					AviLog.Infof("Running AKO in %s namespace", GetAKONamespace())
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
			case INFORMERS_AKO_CLIENT:
				// From main call, this `v` is of type v1beta1 but as there is no field changes
				// so conversion should happen.
				// ako client is used for initializing MCI and SI CRDs which are disabled by default.
				// These CRDs will be removed in future. So this code path of conversion should be hit
				// in normal AKO flow
				akoClient, ok = v.(akocrd.Interface)
				if !ok {
					AviLog.Warnf("arg akoClient is not of type akocrd.Interface")
				}
			default:
				AviLog.Warnf("Unknown Key %s in args", k)
			}
		}
	}

	// In openshift, the secret handling is restricted to the namespace where the AKO is
	// installed if the user sets `handleSecretsFromAKONSOnly` to true.
	if oshiftclient != nil &&
		IsSecretsHandlingRestrictedToAKONS() {
		akoNSBoundInformer = true
	}

	if !instantiateOnce {
		return instantiateInformers(kubeClient, registeredInformers, oshiftclient, akoClient, namespace, akoNSBoundInformer)
	}
	informer.Do(func() {
		informerInstance = instantiateInformers(kubeClient, registeredInformers, oshiftclient, akoClient, namespace, akoNSBoundInformer)
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
			// Important - Panics if slice element points to an unexported struct field
			// see https://golang.org/pkg/reflect/#Value.Interface
			if reflect.DeepEqual(arrV.Index(i).Interface(), elem) {
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

func FindAndRemove(arr []string, item string) (bool, []string) {
	for i, v := range arr {
		if v == item {
			return true, append(arr[:i], arr[i+1:]...)
		}
	}
	return false, arr
}

func RemoveNamespaceName(s []NamespaceName, r NamespaceName) []NamespaceName {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

var globalNSFilterObj *K8ValidNamespaces = &K8ValidNamespaces{}

func GetGlobalNSFilter() *K8ValidNamespaces {
	return globalNSFilterObj
}

func IsNSPresent(namespace string, obj *K8ValidNamespaces) bool {
	obj.validNSList.lock.RLock()
	defer obj.validNSList.lock.RUnlock()
	_, flag := obj.validNSList.nsList[namespace]
	AviLog.Debugf("Namespace %s is accepted : %v", namespace, flag)
	return flag
}

func InitializeNSSync(labelKey, labelVal string) {
	globalNSFilterObj.EnableMigration = true
	globalNSFilterObj.nsFilter.key = labelKey
	globalNSFilterObj.nsFilter.value = labelVal
	globalNSFilterObj.validNSList.nsList = make(map[string]struct{})
}

// Get namespace label filter key and value
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

func AddNamespaceToFilter(namespace string) {
	globalNSFilterObj.validNSList.lock.Lock()
	defer globalNSFilterObj.validNSList.lock.Unlock()
	globalNSFilterObj.validNSList.nsList[namespace] = struct{}{}
}

func DeleteNamespaceFromFilter(namespace string) {
	globalNSFilterObj.validNSList.lock.Lock()
	defer globalNSFilterObj.validNSList.lock.Unlock()
	delete(globalNSFilterObj.validNSList.nsList, namespace)
}

func CheckIfNamespaceAccepted(namespace string, opts ...interface{}) bool {
	//Return true if there is no migration labels mentioned
	if !globalNSFilterObj.EnableMigration {
		return true
	}
	var nsLabels map[string]string = nil
	var nonNSK8ResFlag bool = true
	if len(opts) == 2 {
		nsLabels = opts[0].(map[string]string)
		nonNSK8ResFlag = opts[1].(bool)
	}
	//For k8 resources other than namespace check NS already present or not
	if nonNSK8ResFlag && IsNSPresent(namespace, globalNSFilterObj) {
		return true
	}

	//Following code will be called for Namespace case only from nsevent handler
	if len(nsLabels) != 0 {
		// if namespace have labels
		nsKey, nsValue := GetNSFilter(globalNSFilterObj)
		val, ok := nsLabels[nsKey]
		if ok && val == nsValue {
			AviLog.Debugf("Namespace filter passed for namespace: %s", namespace)
			return true
		}
	}
	return false
}
func IsServiceNSValid(namespace string) bool {
	// L4 Namespace sync not applicable for advance L4
	if !GetAdvancedL4() && !IsVCFCluster() {
		if !CheckIfNamespaceAccepted(namespace) {
			return false
		}
	}
	return true
}

func IsVCFCluster() bool {
	vcfCluster := os.Getenv(VCF_CLUSTER)
	if val, err := strconv.ParseBool(vcfCluster); err == nil {
		return val
	}
	return false
}

// This utility returns a true/false depending on whether
// the user requires advanced L4 functionality
func GetAdvancedL4() bool {
	if ok, _ := strconv.ParseBool(os.Getenv(ADVANCED_L4)); ok {
		return true
	}
	return false
}

// Wrapper function for AKO running in either VDS
// or VCF (WCP with NSX).
func IsWCP() bool {
	if GetAdvancedL4() || IsVCFCluster() {
		return true
	}
	return false
}

// GetAKONamespace returns the namespace of AKO pod.
// In WCP Mode this is vmware-system-ako
// In all other cases this is the namespace in which the
// statefulset runs.
func GetAKONamespace() string {
	akoNS := os.Getenv(POD_NAMESPACE)
	if IsWCP() {
		akoNS = VMWARE_SYSTEM_AKO
	}
	return akoNS
}

func GetTokenFromRestObj(robj interface{}, ctrlAuthToken string) (oldTokenID string, refresh bool, err error) {
	oldTokenID = ""
	refresh = false
	err = nil
	parseError := errors.New("Failed to parse token response obj")

	if _, ok := robj.(map[string]interface{}); !ok {
		err = parseError
		return
	}
	tokenList, ok := robj.(map[string]interface{})["results"].([]interface{})
	if !ok {
		err = parseError
		return
	}
	for _, aviToken := range tokenList {
		if _, ok := aviToken.(map[string]interface{}); !ok {
			err = parseError
			return
		}
		token, ok := aviToken.(map[string]interface{})["token"].(string)
		if !ok {
			err = parseError
			return
		}
		if token == ctrlAuthToken {
			expiry, ok := aviToken.(map[string]interface{})["expires_at"].(string)
			if !ok {
				err = parseError
				return
			}
			layout := "2006-01-02T15:04:05.000000+00:00"
			expiryTime, err2 := time.Parse(layout, expiry)
			if err != nil {
				AviLog.Errorf("Unable to parse token expiry time, err: %+v", err2)
				err = err2
				return
			}
			AviLog.Infof("Expiry time for current token: %+v", expiryTime)
			if expiryTime.Sub(time.Now()) > (RefreshAuthTokenPeriod*AuthTokenExpiry)*time.Hour {
				return
			}
			refresh = true
			if tokenIDToDelete, ok := aviToken.(map[string]interface{})["uuid"].(string); ok {
				oldTokenID = tokenIDToDelete
			}
			return
		}

	}
	refresh = true
	return
}

func GetAuthtokenFromCache() (string, error) {
	ctrlAuthToken, ok := SharedCtrlProp().AviCacheGet(ENV_CTRL_AUTHTOKEN)
	if !ok || ctrlAuthToken == nil {
		return "", errors.New("authToken not updated in cache")
	}
	return ctrlAuthToken.(string), nil
}

func ContainsDuplicate(arr interface{}) bool {
	arrV := reflect.ValueOf(arr)

	if arrV.Kind() == reflect.Slice {
		eleMap := make(map[interface{}]struct{})
		for i := 0; i < arrV.Len(); i++ {
			if _, ok := eleMap[arrV.Index(i).Interface()]; ok {
				return true
			}
			eleMap[arrV.Index(i).Interface()] = struct{}{}
		}
	}

	return false
}

func IsMultiClusterIngressEnabled() bool {
	if ok, _ := strconv.ParseBool(os.Getenv(MCI_ENABLED)); ok {
		return true
	}
	AviLog.Debugf("Multi-cluster ingress is not enabled")
	return false
}

type Version struct {
	subversions []int
}

func (v *Version) Compare(v1 *Version) int {
	/*
		return 0 if v and v1 are equal
		return -1 if v is less than v1
		return 1 if v is greater than 1
	*/
	length := len(v.subversions)
	if len(v1.subversions) < length {
		length = len(v1.subversions)
	}
	for i := 0; i < length; i++ {
		if v.subversions[i] == v1.subversions[i] {
			continue
		}
		if v.subversions[i] < v1.subversions[i] {
			return -1
		}
		return 1
	}
	if len(v.subversions) == len(v1.subversions) {
		return 0
	}
	if len(v.subversions) < len(v1.subversions) {
		return -1
	}
	return 1
}

func NewVersion(version string) (*Version, error) {
	substrings := strings.Split(version, ".")
	v := &Version{
		subversions: make([]int, 0),
	}
	for _, substr := range substrings {
		val, err := strconv.Atoi(substr)
		if err != nil {
			return nil, err
		}
		v.subversions = append(v.subversions, val)
	}
	return v, nil
}

// This utility returns a true/false depending on whether
// the secret handling is restricted to the namespace where the AKO is installed.
func IsSecretsHandlingRestrictedToAKONS() bool {
	ok, err := strconv.ParseBool(os.Getenv(USE_DEFAULT_SECRETS_ONLY))
	if err != nil {
		return false
	}
	return ok
}

var VipNetworkList []akov1beta1.AviInfraSettingVipNetwork

func SetVipNetworkList(vipNetworks []akov1beta1.AviInfraSettingVipNetwork) {
	VipNetworkList = vipNetworks
}

func GetVipNetworkList() []akov1beta1.AviInfraSettingVipNetwork {
	return VipNetworkList
}

func String(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

func GetUriEncoded(uri string) string {
	newUri, err := url.Parse(uri)
	if err != nil {
		AviLog.Errorf("Error while parsing uri: %+v", err)
	}
	queryValues := newUri.Query()
	if len(queryValues) == 0 {
		return uri
	}
	newUri.RawQuery = queryValues.Encode()
	return newUri.String()
}
