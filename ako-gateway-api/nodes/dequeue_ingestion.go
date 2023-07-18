package nodes

import (
	akogatewayapik8s "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/k8s"
	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	akogatewayapiobjects "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
)

func DequeueIngestion(key string, fullsync bool) {
	sharedQueue := utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer)
	objType, namespace, name := lib.ExtractTypeNameNamespace(key)
	akogatewayapilib.AKOControlConfig()

	//Gateway Class
	if objType == lib.GatewayClass {
		gwClass, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayClassInformer.Lister().Get(name)
		if err != nil {
			if !errors.IsNotFound(err) {
				utils.AviLog.Infof("key: %s, got error while getting gateway class: %v", key, err)
				return
			}
			//Delete case
			akogatewayapiobjects.GatewayApiLister().DeleteGatewayClass(name)
			//TODO: optimize with object store
			akogatewayapik8s.SharedGatewayController().FullSyncK8s(false)
			return
		}
		//TODO trigger update for associated gateways
		if akogatewayapik8s.CheckGatewayClassController(*gwClass) {
			//Create/Update case
			akogatewayapiobjects.GatewayApiLister().UpdateGatewayClass(gwClass.Name)
		} else {
			//Delete case, GatewayClass controller changed
			akogatewayapiobjects.GatewayApiLister().DeleteGatewayClass(gwClass.Name)
		}
		//TODO: optimize with object store
		akogatewayapik8s.SharedGatewayController().FullSyncK8s(false)
	}

	// Gateway
	if objType == lib.Gateway {
		// Gateway Class is validated in ingestion
		gw, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayInformer.Lister().Gateways(namespace).Get(name)
		modelName := lib.GetModelName(lib.GetTenant(), akogatewayapilib.GetGatewayParentName(gw.Namespace, gw.Name))
		if err != nil {
			if !errors.IsNotFound(err) {
				utils.AviLog.Infof("key: %s, got error while getting gateway class: %v", key, err)
				return
			}
			//Delete case
			akogatewayapiobjects.GatewayApiLister().DeleteGatewayToGatewayClass(namespace, name)
			if found, _ := objects.SharedAviGraphLister().Get(modelName); found {
				objects.SharedAviGraphLister().Save(modelName, nil)
				if !fullsync {
					nodes.PublishKeyToRestLayer(modelName, key, sharedQueue)
				}
			}
			return
		}
		akogatewayapiobjects.GatewayApiLister().UpdateGatewayToGatewayClass(namespace, name, string(gw.Spec.GatewayClassName))
		aviModelGraph := NewAviObjectGraph()
		aviModelGraph.BuildGatewayVs(gw, key)
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
