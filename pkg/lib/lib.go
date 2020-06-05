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
	"os"
	"strconv"
	"strings"

	"github.com/avinetworks/container-lib/utils"
	"github.com/avinetworks/sdk/go/models"
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

var onlyOneSignalHandler = make(chan struct{})

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
	return GetVrf() + "--" + namespace + "--" + svcName
}

func GetL4VSVipName(svcName, namespace string) string {
	return GetVrf() + "--" + namespace + "--" + svcName
}

func GetL4PoolName(vsName string, port int32) string {
	return vsName + "--" + strconv.Itoa(int(port))
}

func GetL4PGName(vsName string, port int32) string {
	return vsName + "--" + strconv.Itoa(int(port))
}

// All L7 object names.
func GetVsVipName(vsName string) string {
	return GetVrf() + "--" + vsName
}

func GetL7InsecureDSName(vsName string) string {
	return GetVrf() + "--" + vsName
}

func GetL7SharedPGName(vsName string) string {
	return GetVrf() + "--" + vsName
}

func GetL7PoolName(priorityLabel, namespace, ingName string) string {
	priorityLabel = strings.Replace(priorityLabel, "/", "_", 1)
	return GetVrf() + "--" + priorityLabel + "--" + namespace + "--" + ingName
}

func GetL7HttpRedirPolicy(vsName string) string {
	return GetVrf() + "--" + vsName
}

func GetSniNodeName(ingName, namespace, secret string, sniHostName ...string) string {
	if len(sniHostName) > 0 {
		return GetVrf() + "--" + sniHostName[0]
	}
	return GetVrf() + "--" + ingName + "--" + namespace + "--" + secret
}

func GetSniPoolName(ingName, namespace, host, path string) string {
	path = strings.Replace(path, "/", "_", 1)
	return GetVrf() + "--" + namespace + "--" + host + path + "--" + ingName
}

func GetSniHttpPolName(ingName, namespace, host, path string) string {
	path = strings.Replace(path, "/", "_", 1)
	return GetVrf() + "--" + namespace + "--" + host + path + "--" + ingName
}

func GetSniPGName(ingName, namespace, host, path string) string {
	path = strings.Replace(path, "/", "_", 1)
	return GetVrf() + "--" + namespace + "--" + host + path + "--" + ingName
}

func GetTLSKeyCertNodeName(namespace, secret string, sniHostName ...string) string {
	if len(sniHostName) > 0 {
		return GetVrf() + "--" + sniHostName[0]
	}
	return GetVrf() + "--" + namespace + "--" + secret
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

func VrfChecksum(vrfName string, staticRoutes []*models.StaticRoute) uint32 {
	return (utils.Hash(vrfName) + utils.Hash(utils.Stringify(staticRoutes)))
}
