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

package nodes

import (
	"fmt"
	"os"

	"gitlab.eng.vmware.com/orion/container-lib/utils"
)

var shardSizeMap = map[string]uint32{
	"LARGE":  8,
	"MEDIUM": 4,
	"SMALL":  2,
}

func (o *AviObjectGraph) BuildL7VSGraph(namespace string, ingName string, key string) {

	var VsNode *AviVsNode

	// The VS node is decided based on the namespace for now.
	VsNode = o.ConstructAviL7VsNode(namespace, key)
	//o.ConstructAviTCPPGPoolNodes(svcObj, VsNode)
	if VsNode != nil {
		o.AddModelNode(VsNode)
		VsNode.CalculateCheckSum()
		o.GraphChecksum = o.GraphChecksum + VsNode.GetCheckSum()
		utils.AviLog.Info.Printf("key: %s msg: checksum  for AVI VS object %v", key, VsNode.GetCheckSum())
		utils.AviLog.Info.Printf("key: %s msg: computed Graph checksum for VS is: %v", key, o.GraphChecksum)
	}
}

func (o *AviObjectGraph) ConstructAviL7VsNode(namespace string, key string) *AviVsNode {
	// Read the value of the num_shards from the environment variable.
	var vsNum uint32
	shardVsSize := os.Getenv("shard_vs_size")
	shardVsPrefix := os.Getenv("shard_vs_name_prefix")
	if shardVsPrefix == "" {
		shardVsPrefix = DEFAULT_SHARD_VS_PREFIX
	}
	shardSize, ok := shardSizeMap[shardVsSize]
	if ok {
		vsNum = utils.Bkt(namespace, shardSize)
	} else {
		utils.AviLog.Warning.Printf("key: %s msg: the value for shard_vs_size does not match the ENUM values", key)
		return nil
	}
	// Derive the right VS for this update.
	vsName := shardVsPrefix + fmt.Sprint(vsNum)
	var avi_vs_meta *AviVsNode

	// This is a shared VS - always created in the admin namespace for now.
	avi_vs_meta = &AviVsNode{Name: vsName, Tenant: "admin",
		EastWest: false}
	// Hard coded ports for the shared VS
	var portProtocols []AviPortHostProtocol
	httpPort := AviPortHostProtocol{Port: 80, Protocol: HTTP}
	httpsPort := AviPortHostProtocol{Port: 443, Protocol: HTTP}
	portProtocols = append(portProtocols, httpPort)
	portProtocols = append(portProtocols, httpsPort)
	avi_vs_meta.PortProto = portProtocols
	// Default case.
	avi_vs_meta.ApplicationProfile = DEFAULT_L7_APP_PROFILE

	avi_vs_meta.NetworkProfile = DEFAULT_TCP_NW_PROFILE

	return avi_vs_meta
}
