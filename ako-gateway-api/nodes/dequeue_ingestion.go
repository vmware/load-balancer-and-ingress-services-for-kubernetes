/*
 * Copyright 2023-2024 VMware, Inc.
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
	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	akogatewayapiobjects "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
)

func DequeueIngestion(key string, fullsync bool) {
	objType, namespace, name := lib.ExtractTypeNameNamespace(key)

	schema, valid := ConfigDescriptor().GetByType(objType)
	if !valid {
		return
	}
	gatewayList, found := schema.GetGateways(name, namespace, key)
	if !found {
		//returning due to error, cannot delete or update
		return
	}
	handleGateways(gatewayList, fullsync, key)
}

func handleGateways(gatewayList []string, fullsync bool, key string) {
	sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
	for _, gateway := range gatewayList {
		_, namespace, name := lib.ExtractTypeNameNamespace(gateway)

		modelName := lib.GetModelName(lib.GetTenant(), akogatewayapilib.GetGatewayParentName(namespace, name))
		modelFound, _ := objects.SharedAviGraphLister().Get(modelName)

		gatewayObj, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayInformer.Lister().Gateways(namespace).Get(name)
		if err != nil {
			if !errors.IsNotFound(err) {
				utils.AviLog.Infof("key: %s, got error while getting gateway class: %v", key, err)
				continue
			}
			if modelFound {
				objects.SharedAviGraphLister().Save(modelName, nil)
				if !fullsync {
					nodes.PublishKeyToRestLayer(modelName, key, sharedQueue)
				}
			}
			akogatewayapiobjects.GatewayApiLister().DeleteGatewayToGatewayClass(namespace, name)
			continue
		}
		gwClass := string(gatewayObj.Spec.GatewayClassName)
		found, isAkoCtrl := akogatewayapiobjects.GatewayApiLister().IsGatewayClassControllerAKO(gwClass)
		if !found {
			//gateway class deleted
			objects.SharedAviGraphLister().Save(modelName, nil)
			if !fullsync {
				nodes.PublishKeyToRestLayer(modelName, key, sharedQueue)
			}
			continue
		}
		if !isAkoCtrl {
			//AKO is not the controller, do not build model
			continue
		}
		aviModelGraph := NewAviObjectGraph()
		aviModelGraph.BuildGatewayVs(gatewayObj, key)
		if len(aviModelGraph.GetOrderedNodes()) > 0 {
			ok := saveAviModel(modelName, aviModelGraph, key)
			if ok && !fullsync {
				nodes.PublishKeyToRestLayer(modelName, key, sharedQueue)
			}
		}
	}
}

func saveAviModel(modelName string, aviGraph *AviObjectGraph, key string) bool {
	utils.AviLog.Debugf("key: %s, msg: Evaluating model :%s", key, modelName)
	if lib.DisableSync {
		// Note: This is not thread safe, however locking is expensive and the condition for locking should happen rarely
		utils.AviLog.Infof("key: %s, msg: Disable Sync is True, model %s can not be saved", key, modelName)
		return false
	}
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found && aviModel != nil {
		prevChecksum := aviModel.(*AviObjectGraph).GraphChecksum
		utils.AviLog.Debugf("key: %s, msg: the model: %s has a previous checksum: %v", key, modelName, prevChecksum)
		presentChecksum := aviGraph.GetCheckSum()
		utils.AviLog.Debugf("key: %s, msg: the model: %s has a present checksum: %v", key, modelName, presentChecksum)
		if prevChecksum == presentChecksum {
			utils.AviLog.Debugf("key: %s, msg: The model: %s has identical checksums, hence not processing. Checksum value: %v", key, modelName, presentChecksum)
			return false
		}
	}
	// Right before saving the model, let's reset the retry counter for the graph.
	aviGraph.SetRetryCounter()
	aviGraph.CalculateCheckSum()
	objects.SharedAviGraphLister().Save(modelName, aviGraph)
	return true
}
