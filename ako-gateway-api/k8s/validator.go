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

package k8s

import (
	"strings"

	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func IsValidGateway(gateway *gatewayv1beta1.Gateway) bool {
	spec := gateway.Spec
	//is associated with gateway class
	if CheckGatewayController(*gateway) {
		utils.AviLog.Errorf("AKO is not set  found in gateway %+v", gateway.Name)
		return false
	}

	//has 1 or more listeners
	if len(spec.Listeners) == 0 {
		utils.AviLog.Errorf("no listeners found in gateway %+v", gateway.Name)
		return false
	}

	//has 1 or none addresses
	if len(spec.Addresses) > 1 {
		utils.AviLog.Errorf("more than 1 gateway address found in gateway %+v", gateway.Name)
		return false
	}

	for _, listener := range spec.Listeners {
		if !isValidListener(gateway.Name, listener) {
			return false
		}
	}
	return true
}

func isValidListener(gwName string, listener gatewayv1beta1.Listener) bool {
	//has valid name
	if listener.Name == "" {
		utils.AviLog.Errorf("no listener name found in gateway %+v", gwName)
		return false
	}
	//hostname is not wildcard
	if listener.Hostname == nil || strings.Contains(string(*listener.Hostname), "*") {
		utils.AviLog.Errorf("listener hostname with wildcard found in gateway %+v", gwName)
		return false
	}
	//port and protocol valid

	//has valid TLS config
	if listener.TLS != nil {
		if *listener.TLS.Mode != "Terminate" || len(listener.TLS.CertificateRefs) == 0 {
			utils.AviLog.Errorf("tls mode/ref not valid %+v/%+v", gwName, listener.Name)
			return false
		}
	}
	return true
}

func CheckGatewayClassController(gwClass gatewayv1beta1.GatewayClass) bool {
	if gwClass.Spec.ControllerName == lib.AviIngressController {
		return true
	}
	return false
}

func CheckGatewayController(gw gatewayv1beta1.Gateway) bool {
	gwClass := gw.Spec.GatewayClassName
	if gwClass == "" {
		return false
	}
	if !objects.GatewayApiLister().IsGatewayClassPresent(string(gwClass)) {
		return false
	}
	return true
}
