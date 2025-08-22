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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

var (
	L4Rule = GraphSchema{
		Type:              lib.L4Rule,
		GetParentServices: L4RuleToSvc,
	}
)

func L4RuleToSvc(l4RuleName string, namespace string, key string) ([]string, bool) {

	l4Rule, err := lib.AKOControlConfig().CRDInformers().L4RuleInformer.Lister().L4Rules(namespace).Get(l4RuleName)
	if k8serrors.IsNotFound(err) {
		utils.AviLog.Debugf("key: %s, msg: L4Rule %s deleted", key, l4RuleName)
	} else if err != nil {
		utils.AviLog.Errorf("key: %s, msg: Error getting L4Rule: %v", key, err)
		return []string{}, false
	}

	// Note: We should return services even if L4Rule is rejected, so that services can be processed
	// and virtual service properties can be cleaned up when L4Rule becomes invalid
	if l4Rule != nil && l4Rule.Status.Status != lib.StatusAccepted {
		utils.AviLog.Debugf("key: %s, msg: L4Rule %s is in %s state, but still returning services for cleanup", key, l4RuleName, l4Rule.Status.Status)
	}

	// Get all services that are mapped to this L4Rule.
	l4RuleNameWithNamespace := fmt.Sprintf("%s/%s", namespace, l4RuleName)
	services, err := utils.GetInformers().ServiceInformer.Informer().GetIndexer().ByIndex(lib.L4RuleToServicesIndex, l4RuleNameWithNamespace)
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: failed to get the services mapped to L4Rule %s", key, l4RuleNameWithNamespace)
		return []string{}, false
	}

	var allSvcs []string
	for _, svc := range services {
		svcObj, isSvc := svc.(*corev1.Service)
		if isSvc {
			allSvcs = append(allSvcs, svcObj.Namespace+"/"+svcObj.Name)
		}
	}

	utils.AviLog.Debugf("key: %s, msg: total services retrieved from L4Rule: %s", key, allSvcs)
	return allSvcs, true
}
