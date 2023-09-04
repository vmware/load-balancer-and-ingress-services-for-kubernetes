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
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	akogatewayapiobjects "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func DequeueIngestion(key string, fullsync bool) {
	utils.AviLog.Infof("key: %s, msg: starting graph Sync", key)
	objType, namespace, name := lib.ExtractTypeNameNamespace(key)

	schema, valid := ConfigDescriptor().GetByType(objType)
	if !valid {
		return
	}

	gatewayNsNameList, found := schema.GetGateways(namespace, name, key)
	if !found {
		//returning due to error, cannot delete or update
		utils.AviLog.Errorf("key: %s, msg: got error while getting k8s object", key)
		return
	}

	if updatesParent(objType) {
		handleGateway(namespace, name, fullsync, key)
	}

	routeTypeNsNameList, found := schema.GetRoutes(namespace, name, key)
	if !found {
		utils.AviLog.Errorf("key: %s, msg: got error while getting object", key, objType)
		return
	}

	utils.AviLog.Debugf("key: %s, msg: processing gateways %v and routes %v", key, gatewayNsNameList, routeTypeNsNameList)
	for _, gatewayNsName := range gatewayNsNameList {

		parentNs, _, parentName := lib.ExtractTypeNameNamespace(gatewayNsName)
		modelName := lib.GetModelName(lib.GetTenant(), akogatewayapilib.GetGatewayParentName(parentNs, parentName))

		modelFound, modelIntf := objects.SharedAviGraphLister().Get(modelName)
		if !modelFound || modelIntf == nil {
			utils.AviLog.Errorf("key: %s, msg: no model found: %s", key, modelName)
			continue
		}
		model := &AviObjectGraph{modelIntf.(*nodes.AviObjectGraph)}
		for _, routeTypeNsName := range routeTypeNsNameList {
			objType, namespace, name := lib.ExtractTypeNameNamespace(routeTypeNsName)
			utils.AviLog.Infof("key: %s, msg: processing route %s mapped to gateway %s", key, routeTypeNsName, gatewayNsName)

			routeModel, err := NewRouteModel(key, objType, name, namespace)
			if err != nil {
				if k8serrors.IsNotFound(err) {
					utils.AviLog.Infof("key: %s, msg: deleting configurations corresponding to route %s", key, routeTypeNsName)
					model.ProcessRouteDeletion(key, routeModel, fullsync)
				}
				continue
			}

			childVSes := make(map[string]struct{}, 0)

			switch objType {
			case lib.HTTPRoute:
				model.ProcessL7Routes(key, routeModel, gatewayNsName, childVSes)
			default:
				utils.AviLog.Warnf("key: %s, msg: route of type %s not supported", key, objType)
				continue
			}
			model.DeleteStaleChildVSes(key, routeModel, childVSes, fullsync)
		}

		// Only add this node to the list of models if the checksum has changed.
		modelChanged := saveAviModel(modelName, model.AviObjectGraph, key)
		if modelChanged && !fullsync {
			sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
			nodes.PublishKeyToRestLayer(modelName, key, sharedQueue)
		}
	}
}

func handleGateway(namespace, name string, fullsync bool, key string) {
	utils.AviLog.Debugf("key: %s, msg: processing gateway: %s", key, name)

	modelName := lib.GetModelName(lib.GetTenant(), akogatewayapilib.GetGatewayParentName(namespace, name))
	modelFound, _ := objects.SharedAviGraphLister().Get(modelName)
	if modelFound {
		utils.AviLog.Debugf("key: %s, msg: found model: %s", key, modelName)
	} else {
		utils.AviLog.Debugf("key: %s, msg: no model found: %s", key, modelName)
	}

	gatewayObj, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayInformer.Lister().Gateways(namespace).Get(name)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			utils.AviLog.Infof("key: %s, msg: got error while getting gateway class: %v", key, err)
			return
		}
		utils.AviLog.Debugf("key: %s, msg: gateway not found: %s/%s", key, namespace, name)
		if modelFound {
			objects.SharedAviGraphLister().Save(modelName, nil)
			if !fullsync {
				sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
				nodes.PublishKeyToRestLayer(modelName, key, sharedQueue)
			}
		}
		akogatewayapiobjects.GatewayApiLister().DeleteGatewayToGatewayClass(namespace, name)
		return
	}
	gwClass := string(gatewayObj.Spec.GatewayClassName)
	utils.AviLog.Debugf("key: %s, msg: fetching gateway class %s for gateway: %s/%s", key, gwClass, namespace, name)
	found, isAkoCtrl := akogatewayapiobjects.GatewayApiLister().IsGatewayClassControllerAKO(gwClass)
	if !found {
		//gateway class deleted
		utils.AviLog.Debugf("key: %s, msg: gateway class not found: %s", key, gwClass)
		objects.SharedAviGraphLister().Save(modelName, nil)
		if !fullsync {
			sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
			nodes.PublishKeyToRestLayer(modelName, key, sharedQueue)
		}
		return
	}
	utils.AviLog.Debugf("key: %s, msg: fetching gateway class found: %s", key, gwClass)
	if !isAkoCtrl {
		//AKO is not the controller, do not build model
		utils.AviLog.Infof("key: %s, msg: Controller is not AKO for %s, not building VS model", key, modelName)
		return
	}
	aviModelGraph := NewAviObjectGraph()
	aviModelGraph.BuildGatewayVs(gatewayObj, key)

	modelChanged := saveAviModel(modelName, aviModelGraph.AviObjectGraph, key)
	if modelChanged && !fullsync {
		sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
		nodes.PublishKeyToRestLayer(modelName, key, sharedQueue)
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

func (o *AviObjectGraph) ProcessRouteDeletion(key string, routeModel RouteModel, fullsync bool) {

	parentNode := o.GetAviEvhVS()

	found, childVSNames := akogatewayapiobjects.GatewayApiLister().GetRouteToChildVS(routeModel.GetType() + "/" + routeModel.GetNamespace() + "/" + routeModel.GetName())
	if !found {
		utils.AviLog.Warnf("key: %s, msg: no child vs mapped to this route %s", key, routeModel.GetName())
		return
	}
	utils.AviLog.Infof("key: %s, msg: child VSes retrieved for deletion %v", key, childVSNames)

	for _, childVSName := range childVSNames {
		nodes.RemoveEvhInModel(childVSName, parentNode, key)
	}
	akogatewayapiobjects.GatewayApiLister().DeleteRouteToChildVS(routeModel.GetType() + "/" + routeModel.GetNamespace() + "/" + routeModel.GetName())

	modelName := lib.GetTenant() + "/" + parentNode[0].Name
	ok := saveAviModel(modelName, o.AviObjectGraph, key)
	if ok && len(o.AviObjectGraph.GetOrderedNodes()) != 0 && !fullsync {
		sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
		nodes.PublishKeyToRestLayer(modelName, key, sharedQueue)
	}
}

func (o *AviObjectGraph) DeleteStaleChildVSes(key string, routeModel RouteModel, childVSes map[string]struct{}, fullsync bool) {

	parentNode := o.GetAviEvhVS()

	_, storedChildVSes := akogatewayapiobjects.GatewayApiLister().GetRouteToChildVS(routeModel.GetType() + "/" + routeModel.GetNamespace() + "/" + routeModel.GetName())

	var childRemoved bool
	for _, childVSName := range storedChildVSes {
		if _, ok := childVSes[childVSName]; !ok {
			childRemoved = true
			utils.AviLog.Infof("key: %s, msg: child VS retrieved for deletion %v", key, childVSName)
			nodes.RemoveEvhInModel(childVSName, parentNode, key)
			akogatewayapiobjects.GatewayApiLister().DeleteRouteChildVSMappings(routeModel.GetType()+"/"+routeModel.GetNamespace()+"/"+routeModel.GetName(), childVSName)
		}
	}

	if childRemoved {
		modelName := lib.GetTenant() + "/" + parentNode[0].Name
		_ = saveAviModel(modelName, o.AviObjectGraph, key)
	}
}

func updatesParent(objType string) bool {
	return objType == lib.Gateway || objType == utils.Secret
}
