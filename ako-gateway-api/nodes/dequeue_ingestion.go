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
	utils.AviLog.Infof("key: %s, msg: starting graph Sync", key)
	objType, namespace, name := lib.ExtractTypeNameNamespace(key)

	schema, valid := ConfigDescriptor().GetByType(objType)
	if !valid {
		return
	}
	gatewayList, found := schema.GetGateways(namespace, name, key)
	if !found {
		//returning due to error, cannot delete or update
		utils.AviLog.Errorf("key: %s, got error while getting k8s object", key)
		return
	}
	handleGateways(gatewayList, fullsync, key)
}

func handleGateways(gatewayList []string, fullsync bool, key string) {
	sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
	for _, gateway := range gatewayList {
		utils.AviLog.Debugf("key: %s, msg: processing gateway: %s", key, gateway)
		namespace, _, name := lib.ExtractTypeNameNamespace(gateway)

		modelName := lib.GetModelName(lib.GetTenant(), akogatewayapilib.GetGatewayParentName(namespace, name))
		modelFound, _ := objects.SharedAviGraphLister().Get(modelName)
		if modelFound {
			utils.AviLog.Debugf("key: %s, msg: found model: %s", key, modelName)
		} else {
			utils.AviLog.Debugf("key: %s, msg: no model found: %s", key, modelName)
		}

		gatewayObj, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayInformer.Lister().Gateways(namespace).Get(name)
		if err != nil {
			if !errors.IsNotFound(err) {
				utils.AviLog.Infof("key: %s, got error while getting gateway class: %v", key, err)
				continue
			}
			utils.AviLog.Debugf("key: %s, msg: gateway not found: %s/%s", key, namespace, name)
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
		utils.AviLog.Debugf("key: %s, msg: fetching gateway class %s for gateway: %s/%s", key, gwClass, namespace, name)
		found, isAkoCtrl := akogatewayapiobjects.GatewayApiLister().IsGatewayClassControllerAKO(gwClass)
		if !found {
			//gateway class deleted
			utils.AviLog.Debugf("key: %s, msg: gateway class not found: %s", key, gwClass)
			objects.SharedAviGraphLister().Save(modelName, nil)
			if !fullsync {
				nodes.PublishKeyToRestLayer(modelName, key, sharedQueue)
			}
			continue
		}
		utils.AviLog.Debugf("key: %s, msg: fetching gateway class found: %s", key, gwClass)
		if !isAkoCtrl {
			//AKO is not the controller, do not build model
			utils.AviLog.Infof("key: %s, msg: Controller is not AKO for %s, not building VS model", key, modelName)
			continue
		}
		aviModelGraph := NewAviObjectGraph()
		aviModelGraph.BuildGatewayVs(gatewayObj, key)
		aviModelGraph.CalculateCheckSum()

		//extracting the embedded graph after vs is built
		aviModel := aviModelGraph.AviObjectGraph
		if len(aviModel.GetOrderedNodes()) > 0 {
			ok := saveAviModel(modelName, aviModel, key)
			if ok && !fullsync {
				utils.AviLog.Infof("key: %s, msg: Published key with modelName: %s", key, modelName)
				nodes.PublishKeyToRestLayer(modelName, key, sharedQueue)
			}
		}
	}
}

func saveAviModel(modelName string, aviGraph *nodes.AviObjectGraph, key string) bool {
	utils.AviLog.Debugf("key: %s, msg: Evaluating model :%s", key, modelName)
	if lib.DisableSync {
		// Note: This is not thread safe, however locking is expensive and the condition for locking should happen rarely
		utils.AviLog.Infof("key: %s, msg: Disable Sync is True, model %s can not be saved", key, modelName)
		return false
	}
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found && aviModel != nil {
		prevChecksum := aviModel.(*nodes.AviObjectGraph).GraphChecksum
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
