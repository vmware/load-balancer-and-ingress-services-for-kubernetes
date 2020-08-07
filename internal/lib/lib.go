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

package lib

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"ako/pkg/api"
	"ako/pkg/utils"

	"github.com/avinetworks/sdk/go/models"
	oshiftclient "github.com/openshift/client-go/route/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var IngressApiMap = map[string]string{
	"corev1":      utils.CoreV1IngressInformer,
	"extensionv1": utils.ExtV1IngressInformer,
}

var ShardSchemeMap = map[string]string{
	"hostname":  "hostname",
	"namespace": "namespace",
}

var shardSizeMap = map[string]uint32{
	"LARGE":  8,
	"MEDIUM": 4,
	"SMALL":  1,
}

var NamePrefix string

func SetNamePrefix() {
	NamePrefix = GetClusterName() + "--"
}

func GetNamePrefix() string {
	return NamePrefix
}

var AKOUser string

func SetAKOUser() {
	AKOUser = "ako-" + GetClusterName()
	utils.AviLog.Infof("Setting AKOUser: %s for Avi Objects", AKOUser)
}

func GetAKOUser() string {
	return AKOUser
}

func GetshardSize() uint32 {
	shardVsSize := os.Getenv("SHARD_VS_SIZE")
	shardSize, ok := shardSizeMap[shardVsSize]
	if ok {
		return shardSize
	} else {
		return 1
	}
}

func GetModelName(namespace, objectName string) string {
	return namespace + "/" + objectName
}

// All L4 object names.
func GetL4VSName(svcName, namespace string) string {
	return NamePrefix + namespace + "-" + svcName
}

func GetL4VSVipName(svcName, namespace string) string {
	return NamePrefix + namespace + "-" + svcName
}

func GetL4PoolName(vsName string, port int32) string {
	return vsName + "--" + strconv.Itoa(int(port))
}

func GetL4PolicyName(vsName string, port int32) string {
	return GetL4PoolName(vsName, port)
}

func GetL4PGName(vsName string, port int32) string {
	return vsName + "-" + strconv.Itoa(int(port))
}

// All L7 object names.
func GetVsVipName(vsName string) string {
	return vsName
}

func GetL7InsecureDSName(vsName string) string {
	return vsName
}

func GetL7SharedPGName(vsName string) string {
	return vsName
}

func GetL7PoolName(priorityLabel, namespace, ingName string, args ...string) string {
	priorityLabel = strings.Replace(priorityLabel, "/", "_", 1)
	poolName := NamePrefix + priorityLabel + "-" + namespace + "-" + ingName
	if len(args) > 0 {
		svcName := args[0]
		poolName = poolName + "-" + svcName
	}
	return poolName
}

func GetL7HttpRedirPolicy(vsName string) string {
	return vsName
}

func GetSniNodeName(ingName, namespace, secret string, sniHostName ...string) string {
	if len(sniHostName) > 0 {
		return NamePrefix + sniHostName[0]
	}
	return NamePrefix + ingName + "-" + namespace + "-" + secret
}

func GetSniPoolName(ingName, namespace, host, path string, args ...string) string {
	path = strings.Replace(path, "/", "_", 1)
	poolName := NamePrefix + namespace + "-" + host + path + "-" + ingName
	if len(args) > 0 {
		svcName := args[0]
		poolName = poolName + "-" + svcName
	}
	return poolName
}

func GetSniHttpPolName(ingName, namespace, host, path string) string {
	path = strings.ReplaceAll(path, "/", "_")
	return NamePrefix + namespace + "-" + host + path + "-" + ingName
}

func GetSniPGName(ingName, namespace, host, path string) string {
	path = strings.ReplaceAll(path, "/", "_")
	return NamePrefix + namespace + "-" + host + path + "-" + ingName
}

func GetTLSKeyCertNodeName(namespace, secret string, sniHostName ...string) string {
	if len(sniHostName) > 0 {
		return NamePrefix + sniHostName[0]
	}
	return NamePrefix + namespace + "-" + secret
}

func GetCACertNodeName(keycertname string) string {
	return keycertname + "-cacert"
}

func GetPoolTLSKeyCertNodeName(httprule, pathPrefix string) string {
	if pathPrefix == "/" {
		pathPrefix = ""
	}
	return NamePrefix + strings.ReplaceAll(httprule, "/", "-") + strings.ReplaceAll(pathPrefix, "/", "_")
}

var VRFContext string
var VRFUuid string

func SetVrf(vrf string) {
	VRFContext = vrf
}

func SetVrfUuid(uuid string) {
	VRFUuid = uuid
}

func GetVrfUuid() string {
	if VRFUuid == "" {
		utils.AviLog.Warnf("VRF uuid not set")
	}
	return VRFUuid
}

func GetVrf() string {
	if VRFContext == "" {
		return utils.GlobalVRF
	}
	return VRFContext
}

func GetTenant() string {
	// tenant := os.Getenv("CTRL_TENANT")
	// if tenant == "" {
	// 	tenant = utils.ADMIN_NS
	// }
	tenant := utils.ADMIN_NS
	return tenant
}

func GetIngressApi() string {
	ingressApi := os.Getenv(INGRESS_API)
	ingressApi, ok := IngressApiMap[ingressApi]
	if !ok {
		return utils.CoreV1IngressInformer
	}
	return ingressApi
}

func GetShardScheme() string {
	shardScheme := os.Getenv(L7_SHARD_SCHEME)
	shardSchemeName, ok := ShardSchemeMap[shardScheme]
	if !ok {
		return DEFAULT_SHARD_SCHEME
	}
	utils.AviLog.Infof("SHARDING scheme :%s", shardSchemeName)
	return shardSchemeName
}

func GetDefaultIngController() bool {
	defaultIngCtrl := os.Getenv("DEFAULT_ING_CONTROLLER")
	if defaultIngCtrl != "false" {
		return true
	}
	return false
}

func GetNamespaceToSync() string {
	namespace := os.Getenv("SYNC_NAMESPACE")
	if namespace != "" {
		return namespace
	}
	return ""
}

func GetSubnetIP() string {
	subnetIP := os.Getenv(SUBNET_IP)
	if subnetIP != "" {
		return subnetIP
	}
	return ""
}

func GetSubnetPrefix() string {
	subnetPrefix := os.Getenv(SUBNET_PREFIX)
	if subnetPrefix != "" {
		return subnetPrefix
	}
	return ""
}

func GetNetworkName() string {
	networkName := os.Getenv(NETWORK_NAME)
	if networkName != "" {
		return networkName
	}
	return ""
}

func GetDomain() string {
	subDomain := os.Getenv(DEFAULT_DOMAIN)
	if subDomain != "" {
		return subDomain
	}
	return ""
}

func GetClusterName() string {
	clusterName := os.Getenv(CLUSTER_NAME)
	if clusterName != "" {
		return clusterName
	}
	return ""
}

var StaticRouteSyncChan chan struct{}

var akoApi *api.ApiServer

func SetStaticRouteSyncHandler() {
	StaticRouteSyncChan = make(chan struct{})
}

func SetApiServerInstance(akoApiInstance *api.ApiServer) {
	akoApi = akoApiInstance
}

func ShutdownApi() {
	akoApi.ShutDown()
}

func VrfChecksum(vrfName string, staticRoutes []*models.StaticRoute) uint32 {
	return (utils.Hash(vrfName) + utils.Hash(utils.Stringify(staticRoutes)))
}

func DSChecksum(pgrefs []string) uint32 {
	sort.Strings(pgrefs)
	checksum := utils.Hash(utils.Stringify(pgrefs))
	return checksum
}

func InformersToRegister(oclient *oshiftclient.Clientset, kclient *kubernetes.Clientset) []string {
	//allInformers := []string{}
	allInformers := []string{
		utils.ServiceInformer,
		utils.EndpointInformer,
		utils.SecretInformer,
		utils.NSInformer,
		utils.NodeInformer,
		utils.ConfigMapInformer,
	}
	informerTimeout := int64(120)
	_, err := oclient.RouteV1().Routes("").List(metav1.ListOptions{TimeoutSeconds: &informerTimeout})
	if err == nil {
		// Openshift cluster with route support, we will just add route informer
		allInformers = append(allInformers, utils.RouteInformer)
	} else {
		// Kubernetes cluster
		allInformers = append(allInformers, utils.IngressInformer)
	}
	return allInformers
}

func SSLKeyCertChecksum(sslName, certificate, cacert string) uint32 {
	return utils.Hash(sslName + certificate + cacert)
}

func L4PolicyChecksum(ports []int64, protocol string) uint32 {
	var portsInt []int
	for _, port := range ports {
		portsInt = append(portsInt, int(port))
	}
	sort.Ints(portsInt)
	return utils.Hash(utils.Stringify(portsInt)) + utils.Hash(protocol)
}

func IsNodePortMode() bool {
	nodePortType := os.Getenv(SERVICE_TYPE)
	if nodePortType == NODE_PORT {
		return true
	}
	return false
}

func GetNodePortsSelector() map[string]string {
	nodePortsSelectorLabels := make(map[string]string)
	if IsNodePortMode() {
		// If the key/values are kept empty then we select all nodes
		nodePortsSelectorLabels["key"] = os.Getenv(NODE_KEY)
		nodePortsSelectorLabels["value"] = os.Getenv(NODE_VALUE)
	}
	return nodePortsSelectorLabels
}

var CloudType string

func SetCloudType(cloudType string) {
	CloudType = cloudType
}

func GetCloudType() string {
	if CloudType == "" {
		return CLOUD_VCENTER
	}
	return CloudType
}

func IsPublicCloud() bool {

	if GetCloudType() == CLOUD_AZURE || GetCloudType() == CLOUD_AWS {
		return true
	}
	return false
}

func PassthroughShardSize() uint32 {
	shardVsSize := os.Getenv("PASSTHROUGH_SHARD_SIZE")
	shardSize, ok := shardSizeMap[shardVsSize]
	if ok {
		return shardSize
	}
	return 1
}

func GetPassthroughShardVSName(s string, key string) string {
	var vsNum uint32
	shardSize := PassthroughShardSize()
	shardVsPrefix := GetClusterName() + "--" + PassthroughPrefix
	vsNum = utils.Bkt(s, shardSize)
	vsName := shardVsPrefix + fmt.Sprint(vsNum)
	utils.AviLog.Infof("key: %s, msg: ShardVSName: %s", key, vsName)
	return vsName
}
