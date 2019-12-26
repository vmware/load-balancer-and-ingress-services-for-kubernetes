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
	"gitlab.eng.vmware.com/orion/akc/pkg/objects"
	"gitlab.eng.vmware.com/orion/container-lib/utils"
	extensionv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
)

var (
	Service = GraphSchema{
		Type:               "Service",
		GetParentIngresses: SvcToIng,
	}
	Ingress = GraphSchema{
		Type:               "Ingress",
		GetParentIngresses: IngressChanges,
	}
	Endpoint = GraphSchema{
		Type:               "Endpoints",
		GetParentIngresses: EPToIng,
	}
	SupportedGraphTypes = GraphDescriptor{
		Ingress,
		Service,
		Endpoint,
	}
)

type GraphSchema struct {
	Type               string
	GetParentIngresses func(string, string, string) ([]string, bool)
}

type GraphDescriptor []GraphSchema

func IngressChanges(ingName string, namespace string, key string) ([]string, bool) {
	var ingresses []string
	ingresses = append(ingresses, ingName)
	ingObj, err := utils.GetInformers().IngressInformer.Lister().Ingresses(namespace).Get(ingName)
	if err != nil {
		// Detect a delete condition here.
		if errors.IsNotFound(err) {
			// Remove all the Ingress to Services mapping.
			// Remove the references of this ingress from the Services
			objects.SharedSvcLister().IngressMappings(namespace).RemoveIngressMappings(ingName)
		}
	} else {
		services := parseServicesForIngress(ingObj.Spec, key)
		for _, svc := range services {
			utils.AviLog.Info.Printf("key: %s, msg: updating ingress relationship for service:  %s", key, svc)
			objects.SharedSvcLister().IngressMappings(namespace).UpdateIngressMappings(ingName, svc)
		}
	}
	return ingresses, true
}

func SvcToIng(svcName string, namespace string, key string) ([]string, bool) {
	_, ingresses := objects.SharedSvcLister().IngressMappings(namespace).GetSvcToIng(svcName)
	utils.AviLog.Info.Printf("key: %s, msg: total ingresses retrieved:  %s", key, ingresses)
	if len(ingresses) == 0 {
		return nil, false
	}
	return ingresses, true
}

func EPToIng(epName string, namespace string, key string) ([]string, bool) {
	ingresses, found := SvcToIng(epName, namespace, key)
	utils.AviLog.Info.Printf("key: %s, msg: total ingresses retrieved:  %s", key, ingresses)
	return ingresses, found
}

func parseServicesForIngress(ingSpec extensionv1beta1.IngressSpec, key string) []string {
	// Figure out the service names that are part of this ingress
	var services []string
	for _, rule := range ingSpec.Rules {
		for _, path := range rule.IngressRuleValue.HTTP.Paths {
			services = append(services, path.Backend.ServiceName)
		}
	}
	utils.AviLog.Info.Printf("key: %s, msg: total services retrieved:  %s", key, services)
	return services
}
