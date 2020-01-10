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
	"strings"

	"gitlab.eng.vmware.com/orion/akc/pkg/objects"
	"gitlab.eng.vmware.com/orion/container-lib/utils"
	"k8s.io/apimachinery/pkg/api/errors"
)

const nodeObj = "Node"

func DequeueIngestion(key string) {
	// The key format expected here is: objectType/Namespace/ObjKey
	// The assumption is that an update either affects an LB service type or an ingress. It cannot be both.
	ingressFound := false
	var ingressNames []string
	utils.AviLog.Info.Printf("key: %s, msg: starting graph Sync", key)
	sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)

	objType, namespace, name := extractTypeNameNamespace(key)
	if objType == nodeObj {
		processNodeObj(key, name)
		return
	}
	schema, valid := ConfigDescriptor().GetByType(objType)
	if valid {
		// If it's an ingress related change, let's process that.
		ingressNames, ingressFound = schema.GetParentIngresses(name, namespace, key)
	}
	if !ingressFound {
		// If ingress is not found, let's do the other checks.
		if objType == utils.L4LBService {
			// L4 type of services need special handling. We create a dedicated VS in Avi for these.
			if !isServiceDelete(name, namespace, key) {
				utils.AviLog.Warning.Printf("key: %s, msg: service is of type loadbalancer. Will create dedicated VS nodes", key)
				aviModelGraph := NewAviObjectGraph()
				aviModelGraph.BuildL4LBGraph(namespace, name, key)
				if len(aviModelGraph.GetOrderedNodes()) != 0 {
					publishKeyToRestLayer(aviModelGraph, namespace, name, key, sharedQueue)
				}
			} else {
				// This is a DELETE event. The avi graph is set to nil.
				utils.AviLog.Info.Printf("key: %s, msg: received DELETE event for service", key)
				model_name := namespace + "/" + name
				objects.SharedAviGraphLister().Save(model_name, nil)
				bkt := utils.Bkt(model_name, sharedQueue.NumWorkers)
				sharedQueue.Workqueue[bkt].AddRateLimited(model_name)
			}
		} else if objType == utils.Endpoints {
			svcObj, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).Get(name)
			if err != nil {
				utils.AviLog.Info.Printf("key: %s, msg: there was an error in retrieving the service for endpoint", key)
				return
			}
			if svcObj.Spec.Type == utils.LoadBalancer {
				// This endpoint update affects a LB service.
				aviModelGraph := NewAviObjectGraph()
				aviModelGraph.BuildL4LBGraph(namespace, name, key)
				if len(aviModelGraph.GetOrderedNodes()) != 0 {
					publishKeyToRestLayer(aviModelGraph, namespace, name, key, sharedQueue)
				}
			}
		}
	} else {
		shardVsName := DeriveNamespacedShardVS(namespace, key)
		if shardVsName == "" {
			// If we aren't able to derive the ShardVS name, we should return
			return
		}
		model_name := utils.ADMIN_NS + "/" + shardVsName
		for _, ingress := range ingressNames {
			// The assumption is that the ingress names are from the same namespace as the service/ep updates. Kubernetes
			// does not allow cross tenant ingress references.
			found, aviModel := objects.SharedAviGraphLister().Get(model_name)
			if !found {
				utils.AviLog.Info.Printf("key :%s, msg: model not found, generating new model with name: %s", key, model_name)
				aviModel = NewAviObjectGraph()
				aviModel.(*AviObjectGraph).ConstructAviL7VsNode(shardVsName, key)
			}
			aviModel.(*AviObjectGraph).BuildL7VSGraph(shardVsName, namespace, ingress, key)
			if len(aviModel.(*AviObjectGraph).GetOrderedNodes()) != 0 {
				publishKeyToRestLayer(aviModel.(*AviObjectGraph), utils.ADMIN_NS, shardVsName, key, sharedQueue)
			}
		}
	}
}

func processNodeObj(key, nodename string) {
	utils.AviLog.Info.Printf("key: %s, Got node Object %s\n", key, nodename)
	nodeObj, err := utils.GetInformers().NodeInformer.Lister().Get(nodename)
	if err != nil {
		utils.AviLog.Info.Printf("key %s, Error feting object for node %s: %v\n", key, nodename, err)
		return
	}
	utils.AviLog.Info.Printf("key %s, Node Object %v\n", key, nodeObj)
	// TO Do : generate model for adding static route
}

func publishKeyToRestLayer(aviGraph *AviObjectGraph, namespace string, name string, key string, sharedQueue *utils.WorkerQueue) {
	// First see if there's another instance of the same model in the store
	model_name := namespace + "/" + name
	found, aviModel := objects.SharedAviGraphLister().Get(model_name)
	if found && aviModel != nil {
		prevChecksum := aviModel.(*AviObjectGraph).GetCheckSum()
		utils.AviLog.Info.Printf("key :%s, msg: the model: %s has a previous checksum: %v", key, model_name, prevChecksum)
		presentChecksum := aviGraph.GetCheckSum()
		utils.AviLog.Info.Printf("key: %s, msg: the model: %s has a present checksum: %v", key, model_name, presentChecksum)
		if prevChecksum == presentChecksum {
			utils.AviLog.Info.Printf("key: %s, msg: The model: %s has identical checksums, hence not processing. Checksum value: %v", key, model_name, presentChecksum)
			return
		}
	}
	// TODO (sudswas): Lots of checksum optimization goes here
	objects.SharedAviGraphLister().Save(model_name, aviGraph)
	bkt := utils.Bkt(model_name, sharedQueue.NumWorkers)
	sharedQueue.Workqueue[bkt].AddRateLimited(model_name)
	utils.AviLog.Info.Printf("key: %s, msg: Published key with model_name: %s", key, model_name)
}

func isServiceDelete(svcName string, namespace string, key string) bool {
	// If the service is not found we return true.
	_, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).Get(svcName)
	if err != nil {
		utils.AviLog.Warning.Printf("key: %s, msg: could not retrieve the object for service: %s", key, err)
		if errors.IsNotFound(err) {
			return true
		}
	}
	return false
}

// Candidate for utils.
func extractTypeNameNamespace(key string) (string, string, string) {
	segments := strings.Split(key, "/")
	if len(segments) == 3 {
		return segments[0], segments[1], segments[2]
	}
	if len(segments) == 2 {
		return segments[0], "", segments[1]
	}
	return "", "", segments[0]
}

func ConfigDescriptor() GraphDescriptor {
	return SupportedGraphTypes
}

func (descriptor GraphDescriptor) GetByType(name string) (GraphSchema, bool) {
	for _, schema := range descriptor {
		if schema.Type == name {
			return schema, true
		}
	}
	return GraphSchema{}, false
}

func DeriveNamespacedShardVS(namespace string, key string) string {
	// Read the value of the num_shards from the environment variable.
	var vsNum uint32
	shardVsSize := os.Getenv("SHARD_VS_SIZE")
	shardVsPrefix := os.Getenv("SHARD_VS_PREFIX")
	if shardVsPrefix == "" {
		shardVsPrefix = utils.DEFAULT_SHARD_VS_PREFIX
	}
	shardSize, ok := shardSizeMap[shardVsSize]
	if ok {
		vsNum = utils.Bkt(namespace, shardSize)
	} else {
		utils.AviLog.Warning.Printf("key: %s, msg: the value for shard_vs_size does not match the ENUM values", key)
		return ""
	}
	// Derive the right VS for this update.
	vsName := shardVsPrefix + fmt.Sprint(vsNum)
	return vsName
}
