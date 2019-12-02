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
	"gitlab.eng.vmware.com/orion/container-lib/utils"
	corev1 "k8s.io/api/core/v1"
)

const (
	HTTP            = "HTTP"
	HeaderMethod    = ":method"
	HeaderAuthority = ":authority"
	HeaderScheme    = ":scheme"
	TLS             = "TLS"
	HTTPS           = "HTTPS"
	TCP             = "TCP"
)

func contains(s []int32, e int32) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (o *AviObjectGraph) ConstructAviL4VsNode(svcObj *corev1.Service) *AviVsNode {
	var avi_vs_meta *AviVsNode

	// FQDN should come from the cloud. Modify
	avi_vs_meta = &AviVsNode{Name: svcObj.ObjectMeta.Name, Tenant: svcObj.ObjectMeta.Namespace,
		EastWest: false}
	var portProtocols []AviPortHostProtocol
	for _, port := range svcObj.Spec.Ports {
		pp := AviPortHostProtocol{Port: int32(port.Port), Protocol: TCP}
		portProtocols = append(portProtocols, pp)
	}
	avi_vs_meta.PortProto = portProtocols
	// Default case.
	avi_vs_meta.ApplicationProfile = "System-TCP"
	// For HTTP it's always System-TCP-Proxy.
	avi_vs_meta.NetworkProfile = "System-TCP-Proxy"

	return avi_vs_meta
}

func (o *AviObjectGraph) BuildL4LBGraph(namespace string, svcName string) {
	// We use the gateway fields to arrive at various AVI VS Node object.
	var VsNode *AviVsNode
	svcObj, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).Get(svcName)
	if err != nil {
		utils.AviLog.Warning.Printf("Error in obtaining the object for service: %s", svcName)
		return
	}
	VsNode = o.ConstructAviL4VsNode(svcObj)
	o.AddModelNode(VsNode)
	VsNode.CalculateCheckSum()
	o.GraphChecksum = o.GraphChecksum + VsNode.GetCheckSum()
	utils.AviLog.Info.Printf("Checksum  for AVI VS object %v", VsNode.GetCheckSum())
	utils.AviLog.Info.Printf("Computed Graph Checksum for VS is: %v", o.GraphChecksum)
}
