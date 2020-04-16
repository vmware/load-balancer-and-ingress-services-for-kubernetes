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

package lib

import (
	"fmt"
	"os"

	"github.com/avinetworks/container-lib/utils"
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
		return 0
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
	return vsName + "--" + fmt.Sprint(port)
}

func GetL4PGName(vsName string, port int32) string {
	return vsName + "--" + fmt.Sprint(port)
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
	return GetVrf() + "--" + priorityLabel + "--" + namespace + "--" + ingName
}

func GetL7HttpRedirPolicy(vsName string) string {
	return GetVrf() + "--" + vsName
}

func GetSniNodeName(ingName, namespace, secret string, sniHostName ...string) string {
	if len(sniHostName) > 0 {
		return GetVrf() + "--" + ingName + "--" + namespace + "--" + sniHostName[0]
	}
	return GetVrf() + "--" + ingName + "--" + namespace + "--" + secret
}

func GetSniPoolName(ingName, namespace, host, path string) string {
	return GetVrf() + "--" + namespace + "--" + host + path + "--" + ingName
}

func GetSniHttpPolName(ingName, namespace, host, path string) string {
	return GetVrf() + "--" + namespace + "--" + host + path + "--" + ingName
}

func GetSniPGName(ingName, namespace, host, path string) string {
	return GetVrf() + "--" + namespace + "--" + host + path + "--" + ingName
}

func GetTLSKeyCertNodeName(namespace, secret string) string {
	return GetVrf() + "--" + namespace + "--" + secret
}

func GetVrf() string {
	vrfcontext := os.Getenv(utils.VRF_CONTEXT)
	if vrfcontext == "" {
		vrfcontext = utils.GlobalVRF
	}
	return vrfcontext
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
	utils.AviLog.Info.Printf("SHARDING scheme :%s", shardSchemeName)
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
	// Additional checks can be performed here.
	return os.Getenv(SUBNET_IP)

}

func GetCIDR() string {
	// Additional checks can be performed here.
	return os.Getenv(SUBNET_CIDR)

}

func GetNetworkName() string {
	// Additional checks can be performed here.
	return os.Getenv(NETWORK_NAME)

}
