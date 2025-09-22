/*
 * Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
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
	"context"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

	utils.AviLog.Infof("Key: %s, msg: objectType: %s, namespace: %s, name: %s", key, objType, namespace, name)
	schema, valid := ConfigDescriptor().GetByType(objType)
	if !valid {
		return
	}
	if objType == lib.HTTPRoute {
		httpRoute, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().HTTPRouteInformer.Lister().HTTPRoutes(namespace).Get(name)
		if err == nil {
			utils.AviLog.Debugf("key: %s, msg: Successfully retrieved the HTTPRoute object %s", key, name)
			if !IsHTTPRouteValid(key, httpRoute) {
				return
			}
		}
	}

	gatewayNsNameList, found := schema.GetGateways(namespace, name, key)
	if !found {
		//returning due to error, cannot delete or update
		utils.AviLog.Errorf("key: %s, msg: got error while getting k8s object", key)
		return
	}
	utils.AviLog.Infof("key: %s, msg: processing gateways %v", key, gatewayNsNameList)
	if objType == lib.Gateway {
		handleGateway(namespace, name, fullsync, key)
	}

	if objType == utils.Service {
		_, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).Get(name)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				objects.SharedClusterIpLister().Delete(namespace + "/" + name)
			} else {
				utils.AviLog.Errorf("key: %s, msg: got error while getting service object", key)
				return
			}
		}
		objects.SharedClusterIpLister().Save(namespace+"/"+name, name)
	}

	var routeTypeNsNameList []string
	if !(objType == lib.Gateway && fullsync) {
		routeTypeNsNameList, found = schema.GetRoutes(namespace, name, key)
		if !found {
			utils.AviLog.Infof("key: %s, msg: got error while getting object %s", key, objType)
			return
		}
	}
	utils.AviLog.Infof("key: %s, msg: processing gateways %v and routes %v", key, gatewayNsNameList, routeTypeNsNameList)
	for _, gatewayNsName := range gatewayNsNameList {

		parentNs, _, parentName := lib.ExtractTypeNameNamespace(gatewayNsName)
		tenant := objects.SharedNamespaceTenantLister().GetTenantInNamespace(gatewayNsName)
		if tenant == "" {
			tenant = lib.GetTenant()
		}
		modelName := lib.GetModelName(tenant, akogatewayapilib.GetGatewayParentName(parentNs, parentName))

		modelFound, modelIntf := objects.SharedAviGraphLister().Get(modelName)
		// Seq: GW first and the secret created.
		modelNil := !modelFound || modelIntf == nil
		if objType == utils.Secret {
			if modelNil {
				handleGateway(parentNs, parentName, fullsync, key)
				modelFound, modelIntf = objects.SharedAviGraphLister().Get(modelName)
				modelNil = !modelFound || modelIntf == nil
				if modelNil {
					utils.AviLog.Warnf("key: %s, msg: no model found: %s", key, modelName)
					continue
				}
				// Fetch routes for a gateway
				routeTypeNsNameList, found = GatewayToRoutes(parentNs, parentName, key)
				if !found {
					utils.AviLog.Errorf("key: %s, msg: got error while getting route objects for gateway %s/%s", key, parentNs, parentName)
					continue
				}
				utils.AviLog.Infof("key: %s, msg: Routes for gateway %s/%s are: %v", key, parentNs, parentName, utils.Stringify(routeTypeNsNameList))
			} else {
				model := &AviObjectGraph{modelIntf.(*nodes.AviObjectGraph)}
				vsToDelete := handleSecrets(parentNs, parentName, key, model)
				if vsToDelete {
					utils.AviLog.Warnf("key: %s, msg: No valid listener on Gateway %s/%s. Removing Parent VS from Controller", key, parentNs, parentName)
					objects.SharedAviGraphLister().Save(modelName, nil)
					if !fullsync {
						sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
						nodes.PublishKeyToRestLayer(modelName, key, sharedQueue)
						continue
					}
				}
			}
		}
		if modelNil {
			utils.AviLog.Warnf("key: %s, msg: no model found: %s", key, modelName)
			continue
		}

		model := &AviObjectGraph{modelIntf.(*nodes.AviObjectGraph)}
		utils.AviLog.Infof("key: %s, msg: processing routes %v", key, routeTypeNsNameList)
		for _, routeTypeNsName := range routeTypeNsNameList {
			objType, namespace, name := lib.ExtractTypeNameNamespace(routeTypeNsName)
			utils.AviLog.Infof("key: %s, msg: processing route %s mapped to gateway %s", key, routeTypeNsName, gatewayNsName)

			routeModel, err := NewRouteModel(key, objType, name, namespace)
			if err != nil {
				if k8serrors.IsNotFound(err) {
					utils.AviLog.Infof("key: %s, msg: deleting configurations corresponding to route %s", key, routeTypeNsName)
					model.ProcessRouteDeletion(key, gatewayNsName, routeModel, fullsync)
				}
				continue
			}

			childVSes := make(map[string]struct{}, 0)

			switch objType {
			case lib.HTTPRoute:
				model.ProcessL7Routes(key, routeModel, gatewayNsName, childVSes, fullsync)
			default:
				utils.AviLog.Warnf("key: %s, msg: route of type %s not supported", key, objType)
				continue
			}
			model.DeleteStaleChildVSes(key, routeModel, childVSes, fullsync)
		}
		if !akogatewayapilib.IsGatewayInDedicatedMode(parentNs, parentName) {
			model.AddDefaultHTTPPolicySet(key)
		}

		// Only add this node to the list of models if the checksum has changed.
		modelChanged := saveAviModel(modelName, model.AviObjectGraph, key)
		if modelChanged && !fullsync {
			sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
			nodes.PublishKeyToRestLayer(modelName, key, sharedQueue)
		}
	}
	utils.AviLog.Infof("key: %s, msg: finished graph Sync", key)
}
func handleSecrets(gatewayNamespace string, gatewayName string, key string, object *AviObjectGraph) bool {
	_, _, secretName := lib.ExtractTypeNameNamespace(key)
	utils.AviLog.Infof("key: %s, msg: Processing secret update %s has been added.", key, secretName)
	cs := utils.GetInformers().ClientSet
	gatewayObj, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayInformer.Lister().Gateways(gatewayNamespace).Get(gatewayName)
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: unable to get the gateway object. err: %s", key, err)
		return false
	}
	secretObj, err := cs.CoreV1().Secrets(gatewayNamespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil || secretObj == nil {
		utils.AviLog.Warnf("key: %s, msg: secret %s has been deleted, err: %s", key, secretName, err)
		vsToDelete := DeleteTLSNode(key, object, gatewayObj, secretObj)
		return vsToDelete
	} else {
		utils.AviLog.Infof("key: %s, msg: secret %s has been added.", key, secretName)
		AddTLSNode(key, object, gatewayObj, secretObj)
	}
	return false
}
func handleGateway(namespace, name string, fullsync bool, key string) {
	utils.AviLog.Debugf("key: %s, msg: processing gateway: %s", key, name)

	tenant := objects.SharedNamespaceTenantLister().GetTenantInNamespace(namespace + "/" + name)
	if tenant == "" {
		tenant = lib.GetTenant()
	}
	modelName := lib.GetModelName(tenant, akogatewayapilib.GetGatewayParentName(namespace, name))
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
		if !modelFound {
			// try to get model if it was dedicated mode since there is no way to find the annotation once gateway is deleted
			modelName = lib.GetModelName(tenant, lib.GetNamePrefix()+namespace+"-"+name+lib.DedicatedSuffix+"-EVH")
			modelFound, _ = objects.SharedAviGraphLister().Get(modelName)
		}
		if modelFound {
			// As gateway is not present, we need to remove mapping.
			gwNsName := namespace + "/" + name
			akogatewayapiobjects.GatewayApiLister().DeleteGatewayFromStore(gwNsName)
			objects.SharedAviGraphLister().Save(modelName, nil)
			objects.SharedNamespaceTenantLister().RemoveNamespaceToTenantCache(gwNsName)
			if !fullsync {
				sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
				nodes.PublishKeyToRestLayer(modelName, key, sharedQueue)
			}
		}
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

	// Reload the tenant to handle the change in tenant annotation in a Namespace
	tenant = objects.SharedNamespaceTenantLister().GetTenantInNamespace(namespace + "/" + name)
	if tenant == "" {
		tenant = lib.GetTenant()
	}
	modelName = lib.GetModelName(tenant, akogatewayapilib.GetGatewayParentName(namespace, name))
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

func (o *AviObjectGraph) ProcessRouteDeletion(key, parentNsName string, routeModel RouteModel, fullsync bool) {

	parentNode := o.GetAviEvhVS()
	routeTypeNsName := routeModel.GetType() + "/" + routeModel.GetNamespace() + "/" + routeModel.GetName()
	if parentNode[0].Dedicated {
		o.ProcessRouteDeletionForDedicatedMode(key, parentNsName, routeModel, fullsync)

	} else {
		found, childVSNames := akogatewayapiobjects.GatewayApiLister().GetRouteToChildVS(routeTypeNsName)
		if found {
			utils.AviLog.Infof("key: %s, msg: child VSes retrieved for deletion %v", key, childVSNames)

			for _, childVSName := range childVSNames {
				removed := nodes.RemoveEvhInModel(childVSName, parentNode, key)
				if removed {
					akogatewayapiobjects.GatewayApiLister().DeleteRouteChildVSMappings(routeTypeNsName, childVSName)
				}
			}
		}
	}
	updateHostname(key, parentNsName, parentNode[0])
	modelName := parentNode[0].Tenant + "/" + parentNode[0].Name
	ok := saveAviModel(modelName, o.AviObjectGraph, key)
	if ok && len(o.AviObjectGraph.GetOrderedNodes()) != 0 && !fullsync {
		sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
		nodes.PublishKeyToRestLayer(modelName, key, sharedQueue)
	}
}

func (o *AviObjectGraph) ProcessRouteDeletionForDedicatedMode(key, parentNsName string, routeModel RouteModel, fullsync bool) {
	utils.AviLog.Infof("key: %s, msg: Processing route deletion for dedicated mode: %s/%s", key, routeModel.GetNamespace(), routeModel.GetName())

	gatewayVSes := o.GetAviEvhVS()
	if len(gatewayVSes) == 0 {
		utils.AviLog.Errorf("key: %s, msg: No Gateway VS found for dedicated mode deletion", key)
		return
	}

	dedicatedVS := gatewayVSes[0]

	httpPSName := akogatewayapilib.GetHttpPolicySetName(dedicatedVS.AviMarkers.GatewayNamespace, dedicatedVS.AviMarkers.GatewayName, routeModel.GetNamespace(), routeModel.GetName())

	var updatedHttpPolicyRefs []*nodes.AviHttpPolicySetNode
	for _, policy := range dedicatedVS.HttpPolicyRefs {
		if policy.Name != httpPSName {
			updatedHttpPolicyRefs = append(updatedHttpPolicyRefs, policy)
		} else {
			utils.AviLog.Infof("key: %s, msg: Removing HTTP PolicySet %s for route deletion", key, httpPSName)
		}
	}
	dedicatedVS.HttpPolicyRefs = updatedHttpPolicyRefs

	dedicatedVS.PoolGroupRefs = []*nodes.AviPoolGroupNode{}
	dedicatedVS.PoolRefs = []*nodes.AviPoolNode{}

	if dedicatedVS.ServiceMetadata.HTTPRoute == routeModel.GetNamespace()+"/"+routeModel.GetName() {
		dedicatedVS.ServiceMetadata.HTTPRoute = ""
		dedicatedVS.AviMarkers.HTTPRouteName = ""
		dedicatedVS.AviMarkers.HTTPRouteNamespace = ""
	}

	utils.AviLog.Infof("key: %s, msg: Completed route deletion for dedicated mode: %s/%s", key, routeModel.GetNamespace(), routeModel.GetName())
}

func (o *AviObjectGraph) DeleteStaleChildVSes(key string, routeModel RouteModel, childVSes map[string]struct{}, fullsync bool) {

	parentNode := o.GetAviEvhVS()

	_, storedChildVSes := akogatewayapiobjects.GatewayApiLister().GetRouteToChildVS(routeModel.GetType() + "/" + routeModel.GetNamespace() + "/" + routeModel.GetName())

	for _, childVSName := range storedChildVSes {
		if _, ok := childVSes[childVSName]; !ok {
			utils.AviLog.Infof("key: %s, msg: child VS retrieved for deletion %v", key, childVSName)
			removed := nodes.RemoveEvhInModel(childVSName, parentNode, key)
			if removed {
				akogatewayapiobjects.GatewayApiLister().DeleteRouteChildVSMappings(routeModel.GetType()+"/"+routeModel.GetNamespace()+"/"+routeModel.GetName(), childVSName)
			}
		}
	}
}
