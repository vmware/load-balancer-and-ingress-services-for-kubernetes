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
	"errors"
	"strings"

	"github.com/avinetworks/ako/internal/lib"
	"github.com/avinetworks/ako/internal/objects"
	"github.com/avinetworks/ako/pkg/utils"

	advl4v1alpha1pre1 "github.com/vmware-tanzu/service-apis/apis/v1alpha1pre1"
)

func (o *AviObjectGraph) BuildAdvancedL4Graph(namespace string, gatewayName string, key string) {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	var VsNode *AviVsNode
	// TODO: work around gateway object and fetch services
	gateway, err := lib.GetAdvL4Informers().GatewayInformer.Lister().Gateways(namespace).Get(gatewayName)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: error in obtaining networking.x-k8s.io/gateway: %s", key, gatewayName)
		return
	}

	found, services := objects.ServiceGWLister().GetGwToSvcs(namespace + "/" + gatewayName)
	if !found {
		utils.AviLog.Debugf("key: %s, msg: No services mapped to gateway %s/%s", namespace, gatewayName)
		return
	}
	utils.AviLog.Infof("key: %s, msg: Found Services %v for Gateway %s/%s", key, services, gateway.Namespace, gateway.Name)

	var gwSvcNSName string
	for key, _ := range services {
		gwSvcNSName = key
	}
	gwSvcNSNameArr := strings.Split(gwSvcNSName, "/")
	svcNamespace, svcName := gwSvcNSNameArr[0], gwSvcNSNameArr[1]

	svcObj, err := utils.GetInformers().ServiceInformer.Lister().Services(svcNamespace).Get(svcName)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: error in obtaining the object for service %s/%s", key, svcNamespace, svcName)
		return
	}

	// for gatewayName, fetch all valid services for gw.listener
	VsNode = o.ConstructAviL4VsNode(svcObj, key)
	utils.AviLog.Infof("key: %s, msg: Created VSNode with gateway %v", key, utils.Stringify(VsNode))
	o.ConstructAviL4PolPoolNodes(svcObj, VsNode, key)
	o.AddModelNode(VsNode)
	VsNode.CalculateCheckSum()
	o.GraphChecksum = o.GraphChecksum + VsNode.GetCheckSum()
	utils.AviLog.Infof("key: %s, msg: checksum  for AVI VS object %v", key, VsNode.GetCheckSum())
	utils.AviLog.Infof("key: %s, msg: computed Graph checksum for VS is: %v", key, o.GraphChecksum)
}

func validateGatewayObj(key string, gateway *advl4v1alpha1pre1.Gateway) error {
	gwClassObj, err := lib.GetAdvL4Informers().GatewayClassInformer.Lister().Get(gateway.Spec.Class)
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: Unable to fetch corresponding networking.x-k8s.io/gatewayclass %s %v",
			key, gateway.Spec.Class, err)
		return err
	}
	// Additional check to see if the gatewayclass is a valid avi gateway class or not.
	if gwClassObj.Spec.Controller != lib.AviGatewayController {
		// Return an error since this is not our object.
		return errors.New("Unexpected controller")
	}
	return nil
}
