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

package lib

import (
	"context"
	"fmt"

	"os"
	"strings"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func InformersToRegister(kclient *kubernetes.Clientset) ([]string, error) {
	// Initialize the following informers in all AKO deployments. Provide AKO the ability to watch over
	// Services, EndpointSlices, Secrets, ConfigMaps.
	allInformers := []string{
		utils.ServiceInformer,
		utils.SecretInformer,
		utils.ConfigMapInformer,
		utils.NSInformer,
		utils.EndpointSlicesInformer,
	}

	if lib.GetServiceType() == lib.NodePortLocal {
		allInformers = append(allInformers, utils.PodInformer)
	}

	return allInformers, nil
}

// parent vs name format - ako-gw-clustername--gatewayNs-gatewayName-EVH
func GetGatewayParentName(namespace, gwName string) string {
	//clustername > gateway namespace > Gateway-name
	//Adding -EVH prefix to reuse rest layer
	if IsGatewayInDedicatedMode(namespace, gwName) {
		return lib.GetNamePrefix() + namespace + "-" + gwName + lib.DedicatedSuffix + "-EVH"
	}
	return lib.GetNamePrefix() + namespace + "-" + gwName + "-EVH"
}

// child vs name format - ako-gw-clustername--encoded value of ako-gw-clustername--parentNs-parentName-routeNs-routeName-encodedMatch
func GetChildName(parentNs, parentName, routeNs, routeName, matchName string) string {
	name := parentNs + "-" + parentName + "-" + routeNs + "-" + routeName
	if matchName != "" {
		name = fmt.Sprintf("%s-%s", name, utils.Stringify(utils.Hash(matchName)))
	}
	return lib.EncodeWithPrefix(name, lib.EVHVS)
}

func GetPoolName(parentNs, parentName, routeNs, routeName, matchName, backendNs, backendName, backendPort string) string {
	name := parentNs + "-" + parentName + "-" + routeNs + "-" + routeName + "-"
	if matchName != "" {
		name = fmt.Sprintf("%s%s-", name, utils.Stringify(utils.Hash(matchName)))
	}
	name = fmt.Sprintf("%s%s-%s-%s", name, backendNs, backendName, backendPort)
	return lib.EncodeWithPrefix(name, lib.Pool)
}

func GetPoolGroupName(parentNs, parentName, routeNs, routeName, matchName string) string {
	name := parentNs + "-" + parentName + "-" + routeNs + "-" + routeName
	if matchName != "" {
		name = fmt.Sprintf("%s-%s", name, utils.Stringify(utils.Hash(matchName)))
	}
	return lib.EncodeWithPrefix(name, lib.PG)
}

func GetPersistenceProfileName(parentNs, parentName, routeNs, routeName, matchName, sessionPersistenceType string) string {
	name := parentNs + "-" + parentName + "-" + routeNs + "-" + routeName + "-" + sessionPersistenceType
	if matchName != "" {
		name = fmt.Sprintf("%s-%s", name, utils.Stringify(utils.Hash(matchName)))
	}
	return lib.EncodeWithPrefix(name, lib.ApplicationPersistenceProfile)
}

func GetHttpPolicySetName(parentNs, parentName, routeNs, routeName string) string {
	name := parentNs + "-" + parentName + "-" + routeNs + "-" + routeName + "-httproute"
	return lib.EncodeWithPrefix(name, lib.HTTPPS)
}

func GetDedicatedPoolName(poolGroupName, backendNs, backendName string, backendPort int32, backendIndex int) string {
	var name string
	if backendName != "" {
		name = fmt.Sprintf("%s-%s-%s-%d", poolGroupName, backendNs, backendName, backendPort)
	} else {
		name = fmt.Sprintf("%s-backend-%d", poolGroupName, backendIndex)
	}
	return lib.EncodeWithPrefix(name, lib.Pool)
}

func CheckGatewayClassController(controllerName string) bool {
	return controllerName == lib.AviIngressController
}

func FindListenerByName(name string, listener []gatewayv1.Listener) int {
	for i := range listener {
		if string(listener[i].Name) == name {
			return i
		}
	}
	return -1
}

func FindListenerStatusByName(name string, status []gatewayv1.ListenerStatus) int {
	for i := range status {
		if string(status[i].Name) == name {
			return i
		}
	}
	return -1
}

func FindPortName(serviceName, ns string, servicePort int32, key string) string {
	// Query the service and obtain the port name
	svcObj, err := utils.GetInformers().ServiceInformer.Lister().Services(ns).Get(serviceName)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: error while fetching service object: %s", key, err)
		return ""
	}
	for _, port := range svcObj.Spec.Ports {
		// Iterate the ports and find the match for targetPort
		if servicePort == port.Port {
			utils.AviLog.Debugf("key: %s, msg: Found port name %s for Port: %v", key, port.Name, servicePort)
			return port.Name
		}
	}
	utils.AviLog.Warnf("key: %s, msg: Port name not found in service obj: %v", key, svcObj)
	return ""
}
func GetT1LRPath() string {
	return os.Getenv("NSXT_T1_LR")
}

func FindTargetPort(serviceName, ns string, svcPort int32, key string) intstr.IntOrString {
	// Query the service and obtain the targetPort
	svcObj, err := utils.GetInformers().ServiceInformer.Lister().Services(ns).Get(serviceName)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: error while fetching service object: %s", key, err)
		return intstr.IntOrString{}
	}
	if svcObj.Spec.Type == "NodePort" {
		// Service of type NodePorts are not supported with tagertPort info. In such a case, the ports in the ingress must be strings
		return intstr.IntOrString{}
	}
	for _, port := range svcObj.Spec.Ports {
		// Iterate the ports and find the match for targetPort
		if svcPort == port.Port {
			utils.AviLog.Infof("key: %s, msg: Found targetPort %v for Port: %v", key, port.TargetPort.String(), svcPort)
			return port.TargetPort
		}
	}
	return intstr.IntOrString{}
}
func IsListenerInvalid(gwStatus *gatewayv1.GatewayStatus, listenerIndex int) bool {
	if len(gwStatus.Listeners) > int(listenerIndex) && len(gwStatus.Listeners[listenerIndex].Conditions) > 0 && gwStatus.Listeners[listenerIndex].Conditions[0].Type == string(gatewayv1.ListenerConditionAccepted) && gwStatus.Listeners[listenerIndex].Conditions[0].Status == "False" {
		return true
	}
	return false
}
func IsGatewayInvalid(gwStatus *gatewayv1.GatewayStatus) bool {
	if gwStatus.Conditions[0].Type == string(gatewayv1.ListenerConditionAccepted) && gwStatus.Conditions[0].Status == "False" {
		return true
	}
	return false
}

func VerifyHostnameSubdomainMatch(hostname string) bool {
	// Check if a hostname is valid or not by verifying if it has a prefix that
	// matches any of the sub-domains.
	subDomains := nodes.GetDefaultSubDomain()
	if len(subDomains) == 0 {
		// No IPAM DNS configured, we simply pass the hostname
		return true
	} else {
		for _, subd := range subDomains {
			if strings.HasSuffix(hostname, subd) {
				return true
			}
		}
	}
	utils.AviLog.Warnf("Didn't find match for hostname :%s Available sub-domains:%s", hostname, subDomains)
	return false
}

func ProtocolToRoute(proto string) string {
	innerMap := map[string]string{
		"HTTP":  lib.HTTPRoute,
		"HTTPS": lib.HTTPRoute,
		"TCP":   lib.TCPRoute,
		"TLS":   lib.TLSRoute,
		"UDP":   lib.UDPRoute,
	}

	return innerMap[proto]
}

func GetDefaultHTTPPSName() string {
	return Prefix + lib.GetClusterName() + "--" + lib.DefaultPSName
}

func GetTLSKeyCertNodeName(gatewayNameSpace, gatewayName, secretNameSpace, secretName string) string {
	namePrefix := gatewayNameSpace + "-" + gatewayName + "-" + secretNameSpace + "-" + secretName
	return lib.EncodeWithPrefix(namePrefix, lib.TLSKeyCert)
}

func CreateVCFGatewayClass() error {
	gwClass, err := AKOControlConfig().GatewayAPIClientset().GatewayV1().GatewayClasses().Get(context.TODO(), VCFGatewayClassName, metav1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			utils.AviLog.Errorf("Failed to GET Gatewayclass %s, err: %v", VCFGatewayClassName, err)
			return err
		}
		gwClass = nil
	}
	if gwClass != nil && gwClass.Spec.ControllerName == GatewayController {
		return nil
	} else if gwClass != nil {
		// controller Name is an immutable field, need to delete the avi-lb Gateway class and recreate it with the correct controller Name
		err = AKOControlConfig().GatewayAPIClientset().GatewayV1().GatewayClasses().Delete(context.TODO(), VCFGatewayClassName, metav1.DeleteOptions{})
		if err != nil {
			utils.AviLog.Errorf("Failed to DELETE Gatewayclass %s, err: %v", VCFGatewayClassName, err)
			return err
		}
	}

	gwClass = &gatewayv1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: VCFGatewayClassName,
		},
		Spec: gatewayv1.GatewayClassSpec{
			ControllerName: GatewayController,
		},
	}
	_, err = AKOControlConfig().GatewayAPIClientset().GatewayV1().GatewayClasses().Create(context.TODO(), gwClass, metav1.CreateOptions{})
	if err != nil {
		utils.AviLog.Errorf("Failed to CREATE Gatewayclass %s, err: %v", VCFGatewayClassName, err)
		return err
	}
	utils.AviLog.Infof("Successfully created Gatewayclass %s", VCFGatewayClassName)
	return nil
}

func IsGatewayInDedicatedMode(namespace, gatewayName string) bool {
	// Get Gateway object and check for dedicated mode annotation
	gateway, err := AKOControlConfig().GatewayApiInformers().GatewayInformer.Lister().Gateways(namespace).Get(gatewayName)
	if err != nil {
		utils.AviLog.Debugf("Failed to get gateway %s/%s: %v", namespace, gatewayName, err)
		return false
	}

	// Check annotation for dedicated mode
	if annotation, exists := gateway.GetAnnotations()[DedicatedGatewayModeAnnotation]; exists {
		return annotation == "true"
	}

	return false
}

func GetGatewayDedicatedVSName(namespace, gatewayName string) string {
	return lib.GetNamePrefix() + namespace + "-" + gatewayName
}
