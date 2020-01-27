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

const NodeObj = "Node"
const GlobalVRF = "global"

func DequeueIngestion(key string, fullsync bool) {
	// The key format expected here is: objectType/Namespace/ObjKey
	// The assumption is that an update either affects an LB service type or an ingress. It cannot be both.
	ingressFound := false
	var ingressNames []string
	utils.AviLog.Info.Printf("key: %s, msg: starting graph Sync", key)
	sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)

	objType, namespace, name := extractTypeNameNamespace(key)

	// if we get update for object of tyoe k8s node, create vrf graph
	if objType == utils.NodeObj {
		utils.AviLog.Info.Printf("key: %s, msg: processing node obj", key)
		processNodeObj(key, name, sharedQueue)
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
				if len(aviModelGraph.GetOrderedNodes()) != 0 && !fullsync {
					model_name := utils.ADMIN_NS + "/" + aviModelGraph.GetAviVS()[0].Name
					PublishKeyToRestLayer(aviModelGraph, model_name, key, sharedQueue, fullsync)
				}
			} else {
				// This is a DELETE event. The avi graph is set to nil.
				utils.AviLog.Info.Printf("key: %s, msg: received DELETE event for service", key)
				model_name := utils.ADMIN_NS + "/" + name + "--" + namespace
				objects.SharedAviGraphLister().Save(model_name, nil)
				if !fullsync {
					bkt := utils.Bkt(model_name, sharedQueue.NumWorkers)
					sharedQueue.Workqueue[bkt].AddRateLimited(model_name)
				}
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
				if len(aviModelGraph.GetOrderedNodes()) != 0 && !fullsync {
					model_name := utils.ADMIN_NS + "/" + aviModelGraph.GetAviVS()[0].Name
					PublishKeyToRestLayer(aviModelGraph, model_name, key, sharedQueue, fullsync)
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
			if len(aviModel.(*AviObjectGraph).GetOrderedNodes()) != 0 && !fullsync {
				PublishKeyToRestLayer(aviModel.(*AviObjectGraph), model_name, key, sharedQueue, fullsync)
			}
		}
	}
}

func processNodeObj(key, nodename string, sharedQueue *utils.WorkerQueue) {
	utils.AviLog.Info.Printf("key: %s, Got node Object %s\n", key, nodename)
	nodeObj, err := utils.GetInformers().NodeInformer.Lister().Get(nodename)
	if err == nil {
		utils.AviLog.Trace.Printf("key: %s, Node Object %v\n", key, nodeObj)
		objects.SharedNodeLister().AddOrUpdate(nodename, nodeObj)
	} else if errors.IsNotFound(err) {
		utils.AviLog.Info.Printf("key: %s, msg: Node Deleted\n", key)
		objects.SharedNodeLister().Delete(nodename)
	} else {
		utils.AviLog.Error.Printf("key: %s, msg: Error getting node: %v\n", key, err)
		return
	}
	aviModel := NewAviObjectGraph()
	aviModel.IsVrf = true
	vrfcontext := os.Getenv(utils.VRF_CONTEXT)
	if vrfcontext == "" {
		vrfcontext = utils.GlobalVRF
	}
	err = aviModel.BuildVRFGraph(key, vrfcontext)
	if err != nil {
		utils.AviLog.Error.Printf("key: %s, msg: Error creating vrf graph: %v\n", key, err)
		return
	}
	model_name := utils.ADMIN_NS + "/" + vrfcontext
	PublishKeyToRestLayer(aviModel, model_name, key, sharedQueue, false)
}

func PublishKeyToRestLayer(aviGraph *AviObjectGraph, model_name string, key string, sharedQueue *utils.WorkerQueue, fullsync bool) {
	// First see if there's another instance of the same model in the store
	utils.AviLog.Info.Printf("key: %s, msg: Evaluating model :%s", model_name)
	found, aviModel := objects.SharedAviGraphLister().Get(model_name)
	if found && aviModel != nil {
		prevChecksum := aviModel.(*AviObjectGraph).GraphChecksum

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
	if !fullsync {
		bkt := utils.Bkt(model_name, sharedQueue.NumWorkers)
		sharedQueue.Workqueue[bkt].AddRateLimited(model_name)
		utils.AviLog.Info.Printf("key: %s, msg: Published key with model_name: %s", key, model_name)
	}
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
