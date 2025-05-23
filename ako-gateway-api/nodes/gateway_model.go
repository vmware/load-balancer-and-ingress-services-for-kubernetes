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
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/net"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	akogatewayapiobjects "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

type AviObjectGraph struct {
	*nodes.AviObjectGraph
}

func NewAviObjectGraph() *AviObjectGraph {
	return &AviObjectGraph{&nodes.AviObjectGraph{}}
}

func (o *AviObjectGraph) BuildGatewayVs(gateway *gatewayv1.Gateway, key string) {
	o.Lock.Lock()
	defer o.Lock.Unlock()

	vsNode := o.BuildGatewayParent(gateway, key)

	o.AddModelNode(vsNode)
	utils.AviLog.Infof("key: %s, msg: checksum for AVI VS object %v", key, vsNode.GetCheckSum())
}

func (o *AviObjectGraph) BuildGatewayParent(gateway *gatewayv1.Gateway, key string) *nodes.AviEvhVsNode {
	vsName := akogatewayapilib.GetGatewayParentName(gateway.Namespace, gateway.Name)
	parentVsNode := &nodes.AviEvhVsNode{
		Name:               vsName,
		Tenant:             lib.GetTenant(),
		ServiceEngineGroup: lib.GetSEGName(),
		ApplicationProfile: utils.DEFAULT_L7_APP_PROFILE,
		NetworkProfile:     utils.DEFAULT_TCP_NW_PROFILE,
		EVHParent:          true,
		VrfContext:         lib.GetVrf(),
		ServiceMetadata: lib.ServiceMetadataObj{
			Gateway: gateway.Namespace + "/" + gateway.Name,
		},
		Caller: utils.GATEWAY_API, // Always Populate this field to recognise caller at rest layer
	}
	t1LR := lib.GetT1LRPath()
	if t1LR != "" {
		utils.AviLog.Infof("key: %s, msg: T1LR is %s.", key, t1LR)
		parentVsNode.VrfContext = ""
	}
	parentVsNode.PortProto = BuildPortProtocols(gateway, key)

	tlsNodes := BuildTLSNodesForGateway(gateway, key)
	if len(tlsNodes) > 0 {
		parentVsNode.SSLKeyCertRefs = tlsNodes
	}

	vsvipNode := BuildVsVipNodeForGateway(gateway, parentVsNode.Name)
	parentVsNode.VSVIPRefs = []*nodes.AviVSVIPNode{vsvipNode}
	parentVsNode.AviMarkers = utils.AviObjectMarkers{
		GatewayName:      gateway.Name,
		GatewayNamespace: gateway.Namespace,
	}
	return parentVsNode
}

func BuildPortProtocols(gateway *gatewayv1.Gateway, key string) []nodes.AviPortHostProtocol {
	var portProtocols []nodes.AviPortHostProtocol
	gwStatus := akogatewayapiobjects.GatewayApiLister().GetGatewayToGatewayStatusMapping(gateway.Namespace + "/" + gateway.Name)
	for i, listener := range gateway.Spec.Listeners {
		if akogatewayapilib.IsListenerInvalid(gwStatus, i) {
			continue
		}
		pp := nodes.AviPortHostProtocol{Port: int32(listener.Port), Protocol: string(listener.Protocol)}
		//TLS config on listener is present
		if listener.TLS != nil && len(listener.TLS.CertificateRefs) > 0 {
			pp.EnableSSL = true
		}
		if !utils.HasElem(portProtocols, pp) {
			portProtocols = append(portProtocols, pp)
		}

	}
	return portProtocols
}

func BuildTLSNodesForGateway(gateway *gatewayv1.Gateway, key string) []*nodes.AviTLSKeyCertNode {
	var tlsNodes []*nodes.AviTLSKeyCertNode
	var ns, name string
	cs := utils.GetInformers().ClientSet
	gwStatus := akogatewayapiobjects.GatewayApiLister().GetGatewayToGatewayStatusMapping(gateway.Namespace + "/" + gateway.Name)
	for i, listener := range gateway.Spec.Listeners {
		if akogatewayapilib.IsListenerInvalid(gwStatus, i) {
			continue
		}
		if listener.TLS != nil {
			for _, certRef := range listener.TLS.CertificateRefs {
				//kind is validated at ingestion
				if certRef.Namespace == nil || *certRef.Namespace == "" {
					ns = gateway.Namespace
				} else {
					ns = string(*certRef.Namespace)
				}
				name = string(certRef.Name)
				secretObj, err := cs.CoreV1().Secrets(ns).Get(context.TODO(), name, metav1.GetOptions{})
				if err != nil || secretObj == nil {
					utils.AviLog.Warnf("key: %s, msg: secret %s has been deleted, err: %s", key, name, err)
					continue
				}
				tlsNode := TLSNodeFromSecret(secretObj, gateway.Namespace, gateway.Name, ns, name, key)
				if !utils.HasElem(tlsNodes, tlsNode) {
					tlsNodes = append(tlsNodes, tlsNode)
				}
			}
		}
	}
	return tlsNodes
}

func TLSNodeFromSecret(secretObj *corev1.Secret, gatewayNamespace, gatewayName, certificateNamespace, certName, key string) *nodes.AviTLSKeyCertNode {
	keycertMap := secretObj.Data
	tlscert, ok := keycertMap[utils.K8S_TLS_SECRET_CERT]
	if !ok {
		utils.AviLog.Infof("key: %s, msg: certificate not found for secret: %s", key, secretObj.Name)
	}
	tlskey, ok := keycertMap[utils.K8S_TLS_SECRET_KEY]
	if !ok {
		utils.AviLog.Infof("key: %s, msg: key not found for secret: %s", key, secretObj.Name)
	}
	tlsNode := &nodes.AviTLSKeyCertNode{
		Name:   akogatewayapilib.GetTLSKeyCertNodeName(gatewayNamespace, gatewayName, certificateNamespace, certName),
		Tenant: lib.GetTenant(),
		Type:   lib.CertTypeVS,
		Key:    tlskey,
		Cert:   tlscert,
	}
	return tlsNode
}

func BuildVsVipNodeForGateway(gateway *gatewayv1.Gateway, vsName string) *nodes.AviVSVIPNode {
	vsvipNode := &nodes.AviVSVIPNode{
		Name:        lib.GetVsVipName(vsName),
		Tenant:      lib.GetTenant(),
		VrfContext:  lib.GetVrf(),
		VipNetworks: utils.GetVipNetworkList(),
	}
	t1LR := lib.GetT1LRPath()
	if t1LR != "" {
		utils.AviLog.Infof("key: %s, msg: T1LR for vsvip node is: %s.", vsName, t1LR)
		vsvipNode.VrfContext = ""
		vsvipNode.T1Lr = t1LR
	}
	//Type is validated at ingestion
	if len(gateway.Spec.Addresses) == 1 {
		ipAddr := gateway.Spec.Addresses[0].Value
		if net.IsIPv4String(ipAddr) || net.IsIPv6String(ipAddr) {
			vsvipNode.IPAddress = ipAddr
		}
	}
	return vsvipNode
}

func DeleteTLSNode(key string, object *AviObjectGraph, gateway *gatewayv1.Gateway, secretObj *corev1.Secret) bool {
	var tlsNodes []*nodes.AviTLSKeyCertNode
	_, certNamespace, secretName := lib.ExtractTypeNameNamespace(key)
	evhVsCertRefs := object.GetAviEvhVS()[0].SSLKeyCertRefs
	gwStatus := akogatewayapiobjects.GatewayApiLister().GetGatewayToGatewayStatusMapping(gateway.Namespace + "/" + gateway.Name)
	encodedCertName := akogatewayapilib.GetTLSKeyCertNodeName(gateway.Namespace, gateway.Name, certNamespace, secretName)
	for evhVsCertRef := range evhVsCertRefs {
		if evhVsCertRefs[evhVsCertRef].Name != encodedCertName {
			tlsNodes = append(tlsNodes, evhVsCertRefs[evhVsCertRef])
		}
	}
	if len(tlsNodes) > 0 {
		object.GetAviEvhVS()[0].SSLKeyCertRefs = tlsNodes
	} else {
		utils.AviLog.Warnf("key: %s, msg: No certificate present for Parent VS %s", key, object.GetAviEvhVS()[0].Name)
		object.GetAviEvhVS()[0].SSLKeyCertRefs = nil
	}
	utils.AviLog.Infof("key: %s, msg: Updated cert_refs in parentVS: %s", key, object.GetAviEvhVS()[0].Name)
	return akogatewayapilib.IsGatewayInvalid(gwStatus)
}

func AddTLSNode(key string, object *AviObjectGraph, gateway *gatewayv1.Gateway, secretObj *corev1.Secret) {
	_, certNamespace, secretName := lib.ExtractTypeNameNamespace(key)
	tlsNodes := object.GetAviEvhVS()[0].SSLKeyCertRefs
	gwStatus := akogatewayapiobjects.GatewayApiLister().GetGatewayToGatewayStatusMapping(gateway.Namespace + "/" + gateway.Name)
	foundMatchingCertRef := false
	for i, listener := range gateway.Spec.Listeners {
		if akogatewayapilib.IsListenerInvalid(gwStatus, i) {
			continue
		}
		if listener.TLS != nil {
			for _, certRef := range listener.TLS.CertificateRefs {
				name := string(certRef.Name)
				listenerCertRefNamespace := gateway.Namespace
				if certRef.Namespace != nil {
					listenerCertRefNamespace = string(*certRef.Namespace)
				}
				if name == secretName && listenerCertRefNamespace == certNamespace {
					tlsNode := TLSNodeFromSecret(secretObj, gateway.Namespace, gateway.Name, certNamespace, secretName, key)
					indexOfTLSNode := utils.HasElemWithName(tlsNodes, tlsNode)
					if indexOfTLSNode == -1 {
						tlsNodes = append(tlsNodes, tlsNode)
					} else {
						tlsNodes[indexOfTLSNode] = tlsNode
					}
					foundMatchingCertRef = true
					break
				}
			}
		}
		if foundMatchingCertRef {
			break
		}
	}

	utils.AviLog.Infof("key: %s, msg: Updated cert_refs in parentVS: %s", key, object.GetAviEvhVS()[0].Name)
	object.GetAviEvhVS()[0].SSLKeyCertRefs = tlsNodes
}
