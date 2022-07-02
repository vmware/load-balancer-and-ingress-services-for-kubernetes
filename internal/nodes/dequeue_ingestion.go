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

package nodes

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/status"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"k8s.io/apimachinery/pkg/api/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

func DequeueIngestion(key string, fullsync bool) {
	// The key format expected here is: objectType/Namespace/ObjKey
	// The assumption is that an update either affects an LB service type or an ingress. It cannot be both.
	var ingressFound, routeFound, mciFound bool
	var ingressNames, routeNames, mciNames []string
	utils.AviLog.Infof("key: %s, msg: starting graph Sync", key)
	sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)

	objType, namespace, name := lib.ExtractTypeNameNamespace(key)
	if objType == utils.Pod {
		handlePod(key, namespace, name, fullsync)
	}

	schema, valid := ConfigDescriptor().GetByType(objType)
	if valid {
		// If it's an ingress related change, let's process that.
		if utils.GetInformers().IngressInformer != nil && schema.GetParentIngresses != nil {
			ingressNames, ingressFound = schema.GetParentIngresses(name, namespace, key)
		} else if utils.GetInformers().RouteInformer != nil && schema.GetParentRoutes != nil {
			routeNames, routeFound = schema.GetParentRoutes(name, namespace, key)
		}
		// CHECKME: both ingress and mci processing?
		if utils.GetInformers().MultiClusterIngressInformer != nil && schema.GetParentMultiClusterIngresses != nil {
			mciNames, mciFound = schema.GetParentMultiClusterIngresses(name, namespace, key)
		}
	}

	if objType == lib.HostRule &&
		((utils.GetInformers().IngressInformer != nil && len(ingressNames) == 0) ||
			(utils.GetInformers().RouteInformer != nil && len(routeNames) == 0)) {
		// We should be checking for hostrule being possibly connected to a SharedVS
		handleHostRuleForSharedVS(key, fullsync)
	}

	// if we get update for object of type k8s node, create vrf graph
	// if in NodePort Mode we update pool servers
	if objType == utils.NodeObj {
		utils.AviLog.Debugf("key: %s, msg: processing node obj", key)
		processNodeObj(key, name, sharedQueue, fullsync)

		if lib.IsNodePortMode() && !fullsync {
			svcl4Keys, svcl7Keys := lib.GetSvcKeysForNodeCRUD()
			for _, svcl4Key := range svcl4Keys {
				handleL4Service(svcl4Key, fullsync)
			}
			for _, svcl7Key := range svcl7Keys {
				_, namespace, svcName := lib.ExtractTypeNameNamespace(svcl7Key)
				if ingressFound {
					filteredIngressFound, filteredIngressNames := objects.SharedSvcLister().IngressMappings(namespace).GetSvcToIng(svcName)
					if !filteredIngressFound {
						continue
					}
					handleIngress(svcl7Key, fullsync, filteredIngressNames)
				}
				if routeFound {
					filteredRouteFound, filteredRouteNames := objects.OshiftRouteSvcLister().IngressMappings(namespace).GetSvcToIng(svcName)
					if !filteredRouteFound {
						continue
					}
					handleRoute(svcl7Key, fullsync, filteredRouteNames)
				}
				if mciFound {
					filteredMCIFound, filteredMCINames := objects.SharedMultiClusterIngressSvcLister().MultiClusterIngressMappings(namespace).GetSvcToIng(svcName)
					if !filteredMCIFound {
						continue
					}
					handleMultiClusterIngress(svcl7Key, fullsync, filteredMCINames)
				}
			}
		}
		return
	}

	if objType == utils.Service {
		objects.SharedClusterIpLister().Save(namespace+"/"+name, name)
		found, _ := objects.SharedlbLister().Get(namespace + "/" + name)
		// This service is found in the LB list - this means it's a transition from LB to clusterIP or NodePort.
		if found {
			objects.SharedlbLister().Delete(namespace + "/" + name)
			utils.AviLog.Infof("key: %s, msg: service transitioned from type loadbalancer to ClusterIP or NodePort, will delete model", name)
			model_name := lib.GetModelName(lib.GetTenant(), lib.Encode(lib.GetNamePrefix()+namespace+"-"+name, lib.L4VS))
			objects.SharedAviGraphLister().Save(model_name, nil)
			if !fullsync {
				PublishKeyToRestLayer(model_name, key, sharedQueue)
			}
		}
	}

	if routeFound {
		handleRoute(key, fullsync, routeNames)
	}

	// Push Services from InfraSetting updates. Valid for annotation based approach.
	if objType == lib.AviInfraSetting && !lib.UseServicesAPI() {
		svcNames, svcFound := schema.GetParentServices(name, namespace, key)
		if svcFound && utils.CheckIfNamespaceAccepted(namespace) {
			for _, svcNSNameKey := range svcNames {
				handleL4Service(utils.L4LBService+"/"+svcNSNameKey, fullsync)
			}
		}
	}

	if !ingressFound && !lib.GetAdvancedL4() && !mciFound {
		// If ingress is not found, let's do the other checks.
		if objType == lib.SharedVipServiceKey {
			sharedVipKeys, keysFound := schema.GetParentServices(name, namespace, key)
			if keysFound && utils.CheckIfNamespaceAccepted(namespace) {
				for _, sharedVipKey := range sharedVipKeys {
					handleL4SharedVipService(sharedVipKey, key, fullsync)
				}
			}
		} else if objType == utils.L4LBService {
			// L4 type of services need special handling. We create a dedicated VS in Avi for these.
			handleL4Service(key, fullsync)
		} else if objType == utils.Endpoints {
			svcObj, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).Get(name)
			if err != nil {
				utils.AviLog.Debugf("key: %s, msg: there was an error in retrieving the service for endpoint", key)
				return
			}

			// Do not handle service update if it belongs to unaccepted namespace
			if svcObj.Spec.Type == utils.LoadBalancer && !lib.GetLayer7Only() && utils.CheckIfNamespaceAccepted(namespace) {
				// This endpoint update affects a LB service.
				aviModelGraph := NewAviObjectGraph()
				if sharedVipKey, ok := svcObj.Annotations[lib.SharedVipSvcLBAnnotation]; ok && sharedVipKey != "" {
					aviModelGraph.BuildAdvancedL4Graph(namespace, sharedVipKey, key, true)
				} else {
					aviModelGraph.BuildL4LBGraph(namespace, name, key)
				}
				if len(aviModelGraph.GetOrderedNodes()) > 0 {
					model_name := lib.GetModelName(lib.GetTenant(), aviModelGraph.GetAviVS()[0].Name)
					ok := saveAviModel(model_name, aviModelGraph, key)
					if ok && !fullsync {
						PublishKeyToRestLayer(model_name, key, sharedQueue)
					}
				}
			}
		}
	} else {
		if mciFound {
			handleMultiClusterIngress(key, fullsync, mciNames)
		}
		handleIngress(key, fullsync, ingressNames)
	}

	// handle the services APIs
	if (lib.GetAdvancedL4() && objType == utils.L4LBService) ||
		(lib.UseServicesAPI() && (objType == utils.Service || objType == utils.L4LBService)) ||
		((lib.GetAdvancedL4() || lib.UseServicesAPI()) && (objType == lib.Gateway || objType == lib.GatewayClass || objType == utils.Endpoints || objType == lib.AviInfraSetting)) {
		if !valid && objType == utils.L4LBService {
			// Required for advl4 schemas.
			schema, _ = ConfigDescriptor().GetByType(utils.Service)
		}

		gateways, gatewayFound := schema.GetParentGateways(name, namespace, key)
		// For each gateway first verify if it has a valid subscription to the GatewayClass or not.
		// If the gateway does not have a valid gatewayclass relationship, then set the model to nil.
		if gatewayFound {
			for _, gatewayKey := range gateways {
				// Check the gateway has a valid subscription or not. If not, delete it.
				namespace, _, gwName := lib.ExtractTypeNameNamespace(gatewayKey)
				modelName := lib.GetModelName(lib.GetTenant(), lib.Encode(lib.GetNamePrefix()+namespace+"-"+gwName, lib.ADVANCED_L4))
				if isGatewayDelete(gatewayKey, key) {
					// Check if a model corresponding to the gateway exists or not in memory.
					if found, _ := objects.SharedAviGraphLister().Get(modelName); found {
						objects.SharedAviGraphLister().Save(modelName, nil)
						if !fullsync {
							PublishKeyToRestLayer(modelName, key, sharedQueue)
						}
					}
				} else {
					aviModelGraph := NewAviObjectGraph()
					aviModelGraph.BuildAdvancedL4Graph(namespace, gwName, key, false)
					ok := saveAviModel(modelName, aviModelGraph, key)
					if ok && len(aviModelGraph.GetOrderedNodes()) != 0 && !fullsync {
						PublishKeyToRestLayer(modelName, key, sharedQueue)
					}
				}
			}
		}
	}
}

func handleHostRuleForSharedVS(key string, fullsync bool) {
	allModels := []string{}
	_, namespace, hrName := lib.ExtractTypeNameNamespace(key)
	sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
	var fqdn, oldFqdn string
	var fqdnType, oldFqdnType string
	var oldFound bool

	hostrule, err := lib.AKOControlConfig().CRDInformers().HostRuleInformer.Lister().HostRules(namespace).Get(hrName)
	if k8serrors.IsNotFound(err) {
		utils.AviLog.Debugf("key: %s, msg: HostRule Deleted", key)
		oldFound, oldFqdn = objects.SharedCRDLister().GetHostruleToFQDNMapping(namespace + "/" + hrName)
		if strings.Contains(oldFqdn, lib.ShardVSSubstring) {
			objects.SharedCRDLister().DeleteHostruleFQDNMapping(namespace + "/" + hrName)
			oldFqdnType = objects.SharedCRDLister().GetFQDNFQDNTypeMapping(oldFqdn)
			objects.SharedCRDLister().DeleteFQDNFQDNTypeMapping(oldFqdn)
		}
	} else if err != nil {
		utils.AviLog.Errorf("key: %s, msg: Error getting hostrule: %v", key, err)
		return
	} else {
		if hostrule.Status.Status == lib.StatusAccepted {
			fqdn = hostrule.Spec.VirtualHost.Fqdn
			oldFound, oldFqdn = objects.SharedCRDLister().GetHostruleToFQDNMapping(namespace + "/" + hrName)
			if oldFound && strings.Contains(oldFqdn, lib.ShardVSSubstring) {
				objects.SharedCRDLister().DeleteHostruleFQDNMapping(namespace + "/" + hrName)
				oldFqdnType = objects.SharedCRDLister().GetFQDNFQDNTypeMapping(oldFqdn)
			}
			if strings.Contains(fqdn, lib.ShardVSSubstring) {
				objects.SharedCRDLister().UpdateFQDNHostruleMapping(fqdn, namespace+"/"+hrName)
				fqdnType = string(hostrule.Spec.VirtualHost.FqdnType)
				if fqdnType == "" {
					fqdnType = string(akov1alpha1.Exact)
				}
				objects.SharedCRDLister().UpdateFQDNFQDNTypeMapping(fqdn, fqdnType)
			}
		}
	}

	if oldFound && strings.Contains(oldFqdn, lib.ShardVSSubstring) {
		if ok, obj := objects.SharedCRDLister().GetFQDNToSharedVSModelMapping(oldFqdn, oldFqdnType); !ok {
			utils.AviLog.Debugf("key: %s, msg: Couldn't find SharedVS model info for host: %s %s", key, oldFqdn, oldFqdnType)
		} else {
			allModels = append(allModels, obj...)
		}
	}

	if strings.Contains(fqdn, lib.ShardVSSubstring) {
		if ok, obj := objects.SharedCRDLister().GetFQDNToSharedVSModelMapping(fqdn, fqdnType); !ok {
			utils.AviLog.Debugf("key: %s, msg: Couldn't find SharedVS model info for host: %s %s", key, fqdn, fqdnType)
		} else {
			allModels = append(allModels, obj...)
		}
	}

	if len(allModels) == 0 {
		return
	}
	utils.AviLog.Infof("key: %s, msg: Models retrieved from HostRule %v", key, utils.Stringify(allModels))

	uniqueModelSet := make(map[string]bool)
	for _, modelName := range allModels {
		if _, ok := uniqueModelSet[modelName]; ok {
			continue
		}
		uniqueModelSet[modelName] = true
		// Try getting the SharedVS model, update with hostrule properties
		// and publish to the rest layer.
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found || aviModel == nil {
			utils.AviLog.Infof("key: %s, msg: model not found for %s", key, modelName)
		} else {
			aviModelObject := aviModel.(*AviObjectGraph)
			var vsNode AviVsEvhSniModel
			if lib.IsEvhEnabled() {
				if nodes := aviModelObject.GetAviEvhVS(); len(nodes) == 0 {
					continue
				} else {
					vsNode = nodes[0]
				}
			} else {
				if nodes := aviModelObject.GetAviVS(); len(nodes) == 0 {
					continue
				} else {
					vsNode = nodes[0]
				}
			}
			if found, fqdn := objects.SharedCRDLister().GetSharedVSModelFQDNMapping(modelName); found {
				BuildL7HostRule(fqdn, key, vsNode)
				ok := saveAviModel(modelName, aviModelObject, key)
				if ok && len(aviModelObject.GetOrderedNodes()) != 0 && !fullsync {
					PublishKeyToRestLayer(modelName, key, sharedQueue)
				}
			}
		}
	}
}

// handlePod populates NPL annotations for a pod in store.
// It also stores a mapping of Pod to Services for future use
func handlePod(key, namespace, podName string, fullsync bool) {
	utils.AviLog.Debugf("key: %s, msg: handing Pod", key)
	podKey := namespace + "/" + podName
	pod, err := utils.GetInformers().PodInformer.Lister().Pods(namespace).Get(podName)
	if err != nil {
		if !errors.IsNotFound(err) {
			utils.AviLog.Infof("key: %s, got error while getting pod: %v", key, err)
			return
		}

		utils.AviLog.Infof("key: %s, msg: Pod not found, deleting from SharedNPLLister", key)
		objects.SharedNPLLister().Delete(podKey)
		if found, lbSvcIntf := objects.SharedPodToLBSvcLister().Get(podKey); found {
			lbSvcs, ok := lbSvcIntf.([]string)
			if ok {
				//If namespace valid, do L4 service handling
				if utils.IsServiceNSValid(namespace) {
					utils.AviLog.Debugf("key: %s, msg: handling l4 Services %v", key, lbSvcs)
					for _, lbSvc := range lbSvcs {
						lbSvcKey := utils.L4LBService + "/" + lbSvc
						handleL4Service(lbSvcKey, fullsync)
					}
				}
			} else {
				utils.AviLog.Warnf("key: %s, msg: list services for pod is not of type []string", key)
			}
		}
		objects.SharedPodToLBSvcLister().Delete(podKey)
		return
	}
	ann := pod.GetAnnotations()
	var annotations []lib.NPLAnnotation
	if val, ok := ann[lib.NPLPodAnnotation]; ok {
		if err := json.Unmarshal([]byte(val), &annotations); err != nil {
			utils.AviLog.Infof("key: %s, got error while unmarshaling NPL annotations: %v", err)
		}
		objects.SharedNPLLister().Save(podKey, annotations)
		if utils.IsServiceNSValid(namespace) {
			services, lbSvcs := lib.GetServicesForPod(pod)
			if len(services) != 0 {
				objects.SharedPodToSvcLister().Save(podKey, services)
			}
			if len(lbSvcs) != 0 {
				objects.SharedPodToLBSvcLister().Save(podKey, lbSvcs)
			}
			for _, lbSvc := range lbSvcs {
				lbSvcKey := utils.L4LBService + "/" + lbSvc
				utils.AviLog.Debugf("key: %s, msg: handling l4 svc %s", key, lbSvcKey)
				handleL4Service(lbSvcKey, fullsync)
			}
			utils.AviLog.Infof("key: %s, msg: NPL Services retrieved: %s", key, services)
		}
	} else {
		utils.AviLog.Infof("key: %s, NPL annotation not found for Pod", key)
		objects.SharedNPLLister().Delete(podKey)
	}
}

func isGatewayDelete(gatewayKey, key string) bool {
	// parse the gateway name and namespace
	namespace, _, gwName := lib.ExtractTypeNameNamespace(gatewayKey)
	if lib.GetAdvancedL4() {
		gateway, err := lib.AKOControlConfig().AdvL4Informers().GatewayInformer.Lister().Gateways(namespace).Get(gwName)
		if err != nil && errors.IsNotFound(err) {
			return true
		}

		// check if deletiontimesttamp is present to see intended delete
		if gateway.GetDeletionTimestamp() != nil {
			utils.AviLog.Infof("key: %s, deletionTimestamp set on gateway, will be deleting VS", key)
			return true
		}

		// Check if the gateway has a valid gateway class
		err = validateGatewayForClass(key, gateway)
		if err != nil {
			return true
		}
	} else if lib.UseServicesAPI() {
		// If namespace is not accepted, return true to delete model
		if !utils.CheckIfNamespaceAccepted(namespace) {
			return true
		}

		gateway, err := lib.AKOControlConfig().SvcAPIInformers().GatewayInformer.Lister().Gateways(namespace).Get(gwName)
		if err != nil && errors.IsNotFound(err) {
			return true
		}

		// check if deletiontimestamp is present to see intended delete
		if gateway.GetDeletionTimestamp() != nil {
			utils.AviLog.Infof("key: %s, deletionTimestamp set on gateway, will be deleting VS", key)
			return true
		}

		// Check if the gateway has a valid gateway class
		err = validateSvcApiGatewayForClass(key, gateway)
		if err != nil {
			return true
		}
	}
	found, _ := objects.ServiceGWLister().GetGWListeners(namespace + "/" + gwName)
	if !found {
		return true
	}

	return false
}

func handleRoute(key string, fullsync bool, routeNames []string) {
	objType, namespace, _ := lib.ExtractTypeNameNamespace(key)
	sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
	utils.AviLog.Infof("key: %s, msg: route found: %v", key, routeNames)
	for _, route := range routeNames {
		nsroute, nameroute := getIngressNSNameForIngestion(objType, namespace, route)
		utils.AviLog.Infof("key: %s, msg: processing route: %s", key, route)
		HostNameShardAndPublish(utils.OshiftRoute, nameroute, nsroute, key, fullsync, sharedQueue)
	}
}

/*
to test
1. 	key1 svc1 svc2 ; key2 svc3
	change key from key1 to key2 in svc2
	key 1 svc1 ; key2 svc3 svc2

2.	key1 svc1
	change service type to clusterip
	deletes key1 VS
	change servie type to lb
	recreates key1 VS

3. 	key1 svc1	ingress /bar svc1
	change service type to clusterip
	deletes key1 VS, adds to pool in ingress /bar
	change service type to lb
	creates key1 VS, deletes pool of ingress /bar
*/
/*
validations
1.	annotations must not be on service of type non LB
2. 	port/protocol must be unique among all services with annotation key
3. 	preferred IP must be same in all services with annotation key
*/
func handleL4SharedVipService(namespacedVipKey, key string, fullsync bool) {
	if lib.GetLayer7Only() {
		// If the layer 7 only flag is set, then we shouldn't handling layer 4 VSes.
		utils.AviLog.Debugf("key: %s, msg: not handling service of type loadbalancer since AKO is configured to run in layer 7 mode only", key)
		return
	}

	sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
	_, namespace, _ := lib.ExtractTypeNameNamespace(key)
	modelName := lib.GetModelName(lib.GetTenant(), lib.Encode(lib.GetNamePrefix()+strings.ReplaceAll(namespacedVipKey, "/", "-"), lib.ADVANCED_L4))

	found, serviceNSNames := objects.SharedlbLister().GetSharedVipKeyToServices(namespacedVipKey)
	isShareVipKeyDelete := !found || len(serviceNSNames) == 0

	// Check whether all Services have the same preferred VIP setting. If not, delete the VS altogether,
	// assuming bad configuration.
	var sharedVipLBIP string
	for i, serviceNSName := range serviceNSNames {
		svcNSName := strings.Split(serviceNSName, "/")
		svcObj, err := utils.GetInformers().ServiceInformer.Lister().Services(svcNSName[0]).Get(svcNSName[1])
		if err != nil {
			utils.AviLog.Debugf("key: %s, msg: there was an error in retrieving the service", key)
			isShareVipKeyDelete = true
			break
		}

		// Initializing the preferred VIP from the first Service we get, so any other Service
		// that wishes for static IP allocation differently conflicts with this.
		if i == 0 {
			sharedVipLBIP = svcObj.Spec.LoadBalancerIP
		}

		if svcObj.Spec.LoadBalancerIP != sharedVipLBIP {
			utils.AviLog.Errorf("Service loadBalancerIP is not consistent with Services grouped using shared-vip annotation. Conflict found for Services [%s: %s %s: %s]", serviceNSName, svcObj.Spec.LoadBalancerIP, serviceNSNames[0], sharedVipLBIP)
			isShareVipKeyDelete = true
			break
		}
	}

	if isShareVipKeyDelete {
		// Check if a model corresponding to the gateway exists or not in memory.
		if found, _ := objects.SharedAviGraphLister().Get(modelName); found {
			objects.SharedAviGraphLister().Save(modelName, nil)
			if !fullsync {
				PublishKeyToRestLayer(modelName, key, sharedQueue)
			}
		}
	} else {
		aviModelGraph := NewAviObjectGraph()
		vipKey := strings.Split(namespacedVipKey, "/")[1]
		aviModelGraph.BuildAdvancedL4Graph(namespace, vipKey, key, true)
		ok := saveAviModel(modelName, aviModelGraph, key)
		if ok && len(aviModelGraph.GetOrderedNodes()) != 0 && !fullsync {
			PublishKeyToRestLayer(modelName, key, sharedQueue)
		}
	}
}

func handleL4Service(key string, fullsync bool) {
	if lib.GetLayer7Only() {
		// If the layer 7 only flag is set, then we shouldn't handling layer 4 VSes.
		utils.AviLog.Debugf("key: %s, msg: not handling service of type loadbalancer since AKO is configured to run in layer 7 mode only", key)
		return
	}
	_, namespace, name := lib.ExtractTypeNameNamespace(key)
	sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
	if deleteCase := isServiceDelete(name, namespace, key); !deleteCase && utils.CheckIfNamespaceAccepted(namespace) {
		// If Service is Not Annotated with NPL annotation, annotate the service and return.
		if lib.AutoAnnotateNPLSvc() {
			if !status.CheckNPLSvcAnnotation(key, namespace, name) {
				statusOption := status.StatusOptions{
					ObjType:   lib.NPLService,
					Op:        lib.UpdateStatus,
					ObjName:   name,
					Namespace: namespace,
					Key:       key,
				}
				utils.AviLog.Infof("key: %s Publishing to status queue, options: %v", name, utils.Stringify(statusOption))
				status.PublishToStatusQueue(name, statusOption)
				return
			}
		}
		utils.AviLog.Infof("key: %s, msg: service is of type loadbalancer. Will create dedicated VS nodes", key)
		aviModelGraph := NewAviObjectGraph()
		aviModelGraph.BuildL4LBGraph(namespace, name, key)

		// Save the LB service in memory
		objects.SharedlbLister().Save(namespace+"/"+name, name)
		if len(aviModelGraph.GetOrderedNodes()) > 0 {
			model_name := lib.GetModelName(lib.GetTenant(), aviModelGraph.GetAviVS()[0].Name)
			ok := saveAviModel(model_name, aviModelGraph, key)
			if ok && !fullsync {
				PublishKeyToRestLayer(model_name, key, sharedQueue)
			}
		}

		found, _ := objects.SharedClusterIpLister().Get(namespace + "/" + name)
		if found {
			// This is transition from clusterIP to service of type LB
			objects.SharedClusterIpLister().Delete(namespace + "/" + name)
			affectedIngs, _ := SvcToIng(name, namespace, key)
			for _, ingress := range affectedIngs {
				utils.AviLog.Infof("key: %s, msg: transition case from ClusterIP to service of type Loadbalancer: %s", key, ingress)
				HostNameShardAndPublish(utils.Ingress, ingress, namespace, key, fullsync, sharedQueue)
			}
		}
		return
	}
	// This is a DELETE event. The avi graph is set to nil.
	utils.AviLog.Debugf("key: %s, msg: received DELETE event for service", key)
	model_name := lib.GetModelName(lib.GetTenant(), lib.Encode(lib.GetNamePrefix()+namespace+"-"+name, lib.L4VS))
	objects.SharedAviGraphLister().Save(model_name, nil)
	if !fullsync {
		bkt := utils.Bkt(model_name, sharedQueue.NumWorkers)
		sharedQueue.Workqueue[bkt].AddRateLimited(model_name)
	}
}

func handleIngress(key string, fullsync bool, ingressNames []string) {
	objType, namespace, _ := lib.ExtractTypeNameNamespace(key)
	sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
	// The only other shard scheme we support now is hostname sharding.
	for _, ingress := range ingressNames {
		nsing, nameing := getIngressNSNameForIngestion(objType, namespace, ingress)
		utils.AviLog.Debugf("key: %s, msg: processing ingress: %s", key, ingress)
		HostNameShardAndPublish(utils.Ingress, nameing, nsing, key, fullsync, sharedQueue)
	}
}

func handleMultiClusterIngress(key string, fullsync bool, ingressNames []string) {
	objType, namespace, _ := lib.ExtractTypeNameNamespace(key)
	sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
	// The only other shard scheme we support now is hostname sharding.
	for _, ingress := range ingressNames {
		nsing, nameing := getIngressNSNameForIngestion(objType, namespace, ingress)
		utils.AviLog.Debugf("key: %s, msg: processing multi-cluster ingress: %s", key, ingress)
		HostNameShardAndPublish(lib.MultiClusterIngress, nameing, nsing, key, fullsync, sharedQueue)
	}
}

func getIngressNSNameForIngestion(objType, namespace, nsname string) (string, string) {
	if objType == lib.HostRule || objType == lib.HTTPRule || objType == utils.Secret {
		arr := strings.Split(nsname, "/")
		return arr[0], arr[1]
	}

	if objType == utils.IngressClass || objType == lib.AviInfraSetting {
		arr := strings.Split(nsname, "/")
		return arr[0], arr[1]
	}

	return namespace, nsname
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

func processNodeObj(key, nodename string, sharedQueue *utils.WorkerQueue, fullsync bool) {
	utils.AviLog.Debugf("key: %s, Got node Object %s", key, nodename)
	nodeObj, err := utils.GetInformers().NodeInformer.Lister().Get(nodename)
	if err == nil {
		utils.AviLog.Debugf("key: %s, Node Object %v", key, nodeObj)
		objects.SharedNodeLister().AddOrUpdate(nodename, nodeObj)
	} else if errors.IsNotFound(err) {
		utils.AviLog.Debugf("key: %s, msg: Node Deleted", key)
		objects.SharedNodeLister().Delete(nodename)
	} else {
		utils.AviLog.Errorf("key: %s, msg: Error getting node: %v", key, err)
		return
	}
	if lib.IsNodePortMode() {
		return
	}

	// Do not process VRF for non primary AKO
	isPrimaryAKO := lib.AKOControlConfig().GetAKOInstanceFlag()
	if !isPrimaryAKO {
		return
	}
	aviModel := NewAviObjectGraph()
	aviModel.IsVrf = true
	vrfcontext := lib.GetVrf()
	err = aviModel.BuildVRFGraph(key, vrfcontext)
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: Error creating vrf graph: %v", key, err)
		return
	}
	model_name := lib.GetModelName(lib.GetTenant(), vrfcontext)
	ok := saveAviModel(model_name, aviModel, key)
	if ok && !fullsync {
		PublishKeyToRestLayer(model_name, key, sharedQueue)
	}

}

func PublishKeyToRestLayer(modelName string, key string, sharedQueue *utils.WorkerQueue) {
	bkt := utils.Bkt(modelName, sharedQueue.NumWorkers)
	sharedQueue.Workqueue[bkt].AddRateLimited(modelName)
	utils.AviLog.Infof("key: %s, msg: Published key with modelName: %s", key, modelName)

}

func isServiceDelete(svcName string, namespace string, key string) bool {
	// If the service is not found we return true.
	svc, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).Get(svcName)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: could not retrieve the object for service: %s", key, err)
		if errors.IsNotFound(err) {
			return true
		}
	}

	// The annotation for sharedVip might have been added, in which case we should delete the L4
	// dedicated virtual service.
	if svc.Annotations[lib.SharedVipSvcLBAnnotation] != "" {
		return true
	}

	return false
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

func GetShardVSPrefix(key string) string {
	// sample prefix: clusterName--Shared-L7-
	shardVsPrefix := lib.GetNamePrefix() + lib.GetAKOIDPrefix() + lib.ShardVSPrefix + "-"
	utils.AviLog.Debugf("key: %s, msg: ShardVSPrefix: %s", key, shardVsPrefix)
	return shardVsPrefix
}

func GetShardVSName(s string, key string, shardSize uint32, prefix ...string) lib.VSNameMetadata {
	var vsNum uint32
	var vsNameMeta lib.VSNameMetadata
	extraPrefix := strings.Join(prefix, "-")

	if shardSize != 0 {
		vsNum = utils.Bkt(s, shardSize)
		utils.AviLog.Debugf("key: %s, msg: VS number: %v", key, vsNum)
	} else {
		utils.AviLog.Debugf("key: %s, msg: Processing dedicated VS", key)
		vsNameMeta.Dedicated = true
		//format: my-cluster--foo.com-dedicated for dedicated VS. This is to avoid any SNI naming conflicts
		if extraPrefix != "" {
			vsNameMeta.Name = lib.GetNamePrefix() + extraPrefix + "-" + s + lib.DedicatedSuffix

			return vsNameMeta
		}
		vsNameMeta.Name = lib.GetNamePrefix() + s + lib.DedicatedSuffix
		return vsNameMeta
	}
	vsNameMeta.Dedicated = false

	shardVsPrefix := GetShardVSPrefix(key)
	if extraPrefix != "" {
		shardVsPrefix += extraPrefix + "-"
	}
	vsName := shardVsPrefix + strconv.Itoa(int(vsNum))
	vsNameMeta.Name = vsName
	return vsNameMeta
}

// returns old and new models if changed, else just the current one.
func DeriveShardVS(hostname string, key string, routeIgrObj RouteIngressModel) (lib.VSNameMetadata, lib.VSNameMetadata) {
	utils.AviLog.Debugf("key: %s, msg: hostname for sharding: %s", key, hostname)
	var newInfraPrefix, oldInfraPrefix string
	oldShardSize, newShardSize := lib.GetshardSize(), lib.GetshardSize()

	// get stored infrasetting from ingress/route
	// figure out the current infrasetting via class/annotation
	var oldSettingName string
	var found bool
	if found, oldSettingName = objects.InfraSettingL7Lister().GetIngRouteToInfraSetting(routeIgrObj.GetNamespace() + "/" + routeIgrObj.GetName()); found {
		if found, shardSize := objects.InfraSettingL7Lister().GetInfraSettingToShardSize(oldSettingName); found && shardSize != "" {
			oldShardSize = lib.ShardSizeMap[shardSize]
		}
		oldInfraPrefix = oldSettingName
	} else {
		utils.AviLog.Debugf("AviInfraSetting %s not found in cache", oldSettingName)
	}

	newSetting := routeIgrObj.GetAviInfraSetting()
	if !routeIgrObj.Exists() {
		// get the old ones.
		newShardSize = oldShardSize
		newInfraPrefix = oldInfraPrefix
	} else if newSetting != nil {
		if newSetting.Spec.L7Settings != (akov1alpha1.AviInfraL7Settings{}) {
			newShardSize = lib.ShardSizeMap[newSetting.Spec.L7Settings.ShardSize]
		}
		newInfraPrefix = newSetting.Name
	}

	oldVsName, newVsName := GetShardVSName(hostname, key, oldShardSize, oldInfraPrefix), GetShardVSName(hostname, key, newShardSize, newInfraPrefix)
	utils.AviLog.Infof("key: %s, msg: ShardVSNames: %v %v", key, oldVsName, newVsName)
	return oldVsName, newVsName
}
