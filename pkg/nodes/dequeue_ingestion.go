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
	"strings"

	"gitlab.eng.vmware.com/orion/akc/pkg/objects"
	"gitlab.eng.vmware.com/orion/container-lib/utils"
)

func DequeueIngestion(key string) {
	// The key format expected here is: objectType/Namespace/ObjKey
	utils.AviLog.Info.Printf("%s: Starting graph Sync", key)
	objType, namespace, name := extractTypeNameNamespace(key)
	sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
	if objType == "Service" {
		// Check if this Service is of type Loadbalancer.
		if isServiceLBType(name, namespace) {
			utils.AviLog.Warning.Printf("%s service is of type Loadbalancer. Will create dedicated VS nodes", name)
			aviModelGraph := NewAviObjectGraph()
			aviModelGraph.BuildL4LBGraph(namespace, name)
			if len(aviModelGraph.GetOrderedNodes()) != 0 {
				publishKeyToRestLayer(aviModelGraph, namespace, name, sharedQueue)
			}
		}
	}

}

func publishKeyToRestLayer(aviGraph *AviObjectGraph, namespace string, name string, sharedQueue *utils.WorkerQueue) {
	// First see if there's another instance of the same model in the store
	model_name := namespace + "/" + name
	found, aviModel := objects.SharedAviGraphLister().Get(model_name)
	if found && aviModel != nil {
		prevChecksum := aviModel.(*AviObjectGraph).GetCheckSum()
		utils.AviLog.Info.Printf("The model: %s has a previous checksum: %v", model_name, prevChecksum)
		presentChecksum := aviGraph.GetCheckSum()
		utils.AviLog.Info.Printf("The model: %s has a present checksum: %v", model_name, presentChecksum)
		if prevChecksum == presentChecksum {
			utils.AviLog.Info.Printf("The model: %s has identical checksums, hence not processing. Checksum value: %v", model_name, presentChecksum)
			return
		}
	}
	// TODO (sudswas): Lots of checksum optimization goes here
	objects.SharedAviGraphLister().Save(model_name, aviGraph)
	bkt := utils.Bkt(model_name, sharedQueue.NumWorkers)
	sharedQueue.Workqueue[bkt].AddRateLimited(model_name)
}

func BuildAviGraph(gws []string) {
	return
}

func isServiceLBType(svcName string, namespace string) bool {
	svcObj, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).Get(svcName)
	if err != nil {
		utils.AviLog.Warning.Printf("Could not retrieve the object for service: %s", svcName)
		return false
	}
	if svcObj.Spec.Type == "LoadBalancer" {
		return true
	}
	return false
}

func extractTypeNameNamespace(key string) (string, string, string) {
	segments := strings.Split(key, "/")
	if len(segments) == 3 {
		return segments[0], segments[1], segments[2]
	}
	return "", "", segments[0]
}
